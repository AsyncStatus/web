package projectservice

import (
	"api/internal/http/aserr"
	"api/internal/pg"
	"api/repository"
	"api/service"
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/gosimple/slug"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	gonanoid "github.com/matoous/go-nanoid/v2"
)

type Service struct{ service.BaseServiceDeps }

func NewService(baseServiceDeps service.BaseServiceDeps) *Service {
	return &Service{BaseServiceDeps: baseServiceDeps}
}

type CreateProjectForUserInput struct {
	*repository.CreateProjectInput
	UserID uuid.UUID
}

type CreateProjectForUserOutput struct {
	Project           *repository.Project
	ProjectMembership *repository.ProjectMembership
}

func (s *Service) CreateProjectForUser(ctx context.Context, q pg.Querier, input *CreateProjectForUserInput) (*CreateProjectForUserOutput, error) {
	project, err := s.CreateProject(ctx, q, input.CreateProjectInput, 3)
	if err != nil {
		return nil, err
	}

	projectMembership, err := s.Repository().CreateProjectMembership(ctx, q, &repository.CreateProjectMembershipInput{
		UserID:    input.UserID,
		ProjectID: project.ID,
		Role:      "owner",
	})
	if err != nil {
		return nil, err
	}

	return &CreateProjectForUserOutput{project, projectMembership}, nil
}

type CreateProjectTeamForUserInput struct {
	*repository.CreateProjectTeamInput
	UserID uuid.UUID
}

type CreateProjectTeamForUserOutput struct {
	ProjectTeam           *repository.ProjectTeam
	ProjectTeamMembership *repository.ProjectTeamMembership
}

func (s *Service) CreateProjectTeamForUser(ctx context.Context, q pg.Querier, input *CreateProjectTeamForUserInput) (*CreateProjectTeamForUserOutput, error) {
	projectTeam, err := s.Repository().CreateProjectTeam(ctx, q, input.CreateProjectTeamInput)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation && pgErr.ConstraintName == "projects_slug_key" {
			return nil, aserr.ErrProjectTeamConflictName
		}

		return nil, aserr.ErrInternalServerError
	}

	projectTeamMembership, err := s.Repository().CreateProjectTeamMembership(ctx, q, &repository.CreateProjectTeamMembershipInput{
		UserID: input.UserID,
		TeamID: projectTeam.ID,
	})
	if err != nil {
		return nil, err
	}

	return &CreateProjectTeamForUserOutput{projectTeam, projectTeamMembership}, nil
}

func (s *Service) CreateProject(ctx context.Context, q pg.Querier, input *repository.CreateProjectInput, attempts int) (*repository.Project, error) {
	if attempts <= 0 {
		return nil, aserr.ErrInternalServerError
	}

	project, err := s.Repository().CreateProject(ctx, q, input)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation && pgErr.ConstraintName == "projects_slug_key" {
			slugID, err := gonanoid.Generate(nanoIDAlphabet, nanoIDLength)
			if err != nil {
				return nil, aserr.ErrProjectConflictName
			}
			input.Slug = slug.Make(input.Name) + slugID
			return s.CreateProject(ctx, q, input, attempts-1)
		}

		return nil, aserr.ErrInternalServerError
	}

	return project, nil
}

func (s *Service) CreateProjectTeam(ctx context.Context, q pg.Querier, input *repository.CreateProjectTeamInput, attempts int) (*repository.ProjectTeam, error) {
	if attempts <= 0 {
		return nil, aserr.ErrInternalServerError
	}

	team, err := s.Repository().CreateProjectTeam(ctx, q, input)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation && pgErr.ConstraintName == "projects_slug_key" {
			slugID, err := gonanoid.Generate(nanoIDAlphabet, nanoIDLength)
			if err != nil {
				return nil, aserr.ErrProjectTeamConflictName
			}
			input.Slug = slug.Make(input.Name) + slugID
			return s.CreateProjectTeam(ctx, q, input, attempts-1)
		}

		return nil, aserr.ErrInternalServerError
	}

	return team, nil
}

const nanoIDLength = 10
const nanoIDAlphabet = "abcdefghijklmnopqrstuvwxyz0123456789"
