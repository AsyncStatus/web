package repository

import (
	"api/internal/http/aserr"
	"api/internal/pg"
	"api/internal/sql"
	"context"
	"errors"
	"strings"

	"github.com/google/uuid"
	"github.com/guregu/null/v5"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

type User struct {
	ID         uuid.UUID          `json:"id" validate:"required"`
	Name       string             `json:"name" validate:"required"`
	Email      string             `json:"email" validate:"required" format:"email"`
	Referrer   null.String        `json:"referrer" validate:"required" swaggertype:"string" extensions:"x-nullable"`
	AvatarURL  null.String        `json:"avatar_url" validate:"required" swaggertype:"string" extensions:"x-nullable"`
	CreatedAt  pgtype.Timestamptz `json:"created_at" validate:"required" swaggertype:"string"`
	VerifiedAt pgtype.Timestamptz `json:"verified_at" validate:"required" swaggertype:"string" extensions:"x-nullable"`
	UpdatedAt  pgtype.Timestamptz `json:"updated_at" validate:"required" swaggertype:"string" extensions:"x-nullable"`
	ArchivedAt pgtype.Timestamptz `json:"archived_at" validate:"required" swaggertype:"string" extensions:"x-nullable"`
	ApprovedAt pgtype.Timestamptz `json:"approved_at" validate:"required" swaggertype:"string" extensions:"x-nullable"`
} // @name User

type GetUserBody struct {
	ID    *uuid.UUID `json:"id" validate:"omitempty,uuid4" swaggertype:"string"`
	Email *string    `json:"email" validate:"omitempty,email" format:"email"`
} // @name GetUserBody

func (r *Repository) GetUser(ctx context.Context, q pg.Querier, body *GetUserBody) (*User, error) {
	rows, _ := q.Query(ctx, `
		SELECT
			u.*
		FROM
			"user".users u
		WHERE
			(COALESCE(@id::uuid, NULL) IS NULL OR u.id = @id)
			AND (COALESCE(@email::text, NULL) IS NULL OR u.email = @email)
	`, pgx.StrictNamedArgs{"id": body.ID, "email": body.Email})
	user, err := pgx.CollectExactlyOneRow(rows, pgx.RowToAddrOfStructByNameLax[User])
	if err != nil && errors.Is(err, pgx.ErrNoRows) {
		return nil, aserr.ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (r *Repository) ListUsers(ctx context.Context, q pg.Querier, body *sql.PaginatedBody) (*sql.PaginatedData[User], error) {
	listSQL, listArgs, countSQL, countArgs, err := sql.BuildPaginatedListSQL(`"user".users`, User{}, body, nil)
	if err != nil {
		return nil, err
	}

	rows, _ := q.Query(ctx, listSQL, listArgs)
	users, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByNameLax[User])
	if err != nil {
		return nil, err
	}

	rows, _ = q.Query(ctx, countSQL, countArgs)
	count, err := pgx.CollectOneRow(rows, pgx.RowTo[int64])
	if err != nil {
		return nil, err
	}

	return &sql.PaginatedData[User]{Data: users, TotalCount: count}, nil
}

type CreateUserBody struct {
	Name       string           `json:"name" validate:"required" minLength:"3" maxLength:"255"`
	Email      string           `json:"email" validate:"required,email" format:"email"`
	Referrer   null.String      `json:"referrer" swaggertype:"string" extensions:"x-nullable"`
	AvatarURL  null.String      `json:"avatar_url" swaggertype:"string" extensions:"x-nullable"`
	VerifiedAt pgtype.Timestamp `json:"verified_at" swaggertype:"string" extensions:"x-nullable"`
	ArchivedAt pgtype.Timestamp `json:"archived_at" swaggertype:"string" extensions:"x-nullable"`
	ApprovedAt pgtype.Timestamp `json:"approved_at" swaggertype:"string" extensions:"x-nullable"`
} // @name CreateUserBody

func (r *Repository) CreateUser(ctx context.Context, q pg.Querier, body *CreateUserBody) (*User, error) {
	query, args, err := sql.BuildCreateSQL(`"user".users`, body)
	if err != nil {
		return nil, err
	}

	rows, _ := q.Query(ctx, query, args)
	user, err := pgx.CollectExactlyOneRow(rows, pgx.RowToAddrOfStructByNameLax[User])
	if err != nil {
		return nil, err
	}

	return user, nil
}

type UpdateUserBodyFilter struct {
	ID    uuid.UUID `json:"id"`
	Email string    `json:"email" format:"email"`
}

type UpdateUserBodyData struct {
	Name       *string           `json:"name"`
	Email      *string           `json:"email" validate:"email" format:"email"`
	Referrer   *null.String      `json:"referrer" swaggertype:"string" extensions:"x-nullable"`
	AvatarURL  *null.String      `json:"avatar_url" swaggertype:"string" extensions:"x-nullable"`
	VerifiedAt *pgtype.Timestamp `json:"verified_at" swaggertype:"string" extensions:"x-nullable"`
	ArchivedAt *pgtype.Timestamp `json:"archived_at" swaggertype:"string" extensions:"x-nullable"`
	ApprovedAt *pgtype.Timestamp `json:"approved_at" swaggertype:"string" extensions:"x-nullable"`
}

func (r *Repository) UpdateUser(ctx context.Context, q pg.Querier, filter *UpdateUserBodyFilter, body *UpdateUserBodyData) (*User, error) {
	query, args, err := sql.BuildUpdateSQLV2(`"user".users`, filter, body)
	if err != nil {
		return nil, err
	}

	rows, _ := q.Query(ctx, query, args)
	user, err := pgx.CollectExactlyOneRow(rows, pgx.RowToAddrOfStructByNameLax[User])
	if err != nil {
		return nil, err
	}

	return user, nil
}

type DeleteUserBody struct {
	ID    *uuid.UUID `json:"id" validate:"uuid4" swaggertype:"string"`
	Email *string    `json:"email" validate:"email" format:"email"`
} // @name DeleteUserBody

func (r *Repository) DeleteUser(ctx context.Context, q pg.Querier, body *DeleteUserBody) (*User, error) {
	rows, _ := q.Query(ctx, `
		DELETE FROM
			"user".users
		WHERE
			(COALESCE(@id::uuid, NULL) IS NULL OR id = @id)
			AND (COALESCE(@email::text, NULL) IS NULL OR email = @email)
	`, pgx.StrictNamedArgs{"id": body.ID, "email": body.Email})
	user, err := pgx.CollectExactlyOneRow(rows, pgx.RowToAddrOfStructByNameLax[User])
	if err != nil {
		return nil, err
	}

	return user, nil
}

func EmailToName(email string) string {
	parts := strings.Split(email, "@")
	if len(parts) == 0 {
		return ""
	}

	username := parts[0]
	words := strings.Split(username, ".")

	for i, word := range words {
		words[i] = cases.Title(language.Und, cases.NoLower).String(word)
	}

	return strings.Join(words, " ")
}
