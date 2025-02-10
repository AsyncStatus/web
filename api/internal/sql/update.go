package sql

import (
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/google/uuid"
	"github.com/guregu/null/v5"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

// BuildUpdateSQL builds an UPDATE statement that only sets columns
// for which the passed-in struct has a non-nil pointer.
// If a null.* or pgtype.* pointer is non-nil but invalid, we set the DB column to NULL.
func BuildUpdateSQL(
	tableName string,
	primaryKeyCol string,
	primaryKeyVal any,
	input any,
) (string, pgx.StrictNamedArgs, error) {
	if input == nil {
		return "", nil, errors.New("input cannot be nil")
	}

	val := reflect.ValueOf(input)
	typ := reflect.TypeOf(input)

	// We expect input to be a pointer to a struct, e.g. *UpdateUserBody
	if val.Kind() != reflect.Pointer || val.Elem().Kind() != reflect.Struct {
		return "", nil, errors.New("input must be a pointer to a struct")
	}
	val = val.Elem()
	typ = typ.Elem()

	var setClauses []string
	args := pgx.StrictNamedArgs{}

	for i := 0; i < val.NumField(); i++ {
		fieldValue := val.Field(i)
		fieldType := typ.Field(i)

		// Derive the column name from the `json` struct tag, if present,
		// otherwise fall back to the Go field name.
		colName := columnNameFromJSONTag(fieldType)
		if colName == "" {
			continue
		}

		if fieldValue.Kind() == reflect.Ptr {
			if fieldValue.IsNil() {
				// Means user did not provide a value => do not update this column
				continue
			}

			// Non-nil pointer => let's decide how to set it
			// Special handling if it's pointing to a "null.*" or "pgtype.*" type
			underElem := fieldValue.Elem()
			switch actual := fieldValue.Interface().(type) {

			case *uuid.NullUUID:
				setClauses = append(setClauses, fmt.Sprintf(`"%s" = @%s`, colName, colName))
				args[colName] = actual.UUID
				if !actual.Valid {
					args[colName] = nil
				}

			case *null.String:
				setClauses = append(setClauses, fmt.Sprintf(`"%s" = @%s`, colName, colName))
				args[colName] = actual.String
				if !actual.Valid {
					args[colName] = nil
				}

			case *null.Int:
				setClauses = append(setClauses, fmt.Sprintf(`"%s" = @%s`, colName, colName))
				args[colName] = actual.Int64
				if !actual.Valid {
					args[colName] = nil
				}

			case *null.Bool:
				setClauses = append(setClauses, fmt.Sprintf(`"%s" = @%s`, colName, colName))
				args[colName] = actual.Bool
				if !actual.Valid {
					args[colName] = nil
				}

			case *null.Float:
				setClauses = append(setClauses, fmt.Sprintf(`"%s" = @%s`, colName, colName))
				args[colName] = actual.Float64
				if !actual.Valid {
					args[colName] = nil
				}

			case *pgtype.Timestamp:
				setClauses = append(setClauses, fmt.Sprintf(`"%s" = @%s`, colName, colName))
				args[colName] = actual.Time
				if !actual.Valid {
					args[colName] = nil
				}

			case *pgtype.Date:
				setClauses = append(setClauses, fmt.Sprintf(`"%s" = @%s`, colName, colName))
				args[colName] = actual.Time
				if !actual.Valid {
					args[colName] = nil
				}

			default:
				// If it's some other pointer type (e.g. *string, *int, etc.) just dereference
				// the pointer. We assume we want to update with that actual value.
				setClauses = append(setClauses, fmt.Sprintf(`"%s" = @%s`, colName, colName))
				args[colName] = underElem.Interface()
			}

		} else {
			// The field itself is not a pointer; typically for partial updates
			// you'd ignore non-pointer fields or handle them differently.
			//
			// In many "patch" or "partial update" patterns, you won't see
			// non-pointer fields—because you'd have no way to distinguish
			// zero-value from not-provided. So you might skip them here:
			continue
		}
	}

	if len(setClauses) == 0 {
		// No fields to update
		return "", args, errors.New("no fields to update")
	}

	// Construct the final SQL, e.g.:
	// UPDATE users SET "name" = @name, "email" = @email WHERE user_id = @user_id
	setSQL := strings.Join(setClauses, ", ")
	sqlStr := fmt.Sprintf(
		`UPDATE %s SET %s WHERE "%s" = @%s RETURNING *`,
		tableName, setSQL, primaryKeyCol, primaryKeyCol,
	)

	args[primaryKeyCol] = primaryKeyVal

	return sqlStr, args, nil
}

// columnNameFromJSONTag attempts to extract a column name from the field's `db` or `json` tag.
// If none is found (or is '-'), returns an empty string to indicate "skip".
func columnNameFromJSONTag(f reflect.StructField) string {
	dbTagVal := f.Tag.Get("db")
	if dbTagVal != "" {
		return dbTagVal
	}

	tagVal := f.Tag.Get("json")
	if tagVal == "-" {
		return ""
	}

	if tagVal == "" {
		// Fallback to the Go field name if no json tag
		return f.Name
	}

	// Strip off any `,omitempty` or other tag extras
	if idx := strings.Index(tagVal, ","); idx != -1 {
		tagVal = tagVal[:idx]
	}
	return tagVal
}
