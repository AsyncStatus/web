package main

import (
	"api/config"
	"api/config/env"
	"api/handler"
	authemail "api/handler/auth/email"
	authtoken "api/handler/auth/token"
	"api/handler/health"
	"api/handler/project"
	projectteammembership "api/handler/project/membership"
	projectteam "api/handler/project/team"
	"api/handler/session"
	internalstripe "api/handler/stripe"
	"api/handler/user"
	"api/internal/email"
	"api/internal/log"
	"api/internal/middleware"
	"api/internal/pg"
	"api/internal/redis"
	slack "api/internal/slack/client"
	"api/repository"
	"api/service"
	authservice "api/service/auth"
	"context"
	"database/sql/driver"
	"fmt"
	"net/http"
	"reflect"
	"time"

	"github.com/go-chi/chi/v5"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/go-playground/locales/en"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	redisrate "github.com/go-redis/redis_rate/v10"
	"github.com/guregu/null/v5"
	"github.com/jackc/pgx/v5/pgtype"
	_ "github.com/swaggo/http-swagger/example/go-chi/docs"
	"go.uber.org/zap"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
)

// @title		AsyncStatus API
// @version	0.1
// @host		api.asyncstatus.com
func main() {
	time.Local = time.UTC
	cfg := config.MustNewConfig()
	log := log.MustNewLogger(cfg)

	ctx := context.Background()
	pg := pg.MustNewPg(ctx, cfg, log)
	defer pg.Close()

	redisClient := redis.MustNewClient(ctx, cfg)
	defer redisClient.Close()
	redisRateLimiter := redisrate.NewLimiter(redisClient)

	emailClient := email.NewClient(cfg, log)
	slackClient := slack.NewClient(cfg)
	repo := repository.NewRepository(repository.WithConfig(cfg))
	validate := validator.New(validator.WithRequiredStructEnabled())
	validate.RegisterCustomTypeFunc(ValidateValuer, null.String{}, null.Time{}, null.Int64{}, pgtype.Timestamp{}, pgtype.Timestamptz{})
	en := en.New()
	uni := ut.New(en, en)
	translator, _ := uni.GetTranslator("en")

	baseServiceDeps := service.NewBaseServiceDeps(
		service.WithConfig(cfg),
		service.WithLogger(log),
		service.WithPg(pg),
		service.WithRedis(redisClient),
		service.WithSlackClient(slackClient),
		service.WithEmailClient(emailClient),
		service.WithRepository(repo),
	)

	authService := authservice.NewService(baseServiceDeps)

	baseHandlerDeps := handler.NewBaseHandlerDeps(
		handler.WithConfig(cfg),
		handler.WithLogger(log),
		handler.WithPg(pg),
		handler.WithRedis(redisClient),
		handler.WithRedisRateLimiter(redisRateLimiter),
		handler.WithValidate(validate),
		handler.WithSlackClient(slackClient),
		handler.WithEmailClient(emailClient),
		handler.WithTranslator(&translator),
		handler.WithRepository(repo),
	)
	authemailHandler := authemail.NewHandler(
		authemail.WithBaseHandlerDeps(baseHandlerDeps),
		authemail.WithAuthService(authService),
	)
	authtokenHandler := authtoken.NewHandler(
		authtoken.WithBaseHandlerDeps(baseHandlerDeps),
		authtoken.WithAuthService(authService),
	)
	healthHandler := health.NewHandler(baseHandlerDeps)
	userHandler := user.NewHandler(baseHandlerDeps)
	sessionHandler := session.NewHandler(baseHandlerDeps)
	projectHandler := project.NewHandler(baseHandlerDeps)
	projectTeamHandler := projectteam.NewHandler(baseHandlerDeps)
	projectTeamMembershipHandler := projectteammembership.NewHandler(baseHandlerDeps)
	stripeHandler := internalstripe.NewHandler(baseHandlerDeps)

	r := chi.NewRouter()

	r.Use(middleware.Recoverer)
	r.Use(chiMiddleware.RequestID)
	r.Use(chiMiddleware.RealIP)
	r.Use(middleware.NewHTTPLogger(log))
	r.Use(middleware.Cors(cfg))
	if cfg.AppEnv == env.Local {
		r.Use(middleware.Slowdown(time.Millisecond*150, time.Millisecond*600))
	}
	if cfg.AppEnv != env.Local {
		r.Use(chiMiddleware.Timeout(10 * time.Second))
	}

	r.Mount("/health", healthHandler.Router())
	r.Mount("/stripe", stripeHandler.Router())
	r.Route("/users", func(r chi.Router) {
		r.Mount("/", userHandler.Router())
	})
	r.Route("/auth", func(r chi.Router) {
		r.Mount("/email", authemailHandler.Router())
		r.Mount("/token", authtokenHandler.Router())
		r.Mount("/session", sessionHandler.Router())
	})
	r.Route("/projects", func(r chi.Router) {
		r.Mount("/", projectHandler.Router())
		r.Mount("/{projectSlug}/teams", projectTeamHandler.Router())
		r.Mount("/{projectSlug}/teams/{teamSlug}/memberships", projectTeamMembershipHandler.Router())
	})

	addr := fmt.Sprintf("[::]:%s", cfg.Port)
	if cfg.AppEnv == env.Local {
		addr = fmt.Sprintf("localhost:%s", cfg.Port)
	}
	srv2 := &http2.Server{}
	srv := &http.Server{
		Addr:              addr,
		Handler:           h2c.NewHandler(r, srv2),
		ReadHeaderTimeout: 10 * time.Second,
		IdleTimeout:       10 * time.Minute,
	}
	log.Info("server started", zap.String("env", string(cfg.AppEnv)), zap.String("addr", addr))
	if err := srv.ListenAndServe(); err != nil {
		panic(err)
	}
}

func ValidateValuer(field reflect.Value) interface{} {
	if valuer, ok := field.Interface().(driver.Valuer); ok {
		val, err := valuer.Value()
		if err == nil {
			return val
		}
	}

	return nil
}
