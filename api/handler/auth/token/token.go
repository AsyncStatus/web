package authtoken

import (
	"api/handler"
	"api/internal/http/asres"
	"api/internal/utils/authutils"
	"api/repository"
	authservice "api/service/auth"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"
	"go.uber.org/zap"
)

type Handler struct {
	handler.BaseHandlerDeps
	authService *authservice.Service
}
type HandlerOption func(*Handler)

func WithBaseHandlerDeps(deps handler.BaseHandlerDeps) HandlerOption {
	return func(h *Handler) { h.BaseHandlerDeps = deps }
}
func WithAuthService(authService *authservice.Service) HandlerOption {
	return func(h *Handler) { h.authService = authService }
}
func NewHandler(opts ...HandlerOption) *Handler {
	h := &Handler{}
	for _, opt := range opts {
		opt(h)
	}
	return h
}

type SignUpTokenBody struct {
	Token    string `json:"token" validate:"required"`
	Password string `json:"password" validate:"required,min=8,max=255"`
	Timezone string `json:"timezone" validate:"required"`
} //	@name	SignUpTokenBody

// SignUpToken signs up user by token
//
//	@ID			sign-up-token
//	@Summary	signs up user by token
//	@Tags		auth
//	@Accept		json
//	@Produce	json
//	@Param		sign_up_token_body	body		SignUpTokenBody	true	"Sign up token body"
//	@Success	200					{object}	repository.Session
//	@Failure	401					{object}	aserr.ASError
//	@Failure	500					{object}	aserr.ASError
//	@Router		/auth/token/sign-up [post]
func (h *Handler) SignUpToken(w http.ResponseWriter, r *http.Request) (*asres.Response[*repository.Session], error) {
	var body SignUpTokenBody
	if err := h.DecodeAndValidateInputBody(&body, r.Body); err != nil {
		return nil, err
	}

	ctx := r.Context()
	tx, err := h.Pg().Pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, err
	}
	signUpTokenOutput, err := h.authService.SignUpToken(ctx, tx, h.Redis(), &authservice.SignUpTokenInput{
		Token:    body.Token,
		Password: body.Password,
		Timezone: body.Timezone,
	})
	if err != nil {
		if rollbackErr := tx.Rollback(ctx); rollbackErr != nil {
			h.Logger().Error("error rolling back transaction", zap.Error(rollbackErr))
		}

		return nil, err
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	http.SetCookie(w, authutils.NewSessionCookie(h.Config(), signUpTokenOutput.Session, signUpTokenOutput.Session.CreatedAt))

	return asres.NewResponse(signUpTokenOutput.Session, http.StatusOK), nil
}

func (h *Handler) Router() chi.Router {
	r := chi.NewRouter()

	r.Post("/sign-up", asres.ToHandlerFunc(h.SignUpToken))

	return r
}
