package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"plan-api/internal/database"
	"plan-api/internal/models"
	"plan-api/internal/recurrence"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ErrInvalidTimestamp betyder att en tidsstämpel inte kunde tolkas som RFC3339.
var ErrInvalidTimestamp = errors.New("invalid timestamp")

// ErrInvalidDueDate betyder att ett due-datum inte kunde tolkas som "YYYY-MM-DD".
// Det är samma fel som models.ErrInvalidDueDate - models-felet uppstår när
// ett datum avkodas, repot-felet när det skrivs till databasen.
var ErrInvalidDueDate = models.ErrInvalidDueDate

// ErrInvalidStartDate betyder att ett startdatum inte kunde tolkas som
// "YYYY-MM-DD".
var ErrInvalidStartDate = errors.New("ogiltigt startdatum")

// ErrInvalidEstimate betyder att den uppskattade tidsåtgången inte är ett
// positivt antal minuter.
var ErrInvalidEstimate = errors.New("ogiltig tidsåtgång")

// ErrNotRecurring betyder att målet inte är återkommande.
var ErrNotRecurring = errors.New("målet är inte återkommande")

// ErrNoCompletion betyder att det inte finns någon slutförd förekomst att
// avbocka.
var ErrNoCompletion = errors.New("målet har ingen slutförd förekomst")

// dueDateFromTime omvandlar en nullable datumkolumn till ett DueDate.
// Kolumnen är ett DATE, så tiden är alltid midnatt UTC - bara datumet
// används vidare.
func dueDateFromTime(due *time.Time) *models.DueDate {
	if due == nil {
		return nil
	}

	return &models.DueDate{Time: *due}
}

type GoalRepository struct {
	pool *pgxpool.Pool
}

func NewGoalRepository() *GoalRepository {
	return &GoalRepository{
		pool: database.GetPool(),
	}
}

// GetAllGoals returns all goals for a user
func (r *GoalRepository) GetAllGoals(ctx context.Context, userID int64) ([]models.Goal, error) {
	query := `
		SELECT g.id, g.title, COALESCE(g.description, ''), g.icon, g.created, g.started, g.finished, g.due, g.start_date, g.estimated_minutes, g.inbox, g.eternal, g.agent, g."order",
			EXISTS (SELECT 1 FROM goal_relations gr WHERE gr.child_id = g.id) AS has_parent,
			g.recurrence_rule, g.recurrence_mode, g.recurrence_next,
			(SELECT COUNT(*) FROM goal_completions gc WHERE gc.goal_id = g.id) AS recurrence_count
		FROM goals g
		INNER JOIN user_goals ug ON g.id = ug.goal_id
		WHERE ug.user_id = $1
		ORDER BY g.created DESC`

	rows, err := r.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query goals: %w", err)
	}
	defer rows.Close()

	var goals []models.Goal
	for rows.Next() {
		var g models.Goal
		var due *time.Time
		var startDate *time.Time
		var estimatedMinutes *int
		var recurrenceNext *time.Time
		err := rows.Scan(&g.ID, &g.Title, &g.Description, &g.Icon, &g.Created, &g.Started, &g.Finished, &due, &startDate, &estimatedMinutes, &g.Inbox, &g.Eternal, &g.Agent, &g.Order, &g.HasParent,
			&g.RecurrenceRule, &g.RecurrenceMode, &recurrenceNext, &g.RecurrenceCount)
		if err != nil {
			return nil, fmt.Errorf("failed to scan goal: %w", err)
		}
		g.Due = dueDateFromTime(due)
		g.StartDate = dueDateFromTime(startDate)
		g.EstimatedMinutes = estimatedMinutes
		g.RecurrenceNext = dueDateFromTime(recurrenceNext)
		goals = append(goals, g)
	}

	return goals, nil
}

// GetGoalByID returns a single goal by ID
func (r *GoalRepository) GetGoalByID(ctx context.Context, goalID int64) (*models.Goal, error) {
	query := `
		SELECT id, title, COALESCE(description, ''), icon, created, started, finished, due, start_date, estimated_minutes, inbox, eternal, agent, "order",
			EXISTS (SELECT 1 FROM goal_relations gr WHERE gr.child_id = goals.id) AS has_parent,
			recurrence_rule, recurrence_mode, recurrence_next,
			(SELECT COUNT(*) FROM goal_completions gc WHERE gc.goal_id = goals.id) AS recurrence_count
		FROM goals WHERE id = $1`

	var g models.Goal
	var due *time.Time
	var startDate *time.Time
	var estimatedMinutes *int
	var recurrenceNext *time.Time
	err := r.pool.QueryRow(ctx, query, goalID).Scan(
		&g.ID, &g.Title, &g.Description, &g.Icon, &g.Created, &g.Started, &g.Finished, &due, &startDate, &estimatedMinutes, &g.Inbox, &g.Eternal, &g.Agent, &g.Order, &g.HasParent,
		&g.RecurrenceRule, &g.RecurrenceMode, &recurrenceNext, &g.RecurrenceCount)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get goal: %w", err)
	}
	g.Due = dueDateFromTime(due)
	g.StartDate = dueDateFromTime(startDate)
	g.EstimatedMinutes = estimatedMinutes
	g.RecurrenceNext = dueDateFromTime(recurrenceNext)

	return &g, nil
}

