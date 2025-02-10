package migrations

import (
	"context"
	"database/sql"

	"github.com/pressly/goose/v3"
)

func init() {
	goose.AddMigrationContext(upAddproject, downAddproject)
}

func upAddproject(ctx context.Context, tx *sql.Tx) error {
	_, err := tx.ExecContext(ctx, `
create schema if not exists project;

create table if not exists project.projects (
	id uuid not null primary key default uuid_generate_v4 (),
	name varchar(255) not null,
	slug varchar(255) not null unique,
	avatar_url text,
	plan varchar(255) not null default 'free',
	created_at timestamp not null default current_timestamp,
	updated_at timestamp,
	archived_at timestamp
);

create table if not exists project.memberships (
	id uuid not null primary key default uuid_generate_v4 (),
	user_id uuid not null references "user".users on delete cascade,
	project_id uuid not null references project.projects on delete cascade,
	role varchar(50) not null default 'member',
	created_at timestamp not null default current_timestamp,
	updated_at timestamp,
	unique (user_id, project_id)
);

create table if not exists project.teams (
	id uuid not null primary key default uuid_generate_v4 (),
	project_id uuid not null references project.projects on delete cascade,
	name varchar(255) not null,
	emoji varchar(255),
	slug varchar(255) not null,
	parent_team_id uuid references project.teams on delete set null,
	made_private_at timestamp,
	created_at timestamp not null default current_timestamp,
	updated_at timestamp,
	archived_at timestamp,
	unique (project_id, slug)
);

create table if not exists project.invitations (
	id uuid not null primary key default uuid_generate_v4 (),
	project_id uuid not null references project.projects on delete cascade,
	inviter_id uuid not null references "user".users on delete cascade,
	auth_code_id uuid not null references auth.codes on delete cascade,
	team_id uuid references project.teams on delete cascade,
	email varchar(255) not null,
	status varchar(64) not null default 'pending',
	role varchar(64) not null,
	created_at timestamp not null default current_timestamp,
	updated_at timestamp,
	unique (project_id, auth_code_id)
);

create table if not exists project.team_memberships (
	id uuid not null primary key default uuid_generate_v4 (),
	user_id uuid not null references "user".users on delete cascade,
	team_id uuid not null references project.teams on delete cascade,
	position varchar(255),
	created_at timestamp not null default current_timestamp,
	archived_at timestamp,
	unique (user_id, team_id)
);
`)
	return err
}

func downAddproject(ctx context.Context, tx *sql.Tx) error {
	_, err := tx.ExecContext(ctx, `
drop table if exists project.team_memberships;
drop table if exists project.invitations;
drop table if exists project.teams;
drop table if exists project.memberships;
drop table if exists project.projects;
drop schema if exists project;
`)
	return err
}
