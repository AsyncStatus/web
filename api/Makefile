proxy-prod-db:
	fly proxy 15432:5432 -a as-pg

proxy-dev-db:
	fly proxy 10864:5432 -a dev-as-pg

generate-emails:
	go run ./cmd/email/generate.go

build-swagger:
	swag fmt && swag init --pd

migrate-up:
	go run ./cmd/migrate/main.go postgres dir up

migrate-reset:
	go run ./cmd/migrate/main.go postgres dir reset

migrate-down:
	go run ./cmd/migrate/main.go postgres dir down

seed-up:
	go run ./cmd/seed/main.go postgres dir up

seed-down:
	go run ./cmd/seed/main.go postgres dir down

seed-reset:
	go run ./cmd/seed/main.go postgres dir reset

set-prod-secrets-from-env:
	cat .env.production | fly secrets import -a as-api

set-dev-secrets-from-env:
	cat .env.development | fly secrets import -a dev-as-api
