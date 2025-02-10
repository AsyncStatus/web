package sql

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

// PaginatedBodySearchField defines a single condition for a search.
type PaginatedBodySearchField struct {
	Field    string `json:"field"`
	Value    any    `json:"value"`
	Operator string `json:"operator" enum:"=,!=,IN,<,>,<=,>=,~"`
} // @name PaginatedBodySearchField

// PaginatedBodySearch holds multiple fields and a global operator (AND/OR).
type PaginatedBodySearch struct {
	Field    PaginatedBodySearchField `json:"field"`
	Operator string                   `json:"operator" enum:"and,or"`
} // @name PaginatedBodySearch

// PaginatedBody is your request payload for listing resources, possibly with a search.
type PaginatedBody struct {
	Type string `json:"type" enum:"list,search" default:"list"`
	// Stringified JSON of `[]PaginatedBodySearch` type
	Search         string `json:"search" default:""`
	Limit          int64  `json:"limit" validate:"min=1" default:"50"`
	Offset         int64  `json:"offset" validate:"min=0" default:"0"`
	OrderBy        string `json:"orderby"`
	OrderDirection string `json:"orderdirection" enum:"ASC,DESC" default:"DESC"`
} // @name PaginatedBody

type PaginatedData[T any] struct {
	Data       []*T  `json:"results" validate:"required"`
	TotalCount int64 `json:"total_count" validate:"required"`
} // @name PaginatedData

