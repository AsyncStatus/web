package sql

import (
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/jackc/pgx/v5"
)

// BuildCreateSQL builds an INSERT statement by reflecting on the fields
// of `input`. For each struct field:
//
//   - If the field is a pointer and is nil -> Skip (omit from INSERT).
//   - Otherwise, include the field.
//
// The column name is derived from the `json:"..."` struct tag, or uses
// the Go field name if the tag is missing.
//
// Example outcome:
//
//	INSERT INTO "users" ("name", "email") VALUES (:name, :email) RETURNING *
//
// And the returned pgx.StrictNamedArgs might be something like:
//
//	pgx.StrictNamedArgs{"name": "Alice", "email": "alice@example.com"}
func BuildCreateSQL(tableName string, input any) (string, pgx.StrictNamedArgs, error) {
	if input == nil {
		return "", nil, errors.New("input cannot be nil")
	}

	v := reflect.ValueOf(input)
	t := reflect.TypeOf(input)

	// Ensure we have a pointer to a struct
	if v.Kind() != reflect.Pointer || v.Elem().Kind() != reflect.Struct {
		return "", nil, errors.New("input must be a pointer to a struct")
	}
	v = v.Elem()
	t = t.Elem()

	var (
		cols         []string
		placeholders []string
		args         = pgx.StrictNamedArgs{}
	)

	for i := 0; i < v.NumField(); i++ {
		fieldValue := v.Field(i)
		fieldType := t.Field(i)

		// Derive column name from `json` tag
		colName := columnNameFromJSONTag(fieldType)
		if colName == "" {
			// "" indicates skip (e.g. json:"-" or no valid name)
			continue
		}

		// If it's a pointer and nil -> skip
		if fieldValue.Kind() == reflect.Ptr {
			if fieldValue.IsNil() {
				continue
			}
			// Else, dereference and use the underlying value
			cols = append(cols, fmt.Sprintf(`"%s"`, colName))
			placeholders = append(placeholders, fmt.Sprintf("@%s", colName))
			args[colName] = fieldValue.Elem().Interface()
		} else {
			// Non-pointer field -> always insert as-is
			cols = append(cols, fmt.Sprintf(`"%s"`, colName))
			placeholders = append(placeholders, fmt.Sprintf("@%s", colName))
			args[colName] = fieldValue.Interface()
		}
	}

	if len(cols) == 0 {
		return "", nil, errors.New("no columns to insert (all fields skipped)")
	}

	// Construct SQL:
	// INSERT INTO "tableName" ("col1", "col2", ...) VALUES (:col1, :col2, ...) RETURNING *
	sqlStr := fmt.Sprintf(
		`INSERT INTO %s (%s) VALUES (%s) RETURNING *`,
		tableName,
		strings.Join(cols, ", "),
		strings.Join(placeholders, ", "),
	)

	return sqlStr, args, nil
}
