package migrations

import (
	"context"
	"database/sql"

	"github.com/pressly/goose/v3"
)

func init() {
	goose.AddMigrationContext(upBaseconfig, downBaseconfig)
}

func upBaseconfig(ctx context.Context, tx *sql.Tx) error {
	_, err := tx.ExecContext(ctx, `
set time zone 'UTC';
create extension if not exists "uuid-ossp";
create extension if not exists timescaledb cascade;
`)
	return err
}

func downBaseconfig(ctx context.Context, tx *sql.Tx) error {
	_, err := tx.ExecContext(ctx, `
drop extension if exists "uuid-ossp";
reset time zone;
`)
	return err
}
