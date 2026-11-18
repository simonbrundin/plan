package repository

import (
	"context"
	"fmt"

	"plan-api/internal/database"
	"plan-api/internal/models"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// DefaultCalendarColor används när en kalender inte har någon färg satt.
const DefaultCalendarColor = "#6366f1"

type CalendarRepository struct {
	pool *pgxpool.Pool
}

func NewCalendarRepository() *CalendarRepository {
	return &CalendarRepository{pool: database.GetPool()}
}

const calendarColumns = `
	id, user_id, name, color, COALESCE(icon, ''), is_default, sort_order, created_at, updated_at`

func scanCalendar(row pgx.Row) (*models.Calendar, error) {
	var calendar models.Calendar
	err := row.Scan(
		&calendar.ID, &calendar.UserID, &calendar.Name, &calendar.Color,
		&calendar.Icon, &calendar.IsDefault, &calendar.SortOrder, &calendar.CreatedAt, &calendar.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &calendar, nil
}

// List returnerar användarens kalendrar i den ordning de visas i gränssnittet.
func (r *CalendarRepository) List(ctx context.Context, userID int64) ([]models.Calendar, error) {
	query := `SELECT ` + calendarColumns + `
		FROM calendars
		WHERE user_id = $1
		ORDER BY sort_order, id`

	rows, err := r.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to list calendars: %w", err)
	}
	defer rows.Close()

	calendars := make([]models.Calendar, 0)
	for rows.Next() {
		calendar, err := scanCalendar(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan calendar: %w", err)
		}
		calendars = append(calendars, *calendar)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate calendars: %w", err)
	}
	return calendars, nil
}

func (r *CalendarRepository) Get(ctx context.Context, userID, calendarID int64) (*models.Calendar, error) {
	query := `SELECT ` + calendarColumns + `
		FROM calendars WHERE id = $1 AND user_id = $2`
	calendar, err := scanCalendar(r.pool.QueryRow(ctx, query, calendarID, userID))
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get calendar: %w", err)
	}
	return calendar, nil
}

func (r *CalendarRepository) Create(ctx context.Context, userID int64, req models.CreateCalendarRequest) (*models.Calendar, error) {
	color := DefaultCalendarColor
	if req.Color != nil && *req.Color != "" {
		color = *req.Color
	}
	sortOrder := 0
	if req.SortOrder != nil {
		sortOrder = *req.SortOrder
	}

	query := `INSERT INTO calendars (user_id, name, color, icon, is_default, sort_order)
		VALUES ($1, $2, $3, NULLIF($4::text, ''), $5, $6)
		RETURNING ` + calendarColumns

	// En användare kan bara ha en standardkalender. När en ny kalender
	// skapas som standard avmarkeras därför de gamla i samma transaktion.
	if req.IsDefault == nil || !*req.IsDefault {
		return scanCalendar(r.pool.QueryRow(ctx, query, userID, req.Name, color, req.Icon, false, sortOrder))
	}

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, `UPDATE calendars SET is_default = FALSE WHERE user_id = $1 AND is_default`, userID); err != nil {
		return nil, fmt.Errorf("failed to clear default calendar: %w", err)
	}
	calendar, err := scanCalendar(tx.QueryRow(ctx, query, userID, req.Name, color, req.Icon, true, sortOrder))
	if err != nil {
		return nil, fmt.Errorf("failed to create calendar: %w", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}
	return calendar, nil
}

func (r *CalendarRepository) Update(ctx context.Context, userID, calendarID int64, req models.UpdateCalendarRequest) (*models.Calendar, error) {
	query := `UPDATE calendars SET
		name = COALESCE($3, name),
		color = CASE WHEN $4::text IS NULL THEN color WHEN $4::text = '' THEN $5 ELSE $4::text END,
		icon = CASE WHEN $6::text IS NULL THEN icon WHEN $6::text = '' THEN NULL ELSE $6::text END,
		is_default = COALESCE($8::boolean, is_default),
		sort_order = COALESCE($7, sort_order),
		updated_at = CURRENT_TIMESTAMP
		WHERE id = $1 AND user_id = $2
		RETURNING ` + calendarColumns

	// Att sätta en kalender som standard avmarkerar den tidigare
	// standardkalendern. Det sker i en transaktion så att det aldrig kan
	// finnas två standardkalendrar samtidigt.
	if req.IsDefault != nil && *req.IsDefault {
		tx, err := r.pool.Begin(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to begin transaction: %w", err)
		}
		defer tx.Rollback(ctx)

		if _, err := tx.Exec(ctx, `UPDATE calendars SET is_default = FALSE WHERE user_id = $1 AND is_default AND id <> $2`, userID, calendarID); err != nil {
			return nil, fmt.Errorf("failed to clear default calendar: %w", err)
		}
		calendar, err := scanCalendar(tx.QueryRow(ctx, query, calendarID, userID, req.Name, req.Color, DefaultCalendarColor, req.Icon, req.SortOrder, true))
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		if err != nil {
			return nil, fmt.Errorf("failed to update calendar: %w", err)
		}
		if err := tx.Commit(ctx); err != nil {
			return nil, fmt.Errorf("failed to commit transaction: %w", err)
		}
		return calendar, nil
	}

	calendar, err := scanCalendar(r.pool.QueryRow(ctx, query, calendarID, userID, req.Name, req.Color, DefaultCalendarColor, req.Icon, req.SortOrder, req.IsDefault))
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to update calendar: %w", err)
	}
	return calendar, nil
}

// DefaultCalendarID returnerar id:t för användarens standardkalender, eller
// nil om användaren inte har satt någon.
func (r *CalendarRepository) DefaultCalendarID(ctx context.Context, userID int64) (*int64, error) {
	var id int64
	err := r.pool.QueryRow(ctx, `SELECT id FROM calendars WHERE user_id = $1 AND is_default LIMIT 1`, userID).Scan(&id)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to find default calendar: %w", err)
	}
	return &id, nil
}

// Delete tar bort en kalender. Händelserna som legat i den behålls men
// blir kalenderlösa, eftersom kopplingen är ON DELETE SET NULL.
func (r *CalendarRepository) Delete(ctx context.Context, userID, calendarID int64) (bool, error) {
	result, err := r.pool.Exec(ctx, `DELETE FROM calendars WHERE id = $1 AND user_id = $2`, calendarID, userID)
	if err != nil {
		return false, fmt.Errorf("failed to delete calendar: %w", err)
	}
	return result.RowsAffected() > 0, nil
}
