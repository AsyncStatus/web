package repository

import (
	"api/internal/http/aserr"
	"api/internal/pg"
	"api/internal/sql"
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/guregu/null/v5"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
)

type ProjectTeam struct {
	ID            uuid.UUID        `json:"id" validate:"required"`
	ProjectID     uuid.UUID        `json:"project_id" validate:"required"`
	Name          string           `json:"name" validate:"required"`
	Slug          string           `json:"slug" validate:"required"`
	Emoji         null.String      `json:"emoji" validate:"required" swaggertype:"string" extensions:"x-nullable"`
	ParentTeamID  uuid.NullUUID    `json:"parent_team_id" validate:"required" swaggertype:"string" extensions:"x-nullable"`
	MadePrivateAt pgtype.Timestamp `json:"made_private_at" swaggertype:"string" extensions:"x-nullable"`
	CreatedAt     pgtype.Timestamp `json:"created_at" validate:"required" swaggertype:"string"`
	UpdatedAt     pgtype.Timestamp `json:"updated_at" validate:"required" swaggertype:"string" extensions:"x-nullable"`
	ArchivedAt    pgtype.Timestamp `json:"archived_at" validate:"required" swaggertype:"string" extensions:"x-nullable"`
} // @name ProjectTeam

type GetProjectTeamInput struct {
	ProjectSlug string
	TeamSlug    string
}

func (r *Repository) GetProjectTeam(ctx context.Context, q pg.Querier, input GetProjectTeamInput) (*ProjectTeam, error) {
	rows, _ := q.Query(ctx, `
		SELECT
			t.*
		FROM
			project.teams t
		JOIN
			project.projects p ON t.project_id = p.id
		WHERE
			p.slug = @project_slug AND
			t.slug = @team_slug
	`, pgx.StrictNamedArgs{"project_slug": input.ProjectSlug, "team_slug": input.TeamSlug})
	team, err := pgx.CollectExactlyOneRow(rows, pgx.RowToAddrOfStructByNameLax[ProjectTeam])
	if err != nil && errors.Is(err, pgx.ErrNoRows) {
		return nil, aserr.ErrProjectTeamNotFound
	}
	if err != nil {
		return nil, err
	}
	return team, nil
}

type CreateProjectTeamInput struct {
	ProjectID    uuid.UUID     `json:"project_id" validate:"required"`
	Name         string        `json:"name" validate:"required" minLength:"3" maxLength:"255"`
	Slug         string        `json:"slug" validate:"required" minLength:"3" maxLength:"255"`
	Emoji        null.String   `json:"emoji"`
	ParentTeamID uuid.NullUUID `json:"parent_team_id"`
}

func (r *Repository) CreateProjectTeam(ctx context.Context, q pg.Querier, input *CreateProjectTeamInput) (*ProjectTeam, error) {
	createSQL, args, err := sql.BuildCreateSQL(`project.teams`, input)
	if err != nil {
		return nil, err
	}
	rows, _ := q.Query(ctx, createSQL, args)
	team, err := pgx.CollectExactlyOneRow(rows, pgx.RowToAddrOfStructByNameLax[ProjectTeam])
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation && pgErr.ConstraintName == "teams_slug_key" {
			return nil, aserr.ErrProjectTeamConflictName
		}

		return nil, err
	}

	return team, nil
}

func (r *Repository) GetProjectTeams(ctx context.Context, q pg.Querier, projectSlug string) ([]*ProjectTeam, error) {
	rows, _ := q.Query(ctx, `
		SELECT
			t.*
		FROM
			project.teams t
		JOIN
			project.projects p ON t.project_id = p.id
		WHERE
			p.slug = @project_slug
	`, pgx.StrictNamedArgs{"project_slug": projectSlug})
	teams, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByNameLax[ProjectTeam])
	if err != nil {
		return nil, err
	}

	return teams, nil
}

type UpdateProjectTeamFilter struct {
	ID uuid.UUID `json:"id"`
}

type UpdateProjectTeamInput struct {
	Name         *string        `json:"name" validate:"required" minLength:"3" maxLength:"255"`
	Slug         *string        `json:"slug" validate:"required" minLength:"3" maxLength:"255"`
	Emoji        *null.String   `json:"emoji"`
	ParentTeamID *uuid.NullUUID `json:"parent_team_id"`
}

func (r *Repository) UpdateProjectTeam(ctx context.Context, q pg.Querier, filter *UpdateProjectTeamFilter, input *UpdateProjectTeamInput) (*ProjectTeam, error) {
	updateSQL, args, err := sql.BuildUpdateSQLV2(`project.teams`, filter, input)
	if err != nil {
		return nil, err
	}
	rows, _ := q.Query(ctx, updateSQL, args)
	team, err := pgx.CollectExactlyOneRow(rows, pgx.RowToAddrOfStructByNameLax[ProjectTeam])
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation && pgErr.ConstraintName == "teams_slug_key" {
			return nil, aserr.ErrProjectTeamConflictName
		}

		return nil, err
	}

	return team, nil
}
