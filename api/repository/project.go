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

type Project struct {
	ID         uuid.UUID        `json:"id" validate:"required"`
	Name       string           `json:"name" validate:"required"`
	Slug       string           `json:"slug" validate:"required"`
	Plan       string           `json:"plan" validate:"required"`
	AvatarURL  null.String      `json:"avatar_url" validate:"required" swaggertype:"string" extensions:"x-nullable"`
	CreatedAt  pgtype.Timestamp `json:"created_at" validate:"required" swaggertype:"string"`
	UpdatedAt  pgtype.Timestamp `json:"updated_at" validate:"required" swaggertype:"string" extensions:"x-nullable"`
	ArchivedAt pgtype.Timestamp `json:"-" validate:"required" swaggertype:"string" extensions:"x-nullable"`
} // @name Project

func (r *Repository) GetProject(ctx context.Context, q pg.Querier, idOrSlug string) (*Project, error) {
	rows, _ := q.Query(ctx, `SELECT p.* FROM project.projects p WHERE @slug_or_id = ANY(ARRAY[slug, id::text])`, pgx.StrictNamedArgs{"slug_or_id": idOrSlug})
	project, err := pgx.CollectExactlyOneRow(rows, pgx.RowToAddrOfStructByNameLax[Project])
	if err != nil && errors.Is(err, pgx.ErrNoRows) {
		return nil, aserr.ErrProjectNotFound
	}
	if err != nil {
		return nil, err
	}

	return project, nil
}

func (r *Repository) GetProjectByID(ctx context.Context, q pg.Querier, id uuid.UUID) (*Project, error) {
	rows, _ := q.Query(ctx, `SELECT p.* FROM project.projects p WHERE id = $1`, id)
	project, err := pgx.CollectExactlyOneRow(rows, pgx.RowToAddrOfStructByNameLax[Project])
	if err != nil {
		return nil, err
	}

	return project, nil
}

type CreateProjectInput struct {
	Name string `json:"name" validate:"required" minLength:"3" maxLength:"255"`
	Plan string `json:"plan"`
	Slug string `json:"slug"`
} // @name CreateProjectInput

func (r *Repository) CreateProject(ctx context.Context, q pg.Querier, input *CreateProjectInput) (*Project, error) {
	createSQL, args, err := sql.BuildCreateSQL(`project.projects`, input)
	if err != nil {
		return nil, err
	}
	rows, _ := q.Query(ctx, createSQL, args)
	project, err := pgx.CollectExactlyOneRow(rows, pgx.RowToAddrOfStructByNameLax[Project])
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation && pgErr.ConstraintName == "projects_slug_key" {
			return nil, aserr.ErrProjectTeamConflictName
		}

		return nil, err
	}

	return project, nil
}

type GetProjectsByUserIDOpts struct {
	Limit int
}

type GetProjectsByUserIDOptsFunc func(options *GetProjectsByUserIDOpts)

func WithLimit(limit int) GetProjectsByUserIDOptsFunc {
	return func(options *GetProjectsByUserIDOpts) {
		options.Limit = limit
	}
}

func (r *Repository) GetProjectsByUserID(ctx context.Context, q pg.Querier, userID uuid.UUID, opts ...GetProjectsByUserIDOptsFunc) ([]*Project, error) {
	options := GetProjectsByUserIDOpts{Limit: 100}
	for _, opt := range opts {
		opt(&options)
	}

	rows, _ := q.Query(ctx, `
		SELECT
			p.*
		FROM
			project.memberships m
		JOIN
			project.projects p ON p.id = m.project_id
		WHERE
			m.user_id = @user_id
		ORDER BY
			p.created_at ASC
		LIMIT
			@limit
		`, pgx.StrictNamedArgs{"user_id": userID, "limit": options.Limit})
	projects, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByNameLax[Project])
	if err != nil {
		return nil, err
	}

	return projects, nil
}
