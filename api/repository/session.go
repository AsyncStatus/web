package repository

import (
	"api/internal/redis"
	"context"
	"errors"
	"fmt"
	"time"

	"encoding/json"

	goRedis "github.com/gomodule/redigo/redis"
	"github.com/google/uuid"
	"github.com/guregu/null/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

type Session struct {
	ID             uuid.UUID          `json:"id" validate:"required"`
	CreatedAt      time.Time          `json:"created_at" validate:"required" swaggertype:"string"`
	LastActiveAt   time.Time          `json:"last_active_at" validate:"required" swaggertype:"string"`
	RefreshedAt    time.Time          `json:"refreshed_at" validate:"required" swaggertype:"string"`
	UserID         uuid.UUID          `json:"user_id" validate:"required"`
	UserName       string             `json:"user_name" validate:"required"`
	UserEmail      string             `json:"user_email" validate:"required" format:"email"`
	UserAvatarURL  null.String        `json:"user_avatar_url" validate:"required" swaggertype:"string" extensions:"x-nullable"`
	UserTimezone   string             `json:"user_timezone" validate:"required"`
	UserCreatedAt  pgtype.Timestamptz `json:"user_created_at" validate:"required" swaggertype:"string"`
	UserVerifiedAt pgtype.Timestamptz `json:"user_verified_at" validate:"required" swaggertype:"string" extensions:"x-nullable"`
	UserUpdatedAt  pgtype.Timestamptz `json:"user_updated_at" validate:"required" swaggertype:"string" extensions:"x-nullable"`
	UserApprovedAt pgtype.Timestamptz `json:"user_approved_at" validate:"required" swaggertype:"string" extensions:"x-nullable"`
} //@name Session

// Returns session and whether it was refreshed
func (r *Repository) GetSessionByKey(ctx context.Context, redisClient *redis.Client, key string, refresh bool) (*Session, bool, error) {
	redisClient.Rh.SetContext(ctx)
	sessionRes, err := goRedis.Bytes(redisClient.Rh.JSONGet(key, "."))
	if err != nil {
		return nil, false, err
	}

	var session *Session
	err = json.Unmarshal(sessionRes, &session)
	fmt.Println(err)

	if session == nil {
		return nil, false, errors.New("session is nil")
	}

	if refresh && session.CreatedAt.Add(r.cfg.SessionTTL/2).Before(time.Now()) {
		if err := r.UpdateSessionExpiryByKey(ctx, redisClient, key); err != nil {
			return nil, false, err
		}

		session.ID = uuid.New()
		now := time.Now()
		session.RefreshedAt = now
		session.LastActiveAt = now

		sessionKey := GetSessionKey(session.UserID, session.ID)
		res, err := redisClient.Rh.JSONSet(sessionKey, ".", *session)
		if err != nil {
			return nil, false, err
		}
		if res.(string) != "OK" {
			return nil, false, errors.New("failed to set session")
		}

		if err := r.DeleteSessionByKey(ctx, redisClient, key); err != nil {
			return nil, false, err
		}

		return session, true, nil
	}

	return session, false, nil
}

type CreateSessionOptions struct {
	*User
	*UserTimezone
}
type CreateSessionOption func(*CreateSessionOptions)

func SessionWithUser(user *User) CreateSessionOption {
	return func(opts *CreateSessionOptions) {
		opts.User = user
	}
}
func SessionWithUserTimezone(userTimezone *UserTimezone) CreateSessionOption {
	return func(opts *CreateSessionOptions) {
		opts.UserTimezone = userTimezone
	}
}

func (r *Repository) CreateSession(ctx context.Context, redisClient *redis.Client, opts ...CreateSessionOption) (*Session, error) {
	createSessionOptions := &CreateSessionOptions{}

	for _, opt := range opts {
		opt(createSessionOptions)
	}

	now := time.Now()
	session := &Session{
		ID:             uuid.New(),
		CreatedAt:      now,
		RefreshedAt:    now,
		LastActiveAt:   now,
		UserID:         createSessionOptions.User.ID,
		UserName:       createSessionOptions.User.Name,
		UserEmail:      createSessionOptions.User.Email,
		UserAvatarURL:  createSessionOptions.User.AvatarURL,
		UserCreatedAt:  createSessionOptions.User.CreatedAt,
		UserVerifiedAt: createSessionOptions.User.VerifiedAt,
		UserUpdatedAt:  createSessionOptions.User.UpdatedAt,
		UserApprovedAt: createSessionOptions.User.ApprovedAt,
		UserTimezone:   createSessionOptions.UserTimezone.Timezone,
	}

	key := GetSessionKey(session.UserID, session.ID)
	res, err := redisClient.Rh.JSONSet(key, ".", session)
	if err != nil {
		return nil, err
	}
	if res.(string) != "OK" {
		return nil, errors.New("failed to set session")
	}

	_, err = redisClient.Expire(ctx, key, r.cfg.SessionTTL).Result()
	if err != nil {
		return nil, err
	}

	return session, nil
}

func (r *Repository) UpdateSessionExpiryByKey(ctx context.Context, redisClient *redis.Client, key string) error {
	_, err := redisClient.Expire(ctx, key, r.cfg.SessionTTL).Result()
	if err != nil {
		return err
	}

	return nil
}

func (r *Repository) UpdateSessionExpiryByPattern(ctx context.Context, redisClient *redis.Client, pattern string) error {
	keys, err := redisClient.Keys(ctx, pattern).Result()
	if err != nil {
		return err
	}

	if len(keys) == 0 {
		return nil
	}

	for _, key := range keys {
		_, err = redisClient.Expire(ctx, key, r.cfg.SessionTTL).Result()
		if err != nil {
			return err
		}
	}

	return nil
}

func (r *Repository) UpdateSessionByKey(ctx context.Context, redisClient *redis.Client, key string, path string, sessionSlice any) error {
	redisClient.Rh.SetContext(ctx)
	if _, err := redisClient.Rh.JSONSet(key, path, sessionSlice); err != nil {
		return err
	}

	return nil
}

func (r *Repository) UpdateSessionByPattern(ctx context.Context, redisClient *redis.Client, pattern string, path string, sessionSlice any) error {
	redisClient.Rh.SetContext(ctx)
	keys, err := redisClient.Keys(ctx, pattern).Result()
	if err != nil {
		return err
	}

	for _, key := range keys {
		if _, err = redisClient.Rh.JSONSet(key, path, sessionSlice); err != nil {
			return err
		}
	}

	return nil
}

func (r *Repository) DeleteSessionByKey(ctx context.Context, redisClient *redis.Client, key string) error {
	_, err := redisClient.Del(ctx, key).Result()
	if err != nil {
		return err
	}

	return nil
}

func (r *Repository) DeleteSessionByPattern(ctx context.Context, redisClient *redis.Client, keyPattern string) error {
	keys, err := redisClient.Keys(ctx, keyPattern).Result()
	if err != nil {
		return err
	}

	if len(keys) == 0 {
		return nil
	}

	_, err = redisClient.Del(ctx, keys...).Result()
	if err != nil {
		return err
	}

	return nil
}

func GetSessionKey(
	userID any,
	sessionID any,
) string {
	return fmt.Sprintf("sessions:%s:%s", getComponent(userID), getComponent(sessionID))
}

func getComponent(component any) string {
	switch v := component.(type) {
	case string:
		return v
	case uuid.UUID:
		return v.String()
	default:
		return ""
	}
}
