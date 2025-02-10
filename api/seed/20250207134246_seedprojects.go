package seed

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/google/uuid"
	"github.com/gosimple/slug"
	"github.com/pressly/goose/v3"
)

var (
	totalProjects      = 1000
	maxProjectsPerUser = 20
)

func init() {
	goose.AddMigrationContext(upSeedprojects, downSeedprojects)
}

func upSeedprojects(ctx context.Context, tx *sql.Tx) error {
	rows, err := tx.QueryContext(ctx, `SELECT id FROM "user".users`)
	if err != nil {
		return err
	}
	defer rows.Close()

	var userIDs []uuid.UUID
	for rows.Next() {
		var id uuid.UUID
		if err := rows.Scan(&id); err != nil {
			return err
		}
		userIDs = append(userIDs, id)
	}

	projects := make([][]any, 0)
	projectIDs := make([]uuid.UUID, 0)
	projectOwners := make(map[uuid.UUID]uuid.UUID)

	for i := 0; i < totalProjects; i++ {
		projectID := uuid.New()
		projectName := gofakeit.Company()
		projectSlug := slug.Make(fmt.Sprintf("%s-%d", projectName, i))
		ownerID := userIDs[gofakeit.Number(0, len(userIDs)-1)]

		projects = append(projects, []any{
			projectID,
			projectName,
			projectSlug,
			gofakeit.RandomString([]string{"free", "pro", "team", "scale", "enterprise"}),
			time.Now(),
		})
		projectIDs = append(projectIDs, projectID)
		projectOwners[projectID] = ownerID
	}

	err = insertInBatches(ctx, tx, projects, 5, `
		INSERT INTO project.projects (id, name, slug, plan, created_at)
		VALUES %s ON CONFLICT DO NOTHING
	`)
	if err != nil {
		return fmt.Errorf("failed to insert projects: %w", err)
	}

	// Now generate and insert project memberships
	seenMemberships := make(map[string]bool)
	projectMemberships := make([][]any, 0)

	for projectID, ownerID := range projectOwners {
		key := fmt.Sprintf("%s-%s", ownerID, projectID)
		if !seenMemberships[key] {
			projectMemberships = append(projectMemberships, []any{
				uuid.New(),
				ownerID,
				projectID,
				"owner",
				time.Now(),
			})
			seenMemberships[key] = true
		}
	}

	for _, projectID := range projectIDs {
		memberCount := gofakeit.Number(1, 30)

		for i := 0; i < memberCount; i++ {
			userID := userIDs[gofakeit.Number(0, len(userIDs)-1)]
			key := fmt.Sprintf("%s-%s", userID, projectID)

			if !seenMemberships[key] {
				projectMemberships = append(projectMemberships, []any{
					uuid.New(),
					userID,
					projectID,
					"member",
					time.Now(),
				})
				seenMemberships[key] = true
			}
		}
	}

	err = insertInBatches(ctx, tx, projectMemberships, 5, `
		INSERT INTO project.memberships (id, user_id, project_id, role, created_at)
		VALUES %s ON CONFLICT DO NOTHING
	`)
	if err != nil {
		return fmt.Errorf("failed to insert project memberships: %w", err)
	}

	projectTeams := make([][]any, 0)
	teamIDs := make([]uuid.UUID, 0)

	for _, projectID := range projectIDs {
		teamsCount := gofakeit.Number(1, 10)

		for j := 0; j < teamsCount; j++ {
			teamID := uuid.New()
			teamName := gofakeit.JobTitle() + "s"
			teamSlug := slug.Make(fmt.Sprintf("%s-%d", teamName, len(teamIDs)+1))

			projectTeams = append(projectTeams, []any{
				teamID,
				projectID,
				teamName,
				gofakeit.Emoji(),
				teamSlug,
				nil, // we'll add it later parent_team_id
				time.Now(),
			})
			teamIDs = append(teamIDs, teamID)
		}
	}

	err = insertInBatches(ctx, tx, projectTeams, 7, `
		INSERT INTO project.teams (id, project_id, name, emoji, slug, parent_team_id, created_at)
		VALUES %s ON CONFLICT DO NOTHING
	`)
	if err != nil {
		return fmt.Errorf("failed to insert teams: %w", err)
	}

	teamMemberships := make([][]any, 0)
	seenTeamMemberships := make(map[string]bool)

	for _, teamID := range teamIDs {
		memberCount := gofakeit.Number(1, 30)

		for k := 0; k < memberCount; k++ {
			userID := userIDs[gofakeit.Number(0, len(userIDs)-1)]
			key := fmt.Sprintf("%s-%s", userID, teamID)

			if !seenTeamMemberships[key] {
				teamMemberships = append(teamMemberships, []any{
					uuid.New(),
					userID,
					teamID,
					gofakeit.JobTitle(),
					time.Now(),
				})
				seenTeamMemberships[key] = true
			}
		}
	}

	err = insertInBatches(ctx, tx, teamMemberships, 5, `
		INSERT INTO project.team_memberships (id, user_id, team_id, position, created_at)
		VALUES %s ON CONFLICT DO NOTHING
	`)
	if err != nil {
		return fmt.Errorf("failed to insert team memberships: %w", err)
	}

	return nil
}

func insertInBatches[T any](ctx context.Context, tx *sql.Tx, items [][]T, paramsPerItem int, query string) error {
	for i := 0; i < len(items); i += 500 {
		end := i + 500
		if end > len(items) {
			end = len(items)
		}

		batch := items[i:end]
		valueStrings := make([]string, 0, len(batch))
		valueArgs := make([]interface{}, 0, len(batch)*paramsPerItem)

		for j, item := range batch {
			paramRefs := make([]string, paramsPerItem)
			for k := 0; k < paramsPerItem; k++ {
				paramRefs[k] = fmt.Sprintf("$%d", j*paramsPerItem+k+1)
			}
			valueStrings = append(valueStrings, fmt.Sprintf("(%s)", strings.Join(paramRefs, ", ")))

			for _, val := range item {
				valueArgs = append(valueArgs, val)
			}
		}

		batchQuery := fmt.Sprintf(query, strings.Join(valueStrings, ", "))
		if _, err := tx.ExecContext(ctx, batchQuery, valueArgs...); err != nil {
			return err
		}
	}
	return nil
}

func downSeedprojects(ctx context.Context, tx *sql.Tx) error {
	_, err := tx.ExecContext(ctx, `
		DELETE FROM project.team_memberships;
		DELETE FROM project.teams;
		DELETE FROM project.memberships;
		DELETE FROM project.projects;
	`)
	return err
}
