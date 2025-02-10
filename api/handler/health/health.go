package health

import (
	"api/handler"
	"api/internal/http/aserr"
	"api/internal/http/asres"
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
)

type Handler struct {
	handler.BaseHandlerDeps
}

func NewHandler(deps handler.BaseHandlerDeps) *Handler {
	return &Handler{deps}
}

type HealthResponse struct {
	PostgresPing string `json:"postgres_ping" swaggertype:"string"`
	RedisPing    string `json:"redis_ping" swaggertype:"string"`
}

// GetHealth returns health
//
//	@ID			get-health
//	@Summary	get health
//	@Tags		health
//	@Success	200	{object}	HealthResponse
//	@Failure	500	{object}	aserr.ASError
//	@Failure	503	{object}	aserr.ASError
//	@Router		/health [get]
func (h *Handler) GetHealth(w http.ResponseWriter, r *http.Request) (*asres.Response[HealthResponse], error) {
	ctx := r.Context()
	startPG := time.Now()
	if err := h.Pg().Ping(ctx); err != nil {
		h.SlackClient().PostToStatusChannel(ctx, fmt.Sprintf("error pinging postgres: %s", err.Error()))
		return nil, aserr.ErrPgUnavailable
	}
	pgPingTime := time.Since(startPG)

	startRedis := time.Now()
	if err := h.Redis().Ping(ctx).Err(); err != nil {
		h.SlackClient().PostToStatusChannel(ctx, fmt.Sprintf("error pinging redis: %s", err.Error()))
		return nil, aserr.ErrRedisUnavailable
	}
	redisPingTime := time.Since(startRedis)

	return asres.NewResponse(HealthResponse{PostgresPing: fmt.Sprintf("%dµs", pgPingTime.Microseconds()), RedisPing: fmt.Sprintf("%dµs", redisPingTime.Microseconds())}, http.StatusOK), nil
}

func (h *Handler) Router() chi.Router {
	r := chi.NewRouter()

	r.Get("/", asres.ToHandlerFunc(h.GetHealth))

	return r
}