// BuildPaginatedListSQL builds:
//  1. A SELECT query for paginated, optionally filtered rows.
//  2. A matching COUNT query for total rows.
//
// modelStruct: a zero-value (or pointer) of your struct, e.g. User{}, that has `db:"..."` tags.
// selectCols: which columns to SELECT. If empty, defaults to "*".
//
// Example usage:
//
//	type User struct {
//	  ID    string `json:"id" db:"id"`
//	  Name  string `json:"name" db:"name"`
//	  Email string `json:"email" db:"email"`
//	}
//
//	pBody := PaginatedBody{
//	  Type: "search",
//	  Search: PaginatedBodySearch{
//	    Operator: "or",
//	    Fields: []PaginatedBodySearchField{
//	      {Field: "name", Value: "alice", Operator: "~"}, // "~" => partial match
//	      {Field: "email", Value: "@gmail.com", Operator: "~"},
//	    },
//	  },
//	  Limit:  10,
//	  Offset: 0,
//	  OrderBy: "id",
//	  OrderDirection: "ASC",
//	}
//
//	selSQL, selArgs, cntSQL, err := BuildPaginatedListSQL("users", User{}, pBody, []string{"id","name","email"})
//	if err != nil { ... }
//	fmt.Println("SELECT query:\n", selSQL)
//	fmt.Println("ARGS:", selArgs)
//	fmt.Println("COUNT query:\n", cntSQL)
//
// Then use `db.NamedQuery(ctx, selSQL, selArgs)`, etc.
func BuildPaginatedListSQL(
	tableName string,
	modelStruct any,
	body *PaginatedBody,
	selectCols []string,
) (selectSQL string, selArgs pgx.StrictNamedArgs, countSQL string, countArgs pgx.StrictNamedArgs, err error) {
	if tableName == "" {
		return "", nil, "", nil, errors.New("tableName cannot be empty")
	}

	if len(selectCols) == 0 {
		selectCols = []string{"*"}
	}
	selectClause := strings.Join(selectCols, ", ")

	selArgs = make(pgx.StrictNamedArgs)
	countArgs = make(pgx.StrictNamedArgs)

	bodySearch := &[]PaginatedBodySearch{}
	if body.Search == "" {
		bodySearch = nil
	} else {
		err = json.Unmarshal([]byte(body.Search), bodySearch)
		if err != nil {
			return "", nil, "", nil, err
		}
	}

	//------------------------------------------------------------
	// 1) Build WHERE conditions if we're in "search" mode
	//------------------------------------------------------------
	whereParts := []string{}
	whereOperators := []string{}
	if strings.EqualFold(body.Type, "search") {
		// We have multiple fields combined with bodySearch.Operator ("AND" / "OR").
		// op := strings.ToUpper(strings.TrimSpace(bodySearch.Operator))
		// if op != "AND" && op != "OR" {
		// 	op = "AND"
		// }

		for i, sf := range *bodySearch {
			// Attempt to find the DB column from the model struct
			dbCol, kind, found := findDBColumn(modelStruct, sf.Field.Field)
			if !found {
				// If not found, we can either skip or return an error
				// For now, let's skip this field
				continue
			}
			fmt.Println("Field:", sf.Field.Field, "Operator:", sf.Operator, "Value:", sf.Field.Value, "DBCol:", dbCol, "Kind:", kind)

			timestampKind := reflect.TypeOf(pgtype.Timestamp{}).Kind()
			// We'll bind each field’s value to a distinct param, e.g. @search_val_0
			placeholder := fmt.Sprintf("search_val_%d", i)
			whereOperators = append(whereOperators, sf.Operator)

			// Build a condition depending on sf.Operator
			// e.g.: =, !=, ~ (ILIKE?), or you can expand to >, <, etc.
			switch sf.Field.Operator {
			case "=":
				// "dbCol = :search_val_0"
				whereParts = append(whereParts, fmt.Sprintf(`%s = @%s`, dbCol, placeholder))
				selArgs[placeholder] = sf.Field.Value
				countArgs[placeholder] = sf.Field.Value

			case "!=":
				// "dbCol != :search_val_0"
				whereParts = append(whereParts, fmt.Sprintf(`%s != @%s`, dbCol, placeholder))
				selArgs[placeholder] = sf.Field.Value
				countArgs[placeholder] = sf.Field.Value

			case "~":
				// Partial match (ILIKE). We'll cast value to string and wrap in '%...%'
				// If sf.Value is not a string, let's do a naive conversion (or skip).
				valStr, ok := sf.Field.Value.(string)
				if !ok {
					// If we can't convert to string, skip or throw an error.
					continue
				}
				whereParts = append(whereParts, fmt.Sprintf(`%s ILIKE @%s`, dbCol, placeholder))
				selArgs[placeholder] = "%" + valStr + "%"
				countArgs[placeholder] = "%" + valStr + "%"

			case "IN":
				// "dbCol IN (:search_val_0, :search_val_1, :search_val_2)"
				whereParts = append(whereParts, fmt.Sprintf(`%s IN (%s)`, dbCol, placeholder))
				selArgs[placeholder] = sf.Field.Value
				countArgs[placeholder] = sf.Field.Value

			case ">":
				// "dbCol > :search_val_0"
				if kind == timestampKind {
					valueTimeParsed, err := time.Parse(time.RFC3339, sf.Field.Value.(string))
					if err != nil {
						continue
					}
					whereParts = append(whereParts, fmt.Sprintf(`%s > @%s`, dbCol, placeholder))
					selArgs[placeholder] = valueTimeParsed
					countArgs[placeholder] = valueTimeParsed
				} else {
					whereParts = append(whereParts, fmt.Sprintf(`%s > @%s`, dbCol, placeholder))
					selArgs[placeholder] = sf.Field.Value
					countArgs[placeholder] = sf.Field.Value

				}

			case "<":
				// "dbCol < :search_val_0"
				whereParts = append(whereParts, fmt.Sprintf(`%s < @%s`, dbCol, placeholder))
				selArgs[placeholder] = sf.Field.Value
				countArgs[placeholder] = sf.Field.Value

			case ">=":
				// "dbCol >= :search_val_0"
				whereParts = append(whereParts, fmt.Sprintf(`%s >= @%s`, dbCol, placeholder))
				selArgs[placeholder] = sf.Field.Value
				countArgs[placeholder] = sf.Field.Value

			case "<=":
				// "dbCol <= :search_val_0"
				whereParts = append(whereParts, fmt.Sprintf(`%s <= @%s`, dbCol, placeholder))
				selArgs[placeholder] = sf.Field.Value
				countArgs[placeholder] = sf.Field.Value

			default:
				// Unrecognized operator => skip or error
				continue
			}
		}

	}

	//------------------------------------------------------------
	// 2) Combine all WHERE parts with each operator for each part
	//------------------------------------------------------------
	var whereClause string
	if len(whereParts) > 0 {
		for i, part := range whereParts {
			if len(whereOperators) > i {
				whereClause += " " + whereOperators[i] + " "
			}
			whereClause += part
		}

		whereClause = "WHERE " + whereClause
	}

	//------------------------------------------------------------
	// 3) Handle order, limit, offset
	//------------------------------------------------------------
	orderByClause := ""
	if body.OrderBy != "" {
		dbCol, _, found := findDBColumn(modelStruct, body.OrderBy)
		if found && dbCol != "" {
			dir := strings.ToUpper(strings.TrimSpace(body.OrderDirection))
			if dir != "ASC" && dir != "DESC" {
				dir = "ASC"
			}
			orderByClause = fmt.Sprintf(`ORDER BY "%s" %s`, dbCol, dir)
		}
	}

	if body.Limit <= 0 {
		body.Limit = 50 // default
	}
	if body.Offset < 0 {
		body.Offset = 0
	}
	selArgs["limit"] = body.Limit
	selArgs["offset"] = body.Offset

	//------------------------------------------------------------
	// 4) Construct final SELECT query
	//------------------------------------------------------------
	selectSQL = fmt.Sprintf(
		`SELECT * FROM (SELECT %s FROM %s %s %s) AS subquery LIMIT @limit OFFSET @offset`,
		selectClause, tableName, whereClause, orderByClause,
	)

	//------------------------------------------------------------
	// 5) Construct matching COUNT query (no ORDER/LIMIT/OFFSET)
	//------------------------------------------------------------
	countSQL = fmt.Sprintf(
		`SELECT count(*) AS total_count FROM %s %s`,
		tableName, whereClause,
	)

	return selectSQL, selArgs, countSQL, countArgs, nil
}

// findDBColumn tries to match the given fieldName to a struct field with a `db:"..."` tag.
// If found, returns the db column name. If not, returns false.
func findDBColumn(modelStruct any, fieldName string) (string, any, bool) {
	rv := reflect.ValueOf(modelStruct)
	rt := reflect.TypeOf(modelStruct)

	// If it's a pointer, deref it
	if rv.Kind() == reflect.Pointer {
		rv = rv.Elem()
		rt = rt.Elem()
	}
	if rv.Kind() != reflect.Struct {
		return "", "", false
	}

	fieldNameLower := strings.ToLower(fieldName)
	for i := 0; i < rt.NumField(); i++ {
		fType := rt.Field(i)

		colName := fType.Tag.Get("db")
		if colName == "" {
			colName = fType.Tag.Get("json")
			if colName == "" {
				colName = fType.Name // fallback: use struct field name
			}
		}

		// We'll do a case-insensitive match
		if strings.ToLower(colName) == fieldNameLower {
			return colName, fType.Type.Kind(), true
		}
	}

	// No match found
	return "", "", false
}
