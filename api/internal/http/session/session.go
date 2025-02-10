package session

import (
	"api/config"
	"api/internal/http/aserr"
	"api/internal/redis"
	"api/internal/utils/authutils"
	"api/repository"
	"context"
	"encoding/json"
	"net/http"
)

type key int

const (
	sessionCtxKey key = iota
)

type sessionOptions struct {
	refresh             bool
	authorizeUnverified bool
}

type SessionOptionsFunc func(*sessionOptions)

func WithRefresh() SessionOptionsFunc {
	return func(opts *sessionOptions) {
		opts.refresh = true
	}
}

func WithAuthorizeUnverified() SessionOptionsFunc {
	return func(opts *sessionOptions) {
		opts.authorizeUnverified = true
	}
}

func WithSessionCtx(cfg *config.Config, repo *repository.Repository, redisClient *redis.Client, opts ...SessionOptionsFunc) func(next http.Handler) http.Handler {
	var options sessionOptions
	for _, opt := range opts {
		opt(&options)
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			println(r.Cookies())
			unsafeNextAuthCookie, err := r.Cookie(authutils.SessionCookieName)
			if err != nil {
				http.SetCookie(w, authutils.ClearSessionCookie(cfg))
				w.Header().Set("Content-Type", "application/json; charset=utf-8")
				w.WriteHeader(aserr.ErrUnauthorized.HTTPStatusCode)
				json.NewEncoder(w).Encode(aserr.ErrUnauthorized)
				return
			}

			if unsafeNextAuthCookie.Value == "" {
				http.SetCookie(w, authutils.ClearSessionCookie(cfg))
				w.Header().Set("Content-Type", "application/json; charset=utf-8")
				w.WriteHeader(aserr.ErrUnauthorized.HTTPStatusCode)
				json.NewEncoder(w).Encode(aserr.ErrUnauthorized)
				return
			}

			authCookieValue, err := authutils.GetSessionCookieValue(cfg.SecretSession, unsafeNextAuthCookie.Value)
			if err != nil {
				http.SetCookie(w, authutils.ClearSessionCookie(cfg))
				w.Header().Set("Content-Type", "application/json; charset=utf-8")
				w.WriteHeader(aserr.ErrUnauthorized.HTTPStatusCode)
				json.NewEncoder(w).Encode(aserr.ErrUnauthorized)
				return
			}

			session, sessionRefreshed, err := repo.GetSessionByKey(ctx, redisClient, repository.GetSessionKey(authCookieValue.UserID, authCookieValue.ID), options.refresh)
			if err != nil {
				http.SetCookie(w, authutils.ClearSessionCookie(cfg))
				w.Header().Set("Content-Type", "application/json; charset=utf-8")
				w.WriteHeader(aserr.ErrUnauthorized.HTTPStatusCode)
				json.NewEncoder(w).Encode(aserr.ErrUnauthorized)
				return
			}

			if sessionRefreshed {
				http.SetCookie(w, authutils.NewSessionCookie(cfg, session, session.CreatedAt))
			}

			if !options.authorizeUnverified && !session.UserVerifiedAt.Valid {
				w.Header().Set("Content-Type", "application/json; charset=utf-8")
				w.WriteHeader(aserr.ErrUnauthorizedAccountNotVerified.HTTPStatusCode)
				json.NewEncoder(w).Encode(aserr.ErrUnauthorizedAccountNotVerified)
				return
			}

			next.ServeHTTP(w, r.WithContext(context.WithValue(ctx, sessionCtxKey, session)))
		})
	}
}

func GetSessionFromRequest(r *http.Request) *repository.Session {
	session, ok := r.Context().Value(sessionCtxKey).(*repository.Session)
	if !ok {
		panic(aserr.ErrUnauthorized)
	}

	return session
}

func MaybeGetSessionFromRequest(r *http.Request) (*repository.Session, error) {
	session, ok := r.Context().Value(sessionCtxKey).(*repository.Session)
	if !ok {
		return nil, aserr.ErrUnauthorized
	}

	return session, nil
}
