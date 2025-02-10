package middleware

import (
	"api/config"
	"api/config/env"
	"api/internal/http/aserr"
	"api/internal/log"
	"encoding/json"
	"net/http"
	"runtime/debug"
	"time"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"go.uber.org/zap"
	"golang.org/x/exp/rand"
)

func NewHTTPLogger(log *log.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
			t1 := time.Now()

			defer func() {
				reqLogger := log.With(
					zap.String("proto", r.Proto),
					zap.String("method", r.Method),
					zap.String("path", r.URL.Path),
					zap.String("reqId", middleware.GetReqID(r.Context())),
					zap.Duration("lat", time.Since(t1)),
					zap.Int("status", ww.Status()),
					zap.Int("size", ww.BytesWritten()),
					zap.String("ref", r.Header.Get("Referer")),
					zap.String("ua", r.Header.Get("User-Agent")),
				).WithOptions(zap.AddStacktrace(zap.DPanicLevel))
				if ww.Status() <= 599 && ww.Status() >= 400 {
					reqLogger.Error("req")
				} else {
					reqLogger.Info("req")
				}
			}()
			next.ServeHTTP(ww, r)
		}

		return http.HandlerFunc(fn)
	}
}

func Recoverer(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rvr := recover(); rvr != nil {
				if rvr == http.ErrAbortHandler {
					// we don't recover http.ErrAbortHandler so the response
					// to the client is aborted, this should not be logged
					panic(rvr)
				}

				asError, ok := rvr.(*aserr.ASError)
				if ok {
					w.Header().Set("Content-Type", "application/json; charset=utf-8")
					w.WriteHeader(asError.HTTPStatusCode)
					json.NewEncoder(w).Encode(asError)
					return
				}

				logEntry := middleware.GetLogEntry(r)
				if logEntry != nil {
					logEntry.Panic(rvr, debug.Stack())
				} else {
					middleware.PrintPrettyStack(rvr)
				}

				if r.Header.Get("Connection") != "Upgrade" {
					w.WriteHeader(aserr.ErrInternalServerError.HTTPStatusCode)
				}
			}
		}()

		next.ServeHTTP(w, r)
	}

	return http.HandlerFunc(fn)
}

func Slowdown(minDelay time.Duration, maxDelay time.Duration) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			delay := time.Duration(rand.Intn(int(maxDelay-minDelay))) + minDelay //nolint:gosec //it's fine to use rand here
			time.Sleep(delay)
			next.ServeHTTP(w, r)
		})
	}
}

func Cors(cfg *config.Config) func(next http.Handler) http.Handler {
	allowedOrigins := []string{"https://app.asyncstatus.com", "https://dev.app.asyncstatus.com"}
	return cors.Handler(cors.Options{
		AllowOriginFunc: func(r *http.Request, origin string) bool {
			if r.URL.Path == "/health" {
				return true
			}

			if cfg.AppEnv == env.Local {
				return true
			}

			for _, value := range allowedOrigins {
				if value == origin {
					return true
				}
			}

			return false
		},
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"SID", "Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           60,
	})
}