// CreateGoal creates a new goal and associates it with a user
func (r *GoalRepository) CreateGoal(ctx context.Context, userID int64, title, icon string) (*models.Goal, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// Insert goal. Nya grundmål läggs sist i ordningen, så att de inte
	// hamnar ovanför mål som redan har en sparad placering.
	var goal models.Goal
	var due *time.Time
	var estimatedMinutes *int
	query := `INSERT INTO goals (title, icon, "order")
			  SELECT $1, $2, COALESCE((
				  SELECT MAX(g2."order") + 1
				  FROM goals g2
				  INNER JOIN user_goals ug2 ON g2.id = ug2.goal_id
				  WHERE ug2.user_id = $3), 0)
			  RETURNING id, title, COALESCE(description, ''), icon, status_id, created, started, finished, due, estimated_minutes, inbox, eternal, agent, "order"`
	err = tx.QueryRow(ctx, query, title, icon, userID).Scan(
		&goal.ID, &goal.Title, &goal.Description, &goal.Icon, &goal.StatusID, &goal.Created, &goal.Started, &goal.Finished, &due, &estimatedMinutes, &goal.Inbox, &goal.Eternal, &goal.Agent, &goal.Order)
	if err != nil {
		return nil, fmt.Errorf("failed to create goal: %w", err)
	}
	goal.Due = dueDateFromTime(due)
	goal.EstimatedMinutes = estimatedMinutes

	// Associate with user
	_, err = tx.Exec(ctx, `INSERT INTO user_goals (user_id, goal_id) VALUES ($1, $2)`, userID, goal.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to associate goal with user: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return &goal, nil
}

// UpdateGoal updates a goal's fields
func (r *GoalRepository) UpdateGoal(ctx context.Context, goalID int64, req models.UpdateGoalRequest) (*models.Goal, error) {
	// Build dynamic update query
	query := "UPDATE goals SET "
	var args []interface{}
	argNum := 1
	hasUpdate := false

	if req.Title != nil {
		query += fmt.Sprintf("title = $%d, ", argNum)
		args = append(args, *req.Title)
		argNum++
		hasUpdate = true
	}
	if req.Description != nil {
		query += fmt.Sprintf("description = $%d, ", argNum)
		args = append(args, *req.Description)
		argNum++
		hasUpdate = true
	}
	if req.Icon != nil {
		query += fmt.Sprintf("icon = $%d, ", argNum)
		args = append(args, *req.Icon)
		argNum++
		hasUpdate = true
	}
	if req.Started.Set {
		hasUpdate = true
		if req.Started.Value == nil || *req.Started.Value == "" {
			query += "started = NULL, "
		} else {
			t, err := time.Parse(time.RFC3339, *req.Started.Value)
			if err != nil {
				return nil, fmt.Errorf("%w: %q", ErrInvalidTimestamp, *req.Started.Value)
			}
			query += fmt.Sprintf("started = $%d, ", argNum)
			args = append(args, t)
			argNum++
		}
	}
	// Ett evigt mål kan aldrig vara färdigt. När målet markeras som evigt
	// nollställs därför finished, oavsett vad requesten samtidigt säger.
	becomesEternal := req.Eternal != nil && *req.Eternal

	if req.Finished.Set && !becomesEternal {
		hasUpdate = true
		if req.Finished.Value == nil || *req.Finished.Value == "" {
			query += "finished = NULL, "
		} else {
			t, err := time.Parse(time.RFC3339, *req.Finished.Value)
			if err != nil {
				return nil, fmt.Errorf("%w: %q", ErrInvalidTimestamp, *req.Finished.Value)
			}
			query += fmt.Sprintf("finished = $%d, ", argNum)
			args = append(args, t)
			argNum++
		}
	}
	if req.Eternal != nil {
		hasUpdate = true
		query += fmt.Sprintf("eternal = $%d, ", argNum)
		args = append(args, *req.Eternal)
		argNum++
		if becomesEternal {
			query += "finished = NULL, "
		}
	}
	if req.Agent != nil {
		hasUpdate = true
		query += fmt.Sprintf("agent = $%d, ", argNum)
		args = append(args, *req.Agent)
		argNum++
	}
	if req.Due.Set {
		hasUpdate = true
		if req.Due.Value == nil || *req.Due.Value == "" {
			query += "due = NULL, "
		} else {
			t, err := time.Parse(models.DueDateFormat, *req.Due.Value)
			if err != nil {
				return nil, fmt.Errorf("%w: %q", ErrInvalidDueDate, *req.Due.Value)
			}
			query += fmt.Sprintf("due = $%d, ", argNum)
			args = append(args, t)
			argNum++
		}
	}

	// Startdatumet anger från vilket datum målet får dyka upp i
	// prioriteringslistan. Null = målet får synas direkt.
	if req.StartDate.Set {
		hasUpdate = true
		if req.StartDate.Value == nil || *req.StartDate.Value == "" {
			query += "start_date = NULL, "
		} else {
			t, err := time.Parse(models.DueDateFormat, *req.StartDate.Value)
			if err != nil {
				return nil, fmt.Errorf("%w: %q", ErrInvalidStartDate, *req.StartDate.Value)
			}
			query += fmt.Sprintf("start_date = $%d, ", argNum)
			args = append(args, t)
			argNum++
		}
	}

	// Den uppskattade tidsåtgången sätts i minuter. Null rensar den,
	// och negativa värden avvisas - en uppskattning kan inte vara negativ.
	if req.EstimatedMinutes.Set {
		hasUpdate = true
		if req.EstimatedMinutes.Value == nil {
			query += "estimated_minutes = NULL, "
		} else {
			if *req.EstimatedMinutes.Value < 0 {
				return nil, fmt.Errorf("%w: %d", ErrInvalidEstimate, *req.EstimatedMinutes.Value)
			}
			query += fmt.Sprintf("estimated_minutes = $%d, ", argNum)
			args = append(args, *req.EstimatedMinutes.Value)
			argNum++
		}
	}

	if !hasUpdate && !req.RecurrenceRule.Set && !req.RecurrenceMode.Set {
		return r.GetGoalByID(ctx, goalID)
	}

	if hasUpdate {
		// Remove trailing comma and space
		query = query[:len(query)-2]
		query += fmt.Sprintf(" WHERE id = $%d", argNum)
		args = append(args, goalID)

		_, err := r.pool.Exec(ctx, query, args...)
		if err != nil {
			return nil, fmt.Errorf("failed to update goal: %w", err)
		}
	}

	if req.RecurrenceRule.Set || req.RecurrenceMode.Set {
		if err := r.applyRecurrenceUpdate(ctx, goalID, req); err != nil {
			return nil, err
		}
	}

	return r.GetGoalByID(ctx, goalID)
}

// applyRecurrenceUpdate sätter eller tar bort målets upprepning.
//
// En ny regel startar en ny serie - de tidigare slutförandena hör till den
// gamla regeln och tas därför bort. Att bara byta strategi påverkar inte den
// aktuella förekomsten.
func (r *GoalRepository) applyRecurrenceUpdate(ctx context.Context, goalID int64, req models.UpdateGoalRequest) error {
	if req.RecurrenceRule.Set {
		if req.RecurrenceRule.Value == nil {
			return r.ClearRecurrence(ctx, goalID)
		}

		rule := strings.TrimSpace(*req.RecurrenceRule.Value)
		if rule == "" {
			return fmt.Errorf("%w: regeln är tom", recurrence.ErrInvalidRule)
		}

		mode := recurrence.ModeSchedule
		if req.RecurrenceMode.Set && req.RecurrenceMode.Value != nil && *req.RecurrenceMode.Value != "" {
			mode = recurrence.Mode(*req.RecurrenceMode.Value)
		} else {
			existing, err := r.GetGoalByID(ctx, goalID)
			if err != nil {
				return err
			}
			if existing != nil && existing.RecurrenceMode != nil && *existing.RecurrenceMode != "" {
				mode = recurrence.Mode(*existing.RecurrenceMode)
			}
		}

		start := time.Now().UTC()
		if req.RecurrenceStartDate != nil && *req.RecurrenceStartDate != "" {
			parsed, err := time.Parse(models.DueDateFormat, *req.RecurrenceStartDate)
			if err != nil {
				return fmt.Errorf("%w: %q", ErrInvalidDueDate, *req.RecurrenceStartDate)
			}
			start = parsed
		}

		_, err := r.SetRecurrence(ctx, goalID, rule, mode, start)
		return err
	}

	if req.RecurrenceMode.Value == nil || *req.RecurrenceMode.Value == "" {
		return fmt.Errorf("%w: strategin måste anges", recurrence.ErrInvalidMode)
	}

	_, err := r.SetRecurrenceMode(ctx, goalID, recurrence.Mode(*req.RecurrenceMode.Value))
	return err
}

// SetRecurrence gör ett mål återkommande med en ny upprepningsregel.
//
// Serien börjar på start och nästa förekomst räknas fram därifrån. De
// tidigare slutförandena hör till den gamla regeln och tas därför bort.
func (r *GoalRepository) SetRecurrence(ctx context.Context, goalID int64, rule string, mode recurrence.Mode, start time.Time) (*models.Goal, error) {
	if err := recurrence.Validate(rule); err != nil {
		return nil, err
	}
	if err := recurrence.ValidateMode(mode); err != nil {
		return nil, err
	}

	anchor := time.Date(start.Year(), start.Month(), start.Day(), 0, 0, 0, 0, time.UTC)

	first, err := recurrence.FirstOccurrence(rule, anchor)
	if err != nil {
		return nil, err
	}

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx, `
		UPDATE goals
		SET recurrence_rule = $1, recurrence_mode = $2, recurrence_anchor = $3, recurrence_next = $4, finished = NULL
		WHERE id = $5`, rule, string(mode), anchor, first, goalID)
	if err != nil {
		return nil, fmt.Errorf("failed to set recurrence: %w", err)
	}

	if _, err := tx.Exec(ctx, `DELETE FROM goal_completions WHERE goal_id = $1`, goalID); err != nil {
		return nil, fmt.Errorf("failed to clear goal completions: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return r.GetGoalByID(ctx, goalID)
}

// SetRecurrenceMode byter schemaläggningsstrategi för ett återkommande mål.
//
// Den aktuella förekomsten påverkas inte - strategin avgör hur nästa
// förekomst räknas när nästa förekomst slutförs.
func (r *GoalRepository) SetRecurrenceMode(ctx context.Context, goalID int64, mode recurrence.Mode) (*models.Goal, error) {
	if err := recurrence.ValidateMode(mode); err != nil {
		return nil, err
	}

	result, err := r.pool.Exec(ctx, `UPDATE goals SET recurrence_mode = $1 WHERE id = $2 AND recurrence_rule IS NOT NULL`, string(mode), goalID)
	if err != nil {
		return nil, fmt.Errorf("failed to set recurrence mode: %w", err)
	}
	if result.RowsAffected() == 0 {
		return nil, ErrNotRecurring
	}

	return r.GetGoalByID(ctx, goalID)
}

// ClearRecurrence tar bort upprepningen så att målet blir ett vanligt mål.
func (r *GoalRepository) ClearRecurrence(ctx context.Context, goalID int64) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx, `
		UPDATE goals
		SET recurrence_rule = NULL, recurrence_mode = NULL, recurrence_anchor = NULL, recurrence_next = NULL
		WHERE id = $1`, goalID)
	if err != nil {
		return fmt.Errorf("failed to clear recurrence: %w", err)
	}

	if _, err := tx.Exec(ctx, `DELETE FROM goal_completions WHERE goal_id = $1`, goalID); err != nil {
		return fmt.Errorf("failed to clear goal completions: %w", err)
	}

	return tx.Commit(ctx)
}

