package migrations

import (
	"context"
	"database/sql"

	"github.com/pressly/goose/v3"
)

func init() {
	goose.AddMigrationContext(upAddauth, downAddauth)
}

func upAddauth(ctx context.Context, tx *sql.Tx) error {
	_, err := tx.ExecContext(ctx, `
create schema if not exists auth;

create table if not exists auth.accounts (
	id uuid not null primary key default uuid_generate_v4 (),
	user_id uuid not null references "user".users,
	account_id varchar(255) not null, -- either provider user id or this resource id
	provider_id varchar(255) not null, -- provider (email, github, google)
	access_token varchar(255), -- provider
	refresh_token varchar(255), -- provider
	scope varchar(255), -- provider
	password_hash varchar(98),
	created_at timestamp not null default current_timestamp,
	access_token_expires_at timestamp,
	refresh_token_expires_at timestamp,
	updated_at timestamp,
	archived_at timestamp
);

create table if not exists auth.codes (
	id uuid not null primary key default uuid_generate_v4 (),
	user_id uuid references "user".users,
	type varchar(255) not null,
	value varchar(64) not null,
	value_hash varchar(255) not null,
	created_at timestamp not null default current_timestamp,
	expires_at timestamp not null default (current_timestamp + interval '1 days'),
	unique(user_id, type, value)
);

create index idx_auth_codes_value_hash on auth.codes(value_hash);
create index idx_auth_codes_user_type_hash on auth.codes(user_id, type, value_hash);
`)
	return err
}

func downAddauth(ctx context.Context, tx *sql.Tx) error {
	_, err := tx.ExecContext(ctx, `
drop index if exists idx_auth_codes_user_type_hash;
drop index if exists idx_auth_codes_value_hash;
drop table if exists auth.codes;
drop table if exists auth.accounts;
drop schema if exists auth;
`)
	return err
}
