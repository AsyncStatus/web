package limiter

import (
	"api/internal/http/aserr"
	"api/internal/http/asres"
	"api/internal/http/session"
	"api/internal/utils/netutils"
	"api/internal/utils/timeutils"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	redisrate "github.com/go-redis/redis_rate/v10"
)

type RateLimitOptions struct {
	IP            bool
	SessionUserID bool
	Endpoint      bool
}
type RateLimitOpts func(opts *RateLimitOptions)

func WithIP() RateLimitOpts {
	return func(opts *RateLimitOptions) { opts.IP = true }
}
func WithSessionUserID() RateLimitOpts {
	return func(opts *RateLimitOptions) { opts.SessionUserID = true }
}
func WithURLPath() RateLimitOpts {
	return func(opts *RateLimitOptions) { opts.Endpoint = true }
}

func RateLimit[T any](handler asres.HandlerFunc[T], rl *redisrate.Limiter, limit redisrate.Limit, opts ...RateLimitOpts) asres.HandlerFunc[T] {
	return func(w http.ResponseWriter, r *http.Request) (*asres.Response[T], error) {
		o := &RateLimitOptions{}
		for _, opt := range opts {
			opt(o)
		}

		key := ""
		if o.IP {
			ip, realIPErr := netutils.RealIP(r)
			if realIPErr != nil {
				ip, remoteIPErr := netutils.RemoteIP(r)
				if remoteIPErr != nil {
					return nil, errors.Join(realIPErr, remoteIPErr)
				}
				key += fmt.Sprintf("%s:", ip)
			} else {
				key += fmt.Sprintf("%s:", ip)
			}
		}

		ctx := r.Context()
		if o.SessionUserID {
			session := session.GetSessionFromRequest(r)
			key += fmt.Sprintf("%s:", session.UserID)
		}

		if o.Endpoint {
			key += fmt.Sprintf("%s:", r.URL.Path)
		}

		results, _ := rl.Allow(ctx, key, limit)
		if results.Allowed <= 0 {
			w.Header().Set("X-RateLimit-Limit", results.Limit.String())
			w.Header().Set("X-RateLimit-Remaining", strconv.Itoa(results.Remaining))
			w.Header().Set("X-RateLimit-Reset", results.ResetAfter.String())
			w.Header().Set("Retry-After", results.RetryAfter.String())
			return nil, aserr.ErrTooManyRequests.NewWithError(errors.New(aserr.ErrTooManyRequests.Error() + ", try again in " + timeutils.RoundToSingleUnit(results.RetryAfter)))
		}

		return handler(w, r)
	}
}
