package repository

import (
	"context"
	"fmt"
	"time"

	"plan-api/internal/database"
	"plan-api/internal/models"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/teambition/rrule-go"
)

type CalendarEventRepository struct {
	pool *pgxpool.Pool
}

func NewCalendarEventRepository() *CalendarEventRepository {
	return &CalendarEventRepository{pool: database.GetPool()}
}

const calendarEventColumns = `
	id, user_id, goal_id, calendar_id, title, COALESCE(description, ''), COALESCE(icon, ''),
	start_at, end_at, time_zone, all_day, status, recurrence_rule,
	parent_event_id, original_start_at, created_at, updated_at`

func scanCalendarEvent(row pgx.Row) (*models.CalendarEvent, error) {
	var event models.CalendarEvent
	err := row.Scan(
		&event.ID, &event.UserID, &event.GoalID, &event.CalendarID, &event.Title, &event.Description,
		&event.Icon, &event.StartAt, &event.EndAt, &event.TimeZone, &event.AllDay,
		&event.Status, &event.RecurrenceRule, &event.ParentEventID,
		&event.OriginalStartAt, &event.CreatedAt, &event.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &event, nil
}

// Ikonfältet har tre lägen: nil betyder "sätt inte" (behåll nuvarande),
// en tom sträng betyder "töm" och ett värde betyder "sätt". Det görs
// möjligt att både lägga till och ta bort en ikon via PATCH.
func expandCalendarEvent(event models.CalendarEvent, from, to time.Time) ([]models.CalendarEvent, error) {
	if event.RecurrenceRule == nil || *event.RecurrenceRule == "" {
		return []models.CalendarEvent{event}, nil
	}

	rule, err := rrule.StrToRRule(*event.RecurrenceRule)
	if err != nil {
		return nil, fmt.Errorf("invalid recurrence rule for event %d: %w", event.ID, err)
	}
	rule.DTStart(event.StartAt)
	duration := event.EndAt.Sub(event.StartAt)
	occurrences := rule.Between(from.Add(-duration), to, true)
	if len(occurrences) > 1000 {
		occurrences = occurrences[:1000]
	}

	expanded := make([]models.CalendarEvent, 0, len(occurrences))
	for _, occurrenceStart := range occurrences {
		occurrenceEnd := occurrenceStart.Add(duration)
		if !occurrenceStart.Before(to) || !occurrenceEnd.After(from) {
			continue
		}
		instance := event
		instance.OccurrenceID = fmt.Sprintf("%d-%d", event.ID, occurrenceStart.UnixNano())
		instance.OriginalStartAt = &occurrenceStart
		instance.StartAt = occurrenceStart
		instance.EndAt = occurrenceEnd
		expanded = append(expanded, instance)
	}
	return expanded, nil
}

// resolveIcon väljer vilken ikon en händelse ska få.
//
// Om begäran inte innehåller någon ikon ärvs förälderns ikon, vilket
// gör att en override eller en delad serie inte tappar ikon av sig.
// Ett medvetet tomt värde däremot tömmer ikonen.
func resolveIcon(parentIcon string, requested *string) *string {
	if requested != nil {
		return requested
	}
	if parentIcon == "" {
		return nil
	}
	inherited := parentIcon
	return &inherited
}

func (r *CalendarEventRepository) List(ctx context.Context, userID int64, from, to time.Time) ([]models.CalendarEvent, error) {
	query := `SELECT ` + calendarEventColumns + `
		FROM calendar_events
		WHERE user_id = $1
		  AND (status <> 'cancelled' OR parent_event_id IS NOT NULL)
		  AND ((start_at < $3 AND end_at > $2)
		    OR recurrence_rule IS NOT NULL
		    OR (original_start_at >= $2 AND original_start_at < $3))
		ORDER BY start_at, id`

	rows, err := r.pool.Query(ctx, query, userID, from, to)
	if err != nil {
		return nil, fmt.Errorf("failed to list calendar events: %w", err)
	}
	defer rows.Close()

	masters := make([]models.CalendarEvent, 0)
	overrides := make(map[string]models.CalendarEvent)
	for rows.Next() {
		event, err := scanCalendarEvent(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan calendar event: %w", err)
		}
		if event.ParentEventID != nil && event.OriginalStartAt != nil {
			key := fmt.Sprintf("%d:%s", *event.ParentEventID, event.OriginalStartAt.UTC().Format(time.RFC3339Nano))
			overrides[key] = *event
			continue
		}
		masters = append(masters, *event)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate calendar events: %w", err)
	}

	events := make([]models.CalendarEvent, 0)
	for _, event := range masters {
		expanded, err := expandCalendarEvent(event, from, to)
		if err != nil {
			return nil, err
		}
		for _, occurrence := range expanded {
			key := fmt.Sprintf("%d:%s", event.ID, occurrence.StartAt.UTC().Format(time.RFC3339Nano))
			if override, ok := overrides[key]; ok {
				override.OccurrenceID = fmt.Sprintf("%d-%d", event.ID, occurrence.StartAt.UnixNano())
				if override.Status != "cancelled" {
					events = append(events, override)
				}
				delete(overrides, key)
				continue
			}
			events = append(events, occurrence)
		}
	}
	for _, override := range overrides {
		if override.StartAt.Before(to) && override.EndAt.After(from) && override.Status != "cancelled" {
			override.OccurrenceID = fmt.Sprintf("%d-%d", override.ID, override.StartAt.UnixNano())
			events = append(events, override)
		}
	}
	return events, nil
}

// ListAll returns every event the user owns, without expanding recurring
// series.
//
// The window-based List expands recurring series into occurrences, which
// means the client has to ask again every time it moves the time window.
// ListAll instead returns the raw rows: one row per series (with its
// recurrence rule) and one row per occurrence override. The client can then
// expand the series locally and answer "what is visible right now" without
// touching the server.
func (r *CalendarEventRepository) ListAll(ctx context.Context, userID int64) ([]models.CalendarEvent, error) {
	query := `SELECT ` + calendarEventColumns + `
		FROM calendar_events
		WHERE user_id = $1
		  AND (status <> 'cancelled' OR parent_event_id IS NOT NULL)
		ORDER BY start_at, id`

	rows, err := r.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to list calendar events: %w", err)
	}
	defer rows.Close()

	events := make([]models.CalendarEvent, 0)
	for rows.Next() {
		event, err := scanCalendarEvent(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan calendar event: %w", err)
		}
		events = append(events, *event)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate calendar events: %w", err)
	}
	return events, nil
}

