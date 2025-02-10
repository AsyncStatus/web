package repository

import (
	"database/sql"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

type AuthCode struct {
	ID        uuid.UUID        `json:"id" validate:"required"`
	UserID    uuid.UUID        `json:"user_id" validate:"required"`
	Type      string           `json:"type" validate:"required"`
	Value     string           `json:"value" validate:"required"`
	ValueHash string           `json:"value_hash" validate:"required"`
	CreatedAt pgtype.Timestamp `json:"created_at" validate:"required" swaggertype:"string"`
	ExpiresAt pgtype.Timestamp `json:"expires_at" validate:"required" swaggertype:"string"`
}

type GetAuthCodeInput struct {
	UserID    uuid.UUID        `json:"user_id" validate:"required"`
	Type      string           `json:"type" validate:"required"`
	Value     sql.Null[string] `json:"value" validate:"required" swaggertype:"string"`
	ValueHash sql.Null[string] `json:"value_hash" validate:"required" swaggertype:"string"`
}
