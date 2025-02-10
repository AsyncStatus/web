package repository

import (
	"api/internal/pg"
	"api/internal/sql"
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

type UserTimezone struct {
	ID     uuid.UUID `json:"id" validate:"required"`
	UserID uuid.UUID `json:"user_id" validate:"required"`
	// timezone IANA format (e.g., 'America/New_York' or 'Europe/Warsaw')
	Timezone string `json:"timezone" validate:"required"`
	// when this timezone became effective
	ValidFrom pgtype.Timestamp `json:"valid_from" validate:"required" swaggertype:"string"`
	// when this timezone was replaced (null if current)
	ValidUntil pgtype.Timestamp `json:"valid_until" validate:"required" swaggertype:"string" extensions:"x-nullable"`
} // @name UserTimezone

func (r *Repository) GetUserTimezoneByUserID(ctx context.Context, q pg.Querier, userID uuid.UUID) (*UserTimezone, error) {
	rows, _ := q.Query(ctx, `
		SELECT
			*
		FROM
			"user".timezones
		WHERE
			user_id = @user_id AND
			valid_until IS NULL
		ORDER BY
			valid_from DESC
		LIMIT 1
	`, pgx.StrictNamedArgs{"user_id": userID})
	nextUserTimezone, err := pgx.CollectExactlyOneRow(rows, pgx.RowToAddrOfStructByNameLax[UserTimezone])
	if err != nil {
		return nil, err
	}

	return nextUserTimezone, nil
}

type CreateUserTimezoneInput struct {
	UserID     uuid.UUID        `json:"user_id" validate:"required"`
	Timezone   string           `json:"timezone" validate:"required"`
	ValidFrom  pgtype.Timestamp `json:"valid_from" validate:"required" swaggertype:"string"`
	ValidUntil pgtype.Timestamp `json:"valid_until" validate:"required" swaggertype:"string" extensions:"x-nullable"`
}

func (r *Repository) CreateUserTimezone(ctx context.Context, q pg.Querier, input *CreateUserTimezoneInput) (*UserTimezone, error) {
	createSQL, args, err := sql.BuildCreateSQL(`"user".timezones`, input)
	if err != nil {
		return nil, err
	}
	rows, _ := q.Query(ctx, createSQL, args)
	nextUserTimezone, err := pgx.CollectExactlyOneRow(rows, pgx.RowToAddrOfStructByNameLax[UserTimezone])
	if err != nil {
		return nil, err
	}

	return nextUserTimezone, nil
}
