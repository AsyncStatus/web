package userservice

import (
	"api/internal/http/aserr"
	"api/internal/pg"
	"api/internal/utils/authutils"
	"api/repository"
	"api/service"
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
)

type Service struct{ service.BaseServiceDeps }

func NewService(baseServiceDeps service.BaseServiceDeps) *Service {
	return &Service{BaseServiceDeps: baseServiceDeps}
}

type CreateUserInput struct {
	Email    string
	Password string
	Timezone string
}

type CreateUserResult struct {
	User         *repository.User
	Account      *repository.Account
	UserTimezone *repository.UserTimezone
}

func (s *Service) CreateUser(ctx context.Context, q pg.Querier, input *CreateUserInput) (*CreateUserResult, error) {
	name := repository.EmailToName(input.Email)
	user, err := s.Repository().CreateUser(ctx, q, &repository.CreateUserBody{Name: name, Email: input.Email})
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation && pgErr.ConstraintName == "users_email_key" {
			return nil, aserr.ErrUserAlreadyExists
		}

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

	return &CreateUserResult{user, account, userTimezone}, nil
}
