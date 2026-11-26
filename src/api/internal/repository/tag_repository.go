package repository

import (
	"context"
	"errors"
	"fmt"

	"plan/api/internal/database"
	"plan/api/internal/models"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// TagRepository hanterar taggar och kopplingen mellan taggar och mål.
//
// Taggar hör till en användare - två användare kan ha varsin "ute".
type TagRepository struct {
	pool *pgxpool.Pool
}

func NewTagRepository() *TagRepository {
	return &TagRepository{
		pool: database.GetPool(),
	}
}

// List returnerar användarens taggar med antalet mål som har dem.
//
// Taggar som inte sitter på något mål finns med, men med count 0 - de
// kan tas bort genom att skriva om målets tagglista.
func (r *TagRepository) List(ctx context.Context, userID int64) ([]models.Tag, error) {
	query := `SELECT t.name, COUNT(gt.goal_id)
		FROM tags t
		LEFT JOIN goal_tags gt ON gt.tag_id = t.id
		WHERE t.user_id = $1
		GROUP BY t.name
		ORDER BY t.name`

	rows, err := r.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query tags: %w", err)
	}
	defer rows.Close()

	tags := []models.Tag{}
	for rows.Next() {
		var t models.Tag
		if err := rows.Scan(&t.Name, &t.GoalCount); err != nil {
			return nil, fmt.Errorf("failed to scan tag: %w", err)
		}
		tags = append(tags, t)
	}

	return tags, rows.Err()
}

// TagsForGoals returnerar taggarna för flera mål samlat.
//
// Nyckeln är målets id, värdet är taggarnas namn. Mål utan taggar finns
// inte med i kartan - anroparen fyller i en tom lista.
func (r *TagRepository) TagsForGoals(ctx context.Context, goalIDs []int64) (map[int64][]string, error) {
	byGoal := map[int64][]string{}
	if len(goalIDs) == 0 {
		return byGoal, nil
	}

	query := `SELECT gt.goal_id, t.name
		FROM goal_tags gt
		INNER JOIN tags t ON t.id = gt.tag_id
		WHERE gt.goal_id = ANY($1)
		ORDER BY t.name`

	rows, err := r.pool.Query(ctx, query, goalIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to query goal tags: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var goalID int64
		var name string
		if err := rows.Scan(&goalID, &name); err != nil {
			return nil, fmt.Errorf("failed to scan goal tag: %w", err)
		}
		byGoal[goalID] = append(byGoal[goalID], name)
	}

	return byGoal, rows.Err()
}

// GetGoalTags returnerar taggarna på ett mål.
func (r *TagRepository) GetGoalTags(ctx context.Context, goalID int64) ([]string, error) {
	byGoal, err := r.TagsForGoals(ctx, []int64{goalID})
	if err != nil {
		return nil, err
	}

	tags := byGoal[goalID]
	if tags == nil {
		return []string{}, nil
	}

	return tags, nil
}

// SetGoalTags ersätter taggarna på ett mål.
//
// Taggar som inte finns ännu skapas. Namnen normaliseras först, så
// "Ute" och "ute" ger samma tagg. Hela bytet sker i en transaktion -
// antingen får målet alla taggar eller inga.
func (r *TagRepository) SetGoalTags(ctx context.Context, userID, goalID int64, names []string) ([]string, error) {
	normalized, err := models.NormalizeTagNames(names)
	if err != nil {
		return nil, err
	}

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to start transaction: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	if _, err := tx.Exec(ctx, `DELETE FROM goal_tags WHERE goal_id = $1`, goalID); err != nil {
		return nil, fmt.Errorf("failed to clear goal tags: %w", err)
	}

	for _, name := range normalized {
		var tagID int64
		err := tx.QueryRow(ctx,
			`INSERT INTO tags (user_id, name) VALUES ($1, $2)
			 ON CONFLICT (user_id, name) DO UPDATE SET name = EXCLUDED.name
			 RETURNING id`,
			userID, name).Scan(&tagID)
		if err != nil {
			return nil, fmt.Errorf("failed to create tag %q: %w", name, err)
		}

		if _, err := tx.Exec(ctx,
			`INSERT INTO goal_tags (goal_id, tag_id) VALUES ($1, $2)
			 ON CONFLICT DO NOTHING`,
			goalID, tagID); err != nil {
			return nil, fmt.Errorf("failed to link tag %q: %w", name, err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("failed to commit goal tags: %w", err)
	}

	if normalized == nil {
		return []string{}, nil
	}

	return normalized, nil
}

// DeleteGoalTag tar bort en tagg från ett mål, utan att radera taggen
// för andra mål.
func (r *TagRepository) DeleteGoalTag(ctx context.Context, userID, goalID int64, name string) ([]string, error) {
	normalized, err := models.NormalizeTagName(name)
	if err != nil {
		return nil, err
	}

	tag, err := tagIDFor(ctx, r.pool, userID, normalized)
	if err != nil {
		return nil, err
	}
	if tag == nil {
		return r.GetGoalTags(ctx, goalID)
	}

	if _, err := r.pool.Exec(ctx,
		`DELETE FROM goal_tags WHERE goal_id = $1 AND tag_id = $2`,
		goalID, *tag); err != nil {
		return nil, fmt.Errorf("failed to delete goal tag: %w", err)
	}

	return r.GetGoalTags(ctx, goalID)
}

// tagIDFor slår upp en taggs id för användaren. Nil om taggen inte finns.
func tagIDFor(ctx context.Context, pool *pgxpool.Pool, userID int64, name string) (*int64, error) {
	var id int64
	err := pool.QueryRow(ctx,
		`SELECT id FROM tags WHERE user_id = $1 AND name = $2`,
		userID, name).Scan(&id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to query tag: %w", err)
	}
	return &id, nil
}
