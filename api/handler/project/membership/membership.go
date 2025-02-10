package projectteammembership

import (
	"api/handler"
	"api/internal/http/aserr"
	"api/internal/http/asres"
	"api/repository"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/guregu/null/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

type Handler struct {
	handler.BaseHandlerDeps
}

func NewHandler(deps handler.BaseHandlerDeps) *Handler {
	return &Handler{deps}
}

// GetProjectTeamMemberships returns project team memberships
//
//	@ID			get-project-team-memberships
//	@Summary	get project team memberships
//	@Tags		projectteammembership
//	@Accept		json
//	@Produce	json
//	@Param		projectSlug	path		string	true	"Project slug"
//	@Param		teamSlug		path		string	true	"Team slug"
//	@Success	200				{object}	[]repository.ProjectTeamMembershipWithUser
//	@Failure	400				{object}	aserr.ASError
//	@Failure	401				{object}	aserr.ASError
//	@Failure	500				{object}	aserr.ASError
//	@Router		/projects/{projectSlug}/teams/{teamSlug}/memberships [get]
func (h *Handler) GetProjectTeamMemberships(w http.ResponseWriter, r *http.Request) (*asres.Response[[]repository.ProjectTeamMembershipWithUser], error) {
	ctx := r.Context()
	projectSlug := chi.URLParam(r, "projectSlug")
	teamSlug := chi.URLParam(r, "teamSlug")
	projectTeams, err := h.Repository().GetProjectTeamMemberships(ctx, h.Pg().Pool, repository.GetProjectTeamMembershipsInput{
		ProjectSlug: projectSlug,
		TeamSlug:    teamSlug,
	})
	if err != nil {
		return nil, err
	}

	return asres.NewResponse(projectTeams, http.StatusOK), nil
}

type CreateProjectTeamMembershipBody struct {
	UserID   uuid.UUID   `json:"userId" validate:"required,uuid4" swaggertype:"string"`
	Position null.String `json:"position,omitempty" validate:"omitempty,min=3,max=255" swaggertype:"string" extensions:"x-nullable"`
} //	@name	CreateProjectTeamMembershipBody

// CreateProjectTeamMembership creates project team membership
//
//	@ID			create-project-team-membership
//	@Summary	create project team membership
//	@Tags		projectteammembership
//	@Accept		json
//	@Produce	json
//	@Param		body			body		CreateProjectTeamMembershipBody	true	"Body"
//	@Param		projectSlug	path		string							true	"Project slug"
//	@Param		teamSlug		path		string							true	"Team slug"
//	@Success	200				{object}	repository.ProjectTeamMembership
//	@Failure	400				{object}	aserr.ASError
//	@Failure	401				{object}	aserr.ASError
//	@Failure	500				{object}	aserr.ASError
//	@Router		/projects/{projectSlug}/teams/{teamSlug}/memberships [post]
func (h *Handler) CreateProjectTeamMembership(w http.ResponseWriter, r *http.Request) (*asres.Response[*repository.ProjectTeamMembership], error) {
	ctx := r.Context()
	var body CreateProjectTeamMembershipBody
	if err := h.DecodeAndValidateInputBody(&body, r.Body); err != nil {
		return nil, err
	}

	projectSlug := chi.URLParam(r, "projectSlug")
	teamSlug := chi.URLParam(r, "teamSlug")
	team, err := h.Repository().GetProjectTeam(ctx, h.Pg().Pool, repository.GetProjectTeamInput{
		TeamSlug:    teamSlug,
		ProjectSlug: projectSlug,
	})
	if err != nil {
		return nil, err
	}

	projectTeam, err := h.Repository().CreateProjectTeamMembership(ctx, h.Pg().Pool, &repository.CreateProjectTeamMembershipInput{
		TeamID:    team.ID,
		UserID:    body.UserID,
		Position:  body.Position,
		CreatedAt: pgtype.Timestamp{Time: time.Now(), Valid: true},
	})
	if err != nil {
		return nil, err
	}

	return asres.NewResponse(projectTeam, http.StatusCreated), nil
}

// DeleteProjectTeamMembership deletes project team membership
//
//	@ID			delete-project-team-membership
//	@Summary	Delete project team membership
//	@Tags		ProjectTeamMembership
//	@Produce	json
//	@Param		id	path		string	true	"Project team membership ID"
//	@Success	200	{object}	repository.ProjectTeamMembership
//	@Failure	400	{object}	aserr.ASError
//	@Failure	401	{object}	aserr.ASError
//	@Failure	500	{object}	aserr.ASError
//	@Router		/projects/{project_slug}/teams/{team_slug}/memberships/{id} [delete]
func (h *Handler) DeleteProjectTeamMembership(w http.ResponseWriter, r *http.Request) (*asres.Response[*repository.ProjectTeamMembership], error) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		return nil, aserr.ErrBadRequest.NewWithError(err)
	}

	ctx := r.Context()
	projectTeamMembership, err := h.Repository().DeleteProjectTeamMembership(ctx, h.Pg().Pool, id)
	if err != nil {
		return nil, err
	}

	return asres.NewResponse(projectTeamMembership, http.StatusNoContent), nil
}

func (h *Handler) Router() chi.Router {
	r := chi.NewRouter()

	r.Use(h.WithSessionCtx())
	r.Get("/", asres.ToHandlerFunc(h.GetProjectTeamMemberships))
	r.Post("/", asres.ToHandlerFunc(h.CreateProjectTeamMembership))
	r.Delete("/{id}", asres.ToHandlerFunc(h.DeleteProjectTeamMembership))

	return r
}