func (r *CalendarEventRepository) Get(ctx context.Context, userID, eventID int64) (*models.CalendarEvent, error) {
	query := `SELECT ` + calendarEventColumns + `
		FROM calendar_events WHERE id = $1 AND user_id = $2`
	event, err := scanCalendarEvent(r.pool.QueryRow(ctx, query, eventID, userID))
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get calendar event: %w", err)
	}
	return event, nil
}

func (r *CalendarEventRepository) Create(ctx context.Context, userID int64, req models.CreateCalendarEventRequest, startAt, endAt time.Time) (*models.CalendarEvent, error) {
	status := req.Status
	if status == "" {
		status = "confirmed"
	}
	timeZone := req.TimeZone
	if timeZone == "" {
		timeZone = "UTC"
	}

	query := `INSERT INTO calendar_events
		(user_id, goal_id, calendar_id, title, description, icon, start_at, end_at, time_zone, all_day, status, recurrence_rule)
		VALUES ($1, $2, $3, $4, $5, NULLIF($6::text, ''), $7, $8, $9, $10, $11, $12)
		RETURNING ` + calendarEventColumns
	return scanCalendarEvent(r.pool.QueryRow(ctx, query, userID, req.GoalID, req.CalendarID, req.Title, req.Description, req.Icon, startAt, endAt, timeZone, req.AllDay, status, req.RecurrenceRule))
}

func (r *CalendarEventRepository) Update(ctx context.Context, userID, eventID int64, req models.UpdateCalendarEventRequest, startAt, endAt *time.Time) (*models.CalendarEvent, error) {
	// calendar_id har samma tre-state-semantik som ikon: nil betyder "behåll
	// nuvarande kalender" och 0 betyder "ta bort kalendern".

	query := `UPDATE calendar_events SET
		goal_id = COALESCE($3, goal_id),
		title = COALESCE($4, title),
		description = COALESCE($5, description),
		icon = CASE WHEN $12::text IS NULL THEN icon WHEN $12::text = '' THEN NULL ELSE $12::text END,
		start_at = COALESCE($6, start_at),
		end_at = COALESCE($7, end_at),
		time_zone = COALESCE($8, time_zone),
		all_day = COALESCE($9, all_day),
		status = COALESCE($10, status),
		recurrence_rule = COALESCE($11, recurrence_rule),
		calendar_id = CASE WHEN $13::bigint IS NULL THEN calendar_id ELSE NULLIF($13::bigint, 0) END,
		updated_at = CURRENT_TIMESTAMP
		WHERE id = $1 AND user_id = $2
		RETURNING ` + calendarEventColumns

	event, err := scanCalendarEvent(r.pool.QueryRow(ctx, query, eventID, userID, req.GoalID, req.Title, req.Description, startAt, endAt, req.TimeZone, req.AllDay, req.Status, req.RecurrenceRule, req.Icon, req.CalendarID))
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to update calendar event: %w", err)
	}
	return event, nil
}

