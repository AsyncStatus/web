package projectteam

import (
	"api/handler"
	"api/internal/http/aserr"
	"api/internal/http/asres"
	"api/internal/http/session"
	"api/repository"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/gosimple/slug"
	"github.com/guregu/null/v5"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"go.uber.org/zap"
)

type Handler struct {
	handler.BaseHandlerDeps
}

func NewHandler(deps handler.BaseHandlerDeps) *Handler {
	return &Handler{deps}
}

// GetProjectTeam returns project team
//
//	@ID			get-project-team
//	@Summary	get project team
//	@Tags		projectteam
//	@Accept		json
//	@Produce	json
//	@Param		projectSlug	path		string	true	"Project slug"
//	@Param		teamSlug	path		string	true	"Team slug"
//	@Success	200			{object}	repository.ProjectTeam
//	@Failure	400			{object}	aserr.ASError
//	@Failure	401			{object}	aserr.ASError
//	@Failure	500			{object}	aserr.ASError
//	@Router		/projects/{projectSlug}/teams/{teamSlug} [get]
func (h *Handler) GetProjectTeam(w http.ResponseWriter, r *http.Request) (*asres.Response[*repository.ProjectTeam], error) {
	ctx := r.Context()
	projectSlug := chi.URLParam(r, "projectSlug")
	teamSlug := chi.URLParam(r, "teamSlug")
	projectTeam, err := h.Repository().GetProjectTeam(ctx, h.Pg().Pool, repository.GetProjectTeamInput{
		ProjectSlug: projectSlug,
		TeamSlug:    teamSlug,
	})
	if err != nil {
		return nil, err
	}

	return asres.NewResponse(projectTeam, http.StatusOK), nil
}

// GetProjectTeams returns project teams
//
//	@ID			get-project-teams
//	@Summary	get project teams
//	@Tags		projectteam
//	@Accept		json
//	@Produce	json
//	@Param		projectSlug	path		string	true	"Project slug"
//	@Success	200			{object}	[]repository.ProjectTeam
//	@Failure	400			{object}	aserr.ASError
//	@Failure	401			{object}	aserr.ASError
//	@Failure	500			{object}	aserr.ASError
//	@Router		/projects/{projectSlug}/teams [get]
func (h *Handler) GetProjectTeams(w http.ResponseWriter, r *http.Request) (*asres.Response[[]*repository.ProjectTeam], error) {
	ctx := r.Context()
	projectSlug := chi.URLParam(r, "projectSlug")
	projectTeams, err := h.Repository().GetProjectTeams(ctx, h.Pg().Pool, projectSlug)
	if err != nil {
		return nil, err
	}

	return asres.NewResponse(projectTeams, http.StatusOK), nil
}

type CreateProjectTeamBody struct {
	Name  string      `json:"name" validate:"required" minLength:"3" maxLength:"255"`
	Emoji null.String `json:"emoji" swaggertype:"string" extensions:"x-nullable"`
}

// CreateProjectTeam creates project team
//
//	@ID			create-project-team
//	@Summary	create project team
//	@Tags		projectteam
//	@Accept		json
//	@Produce	json
//	@Param		request		body		CreateProjectTeamBody	true	"Create project team body"
//	@Param		projectSlug	path		string					true	"Project slug"
//	@Success	200			{object}	repository.ProjectTeam
//	@Failure	400			{object}	aserr.ASError
//	@Failure	401			{object}	aserr.ASError
//	@Failure	500			{object}	aserr.ASError
//	@Router		/projects/{projectSlug}/teams [post]
func (h *Handler) CreateProjectTeam(w http.ResponseWriter, r *http.Request) (*asres.Response[*repository.ProjectTeam], error) {
	ctx := r.Context()
	projectSlug := chi.URLParam(r, "projectSlug")
	var body CreateProjectTeamBody
	if err := h.DecodeAndValidateInputBody(&body, r.Body); err != nil {
		return nil, err
	}

	tx, err := h.Pg().Pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, err
	}
	defer func() {
		if err != nil {
			if rollbackErr := tx.Rollback(ctx); rollbackErr != nil {
				h.Logger().Error("error rolling back transaction", zap.Error(rollbackErr))
			}
		}
	}()

	project, err := h.Repository().GetProject(ctx, tx, projectSlug)
	if err != nil {
		return nil, err
	}

	projectTeam, err := h.Repository().CreateProjectTeam(ctx, tx, &repository.CreateProjectTeamInput{
		ProjectID: project.ID,
		Name:      body.Name,
		Emoji:     body.Emoji,
		Slug:      slug.Make(body.Name),
	})
	if err != nil {
		return nil, err
	}

	if _, err := h.Repository().CreateProjectTeamMembership(ctx, tx, &repository.CreateProjectTeamMembershipInput{
		UserID:    session.GetSessionFromRequest(r).UserID,
		TeamID:    projectTeam.ID,
		Position:  null.StringFromPtr(nil),
		CreatedAt: pgtype.Timestamp{Time: time.Now(), Valid: true},
	}); err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	return asres.NewResponse(projectTeam, http.StatusOK), nil
}

type UpdateProjectTeamBody struct {
	Name  string      `json:"name" validate:"required,min=3,max=255"`
	Emoji null.String `json:"emoji" swaggertype:"string" extensions:"x-nullable"`
} //	@name	UpdateProjectTeamBody

// UpdateProjectTeam updates project team
//
//	@ID			update-project-team
//	@Summary	update project team
//	@Tags		projectteam
//	@Accept		json
//	@Produce	json
//	@Param		id		path		string					true	"Project team ID"
//	@Param		body	body		UpdateProjectTeamBody	true	"Project team data"
//	@Success	200		{object}	repository.ProjectTeam
//	@Failure	400		{object}	aserr.ASError
//	@Failure	401		{object}	aserr.ASError
//	@Failure	500		{object}	aserr.ASError
//	@Router		/projects/{project_slug}/teams/{id} [put]
func (h *Handler) UpdateProjectTeam(w http.ResponseWriter, r *http.Request) (*asres.Response[*repository.ProjectTeam], error) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		return nil, aserr.ErrBadRequest.NewWithError(err)
	}

	var body UpdateProjectTeamBody
	if err := h.DecodeAndValidateInputBody(&body, r.Body); err != nil {
		return nil, err
	}

	ctx := r.Context()
	slug := slug.Make(body.Name)
	projectTeam, err := h.Repository().UpdateProjectTeam(
		ctx,
		h.Pg().Pool,
		&repository.UpdateProjectTeamFilter{ID: id},
		&repository.UpdateProjectTeamInput{Name: &body.Name, Emoji: &body.Emoji, Slug: &slug},
	)
	if err != nil {
		return nil, err
	}

	return asres.NewResponse(projectTeam, http.StatusOK), nil
}

func (h *Handler) Router() chi.Router {
	r := chi.NewRouter()

	r.Use(h.WithSessionCtx())
	r.Get("/{teamSlug}", asres.ToHandlerFunc(h.GetProjectTeam))
	r.Get("/", asres.ToHandlerFunc(h.GetProjectTeams))
	r.Post("/", asres.ToHandlerFunc(h.CreateProjectTeam))
	r.Put("/{id}", asres.ToHandlerFunc(h.UpdateProjectTeam))

	return r
}
