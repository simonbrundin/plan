package repository

import (
	"context"
	"fmt"
	"time"

	"plan-api/internal/database"
	"plan-api/internal/models"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

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
		SELECT g.id, g.title, g.icon, g.created, g.started, g.finished, g.inbox 
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
		err := rows.Scan(&g.ID, &g.Title, &g.Icon, &g.Created, &g.Started, &g.Finished, &g.Inbox)
		if err != nil {
			return nil, fmt.Errorf("failed to scan goal: %w", err)
		}
		goals = append(goals, g)
	}

	return goals, nil
}

// GetGoalByID returns a single goal by ID
func (r *GoalRepository) GetGoalByID(ctx context.Context, goalID int64) (*models.Goal, error) {
	query := `SELECT id, title, icon, created, started, finished, inbox FROM goals WHERE id = $1`

	var g models.Goal
	err := r.pool.QueryRow(ctx, query, goalID).Scan(
		&g.ID, &g.Title, &g.Icon, &g.Created, &g.Started, &g.Finished, &g.Inbox)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get goal: %w", err)
	}

	return &g, nil
}

// CreateGoal creates a new goal and associates it with a user
func (r *GoalRepository) CreateGoal(ctx context.Context, userID int64, title, icon string) (*models.Goal, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// Insert goal
	var goal models.Goal
	query := `INSERT INTO goals (title, icon) VALUES ($1, $2) 
			  RETURNING id, title, icon, created, started, finished, inbox`
	err = tx.QueryRow(ctx, query, title, icon).Scan(
		&goal.ID, &goal.Title, &goal.Icon, &goal.Created, &goal.Started, &goal.Finished, &goal.Inbox)
	if err != nil {
		return nil, fmt.Errorf("failed to create goal: %w", err)
	}

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

	if req.Title != nil {
		query += fmt.Sprintf("title = $%d, ", argNum)
		args = append(args, *req.Title)
		argNum++
	}
	if req.Icon != nil {
		query += fmt.Sprintf("icon = $%d, ", argNum)
		args = append(args, *req.Icon)
		argNum++
	}
	if req.Started != nil {
		if *req.Started == "" {
			query += "started = NULL, "
		} else {
			query += fmt.Sprintf("started = $%d, ", argNum)
			t, _ := time.Parse(time.RFC3339, *req.Started)
			args = append(args, t)
			argNum++
		}
	}
	if req.Finished != nil {
		if *req.Finished == "" {
			query += "finished = NULL, "
		} else {
			query += fmt.Sprintf("finished = $%d, ", argNum)
			t, _ := time.Parse(time.RFC3339, *req.Finished)
			args = append(args, t)
			argNum++
		}
	}

	if len(args) == 0 {
		return r.GetGoalByID(ctx, goalID)
	}

	// Remove trailing comma and space
	query = query[:len(query)-2]
	query += fmt.Sprintf(" WHERE id = $%d", argNum)
	args = append(args, goalID)

	_, err := r.pool.Exec(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to update goal: %w", err)
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
		SELECT g.id, g.title, g.icon, g.created, g.started, g.finished, g.inbox
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
		if err := rows.Scan(&g.ID, &g.Title, &g.Icon,
			&g.Created, &g.Started, &g.Finished, &g.Inbox); err != nil {
			return nil, fmt.Errorf("failed to scan goal: %w", err)
		}
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
		SELECT g.id, g.title, g.icon, g.created, g.started, g.finished, g.inbox
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
		if err := rows.Scan(&g.ID, &g.Title, &g.Icon, &g.Created, &g.Started, &g.Finished, &g.Inbox); err != nil {
			return nil, err
		}
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
			g.id, g.title, g.icon, g.created, g.started, g.finished, g.inbox,
			COALESCE((SELECT MAX(weight) FROM goal_relations WHERE child_id = g.id), 10) AS weight,
			(SELECT g2.title FROM goal_relations gr INNER JOIN goals g2 ON g2.id = gr.parent_id WHERE gr.child_id = g.id ORDER BY gr.weight DESC LIMIT 1) AS parent_title,
			(SELECT gr.parent_id FROM goal_relations gr WHERE gr.child_id = g.id ORDER BY gr.weight DESC LIMIT 1) AS parent_id
		FROM goals g
		INNER JOIN user_goals ug ON g.id = ug.goal_id
		WHERE ug.user_id = $1
		  AND g.finished IS NULL
		  AND g.id <> 1
		ORDER BY weight DESC, g.created DESC`

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
		if err := rows.Scan(&g.ID, &g.Title, &g.Icon, &g.Created, &g.Started, &g.Finished, &g.Inbox,
			&g.Weight, &parentTitle, &parentID); err != nil {
			return nil, err
		}
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
		fromStatusName = fromStatusName
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
