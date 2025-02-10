package migrations

import (
	"context"
	"database/sql"

	"github.com/pressly/goose/v3"
)

func init() {
	goose.AddMigrationContext(upAddusers, downAddusers)
}

func upAddusers(ctx context.Context, tx *sql.Tx) error {
	_, err := tx.ExecContext(ctx, `
create schema if not exists "user";

create table if not exists "user".users (
	id uuid not null primary key default uuid_generate_v4 (),
	name varchar(255) not null,
	email varchar(255) not null unique,
	avatar_url text,
	referrer varchar(255),
	created_at timestamptz not null default current_timestamp,
	verified_at timestamptz,
	updated_at timestamptz,
	archived_at timestamptz,
	approved_at timestamptz
);

create table if not exists "user".timezones (
	id uuid not null primary key default uuid_generate_v4 (),
	user_id uuid not null references "user".users (id) on delete cascade,
	timezone varchar(255) not null, -- timezone IANA format (e.g., 'America/New_York')
	valid_from timestamp not null default current_timestamp, -- When this timezone became effective
	valid_until timestamp -- When this timezone was replaced (null if current)
);

-- Ensure only one current timezone per user
create unique index if not exists user_current_timezone_idx on "user".timezones (user_id)
where valid_until is null;
`)
	return err
}

func downAddusers(ctx context.Context, tx *sql.Tx) error {
	_, err := tx.ExecContext(ctx, `
drop table if exists "user".timezones;
drop table if exists "user".users;
drop schema if exists "user";
`)
	return err
}
