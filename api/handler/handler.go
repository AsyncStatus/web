package handler

import (
	"api/config"
	"api/internal/email"
	"api/internal/http/aserr"
	"api/internal/http/session"
	"api/internal/log"
	"api/internal/pg"
	"api/internal/redis"
	slack "api/internal/slack/client"
	"api/internal/utils/netutils"
	"api/repository"
	"encoding/json"
	"io"
	"net/http"

	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	redisrate "github.com/go-redis/redis_rate/v10"
	"github.com/gorilla/schema"
)

type BaseHandlerDeps interface {
	Config() *config.Config
	Logger() *log.Logger
	Pg() *pg.Pg
	Redis() *redis.Client
	RedisRateLimiter() *redisrate.Limiter
	Validate() *validator.Validate
	Translator() *ut.Translator
	Repository() *repository.Repository
	SlackClient() *slack.Client
	EmailClient() *email.Client
	DecodeAndValidateInputQuery(dst interface{}, src map[string][]string) *aserr.ASError
	DecodeAndValidateInputBody(dst interface{}, src io.ReadCloser) *aserr.ASError
	RealIP(r *http.Request) (string, error)
	WithSessionCtx(opts ...session.SessionOptionsFunc) func(next http.Handler) http.Handler
}

type baseHandlerDeps struct {
	cfg          *config.Config
	log          *log.Logger
	pg           *pg.Pg
	redis        *redis.Client
	redisLimiter *redisrate.Limiter
	validate     *validator.Validate
	translator   *ut.Translator
	slackClient  *slack.Client
	emailClient  *email.Client
	repository   *repository.Repository
}

func (h *baseHandlerDeps) Config() *config.Config {
	return h.cfg
}

func (h *baseHandlerDeps) Logger() *log.Logger {
	return h.log
}

func (h *baseHandlerDeps) Pg() *pg.Pg {
	return h.pg
}

func (h *baseHandlerDeps) Redis() *redis.Client {
	return h.redis
}

func (h *baseHandlerDeps) RedisRateLimiter() *redisrate.Limiter {
	return h.redisLimiter
}

func (h *baseHandlerDeps) Validate() *validator.Validate {
	return h.validate
}

func (h *baseHandlerDeps) Translator() *ut.Translator {
	return h.translator
}

func (h *baseHandlerDeps) Repository() *repository.Repository {
	return h.repository
}

func (h *baseHandlerDeps) SlackClient() *slack.Client {
	return h.slackClient
}

func (h *baseHandlerDeps) EmailClient() *email.Client {
	return h.emailClient
}

func (h *baseHandlerDeps) DecodeAndValidateInputQuery(dst interface{}, src map[string][]string) *aserr.ASError {
	_ = schema.NewDecoder().Decode(dst, src)
	err := h.validate.Struct(dst)
	if err != nil {
		validationErrors, ok := err.(validator.ValidationErrors)
		if ok {
			return aserr.ErrBadRequest.NewWithErrorMap(err, validationErrors.Translate(*h.translator))
		}

		return aserr.ErrBadRequest.NewWithError(err)
	}

	return nil
}

func (h *baseHandlerDeps) DecodeAndValidateInputBody(dst interface{}, src io.ReadCloser) *aserr.ASError {
	json.NewDecoder(src).Decode(dst)
	err := h.validate.Struct(dst)
	if err != nil {
		validationErrors, ok := err.(validator.ValidationErrors)
		if ok {
			return aserr.ErrBadRequest.NewWithErrorMap(err, validationErrors.Translate(*h.translator))
		}

		return aserr.ErrBadRequest.NewWithError(err)
	}

	return nil
}

func (h *baseHandlerDeps) WithSessionCtx(opts ...session.SessionOptionsFunc) func(next http.Handler) http.Handler {
	return session.WithSessionCtx(h.Config(), h.Repository(), h.Redis(), opts...)
}

func (h *baseHandlerDeps) RealIP(r *http.Request) (string, error) {
	ip, err := netutils.RealIP(r)
	if err != nil {
		ip, err = netutils.RemoteIP(r)
		if err != nil {
			return "", err
		}
	}

	return ip, nil
}

type BaseHandlerDepsOption func(*baseHandlerDeps)

func WithConfig(cfg *config.Config) BaseHandlerDepsOption {
	return func(h *baseHandlerDeps) { h.cfg = cfg }
}

func WithLogger(logger *log.Logger) BaseHandlerDepsOption {
	return func(h *baseHandlerDeps) { h.log = logger }
}

func WithPg(pg *pg.Pg) BaseHandlerDepsOption {
	return func(h *baseHandlerDeps) { h.pg = pg }
}

func WithRedis(redis *redis.Client) BaseHandlerDepsOption {
	return func(h *baseHandlerDeps) { h.redis = redis }
}

func WithRedisRateLimiter(redisLimiter *redisrate.Limiter) BaseHandlerDepsOption {
	return func(h *baseHandlerDeps) { h.redisLimiter = redisLimiter }
}

func WithValidate(validate *validator.Validate) BaseHandlerDepsOption {
	return func(h *baseHandlerDeps) { h.validate = validate }
}

func WithTranslator(translator *ut.Translator) BaseHandlerDepsOption {
	return func(h *baseHandlerDeps) { h.translator = translator }
}

func WithSlackClient(slackClient *slack.Client) BaseHandlerDepsOption {
	return func(h *baseHandlerDeps) { h.slackClient = slackClient }
}

func WithEmailClient(emailClient *email.Client) BaseHandlerDepsOption {
	return func(h *baseHandlerDeps) { h.emailClient = emailClient }
}

func WithRepository(repository *repository.Repository) BaseHandlerDepsOption {
	return func(h *baseHandlerDeps) { h.repository = repository }
}

func NewBaseHandlerDeps(opts ...BaseHandlerDepsOption) BaseHandlerDeps {
	h := &baseHandlerDeps{}
	for _, opt := range opts {
		opt(h)
	}
	return h
}
