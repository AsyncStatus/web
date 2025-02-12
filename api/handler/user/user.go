package user

import (
	"api/handler"
	"api/internal/http/aserr"
	"api/internal/http/asres"
	"api/internal/sql"
	"api/repository"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type Handler struct {
	handler.BaseHandlerDeps
}

func NewHandler(deps handler.BaseHandlerDeps) *Handler {
	return &Handler{deps}
}

// GetUser returns user
//
//	@ID			get-user
//	@Summary	get user
//	@Tags		user
//	@Accept		json
//	@Produce	json
//	@Param		get_user_body	query		repository.GetUserBody	true	"Get user body"
//	@Success	200				{object}	repository.User
//	@Failure	400				{object}	aserr.ASError
//	@Router		/users [get]
func (h *Handler) GetUser(w http.ResponseWriter, r *http.Request) (*asres.Response[*repository.User], error) {
	var body repository.GetUserBody
	if err := h.DecodeAndValidateInputQuery(&body, r.URL.Query()); err != nil {
		return nil, err
	}

	ctx := r.Context()
	user, err := h.Repository().GetUser(ctx, h.Pg().Pool, &body)
	if err != nil {
		return nil, err
	}

	return asres.NewResponse(user, http.StatusOK), nil
}

// ListUsers lists users
//
//	@ID			list-users
//	@Summary	list users
//	@Tags		user
//	@Accept		json
//	@Produce	json
//	@Param		paginated_body	query		sql.PaginatedBody	true	"Paginated body"
//	@Success	200				{object}	sql.PaginatedData[repository.User]
//	@Router		/users/list [get]
func (h *Handler) ListUsers(w http.ResponseWriter, r *http.Request) (*asres.Response[*sql.PaginatedData[repository.User]], error) {
	var body sql.PaginatedBody
	if err := h.DecodeAndValidateInputQuery(&body, r.URL.Query()); err != nil {
		return nil, err
	}

	ctx := r.Context()
	list, err := h.Repository().ListUsers(ctx, h.Pg().Pool, &body)
	if err != nil {
		return nil, err
	}

	return asres.NewResponse(list, http.StatusOK), nil
}

type GetUserProjectsBody struct {
	Limit int `json:"limit" validate:"min=1,max=100"`
}

// GetUserProjects returns projects for user
//
//	@ID			get-user-projects
//	@Summary	get user projects
//	@Tags		user
//	@Accept		json
//	@Produce	json
//	@Param		userId	path		string				true	"User ID"
//	@Param		body	query		GetUserProjectsBody	true	"Get user projects body"
//	@Success	200		{object}	[]repository.Project
//	@Failure	401		{object}	aserr.ASError
//	@Failure	500		{object}	aserr.ASError
//	@Router		/users/{userId}/projects [get]
func (h *Handler) GetUserProjects(w http.ResponseWriter, r *http.Request) (*asres.Response[[]*repository.Project], error) {
	var body GetUserProjectsBody
	if err := h.DecodeAndValidateInputQuery(&body, r.URL.Query()); err != nil {
		return nil, err
	}

	userID, err := uuid.Parse(chi.URLParam(r, "userId"))
	if err != nil {
		return nil, aserr.ErrBadRequest.NewWithError(err)
	}

	projects, err := h.Repository().GetProjectsByUserID(r.Context(), h.Pg().Pool, userID, repository.WithLimit(body.Limit))
	if err != nil {
		return nil, err
	}

	return asres.NewResponse(projects, http.StatusOK), nil
}

// GetUserTeamMembership returns team membership for user
//
//	@ID			get-user-team-membership
//	@Summary	get user team membership
//	@Tags		user
//	@Accept		json
//	@Produce	json
//	@Param		userId		path		string	true	"User ID"
//	@Param		teamSlug	path		string	true	"Team slug"
//	@Param		projectSlug	path		string	true	"Project slug"
//	@Success	200			{object}	repository.ProjectTeamMembership
//	@Failure	401			{object}	aserr.ASError
//	@Failure	500			{object}	aserr.ASError
//	@Router		/users/{userId}/projects/{projectSlug}/memberships/{teamSlug} [get]
func (h *Handler) GetUserTeamMembership(w http.ResponseWriter, r *http.Request) (*asres.Response[*repository.ProjectTeamMembership], error) {
	userID, err := uuid.Parse(chi.URLParam(r, "userId"))
	if err != nil {
		return nil, aserr.ErrBadRequest.NewWithError(err)
	}

	teamSlug := chi.URLParam(r, "teamSlug")
	projectSlug := chi.URLParam(r, "projectSlug")
	projectTeam, err := h.Repository().GetProjectTeamMembership(r.Context(), h.Pg().Pool, repository.GetProjectTeamMembershipInput{
		UserID:      userID,
		ProjectSlug: projectSlug,
		TeamSlug:    teamSlug,
	})
	if err != nil {
		return nil, err
	}

	return asres.NewResponse(projectTeam, http.StatusOK), nil
}

func (h *Handler) Router() chi.Router {
	r := chi.NewRouter()

	r.Use(h.WithSessionCtx())
	r.Get("/", asres.ToHandlerFunc(h.GetUser))
	r.Get("/list", asres.ToHandlerFunc(h.ListUsers))
	r.Get("/{userId}/projects", asres.ToHandlerFunc(h.GetUserProjects))
	r.Get("/{userId}/projects/{projectSlug}/memberships/{teamSlug}", asres.ToHandlerFunc(h.GetUserTeamMembership))

	return r
}