func (r *CalendarEventRepository) CreateOverride(ctx context.Context, userID, parentEventID int64, req models.CreateCalendarOccurrenceRequest, originalStart, startAt, endAt time.Time) (*models.CalendarEvent, error) {
	parent, err := r.Get(ctx, userID, parentEventID)
	if err != nil {
		return nil, err
	}
	if parent == nil || parent.RecurrenceRule == nil {
		return nil, pgx.ErrNoRows
	}

	title := req.Title
	if title == "" {
		title = parent.Title
	}
	description := req.Description
	if description == "" {
		description = parent.Description
	}
	timeZone := req.TimeZone
	if timeZone == "" {
		timeZone = parent.TimeZone
	}
	allDay := parent.AllDay
	if req.AllDay != nil {
		allDay = *req.AllDay
	}
	status := req.Status
	if status == "" {
		status = "confirmed"
	}
	goalID := req.GoalID
	if goalID == nil {
		goalID = parent.GoalID
	}
	calendarID := req.CalendarID
	if calendarID == nil {
		calendarID = parent.CalendarID
	}
	if calendarID != nil && *calendarID == 0 {
		calendarID = nil
	}
	icon := resolveIcon(parent.Icon, req.Icon)

	query := `INSERT INTO calendar_events
		(user_id, goal_id, calendar_id, title, description, icon, start_at, end_at, time_zone, all_day, status, parent_event_id, original_start_at)
		VALUES ($1, $2, $3, $4, $5, NULLIF($6::text, ''), $7, $8, $9, $10, $11, $12, $13)
		ON CONFLICT DO NOTHING
		RETURNING ` + calendarEventColumns
	return scanCalendarEvent(r.pool.QueryRow(ctx, query, userID, goalID, calendarID, title, description, icon, startAt, endAt, timeZone, allDay, status, parentEventID, originalStart))
}

func (r *CalendarEventRepository) SplitSeries(ctx context.Context, userID, parentEventID int64, req models.SplitCalendarEventRequest, originalStart, startAt, endAt time.Time) (*models.CalendarEvent, error) {
	parent, err := r.Get(ctx, userID, parentEventID)
	if err != nil {
		return nil, err
	}
	if parent == nil || parent.RecurrenceRule == nil {
		return nil, pgx.ErrNoRows
	}
	newRule, err := rrule.StrToRRule(req.RecurrenceRule)
	if err != nil {
		return nil, err
	}
	_ = newRule
	oldOptions, err := rrule.StrToROption(*parent.RecurrenceRule)
	if err != nil {
		return nil, err
	}
	oldOptions.Count = 0
	oldOptions.Until = originalStart.Add(-time.Nanosecond)
	truncatedRule := oldOptions.RRuleString()

	title := parent.Title
	if req.Title != nil {
		title = *req.Title
	}
	description := parent.Description
	if req.Description != nil {
		description = *req.Description
	}
	timeZone := parent.TimeZone
	if req.TimeZone != nil {
		timeZone = *req.TimeZone
	}
	allDay := parent.AllDay
	if req.AllDay != nil {
		allDay = *req.AllDay
	}
	goalID := parent.GoalID
	if req.GoalID != nil {
		goalID = req.GoalID
	}
	calendarID := parent.CalendarID
	if req.CalendarID != nil {
		calendarID = req.CalendarID
	}
	if calendarID != nil && *calendarID == 0 {
		calendarID = nil
	}
	icon := resolveIcon(parent.Icon, req.Icon)

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)
	if _, err := tx.Exec(ctx, `UPDATE calendar_events SET recurrence_rule = $1, updated_at = CURRENT_TIMESTAMP WHERE id = $2 AND user_id = $3`, truncatedRule, parentEventID, userID); err != nil {
		return nil, err
	}
	query := `INSERT INTO calendar_events
		(user_id, goal_id, calendar_id, title, description, icon, start_at, end_at, time_zone, all_day, status, recurrence_rule)
		VALUES ($1, $2, $3, $4, $5, NULLIF($6::text, ''), $7, $8, $9, $10, $11, $12)
		RETURNING ` + calendarEventColumns
	created, err := scanCalendarEvent(tx.QueryRow(ctx, query, userID, goalID, calendarID, title, description, icon, startAt, endAt, timeZone, allDay, parent.Status, req.RecurrenceRule))
	if err != nil {
		return nil, err
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return created, nil
}

func (r *CalendarEventRepository) UserHasGoal(ctx context.Context, userID, goalID int64) (bool, error) {
	var exists bool
	err := r.pool.QueryRow(ctx, `SELECT EXISTS (SELECT 1 FROM user_goals WHERE user_id = $1 AND goal_id = $2)`, userID, goalID).Scan(&exists)
	return exists, err
}

// UserHasCalendar kontrollerar att kalendern finns och tillhör användaren.
func (r *CalendarEventRepository) UserHasCalendar(ctx context.Context, userID, calendarID int64) (bool, error) {
	var exists bool
	err := r.pool.QueryRow(ctx, `SELECT EXISTS (SELECT 1 FROM calendars WHERE id = $1 AND user_id = $2)`, calendarID, userID).Scan(&exists)
	return exists, err
}

func (r *CalendarEventRepository) Delete(ctx context.Context, userID, eventID int64) (bool, error) {
	result, err := r.pool.Exec(ctx, `DELETE FROM calendar_events WHERE id = $1 AND user_id = $2`, eventID, userID)
	if err != nil {
		return false, fmt.Errorf("failed to delete calendar event: %w", err)
	}
	return result.RowsAffected() > 0, nil
}
