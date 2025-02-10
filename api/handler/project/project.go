package project

import (
	"api/handler"
	"api/internal/http/aserr"
	"api/internal/http/asres"
	"api/internal/http/session"
	"api/repository"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/gosimple/slug"
	"github.com/jackc/pgx/v5"
	"go.uber.org/zap"
)

type Handler struct {
	handler.BaseHandlerDeps
}

func NewHandler(deps handler.BaseHandlerDeps) *Handler {
	return &Handler{deps}
}

// GetProject returns project
//
//	@ID			get-project
//	@Summary	get project
//	@Tags		project
//	@Accept		json
//	@Produce	json
//	@Param		project_slug	path		string	true	"Project slug"
//	@Success	200				{object}	repository.Project
//	@Failure	400				{object}	aserr.ASError
//	@Router		/projects/{project_slug} [get]
func (h *Handler) GetProject(w http.ResponseWriter, r *http.Request) (*asres.Response[*repository.Project], error) {
	ctx := r.Context()
	slug := chi.URLParam(r, "project_slug")
	project, err := h.Repository().GetProject(ctx, h.Pg().Pool, slug)
	if err != nil {
		return nil, err
	}

	return asres.NewResponse(project, http.StatusOK), nil
}

type CreateProjectBody struct {
	Name string `json:"name" validate:"required" minLength:"3" maxLength:"255"`
} //	@name	CreateProjectBody

// CreateProject creates project
//
//	@ID			create-project
//	@Summary	create project
//	@Tags		project
//	@Accept		json
//	@Produce	json
//	@Param		request	body		CreateProjectBody	true	"Create project body"
//	@Success	200		{object}	repository.Project
//	@Failure	400		{object}	aserr.ASError
//	@Router		/projects [post]
func (h *Handler) CreateProject(w http.ResponseWriter, r *http.Request) (*asres.Response[*repository.Project], error) {
	ctx := r.Context()
	var body CreateProjectBody
	if err := h.DecodeAndValidateInputBody(&body, r.Body); err != nil {
		return nil, err
	}

	tx, err := h.Pg().Pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, aserr.ErrInternalServerError
	}
	defer func() {
		if err != nil {
			if rollbackErr := tx.Rollback(ctx); rollbackErr != nil {
				h.Logger().Error("error rolling back transaction", zap.Error(rollbackErr))
			}
		}
	}()

	project, err := h.Repository().CreateProject(ctx, tx, &repository.CreateProjectInput{
		Name: body.Name,
		Slug: slug.Make(body.Name),
		Plan: "free",
	})
	if err != nil {
		return nil, err
	}

	if _, err := h.Repository().CreateProjectMembership(ctx, tx, &repository.CreateProjectMembershipInput{
		UserID:    session.GetSessionFromRequest(r).UserID,
		ProjectID: project.ID,
		Role:      "owner",
	}); err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	return asres.NewResponse(project, http.StatusOK), nil
}

func (h *Handler) Router() chi.Router {
	r := chi.NewRouter()

	r.Use(h.WithSessionCtx())
	r.Get("/{project_slug}", asres.ToHandlerFunc(h.GetProject))
	r.Post("/", asres.ToHandlerFunc(h.CreateProject))

	return r
}
