package seed

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/google/uuid"
	"github.com/pressly/goose/v3"
)

var usersCount int = 2000

func init() {
	goose.AddMigrationContext(upSeedusers, downSeedusers)
}

func upSeedusers(ctx context.Context, tx *sql.Tx) error {
	localUserID := uuid.New()
	localNow := time.Now()
	users := [][]any{{localUserID, "Local User", "local@asyncstatus.com", localNow}}
	usersTimezones := [][]any{{localUserID, "Europe/Warsaw", localNow}}

	for i := 0; i < usersCount; i++ {
		now := gofakeit.DateRange(time.Now().AddDate(0, -12, 0), time.Now())
		userID := uuid.New()
		users = append(users, []any{userID, gofakeit.Name(), gofakeit.Email(), now})
		usersTimezones = append(usersTimezones, []any{userID, gofakeit.TimeZoneRegion(), now})
	}

	valueStrings := make([]string, 0, len(users))
	valueArgs := make([]interface{}, 0, len(users)*6)
	for i, user := range users {
		valueStrings = append(valueStrings, fmt.Sprintf("($%d, $%d, $%d, $%d, $%d, $%d)", i*6+1, i*6+2, i*6+3, i*6+4, i*6+5, i*6+6))
		valueArgs = append(valueArgs, user[0], user[1], user[2], user[3], user[3], user[3])
	}

	query := fmt.Sprintf(`
		INSERT INTO "user".users (id, name, email, created_at, approved_at, verified_at)
		VALUES %s
	`, strings.Join(valueStrings, ", "))

	_, err := tx.ExecContext(ctx, query, valueArgs...)
	if err != nil {
		return err
	}

	valueStrings = make([]string, 0, len(usersTimezones))
	valueArgs = make([]interface{}, 0, len(usersTimezones)*3)
	for i, user := range usersTimezones {
		valueStrings = append(valueStrings, fmt.Sprintf("($%d, $%d, $%d)", i*3+1, i*3+2, i*3+3))
		valueArgs = append(valueArgs, user[0], user[1], user[2])
	}

	query = fmt.Sprintf(`
		INSERT INTO "user".timezones (user_id, timezone, valid_from)
		VALUES %s
	`, strings.Join(valueStrings, ", "))

	_, err = tx.ExecContext(ctx, query, valueArgs...)
	if err != nil {
		return err
	}

	return nil
}

func downSeedusers(ctx context.Context, tx *sql.Tx) error {
	_, err := tx.ExecContext(ctx, `
	DELETE FROM "user".timezones;
	DELETE FROM "user".users;
`)
	return err
}
