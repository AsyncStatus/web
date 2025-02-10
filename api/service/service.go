package service

import (
	"api/config"
	"api/internal/email"
	"api/internal/log"
	"api/internal/pg"
	"api/internal/redis"
	slack "api/internal/slack/client"
	"api/repository"
)

type BaseServiceDeps interface {
	Config() *config.Config
	Logger() *log.Logger
	Pg() *pg.Pg
	Redis() *redis.Client
	SlackClient() *slack.Client
	EmailClient() *email.Client
	Repository() *repository.Repository
}

type baseServiceDeps struct {
	cfg         *config.Config
	log         *log.Logger
	pg          *pg.Pg
	redis       *redis.Client
	slackClient *slack.Client
	emailClient *email.Client
	repository  *repository.Repository
}

func (h *baseServiceDeps) Config() *config.Config {
	return h.cfg
}

func (h *baseServiceDeps) Logger() *log.Logger {
	return h.log
}

func (h *baseServiceDeps) Pg() *pg.Pg {
	return h.pg
}

func (h *baseServiceDeps) Redis() *redis.Client {
	return h.redis
}

func (h *baseServiceDeps) Repository() *repository.Repository {
	return h.repository
}

func (h *baseServiceDeps) SlackClient() *slack.Client {
	return h.slackClient
}

func (h *baseServiceDeps) EmailClient() *email.Client {
	return h.emailClient
}

type BaseServiceDepsOption func(*baseServiceDeps)

func WithConfig(cfg *config.Config) BaseServiceDepsOption {
	return func(h *baseServiceDeps) { h.cfg = cfg }
}

func WithLogger(logger *log.Logger) BaseServiceDepsOption {
	return func(h *baseServiceDeps) { h.log = logger }
}

func WithPg(pg *pg.Pg) BaseServiceDepsOption {
	return func(h *baseServiceDeps) { h.pg = pg }
}

func WithRedis(redis *redis.Client) BaseServiceDepsOption {
	return func(h *baseServiceDeps) { h.redis = redis }
}

func WithSlackClient(slackClient *slack.Client) BaseServiceDepsOption {
	return func(h *baseServiceDeps) { h.slackClient = slackClient }
}

func WithEmailClient(emailClient *email.Client) BaseServiceDepsOption {
	return func(h *baseServiceDeps) { h.emailClient = emailClient }
}

func WithRepository(repository *repository.Repository) BaseServiceDepsOption {
	return func(h *baseServiceDeps) { h.repository = repository }
}

func NewBaseServiceDeps(opts ...BaseServiceDepsOption) BaseServiceDeps {
	h := &baseServiceDeps{}
	for _, opt := range opts {
		opt(h)
	}
	return h
}
