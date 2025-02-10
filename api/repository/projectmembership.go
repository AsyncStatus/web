package repository

import (
	"api/internal/pg"
	"api/internal/sql"
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

type ProjectMembership struct {
	ID        uuid.UUID        `json:"id" validate:"required"`
	UserID    uuid.UUID        `json:"user_id" validate:"required"`
	ProjectID uuid.UUID        `json:"project_id" validate:"required"`
	Role      string           `json:"role" validate:"required"`
	CreatedAt pgtype.Timestamp `json:"created_at" validate:"required" swaggertype:"string"`
	UpdatedAt pgtype.Timestamp `json:"updated_at" validate:"required" swaggertype:"string" extensions:"x-nullable"`
} // @name ProjectMembership

type CreateProjectMembershipInput struct {
	UserID    uuid.UUID `json:"user_id" validate:"required"`
	ProjectID uuid.UUID `json:"project_id" validate:"required"`
	Role      string    `json:"role" validate:"required" minLength:"3" maxLength:"50"`
}

func (r *Repository) CreateProjectMembership(ctx context.Context, q pg.Querier, input *CreateProjectMembershipInput) (*ProjectMembership, error) {
	createSQL, args, err := sql.BuildCreateSQL(`project.memberships`, input)
	if err != nil {
		return nil, err
	}
	rows, _ := q.Query(ctx, createSQL, args)
	membership, err := pgx.CollectExactlyOneRow(rows, pgx.RowToAddrOfStructByNameLax[ProjectMembership])
	if err != nil {
		return nil, err
	}

	return membership, nil
}

func (r *Repository) GetProjectMembershipsByUserID(ctx context.Context, q pg.Querier, userID uuid.UUID) ([]ProjectMembership, error) {
	rows, _ := q.Query(ctx, `SELECT * FROM project.memberships WHERE @user_id = $1`, pgx.StrictNamedArgs{"user_id": userID})
	return pgx.CollectRows(rows, pgx.RowToStructByName[ProjectMembership])
}

func (r *Repository) GetUserProjectOwner(ctx context.Context, q pg.Querier, projectSlug string) (*User, error) {
	rows, _ := q.Query(ctx, `
		SELECT
			u.*
		FROM
			project.memberships m
		JOIN
			project.projects p ON m.project_id = p.id
		JOIN
			"user".users u ON m.user_id = u.id
		WHERE
			m.role = 'owner' AND
			p.slug = @project_slug
	`, pgx.StrictNamedArgs{"project_slug": projectSlug})
	user, err := pgx.CollectExactlyOneRow(rows, pgx.RowToAddrOfStructByNameLax[User])
	if err != nil {
		return nil, err
	}

	return user, nil
}
