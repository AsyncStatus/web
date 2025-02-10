package sql

import (
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/guregu/null/v5"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

// BuildUpdateSQL constructs a SQL UPDATE statement by reflecting over two structs:
//
//   - whereStruct -> fields used in the WHERE clause
//
//   - dataStruct  -> fields used in the SET clause
//
//   - For the WHERE struct (whereStruct), we include **all** non-zero (or non-empty) fields as equality conditions.
//     (If you want to skip zero-values for non-pointer fields, adjust accordingly.)
//
//   - For the DATA struct (dataStruct), we include only **non-nil pointer** fields in the SET clause.
//     (If it's nil, we skip updating that column.)
//
// Example usage:
//
//	query, args, err := BuildUpdateSQL(`"user".users`, body.Select, body.Data)
//	// => UPDATE "user".users SET ... WHERE ... RETURNING *
//
// We'll name placeholders as @where_<field> for the WHERE struct, and @data_<field> for the data struct
// so there's no collision if the same column is used in both places.
//
// Adjust or simplify as needed for your domain.
func BuildUpdateSQLV2(tableName string, whereStruct any, dataStruct any) (string, pgx.StrictNamedArgs, error) {
	if whereStruct == nil {
		return "", nil, errors.New("whereStruct cannot be nil")
	}
	if dataStruct == nil {
		return "", nil, errors.New("dataStruct cannot be nil")
	}

	whereSQL, whereArgs, err := buildWhereClause(whereStruct)
	if err != nil {
		return "", nil, err
	}
	if len(whereSQL) == 0 {
		return "", nil, errors.New("no WHERE conditions generated; refused to do unscoped update")
	}

	setSQL, setArgs, err := buildSetClause(dataStruct)
	if err != nil {
		return "", nil, err
	}
	if len(setSQL) == 0 {
		return "", nil, errors.New("no SET columns to update")
	}

	query := fmt.Sprintf(
		`UPDATE %s SET %s WHERE %s RETURNING *`,
		tableName, strings.Join(setSQL, ", "), strings.Join(whereSQL, " AND "),
	)

	args := pgx.StrictNamedArgs{}
	for k, v := range whereArgs {
		args[k] = v
	}
	for k, v := range setArgs {
		args[k] = v
	}

	return query, args, nil
}

// buildWhereClause reflects over whereStruct, building a list of conditions:
//
//	"<colName> = @where_<field>" AND ...
//
// We include both pointer and non-pointer fields, as long as they are non-zero.
// (Adjust logic if you want to skip zero-values for non-pointer fields, etc.)
func buildWhereClause(whereStruct any) ([]string, pgx.StrictNamedArgs, error) {
	var clauses []string
	args := pgx.StrictNamedArgs{}

	rv := reflect.ValueOf(whereStruct)
	rt := reflect.TypeOf(whereStruct)

	// If pointer, dereference
	if rv.Kind() == reflect.Pointer {
		rv = rv.Elem()
		rt = rt.Elem()
	}
	if rv.Kind() != reflect.Struct {
		return nil, nil, errors.New("whereStruct must be a struct or pointer to struct")
	}

	for i := 0; i < rv.NumField(); i++ {
		fieldVal := rv.Field(i)
		fieldTyp := rt.Field(i)

		// Derive column name from `json` tag or fallback to field name
		colName := columnNameFromJSONTag(fieldTyp)
		if colName == "" {
			continue
		}

		paramKey := "where_" + colName // e.g. @where_email

		// We'll handle pointer vs non-pointer
		if fieldVal.Kind() == reflect.Ptr {
			if fieldVal.IsNil() {
				// Skip if pointer is nil
				continue
			}
			// Non-nil pointer => use the value
			clauses = append(clauses, fmt.Sprintf(`"%s" = @%s`, colName, paramKey))
			args[paramKey] = fieldVal.Elem().Interface()
		} else {
			// Non-pointer field => skip if it's zero?
			zeroVal := reflect.Zero(fieldVal.Type()).Interface()
			currVal := fieldVal.Interface()
			if isEqual(currVal, zeroVal) {
				// It's zero => skip or use it as a condition?
				// Typically, if you do want to filter by "0" or empty string,
				// you'd keep it. Let's keep it here for safety.
				continue
			}
			clauses = append(clauses, fmt.Sprintf(`"%s" = @%s`, colName, paramKey))
			args[paramKey] = currVal
		}
	}

	return clauses, args, nil
}

// buildSetClause reflects over dataStruct, building a list of expressions:
//
//	"<colName> = @data_<field>"
//
// Only includes pointer fields that are non-nil. If it's a pointer to null.* or pgtype.*,
// we check if it's "Valid" (or "Set") vs nil, and handle accordingly.
//
// Adjust as needed for your domain logic.
func buildSetClause(dataStruct any) ([]string, pgx.StrictNamedArgs, error) {
	var setClauses []string
	args := pgx.StrictNamedArgs{}

	rv := reflect.ValueOf(dataStruct)
	rt := reflect.TypeOf(dataStruct)

	// If pointer, dereference
	if rv.Kind() == reflect.Pointer {
		rv = rv.Elem()
		rt = rt.Elem()
	}
	if rv.Kind() != reflect.Struct {
		return nil, nil, errors.New("dataStruct must be a struct or pointer to struct")
	}

	for i := 0; i < rv.NumField(); i++ {
		fieldVal := rv.Field(i)
		fieldTyp := rt.Field(i)

		colName := columnNameFromJSONTag(fieldTyp)
		if colName == "" {
			continue
		}

		// We only update if it's a pointer and non-nil
		if fieldVal.Kind() == reflect.Ptr {
			if fieldVal.IsNil() {
				// not set => skip
				continue
			}
			// non-nil pointer => figure out if it's e.g. *string, *null.String, *pgtype.Timestamp, etc.
			placeholder := "data_" + colName
			switch x := fieldVal.Interface().(type) {
			case *null.String:
				setClauses = append(setClauses, fmt.Sprintf(`"%s" = @%s`, colName, placeholder))
				args[placeholder] = x.String
				if !x.Valid {
					args[placeholder] = nil
				}
			case *pgtype.Timestamp:
				setClauses = append(setClauses, fmt.Sprintf(`"%s" = @%s`, colName, placeholder))
				args[placeholder] = x.Time
				if !x.Valid {
					args[placeholder] = nil
				}
			// add other null.* / pgtype.* if needed
			default:
				// e.g. *string, *int, etc.
				setClauses = append(setClauses, fmt.Sprintf(`"%s" = @%s`, colName, placeholder))
				args[placeholder] = fieldVal.Elem().Interface()
			}
		} else {
			// If it's not a pointer, we skip or set unconditionally?
			// Typically for partial updates, you want pointer fields so you know if user set them or not.
			// We'll skip non-pointer here. Or do you want to always set it?
			continue
		}
	}

	return setClauses, args, nil
}

// isEqual is a helper to compare interface{} values for equality
// in a naive way. We do a basic reflect.DeepEqual here.
func isEqual(a, b any) bool {
	return reflect.DeepEqual(a, b)
}