// CompleteOccurrence slutför den aktuella förekomsten av ett återkommande mål.
//
// Förekomsten loggas i goal_completions och nästa förekomst räknas fram enligt
// målets strategi. När serien tar slut markeras målet som klart.
//
// occurrenceDate anger vilken förekomst som slutförs. Utan det slutförs
// målets aktuella förekomst. En redan slutförd förekomst ändrar ingenting -
// anropet är då en dubblett.
func (r *GoalRepository) CompleteOccurrence(ctx context.Context, goalID int64, occurrenceDate *time.Time, completedOn time.Time) (*models.Goal, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	var rule *string
	var mode *string
	var anchor *time.Time
	var next *time.Time

	err = tx.QueryRow(ctx, `
		SELECT recurrence_rule, recurrence_mode, recurrence_anchor, recurrence_next
		FROM goals WHERE id = $1 FOR UPDATE`, goalID).Scan(&rule, &mode, &anchor, &next)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to load goal recurrence: %w", err)
	}

	if rule == nil || mode == nil || next == nil {
		return nil, ErrNotRecurring
	}

	occurrence := startOfDay(next.UTC())

	if occurrenceDate != nil {
		requested := startOfDay(occurrenceDate.UTC())
		if !requested.Equal(occurrence) {
			// En redan slutförd förekomst är en dubblett - inget ändras.
			var completed bool
			err := tx.QueryRow(ctx, `
				SELECT EXISTS(SELECT 1 FROM goal_completions WHERE goal_id = $1 AND occurrence_date = $2)`,
				goalID, requested).Scan(&completed)
			if err != nil {
				return nil, fmt.Errorf("failed to check goal completion: %w", err)
			}
			if completed {
				if err := tx.Commit(ctx); err != nil {
					return nil, fmt.Errorf("failed to commit transaction: %w", err)
				}
				return r.GetGoalByID(ctx, goalID)
			}

			return nil, fmt.Errorf("%w: %s är inte den aktuella förekomsten", recurrence.ErrNoOccurrence, requested.Format(time.DateOnly))
		}
		occurrence = requested
	}

	completedOn = startOfDay(completedOn)

	var completionID int64
	err = tx.QueryRow(ctx, `
		INSERT INTO goal_completions (goal_id, occurrence_date)
		VALUES ($1, $2)
		ON CONFLICT (goal_id, occurrence_date) DO NOTHING
		RETURNING id`, goalID, occurrence).Scan(&completionID)
	if err != nil {
		if err == pgx.ErrNoRows {
			if err := tx.Commit(ctx); err != nil {
				return nil, fmt.Errorf("failed to commit transaction: %w", err)
			}
			return r.GetGoalByID(ctx, goalID)
		}
		return nil, fmt.Errorf("failed to log goal completion: %w", err)
	}

	seriesAnchor := occurrence
	if anchor != nil {
		seriesAnchor = startOfDay(anchor.UTC())
	}

	nextOccurrence, err := recurrence.NextOccurrence(*rule, recurrence.Mode(*mode), seriesAnchor, occurrence, completedOn)
	if err != nil {
		if !errors.Is(err, recurrence.ErrNoOccurrence) {
			return nil, err
		}

		// Serien är slut - målet har inget mer att göra.
		_, err = tx.Exec(ctx, `
			UPDATE goals SET recurrence_next = NULL, finished = CURRENT_TIMESTAMP
			WHERE id = $1`, goalID)
		if err != nil {
			return nil, fmt.Errorf("failed to finish goal: %w", err)
		}
	} else {
		_, err = tx.Exec(ctx, `UPDATE goals SET recurrence_next = $1 WHERE id = $2`, nextOccurrence, goalID)
		if err != nil {
			return nil, fmt.Errorf("failed to update recurrence: %w", err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return r.GetGoalByID(ctx, goalID)
}

// UndoOccurrence avbockar den senast slutförda förekomsten.
func (r *GoalRepository) UndoOccurrence(ctx context.Context, goalID int64) (*models.Goal, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	var rule *string
	err = tx.QueryRow(ctx, `SELECT recurrence_rule FROM goals WHERE id = $1 FOR UPDATE`, goalID).Scan(&rule)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to load goal recurrence: %w", err)
	}
	if rule == nil {
		return nil, ErrNotRecurring
	}

	var occurrence time.Time
	err = tx.QueryRow(ctx, `
		SELECT occurrence_date FROM goal_completions
		WHERE goal_id = $1
		ORDER BY completed_at DESC, id DESC
		LIMIT 1`, goalID).Scan(&occurrence)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, ErrNoCompletion
		}
		return nil, fmt.Errorf("failed to load goal completion: %w", err)
	}

	if _, err := tx.Exec(ctx, `DELETE FROM goal_completions WHERE goal_id = $1 AND occurrence_date = $2`, goalID, occurrence); err != nil {
		return nil, fmt.Errorf("failed to delete goal completion: %w", err)
	}

	_, err = tx.Exec(ctx, `
		UPDATE goals SET recurrence_next = $1, finished = NULL
		WHERE id = $2`, occurrence, goalID)
	if err != nil {
		return nil, fmt.Errorf("failed to restore recurrence: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return r.GetGoalByID(ctx, goalID)
}

// DeleteGoal deletes a goal and all related data
func (r *GoalRepository) DeleteGoal(ctx context.Context, goalID int64) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// Delete in correct order to avoid FK violations
	_, err = tx.Exec(ctx, `DELETE FROM user_goals WHERE goal_id = $1`, goalID)
	if err != nil {
		return fmt.Errorf("failed to delete user_goals: %w", err)
	}

	_, err = tx.Exec(ctx, `DELETE FROM goal_relations WHERE parent_id = $1 OR child_id = $1`, goalID)
	if err != nil {
		return fmt.Errorf("failed to delete goal_relations: %w", err)
	}

	_, err = tx.Exec(ctx, `DELETE FROM goal_dependencies WHERE goal_id = $1 OR depends_on_id = $1`, goalID)
	if err != nil {
		return fmt.Errorf("failed to delete goal_dependencies: %w", err)
	}

	_, err = tx.Exec(ctx, `DELETE FROM goals WHERE id = $1`, goalID)
	if err != nil {
		return fmt.Errorf("failed to delete goal: %w", err)
	}

	return tx.Commit(ctx)
}

