package authservice

import (
	"api/internal/http/aserr"
	"api/internal/pg"
	"api/internal/redis"
	"api/internal/utils/authutils"
	"api/repository"
	"api/service"
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
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

type SignUpTokenInput struct {
	Token    string
	Password string
	Timezone string
}

type SignUpTokenOutput struct {
	User    *repository.User
	Account *repository.Account
	Session *repository.Session
}

func (s *Service) SignUpToken(ctx context.Context, q pg.Querier, redisClient *redis.Client, input *SignUpTokenInput) (*SignUpTokenOutput, error) {
	authCode, err := s.Repository().GetAuthCode(
		ctx, q,
		repository.WithGetAuthCodeInputType(repository.AuthCodeTypeCreateAccount),
		repository.WithGetAuthCodeInputValueHash(input.Token),
	)
	if err != nil {
		return nil, err
	}
	userID, err := s.VerifyAuthCode(ctx, q, authCode)
	if err != nil {
		return nil, err
	}

	user, err := s.Repository().GetUser(ctx, q, &repository.GetUserBody{ID: &userID})
	if err != nil {
		return nil, err
	}

	userTimezone, err := s.Repository().CreateUserTimezone(ctx, q, &repository.CreateUserTimezoneInput{
		UserID:    user.ID,
		Timezone:  input.Timezone,
		ValidFrom: pgtype.Timestamp{Time: time.Now(), Valid: true},
	})
	if err != nil {
		return nil, err
	}

	passwordHash, err := authutils.CreateHash(input.Password, authutils.DefaultParams)
	if err != nil {
		return nil, err
	}

	accountId := uuid.New()
	account, err := s.Repository().CreateAccount(
		ctx,
		q,
		&repository.CreateAccountInput{
			ID:           accountId,
			UserID:       user.ID,
			AccountID:    accountId.String(),
			ProviderID:   sql.NullString{String: "email", Valid: true},
			PasswordHash: sql.NullString{String: passwordHash, Valid: true},
		},
	)
	if err != nil {
		return nil, err
	}

	session, err := s.Repository().CreateSession(ctx, redisClient, repository.SessionWithUser(user), repository.SessionWithUserTimezone(userTimezone))
	if err != nil {
		return nil, err
	}

	return &SignUpTokenOutput{User: user, Account: account, Session: session}, nil
}

func (s *Service) VerifyAuthCode(ctx context.Context, q pg.Querier, authCode *repository.AuthCode) (uuid.UUID, error) {
	now := time.Now()
	if now.After(authCode.ExpiresAt.Time) {
		return uuid.Nil, aserr.ErrVerificationTokenExpired
	}
	if !authutils.VerifyCodeHash(s.Config().SecretAuthCodeValueHash, authCode.Value, authCode.ValueHash) {
		return uuid.Nil, aserr.ErrVerificationTokenInvalid
	}
	if _, err := s.Repository().DeleteAuthCode(ctx, q, authCode.ID); err != nil {
		return uuid.Nil, err
	}

	return authCode.UserID, nil
}
