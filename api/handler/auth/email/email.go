package authemail

import (
	"api/handler"
	"api/internal/http/aserr"
	"api/internal/http/asres"
	"api/internal/utils/authutils"
	"api/internal/utils/timeutils"
	"api/repository"
	authservice "api/service/auth"
	"errors"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-redis/redis_rate/v10"
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

type SignUpEmailBody struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
	Timezone string `json:"timezone" validate:"required"`
} //	@name	SignUpEmailBody

func (h *Handler) SignUpEmail(w http.ResponseWriter, r *http.Request) (*asres.Response[any], error) {
	return nil, nil
}

type SignInEmailBody struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
} //	@name	SignInEmailBody

// SignInEmail signs in user by email
//
//	@ID			sign-in-email
//	@Summary	signs in user by email
//	@Tags		auth
//	@Accept		json
//	@Produce	json
//	@Param		sign_in_email_body	body		SignInEmailBody	true	"Sign in email body"
//	@Success	200					{object}	repository.Session
//	@Failure	401					{object}	aserr.ASError
//	@Failure	500					{object}	aserr.ASError
//	@Router		/auth/email/sign-in [post]
func (h *Handler) SignInEmail(w http.ResponseWriter, r *http.Request) (*asres.Response[*repository.Session], error) {
	var body SignInEmailBody
	if err := h.DecodeAndValidateInputBody(&body, r.Body); err != nil {
		return nil, err
	}

	ip, err := h.RealIP(r)
	if err != nil {
		return nil, err
	}

	ctx := r.Context()
	authenticateOutput, err := h.authService.Authenticate(ctx, h.Pg().Pool, h.Redis(), &authservice.AuthenticateInput{
		Email:    body.Email,
		Password: body.Password,
	})
	if err != nil && errors.Is(err, aserr.ErrInvalidEmailOrPassword) {
		limitOutput, err := h.RedisRateLimiter().AllowN(ctx, fmt.Sprintf("%s:sign-in-email", ip), redis_rate.PerMinute(10), 1)
		if err != nil {
			return nil, err
		}
		if limitOutput.Allowed <= 0 {
			return nil, aserr.ErrTooManyRequests.NewWithError(errors.New(aserr.ErrTooManyRequests.Error() + ", try again in " + timeutils.RoundToSingleUnit(limitOutput.RetryAfter)))
		}

		return nil, aserr.ErrInvalidEmailOrPassword
	}
	if err != nil {
		return nil, err
	}

	http.SetCookie(w, authutils.NewSessionCookie(h.Config(), authenticateOutput.Session, authenticateOutput.Session.CreatedAt))
	h.RedisRateLimiter().Reset(ctx, fmt.Sprintf("%s:sign-in-email", ip))

	return asres.NewResponse(authenticateOutput.Session, http.StatusOK), nil
}

func (h *Handler) Router() chi.Router {
	r := chi.NewRouter()

	r.Post("/sign-up", asres.ToHandlerFunc(h.SignUpEmail))
	r.Post("/sign-in", asres.ToHandlerFunc(h.SignInEmail))

	return r
}