// UserHasAccess checks if a user has access to a goal
func (r *GoalRepository) UserHasAccess(ctx context.Context, userID, goalID int64) (bool, error) {
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM user_goals WHERE user_id = $1 AND goal_id = $2)`
	err := r.pool.QueryRow(ctx, query, userID, goalID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check access: %w", err)
	}
	return exists, nil
}

// GetAllRelations returns every parent-child relation among a user's goals.
//
// Alla relationer hämtas i en fråga, så att klienten kan räkna ut barn och
// föräldrar lokalt i stället för att fråga per mål.
func (r *GoalRepository) GetAllRelations(ctx context.Context, userID int64) ([]models.GoalRelation, error) {
	query := `
		SELECT gr.parent_id, gr.child_id, gr."order", gr.weight
		FROM goal_relations gr
		INNER JOIN user_goals ug ON ug.goal_id = gr.child_id
		WHERE ug.user_id = $1
		ORDER BY gr."order" ASC`

	rows, err := r.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get relations: %w", err)
	}
	defer rows.Close()

	var relations []models.GoalRelation
	for rows.Next() {
		var rel models.GoalRelation
		if err := rows.Scan(&rel.ParentID, &rel.ChildID, &rel.Order, &rel.Weight); err != nil {
			return nil, fmt.Errorf("failed to scan relation: %w", err)
		}
		relations = append(relations, rel)
	}

	return relations, rows.Err()
}

// GetAllDependencies returns every dependency among a user's goals.
func (r *GoalRepository) GetAllDependencies(ctx context.Context, userID int64) ([]models.GoalDependency, error) {
	query := `
		SELECT gd.id, gd.goal_id, gd.depends_on_id, gd.created
		FROM goal_dependencies gd
		INNER JOIN user_goals ug ON ug.goal_id = gd.goal_id
		WHERE ug.user_id = $1`

	rows, err := r.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get dependencies: %w", err)
	}
	defer rows.Close()

	var deps []models.GoalDependency
	for rows.Next() {
		var d models.GoalDependency
		if err := rows.Scan(&d.ID, &d.GoalID, &d.DependsOnID, &d.Created); err != nil {
			return nil, fmt.Errorf("failed to scan dependency: %w", err)
		}
		deps = append(deps, d)
	}

	return deps, rows.Err()
}

// GetChildRelations returns child relations for a goal
func (r *GoalRepository) GetChildRelations(ctx context.Context, parentID int64) ([]models.ChildRelation, error) {
	query := `SELECT child_id, "order", weight FROM goal_relations WHERE parent_id = $1 ORDER BY "order" ASC`
	rows, err := r.pool.Query(ctx, query, parentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get child relations: %w", err)
	}
	defer rows.Close()

	var relations []models.ChildRelation
	for rows.Next() {
		var r models.ChildRelation
		if err := rows.Scan(&r.ChildID, &r.Order, &r.Weight); err != nil {
			return nil, fmt.Errorf("failed to scan relation: %w", err)
		}
		relations = append(relations, r)
	}

	return relations, nil
}

// GetParentRelations returns parent goals for a goal
func (r *GoalRepository) GetParentRelations(ctx context.Context, childID int64) ([]models.Goal, error) {
	query := `
		SELECT g.id, g.title, COALESCE(g.description, ''), g.icon, g.created, g.started, g.finished, g.due, g.estimated_minutes, g.inbox, g.eternal, g.agent,
			g.recurrence_rule, g.recurrence_mode, g.recurrence_next,
			(SELECT COUNT(*) FROM goal_completions gc WHERE gc.goal_id = g.id) AS recurrence_count
		FROM goal_relations gr
		INNER JOIN goals g ON g.id = gr.parent_id
		WHERE gr.child_id = $1`

	rows, err := r.pool.Query(ctx, query, childID)
	if err != nil {
		return nil, fmt.Errorf("failed to get parent relations: %w", err)
	}
	defer rows.Close()

	var goals []models.Goal
	for rows.Next() {
		var g models.Goal
		var due *time.Time
		var estimatedMinutes *int
		var recurrenceNext *time.Time
		if err := rows.Scan(&g.ID, &g.Title, &g.Description, &g.Icon,
			&g.Created, &g.Started, &g.Finished, &due, &estimatedMinutes, &g.Inbox, &g.Eternal, &g.Agent,
			&g.RecurrenceRule, &g.RecurrenceMode, &recurrenceNext, &g.RecurrenceCount); err != nil {
			return nil, fmt.Errorf("failed to scan goal: %w", err)
		}
		g.Due = dueDateFromTime(due)
		g.EstimatedMinutes = estimatedMinutes
		g.RecurrenceNext = dueDateFromTime(recurrenceNext)
		goals = append(goals, g)
	}

	return goals, nil
}

// CreateRelation creates a parent-child relation
func (r *GoalRepository) CreateRelation(ctx context.Context, parentID, childID int64, order, weight int) (*models.ChildRelation, error) {
	query := `
		INSERT INTO goal_relations (parent_id, child_id, "order", weight)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (parent_id, child_id) DO UPDATE SET "order" = $3, weight = $4
		RETURNING child_id, "order", weight`

	var rel models.ChildRelation
	err := r.pool.QueryRow(ctx, query, parentID, childID, order, weight).Scan(&rel.ChildID, &rel.Order, &rel.Weight)
	if err != nil {
		return nil, fmt.Errorf("failed to create relation: %w", err)
	}

	return &rel, nil
}

// UpdateRelation updates a relation's order and weight
func (r *GoalRepository) UpdateRelation(ctx context.Context, parentID, childID int64, order, weight *int) error {
	query := "UPDATE goal_relations SET "
	var args []interface{}
	argNum := 1

	if order != nil {
		query += fmt.Sprintf("\"order\" = $%d, ", argNum)
		args = append(args, *order)
		argNum++
	}
	if weight != nil {
		query += fmt.Sprintf("weight = $%d, ", argNum)
		args = append(args, *weight)
		argNum++
	}

	if len(args) == 0 {
		return nil
	}

	query = query[:len(query)-2]
	query += fmt.Sprintf(" WHERE parent_id = $%d AND child_id = $%d", argNum, argNum+1)
	args = append(args, parentID, childID)

	_, err := r.pool.Exec(ctx, query, args...)
	return err
}

// ReorderRelations sets a new order for the children of a goal.
// The children in childIDs get order 0, 1, 2... in the given order. Any child
// that is not listed keeps its relative position and is placed after the
// listed ones, so a stale client can never leave gaps or duplicates behind.
func (r *GoalRepository) ReorderRelations(ctx context.Context, parentID int64, childIDs []int64) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// Fetch the children that are not part of the request so they can be
	// appended instead of being left with a colliding order value.
	rows, err := tx.Query(ctx, `SELECT child_id FROM goal_relations WHERE parent_id = $1 ORDER BY "order" ASC, child_id ASC`, parentID)
	if err != nil {
		return fmt.Errorf("failed to fetch child relations: %w", err)
	}
	var existing []int64
	for rows.Next() {
		var childID int64
		if err := rows.Scan(&childID); err != nil {
			rows.Close()
			return fmt.Errorf("failed to scan child relation: %w", err)
		}
		existing = append(existing, childID)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return fmt.Errorf("failed to read child relations: %w", err)
	}

	for i, childID := range mergeOrder(existing, childIDs) {
		if _, err := tx.Exec(ctx, `UPDATE goal_relations SET "order" = $1 WHERE parent_id = $2 AND child_id = $3`, i, parentID, childID); err != nil {
			return fmt.Errorf("failed to update order for child %d: %w", childID, err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit reorder: %w", err)
	}
	return nil
}

// ReorderRootGoals sets a new order for a user's root goals (goals without a
// parent). The goals in goalIDs get order 0, 1, 2... in the given order. Any
// root goal that is not listed keeps its relative position and is placed
// after the listed ones, so a stale client can never leave gaps or
// duplicates behind. Goals that belong to another user are ignored.
func (r *GoalRepository) ReorderRootGoals(ctx context.Context, userID int64, goalIDs []int64) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// Fetch the root goals that are not part of the request so they can be
	// appended instead of being left with a colliding order value.
	rows, err := tx.Query(ctx, `
		SELECT g.id
		FROM goals g
		INNER JOIN user_goals ug ON ug.goal_id = g.id
		WHERE ug.user_id = $1
			AND NOT EXISTS (SELECT 1 FROM goal_relations gr WHERE gr.child_id = g.id)
		ORDER BY g."order" ASC, g.created DESC, g.id ASC`, userID)
	if err != nil {
		return fmt.Errorf("failed to fetch root goals: %w", err)
	}
	var existing []int64
	for rows.Next() {
		var goalID int64
		if err := rows.Scan(&goalID); err != nil {
			rows.Close()
			return fmt.Errorf("failed to scan root goal: %w", err)
		}
		existing = append(existing, goalID)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return fmt.Errorf("failed to read root goals: %w", err)
	}

	for i, goalID := range mergeOrder(existing, goalIDs) {
		if _, err := tx.Exec(ctx, `
			UPDATE goals SET "order" = $1
			WHERE id = $2
				AND id IN (SELECT goal_id FROM user_goals WHERE user_id = $3)
				AND NOT EXISTS (SELECT 1 FROM goal_relations gr WHERE gr.child_id = goals.id)`,
			i, goalID, userID); err != nil {
			return fmt.Errorf("failed to update order for goal %d: %w", goalID, err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit reorder: %w", err)
	}
	return nil
}

// ReorderPriorityGoals sets the order of all active goals in the priority list.
func (r *GoalRepository) ReorderPriorityGoals(ctx context.Context, userID int64, goalIDs []int64) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	rows, err := tx.Query(ctx, `
		SELECT g.id
		FROM goals g
		INNER JOIN user_goals ug ON ug.goal_id = g.id
		WHERE ug.user_id = $1 AND g.finished IS NULL AND g.id <> 1
		ORDER BY g.priority_order ASC, g.created DESC, g.id ASC`, userID)
	if err != nil {
		return fmt.Errorf("failed to fetch priority goals: %w", err)
	}
	var existing []int64
	for rows.Next() {
		var goalID int64
		if err := rows.Scan(&goalID); err != nil {
			rows.Close()
			return fmt.Errorf("failed to scan priority goal: %w", err)
		}
		existing = append(existing, goalID)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return fmt.Errorf("failed to read priority goals: %w", err)
	}

	for i, goalID := range mergeOrder(existing, goalIDs) {
		if _, err := tx.Exec(ctx, `
			UPDATE goals SET priority_order = $1
			WHERE id = $2 AND id IN (SELECT goal_id FROM user_goals WHERE user_id = $3)
				AND finished IS NULL AND id <> 1`, i, goalID, userID); err != nil {
			return fmt.Errorf("failed to update priority order for goal %d: %w", goalID, err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit priority reorder: %w", err)
	}
	return nil
}

// mergeOrder combines the requested order with the items that exist on the
// server but were not part of the request. The requested items come first, in
// the requested order, and the rest keep their relative order. Items that do
// not exist on the server are ignored, so the resulting order always has
// exactly one slot per existing item.
func mergeOrder(existing []int64, requested []int64) []int64 {
	known := make(map[int64]bool, len(existing))
	for _, id := range existing {
		known[id] = true
	}

	ordered := make([]int64, 0, len(existing))
	seen := make(map[int64]bool, len(existing))

	for _, id := range requested {
		if seen[id] || !known[id] {
			continue
		}
		seen[id] = true
		ordered = append(ordered, id)
	}

	for _, id := range existing {
		if seen[id] {
			continue
		}
		seen[id] = true
		ordered = append(ordered, id)
	}

	return ordered
}

// DeleteRelation deletes a parent-child relation
func (r *GoalRepository) DeleteRelation(ctx context.Context, parentID, childID int64) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM goal_relations WHERE parent_id = $1 AND child_id = $2`, parentID, childID)
	return err
}

