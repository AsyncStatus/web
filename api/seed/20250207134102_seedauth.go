package seed

import (
	"api/internal/utils/authutils"
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/pressly/goose/v3"
)

func init() {
	goose.AddMigrationContext(upSeedauth, downSeedauth)
}

func upSeedauth(ctx context.Context, tx *sql.Tx) error {
	rows, err := tx.QueryContext(ctx, `
		SELECT id, email 
		FROM "user".users 
		LIMIT $1
	`, usersCount+1) // add local user
	if err != nil {
		return fmt.Errorf("selecting random users: %w", err)
	}
	defer rows.Close()

	passwordHash, err := authutils.CreateHash("password", authutils.DefaultParams)
	if err != nil {
		return err
	}

	type userData struct {
		id    uuid.UUID
		email string
	}
	var users []userData
	for rows.Next() {
		var u userData
		if err := rows.Scan(&u.id, &u.email); err != nil {
			return fmt.Errorf("scanning user row: %w", err)
		}
		users = append(users, u)
	}
	if err := rows.Err(); err != nil {
		return err
	}

	if len(users) == 0 {
		return nil
	}

	valueStrings := make([]string, 0, len(users))
	valueArgs := make([]interface{}, 0, len(users)*4)
	for i, u := range users {
		valueStrings = append(valueStrings, fmt.Sprintf("($%d, $%d, $%d, $%d)", i*4+1, i*4+2, i*4+3, i*4+4))
		valueArgs = append(valueArgs, u.id, u.id, "email", passwordHash)
	}

	insertQuery := fmt.Sprintf(`
		INSERT INTO "auth".accounts (user_id, account_id, provider_id, password_hash)
		VALUES %s
	`, strings.Join(valueStrings, ", "))

	if _, err := tx.ExecContext(ctx, insertQuery, valueArgs...); err != nil {
		return fmt.Errorf("inserting auth accounts: %w", err)
	}

	return nil
}

func downSeedauth(ctx context.Context, tx *sql.Tx) error {
	_, err := tx.ExecContext(ctx, `
	DELETE FROM "auth".accounts WHERE provider_id = 'email';
`)
	return err
}
