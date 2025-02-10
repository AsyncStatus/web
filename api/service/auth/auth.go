package authservice

import (
	"api/internal/http/aserr"
	"api/internal/pg"
	"api/internal/redis"
	"api/internal/utils/authutils"
	"api/repository"
	"api/service"
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
)

type Service struct{ service.BaseServiceDeps }

func NewService(baseServiceDeps service.BaseServiceDeps) *Service {
	return &Service{BaseServiceDeps: baseServiceDeps}
}

type AuthenticateInput struct {
	Email    string
	Password string
}

type AuthenticateOutput struct {
	User    *repository.User
	Account *repository.Account
	Session *repository.Session
}

func (s *Service) Authenticate(ctx context.Context, q pg.Querier, redisClient *redis.Client, body *AuthenticateInput) (*AuthenticateOutput, error) {
	user, err := s.Repository().GetUser(ctx, q, &repository.GetUserBody{Email: &body.Email})
	if err != nil && errors.Is(err, pgx.ErrNoRows) {
		return nil, aserr.ErrInvalidEmailOrPassword
	}
	if err != nil {
		return nil, err
	}

	account, err := s.Repository().GetAccount(ctx, q, &repository.GetAccountInput{UserID: &user.ID})
	if err != nil {
		return nil, err
	}

	passwordHash, err := account.PasswordHash.Value()
	if err != nil {
		return nil, aserr.ErrInvalidEmailOrPassword
	}
	passwordHashStr, ok := passwordHash.(string)
	if !ok {
		return nil, aserr.ErrInvalidEmailOrPassword
	}

	matches, err := authutils.ComparePasswordAndHash(body.Password, passwordHashStr)
	if err != nil {
		return nil, err
	}

	if !matches {
		return nil, aserr.ErrInvalidEmailOrPassword
	}

	if !user.ApprovedAt.Valid {
		return nil, aserr.ErrUserNotFoundSignIn
	}

	userTimezone, err := s.Repository().GetUserTimezoneByUserID(ctx, q, user.ID)
	if err != nil {
		return nil, err
	}

	session, err := s.Repository().CreateSession(ctx, redisClient, repository.SessionWithUser(user), repository.SessionWithUserTimezone(userTimezone))
	if err != nil {
		return nil, err
	}

	return &AuthenticateOutput{User: user, Account: account, Session: session}, nil
}

type SignUpEmailInput struct {
	Email    string
	Password string
	Timezone string
}

type SignUpEmailOutput struct {
	User    *repository.User
	Account *repository.Account
	Session *repository.Session
}