// GetDependencies returns dependencies for a goal
func (r *GoalRepository) GetDependencies(ctx context.Context, goalID int64) ([]models.GoalDependency, error) {
	query := `SELECT id, goal_id, depends_on_id, created FROM goal_dependencies WHERE goal_id = $1`
	rows, err := r.pool.Query(ctx, query, goalID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var deps []models.GoalDependency
	for rows.Next() {
		var d models.GoalDependency
		if err := rows.Scan(&d.ID, &d.GoalID, &d.DependsOnID, &d.Created); err != nil {
			return nil, err
		}
		deps = append(deps, d)
	}

	return deps, nil
}

// GetBlockingGoals returns goals that depend on this goal
func (r *GoalRepository) GetBlockingGoals(ctx context.Context, goalID int64) ([]models.Goal, error) {
	query := `
		SELECT g.id, g.title, COALESCE(g.description, ''), g.icon, g.created, g.started, g.finished, g.due, g.estimated_minutes, g.inbox, g.eternal, g.agent,
			g.recurrence_rule, g.recurrence_mode, g.recurrence_next,
			(SELECT COUNT(*) FROM goal_completions gc WHERE gc.goal_id = g.id) AS recurrence_count
		FROM goal_dependencies gd
		INNER JOIN goals g ON g.id = gd.goal_id
		WHERE gd.depends_on_id = $1`

	rows, err := r.pool.Query(ctx, query, goalID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var goals []models.Goal
	for rows.Next() {
		var g models.Goal
		var due *time.Time
		var estimatedMinutes *int
		var recurrenceNext *time.Time
		if err := rows.Scan(&g.ID, &g.Title, &g.Description, &g.Icon, &g.Created, &g.Started, &g.Finished, &due, &estimatedMinutes, &g.Inbox, &g.Eternal, &g.Agent,
			&g.RecurrenceRule, &g.RecurrenceMode, &recurrenceNext, &g.RecurrenceCount); err != nil {
			return nil, err
		}
		g.Due = dueDateFromTime(due)
		g.EstimatedMinutes = estimatedMinutes
		g.RecurrenceNext = dueDateFromTime(recurrenceNext)
		goals = append(goals, g)
	}

	return goals, nil
}

// CreateDependency creates a dependency between goals
func (r *GoalRepository) CreateDependency(ctx context.Context, goalID, dependsOnID int64) (*models.GoalDependency, error) {
	query := `INSERT INTO goal_dependencies (goal_id, depends_on_id) VALUES ($1, $2) RETURNING id, goal_id, depends_on_id, created`
	var dep models.GoalDependency
	err := r.pool.QueryRow(ctx, query, goalID, dependsOnID).Scan(&dep.ID, &dep.GoalID, &dep.DependsOnID, &dep.Created)
	if err != nil {
		return nil, err
	}
	return &dep, nil
}

// DeleteDependency deletes a dependency
func (r *GoalRepository) DeleteDependency(ctx context.Context, goalID, dependsOnID int64) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM goal_dependencies WHERE goal_id = $1 AND depends_on_id = $2`, goalID, dependsOnID)
	return err
}

// CheckForCycle checks if adding a dependency would create a cycle
func (r *GoalRepository) CheckForCycle(ctx context.Context, goalID, dependsOnID int64) (bool, []int64, error) {
	// Check if there's a path from dependsOnID to goalID
	// If yes, adding goalID -> dependsOnID would create a cycle
	visited := make(map[int64]bool)
	path := make([]int64, 0)

	var dfs func(currentID int64) bool
	dfs = func(currentID int64) bool {
		if currentID == goalID {
			return true
		}
		if visited[currentID] {
			return false
		}
		visited[currentID] = true
		path = append(path, currentID)

		query := `SELECT depends_on_id FROM goal_dependencies WHERE goal_id = $1`
		rows, err := r.pool.Query(ctx, query, currentID)
		if err != nil {
			return false
		}
		defer rows.Close()

		for rows.Next() {
			var nextID int64
			if err := rows.Scan(&nextID); err != nil {
				continue
			}
			if dfs(nextID) {
				return true
			}
		}

		path = path[:len(path)-1]
		return false
	}

	hasCycle := dfs(dependsOnID)
	return hasCycle, path, nil
}

// GetPrioritizedGoals returns goals sorted by priority
func (r *GoalRepository) GetPrioritizedGoals(ctx context.Context, userID int64) ([]models.PrioritizedGoal, error) {
	query := `
		SELECT
			g.id, g.title, COALESCE(g.description, ''), g.icon, g.created, g.started, g.finished, g.due, g.start_date, g.estimated_minutes, g.inbox, g.eternal, g.agent,
			g.priority_order,
			COALESCE((SELECT MAX(weight) FROM goal_relations WHERE child_id = g.id), 10) AS weight,
			(SELECT g2.title FROM goal_relations gr INNER JOIN goals g2 ON g2.id = gr.parent_id WHERE gr.child_id = g.id ORDER BY gr.weight DESC LIMIT 1) AS parent_title,
			(SELECT gr.parent_id FROM goal_relations gr WHERE gr.child_id = g.id ORDER BY gr.weight DESC LIMIT 1) AS parent_id,
			g.recurrence_rule, g.recurrence_mode, g.recurrence_next,
			(SELECT COUNT(*) FROM goal_completions gc WHERE gc.goal_id = g.id) AS recurrence_count
		FROM goals g
		INNER JOIN user_goals ug ON g.id = ug.goal_id
		WHERE ug.user_id = $1
		  AND g.finished IS NULL
		  AND g.id <> 1
		  AND (g.start_date IS NULL OR g.start_date <= CURRENT_DATE)
		ORDER BY g.priority_order ASC, weight DESC, g.created DESC`

	rows, err := r.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var goals []models.PrioritizedGoal
	for rows.Next() {
		var g models.PrioritizedGoal
		var parentTitle *string
		var parentID *int64
		var due *time.Time
		var startDate *time.Time
		var estimatedMinutes *int
		var recurrenceNext *time.Time
		if err := rows.Scan(&g.ID, &g.Title, &g.Description, &g.Icon, &g.Created, &g.Started, &g.Finished, &due, &startDate, &estimatedMinutes, &g.Inbox, &g.Eternal, &g.Agent,
			&g.PriorityOrder, &g.Weight, &parentTitle, &parentID,
			&g.RecurrenceRule, &g.RecurrenceMode, &recurrenceNext, &g.RecurrenceCount); err != nil {
			return nil, err
		}
		g.Due = dueDateFromTime(due)
		g.StartDate = dueDateFromTime(startDate)
		g.EstimatedMinutes = estimatedMinutes
		g.RecurrenceNext = dueDateFromTime(recurrenceNext)
		g.ParentTitle = parentTitle
		g.ParentID = parentID
		goals = append(goals, g)
	}

	return goals, nil
}

// GetAllStatuses returns all available statuses
func (r *GoalRepository) GetAllStatuses(ctx context.Context) ([]models.Status, error) {
	query := `SELECT id, name, label FROM statuses ORDER BY id`
	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get statuses: %w", err)
	}
	defer rows.Close()

	var statuses []models.Status
	for rows.Next() {
		var s models.Status
		if err := rows.Scan(&s.ID, &s.Name, &s.Label); err != nil {
			return nil, fmt.Errorf("failed to scan status: %w", err)
		}
		statuses = append(statuses, s)
	}

	return statuses, nil
}

// GetStatusByID returns a status by ID
func (r *GoalRepository) GetStatusByID(ctx context.Context, statusID int64) (*models.Status, error) {
	query := `SELECT id, name, label FROM statuses WHERE id = $1`
	var s models.Status
	err := r.pool.QueryRow(ctx, query, statusID).Scan(&s.ID, &s.Name, &s.Label)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get status: %w", err)
	}
	return &s, nil
}

// GetGoalStatus returns the current status of a goal
func (r *GoalRepository) GetGoalStatus(ctx context.Context, goalID int64) (*models.Status, error) {
	query := `
		SELECT s.id, s.name, s.label
		FROM goals g
		INNER JOIN statuses s ON s.id = g.status_id
		WHERE g.id = $1`

	var s models.Status
	err := r.pool.QueryRow(ctx, query, goalID).Scan(&s.ID, &s.Name, &s.Label)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get goal status: %w", err)
	}
	return &s, nil
}

// UpdateGoalStatus updates a goal's status and logs the change
func (r *GoalRepository) UpdateGoalStatus(ctx context.Context, goalID int64, newStatusID int64) (*models.StatusUpdate, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// Get current status
	var currentStatusID *int64
	var fromStatusName string
	query := `SELECT status_id FROM goals WHERE id = $1`
	var currentID int64
	err = tx.QueryRow(ctx, query, goalID).Scan(&currentID)
	if err != nil {
		if err != pgx.ErrNoRows {
			return nil, fmt.Errorf("failed to get current status: %w", err)
		}
	} else {
		currentStatusID = &currentID
		// Get from status name
		query = `SELECT name FROM statuses WHERE id = $1`
		_ = tx.QueryRow(ctx, query, currentID).Scan(&fromStatusName)
	}

	// Update goal status
	query = `UPDATE goals SET status_id = $1 WHERE id = $2`
	_, err = tx.Exec(ctx, query, newStatusID, goalID)
	if err != nil {
		return nil, fmt.Errorf("failed to update goal status: %w", err)
	}

	// Get to status name
	var toStatusName string
	query = `SELECT name FROM statuses WHERE id = $1`
	_ = tx.QueryRow(ctx, query, newStatusID).Scan(&toStatusName)

	// Log status change
	query = `INSERT INTO status_updates (goal_id, from_status_id, to_status_id) VALUES ($1, $2, $3) RETURNING id, changed_at`
	var update models.StatusUpdate
	err = tx.QueryRow(ctx, query, goalID, currentStatusID, newStatusID).Scan(&update.ID, &update.ChangedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to log status update: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	update.GoalID = goalID
	update.FromStatusID = currentStatusID
	update.ToStatusID = newStatusID
	update.FromStatusName = fromStatusName
	update.ToStatusName = toStatusName

	return &update, nil
}

// GetStatusHistory returns status update history for a goal
func (r *GoalRepository) GetStatusHistory(ctx context.Context, goalID int64) ([]models.StatusUpdate, error) {
	query := `
		SELECT 
			su.id, 
			su.goal_id, 
			su.from_status_id, 
			su.to_status_id, 
			su.changed_at,
			COALESCE(fs.name, '') as from_status_name,
			ts.name as to_status_name
		FROM status_updates su
		LEFT JOIN statuses fs ON fs.id = su.from_status_id
		INNER JOIN statuses ts ON ts.id = su.to_status_id
		WHERE su.goal_id = $1
		ORDER BY su.changed_at DESC`

	rows, err := r.pool.Query(ctx, query, goalID)
	if err != nil {
		return nil, fmt.Errorf("failed to get status history: %w", err)
	}
	defer rows.Close()

	var history []models.StatusUpdate
	for rows.Next() {
		var s models.StatusUpdate
		var fromStatusName string
		err := rows.Scan(&s.ID, &s.GoalID, &s.FromStatusID, &s.ToStatusID, &s.ChangedAt, &fromStatusName, &s.ToStatusName)
		if err != nil {
			return nil, fmt.Errorf("failed to scan status update: %w", err)
		}
		s.FromStatusName = fromStatusName
		history = append(history, s)
	}

	return history, nil
}

// startOfDay gör om ett klockslag till det datum det tillhör.
func startOfDay(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)
}
