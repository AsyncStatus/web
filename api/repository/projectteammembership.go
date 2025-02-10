package repository

import (
	"api/internal/http/aserr"
	"api/internal/pg"
	"api/internal/sql"
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/guregu/null/v5"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

type ProjectTeamMembership struct {
	ID         uuid.UUID        `json:"id" validate:"required"`
	UserID     uuid.UUID        `json:"user_id" validate:"required"`
	TeamID     uuid.UUID        `json:"team_id" validate:"required"`
	Position   null.String      `json:"position" validate:"required" swaggertype:"string" extensions:"x-nullable"`
	CreatedAt  pgtype.Timestamp `json:"created_at" validate:"required" swaggertype:"string"`
	ArchivedAt pgtype.Timestamp `json:"archived_at" validate:"required" swaggertype:"string" extensions:"x-nullable"`
} // @name ProjectTeamMembership

type ProjectTeamMembershipWithUser struct {
	ID         uuid.UUID        `json:"id" validate:"required"`
	UserID     uuid.UUID        `json:"user_id" validate:"required"`
	User       User             `json:"user" validate:"required"`
	TeamID     uuid.UUID        `json:"team_id" validate:"required"`
	Position   null.String      `json:"position" validate:"required" swaggertype:"string" extensions:"x-nullable"`
	CreatedAt  pgtype.Timestamp `json:"created_at" validate:"required" swaggertype:"string"`
	ArchivedAt pgtype.Timestamp `json:"archived_at" validate:"required" swaggertype:"string" extensions:"x-nullable"`
} // @name ProjectTeamMembershipWithUser

type GetProjectTeamMembershipInput struct {
	UserID      uuid.UUID
	ProjectSlug string
	TeamSlug    string
}

func (r *Repository) GetProjectTeamMembership(ctx context.Context, q pg.Querier, input GetProjectTeamMembershipInput) (*ProjectTeamMembership, error) {
	rows, _ := q.Query(ctx, `
		SELECT
			tm.*
		FROM
			project.team_memberships tm
		JOIN
			project.teams t ON tm.team_id = t.id
		JOIN
			project.projects p ON t.project_id = p.id
		WHERE
			p.slug = @project_slug AND
			t.slug = @team_slug AND
			tm.user_id = @user_id
	`, pgx.StrictNamedArgs{"project_slug": input.ProjectSlug, "team_slug": input.TeamSlug, "user_id": input.UserID})
	team, err := pgx.CollectExactlyOneRow(rows, pgx.RowToAddrOfStructByNameLax[ProjectTeamMembership])
	if err != nil && errors.Is(err, pgx.ErrNoRows) {
		return nil, aserr.ErrProjectTeamMembershipNotFound
	}
	if err != nil {
		return nil, err
	}
	return team, nil
}

type GetProjectTeamMembershipsInput struct {
	ProjectSlug, TeamSlug string
}

func (r *Repository) GetProjectTeamMemberships(ctx context.Context, q pg.Querier, input GetProjectTeamMembershipsInput) ([]ProjectTeamMembershipWithUser, error) {
	rows, _ := q.Query(ctx, `
		SELECT
			tm.*,
			row(u.id, u.name, u.email, u.referrer, u.avatar_url, u.created_at, u.verified_at, u.updated_at, u.archived_at, u.approved_at) as "user"
		FROM
			project.team_memberships tm
		JOIN
			project.teams t ON tm.team_id = t.id
		JOIN
			"user".users u ON tm.user_id = u.id
		JOIN
			project.projects p ON p.id = t.project_id
		WHERE
			p.slug = @project_slug AND
			t.slug = @team_slug
	`, pgx.StrictNamedArgs{"project_slug": input.ProjectSlug, "team_slug": input.TeamSlug})
	teams, err := pgx.CollectRows(rows, pgx.RowToStructByNameLax[ProjectTeamMembershipWithUser])
	if err != nil {
		return nil, err
	}
	return teams, nil
}

type CreateProjectTeamMembershipInput struct {
	UserID     uuid.UUID        `json:"user_id" validate:"required"`
	TeamID     uuid.UUID        `json:"team_id" validate:"required"`
	Position   null.String      `json:"position" validate:"required" swaggertype:"string" extensions:"x-nullable"`
	CreatedAt  pgtype.Timestamp `json:"created_at" validate:"required" swaggertype:"string" extensions:"x-nullable"`
	ArchivedAt pgtype.Timestamp `json:"archived_at" swaggertype:"string" extensions:"x-nullable"`
}

func (r *Repository) CreateProjectTeamMembership(ctx context.Context, q pg.Querier, input *CreateProjectTeamMembershipInput) (*ProjectTeamMembership, error) {
	createSQL, args, err := sql.BuildCreateSQL(`project.team_memberships`, input)
	if err != nil {
		return nil, err
	}
	rows, _ := q.Query(ctx, createSQL, args)
	team, err := pgx.CollectExactlyOneRow(rows, pgx.RowToAddrOfStructByNameLax[ProjectTeamMembership])
	if err != nil {
		return nil, err
	}

	return team, nil
}

func (r *Repository) DeleteProjectTeamMembership(ctx context.Context, q pg.Querier, id uuid.UUID) (*ProjectTeamMembership, error) {
	rows, _ := q.Query(ctx, `
		DELETE FROM
			project.team_memberships
		WHERE
			id = @id
		RETURNING *
	`, pgx.StrictNamedArgs{"id": id})
	team, err := pgx.CollectExactlyOneRow(rows, pgx.RowToAddrOfStructByNameLax[ProjectTeamMembership])
	if err != nil {
		return nil, err
	}
	return team, nil
}
