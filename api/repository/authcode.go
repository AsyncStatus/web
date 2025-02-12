package repository

import (
	"api/internal/http/aserr"
	"api/internal/pg"
	"api/internal/sql"
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

type AuthCode struct {
	ID        uuid.UUID          `json:"id" validate:"required"`
	UserID    uuid.UUID          `json:"user_id" validate:"required" swaggertype:"string"`
	Type      string             `json:"type" validate:"required"`
	Value     string             `json:"value" validate:"required"`
	ValueHash string             `json:"value_hash" validate:"required"`
	CreatedAt pgtype.Timestamptz `json:"created_at" validate:"required" swaggertype:"string"`
	ExpiresAt pgtype.Timestamptz `json:"expires_at" validate:"required" swaggertype:"string"`
} // @name AuthCode

type AuthCodeType string

const (
	AuthCodeTypeCreateAccount AuthCodeType = "create-account"
)

type GetAuthCodeInputOpts struct {
	UserID    *uuid.UUID    `json:"user_id" validate:"required"`
	Type      *AuthCodeType `json:"type" validate:"required"`
	ValueHash *string       `json:"value_hash" validate:"required" swaggertype:"string"`
}

type GetAuthCodeInputOpt func(input *GetAuthCodeInputOpts)

func WithGetAuthCodeInputUserID(userID uuid.UUID) GetAuthCodeInputOpt {
	return func(input *GetAuthCodeInputOpts) {
		input.UserID = &userID
	}
}

func WithGetAuthCodeInputType(authCodeType AuthCodeType) GetAuthCodeInputOpt {
	return func(input *GetAuthCodeInputOpts) {
		input.Type = &authCodeType
	}
}

func WithGetAuthCodeInputValueHash(valueHash string) GetAuthCodeInputOpt {
	return func(input *GetAuthCodeInputOpts) {
		input.ValueHash = &valueHash
	}
}

func (r *Repository) GetAuthCode(ctx context.Context, q pg.Querier, opts ...GetAuthCodeInputOpt) (*AuthCode, error) {
	input := &GetAuthCodeInputOpts{}
	for _, opt := range opts {
		opt(input)
	}

	wheres := make([]string, 0)
	args := pgx.StrictNamedArgs{}

	if input.UserID != nil {
		wheres = append(wheres, "user_id = @user_id")
		args["user_id"] = *input.UserID
	}

	if input.Type != nil {
		wheres = append(wheres, "type = @type")
		args["type"] = *input.Type
	}

	if input.ValueHash != nil {
		wheres = append(wheres, "value_hash = @value_hash")
		args["value_hash"] = *input.ValueHash
	}

	rows, _ := q.Query(ctx, fmt.Sprintf(`SELECT * FROM auth.codes WHERE %s`, strings.Join(wheres, " AND ")), args)
	authCode, err := pgx.CollectExactlyOneRow(rows, pgx.RowToAddrOfStructByNameLax[AuthCode])
	if err != nil && errors.Is(err, pgx.ErrNoRows) {
		return nil, aserr.ErrVerificationTokenInvalid
	}
	if err != nil {
		return nil, err
	}
	return authCode, nil
}

type CreateAuthCodeInput struct {
	UserID    uuid.UUID          `json:"user_id" validate:"required"`
	Type      AuthCodeType       `json:"type" validate:"required"`
	Value     string             `json:"value" validate:"required"`
	ValueHash string             `json:"value_hash" validate:"required"`
	CreatedAt pgtype.Timestamptz `json:"created_at" validate:"required" swaggertype:"string"`
	ExpiresAt pgtype.Timestamptz `json:"expires_at" validate:"required" swaggertype:"string"`
}

func (r *Repository) CreateAuthCode(ctx context.Context, q pg.Querier, input *CreateAuthCodeInput) (*AuthCode, error) {
	createSQL, args, err := sql.BuildCreateSQL("auth.codes", input)
	if err != nil {
		return nil, err
	}
	rows, _ := q.Query(ctx, createSQL, args)
	authCode, err := pgx.CollectExactlyOneRow(rows, pgx.RowToAddrOfStructByNameLax[AuthCode])
	if err != nil {
		return nil, err
	}

	return authCode, nil
}

func (r *Repository) DeleteAuthCode(ctx context.Context, q pg.Querier, id uuid.UUID) (*AuthCode, error) {
	rows, _ := q.Query(ctx, `DELETE FROM auth.codes WHERE id = $1 RETURNING *`, id)
	authCode, err := pgx.CollectExactlyOneRow(rows, pgx.RowToAddrOfStructByNameLax[AuthCode])
	if err != nil {
		return nil, err
	}

	return authCode, nil
}
