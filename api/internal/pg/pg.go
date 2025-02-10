package pg

import (
	"api/config"
	"api/config/env"
	"api/internal/log"
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Pg struct{ *pgxpool.Pool }

func MustNewPg(ctx context.Context, cfg *config.Config, logger *log.Logger) *Pg {
	connCfg, err := pgxpool.ParseConfig(cfg.PgDBURL)
	if err != nil {
		panic(err)
	}

	if cfg.AppEnv == env.Local {
		connCfg.ConnConfig.Tracer = &Tracer{logger}
	}

	pool, err := pgxpool.NewWithConfig(ctx, connCfg)
	if err != nil {
		panic(err)
	}

	err = pool.Ping(ctx)
	if err != nil {
		panic(err)
	}

	return &Pg{pool}
}
