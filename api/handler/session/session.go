package session

import (
	"api/handler"
	"api/internal/http/aserr"
	"api/internal/http/asres"
	"api/internal/http/session"
	"api/internal/utils/authutils"
	"api/repository"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/guregu/null/v5"
)

type Handler struct {
	handler.BaseHandlerDeps
}

func NewHandler(deps handler.BaseHandlerDeps) *Handler {
	return &Handler{deps}
}

// GetSession returns current user session
//
//	@ID			get-session
//	@Summary	returns current user session
//	@Tags		auth
//	@Accept		json
//	@Produce	json
//	@Success	200	{object}	repository.Session
//	@Failure	401	{object}	aserr.ASError
//	@Failure	500	{object}	aserr.ASError
//	@Router		/auth/session [get]
func (h *Handler) GetSession(w http.ResponseWriter, r *http.Request) (*asres.Response[*repository.Session], error) {
	return asres.NewResponse(session.GetSessionFromRequest(r), http.StatusOK), nil
}

type CurrentUser struct {
	ID        uuid.UUID   `json:"id" validate:"required"`
	Name      string      `json:"name" validate:"required"`
	Email     string      `json:"email" validate:"required"`
	AvatarURL null.String `json:"avatar_url" validate:"required" swaggertype:"string" extensions:"x-nullable"`
	Timezone  string      `json:"timezone" validate:"required"`
} // @name CurrentUser

// GetUser returns current user from session
//
//	@ID			get-current-user
//	@Summary	returns current user from session, only call it when you sure that user is authenticated (e.g. you fetch get-session in your middleware)
//	@Tags		auth
//	@Accept		json
//	@Produce	json
//	@Success	200	{object}	CurrentUser
//	@Failure	401	{object}	aserr.ASError
//	@Failure	500	{object}	aserr.ASError
//	@Router		/auth/session/user [get]
func (h *Handler) GetUserFromSession(w http.ResponseWriter, r *http.Request) (*asres.Response[*CurrentUser], error) {
	session := session.GetSessionFromRequest(r)

	return asres.NewResponse(&CurrentUser{
		ID:        session.UserID,
		Name:      session.UserName,
		Email:     session.UserEmail,
		AvatarURL: session.UserAvatarURL,
		Timezone:  session.UserTimezone,
	}, http.StatusOK), nil
}

// Logout logs out the current user
//
//	@ID			logout
//	@Summary	logout the current user
//	@Tags		auth
//	@Accept		json
//	@Produce	json
//	@Success	200	{string}	string
//	@Failure	401	{object}	aserr.ASError
//	@Failure	500	{object}	aserr.ASError
//	@Router		/auth/session [delete]
func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) (*asres.Response[string], error) {
	session := session.GetSessionFromRequest(r)
	if err := h.Repository().DeleteSessionByKey(r.Context(), h.Redis(), repository.GetSessionKey(session.UserID.String(), session.ID.String())); err != nil {
		return nil, aserr.ErrInternalServerError
	}
	http.SetCookie(w, authutils.ClearSessionCookie(h.Config()))
	return asres.NewResponse("ok", http.StatusOK), nil
}

func (h *Handler) Router() chi.Router {
	r := chi.NewRouter()

	r.Group(func(r chi.Router) {
		r.Use(h.WithSessionCtx(session.WithAuthorizeUnverified(), session.WithRefresh()))
		r.Get("/", asres.ToHandlerFunc(h.GetSession))
	})
	r.Group(func(r chi.Router) {
		r.Use(h.WithSessionCtx(session.WithAuthorizeUnverified()))
		r.Get("/user", asres.ToHandlerFunc(h.GetUserFromSession))
		r.Delete("/", asres.ToHandlerFunc(h.Logout))
	})

	return r
}
