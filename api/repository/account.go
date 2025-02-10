package repository

import (
	"api/internal/pg"
	"api/internal/sql"
	"context"
	sqltype "database/sql"

	"github.com/google/uuid"
	"github.com/guregu/null/v5"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

type Account struct {
	ID                    uuid.UUID        `json:"id" validate:"required"`
	UserID                uuid.UUID        `json:"user_id" validate:"required"`
	AccountID             string           `json:"account_id" validate:"required"`
	ProviderID            string           `json:"provider_id" validate:"required"`
	AccessToken           null.String      `json:"access_token" validate:"required" swaggertype:"string" extensions:"x-nullable"`
	RefreshToken          null.String      `json:"refresh_token" validate:"required" swaggertype:"string" extensions:"x-nullable"`
	Scope                 null.String      `json:"scope" validate:"required" swaggertype:"string" extensions:"x-nullable"`
	PasswordHash          null.String      `json:"-" validate:"required" swaggertype:"string" extensions:"x-nullable"`
	CreatedAt             pgtype.Timestamp `json:"created_at" validate:"required" swaggertype:"string"`
	AccessTokenExpiresAt  pgtype.Timestamp `json:"access_token_expires_at" validate:"required" swaggertype:"string" extensions:"x-nullable"`
	RefreshTokenExpiresAt pgtype.Timestamp `json:"refresh_token_expires_at" validate:"required" swaggertype:"string" extensions:"x-nullable"`
	UpdatedAt             pgtype.Timestamp `json:"updated_at" validate:"required" swaggertype:"string" extensions:"x-nullable"`
	ArchivedAt            pgtype.Timestamp `json:"-" validate:"required" swaggertype:"string" extensions:"x-nullable"`
	VerifiedAt            pgtype.Timestamp `json:"verified_at" validate:"required" swaggertype:"string" extensions:"x-nullable"`
} // @name Account

type GetAccountInput struct {
	ID     *uuid.UUID `json:"id" validate:"omitempty,uuid4" swaggertype:"string"`
	UserID *uuid.UUID `json:"user_id" validate:"omitempty,uuid4" swaggertype:"string"`
}

func (r *Repository) GetAccount(ctx context.Context, q pg.Querier, input *GetAccountInput) (*Account, error) {
	rows, _ := q.Query(ctx, `
		SELECT
			a.*
		FROM
			auth.accounts a
		WHERE
			(COALESCE(@id::uuid, NULL) IS NULL OR a.id = @id)
			AND (COALESCE(@user_id::uuid, NULL) IS NULL OR a.user_id = @user_id)
	`, pgx.StrictNamedArgs{"id": input.ID, "user_id": input.UserID})
	account, err := pgx.CollectExactlyOneRow(rows, pgx.RowToAddrOfStructByNameLax[Account])
	if err != nil {
		return nil, err
	}

	return account, nil
}

type CreateAccountInput struct {
	ID                    uuid.UUID          `json:"id" validate:"required"`
	UserID                uuid.UUID          `json:"user_id" validate:"required"`
	AccountID             string             `json:"account_id" validate:"required"`
	ProviderID            sqltype.NullString `json:"provider_id"`
	AccessToken           sqltype.NullString `json:"access_token"`
	RefreshToken          sqltype.NullString `json:"refresh_token"`
	Scope                 sqltype.NullString `json:"scope"`
	PasswordHash          sqltype.NullString `json:"password_hash"`
	AccessTokenExpiresAt  pgtype.Timestamp   `json:"access_token_expires_at" swaggertype:"string" extensions:"x-nullable"`
	RefreshTokenExpiresAt pgtype.Timestamp   `json:"refresh_token_expires_at" swaggertype:"string" extensions:"x-nullable"`
}

func (r *Repository) CreateAccount(ctx context.Context, q pg.Querier, input *CreateAccountInput) (*Account, error) {
	createSQL, args, err := sql.BuildCreateSQL("auth.accounts", input)
	if err != nil {
		return nil, err
	}
	rows, _ := q.Query(ctx, createSQL, args)
	account, err := pgx.CollectExactlyOneRow(rows, pgx.RowToAddrOfStructByNameLax[Account])
	if err != nil {
		return nil, err
	}

	return account, nil
}
