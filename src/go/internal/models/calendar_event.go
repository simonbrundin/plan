package models

import "time"

// CalendarEvent represents a user-owned calendar event.
type CalendarEvent struct {
	ID           int64  `json:"id"`
	OccurrenceID string `json:"occurrence_id,omitempty"`
	UserID       int64  `json:"user_id"`
	GoalID       *int64 `json:"goal_id,omitempty"`
	// CalendarID pekar på den kalender händelsen ligger i. Nil betyder att
	// händelsen inte tillhör någon kalender.
	CalendarID  *int64 `json:"calendar_id,omitempty"`
	Title       string `json:"title"`
	Description string `json:"description"`
	// Icon är en emoji (t.ex. "🏃") eller ett Iconify-namn (t.ex.
	// "heroicons:star"). Det är valfritt och används bara för visning.
	Icon            string     `json:"icon,omitempty"`
	Color           string     `json:"color,omitempty"`
	StartAt         time.Time  `json:"start_at"`
	EndAt           time.Time  `json:"end_at"`
	TimeZone        string     `json:"time_zone"`
	AllDay          bool       `json:"all_day"`
	Status          string     `json:"status"`
	RecurrenceRule  *string    `json:"recurrence_rule,omitempty"`
	ParentEventID   *int64     `json:"parent_event_id,omitempty"`
	OriginalStartAt *time.Time `json:"original_start_at,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

type CreateCalendarEventRequest struct {
	Title          string  `json:"title" binding:"required"`
	Description    string  `json:"description"`
	Icon           *string `json:"icon"`
	Color          *string `json:"color"`
	StartAt        string  `json:"start_at" binding:"required"`
	EndAt          string  `json:"end_at" binding:"required"`
	TimeZone       string  `json:"time_zone"`
	AllDay         bool    `json:"all_day"`
	Status         string  `json:"status"`
	RecurrenceRule *string `json:"recurrence_rule"`
	GoalID         *int64  `json:"goal_id"`
	// CalendarID anger vilken kalender händelsen ska ligga i. Nil lämnar
	// händelsen utan kalender.
	CalendarID *int64 `json:"calendar_id"`
}

type UpdateCalendarEventRequest struct {
	Title          *string `json:"title"`
	Description    *string `json:"description"`
	Icon           *string `json:"icon"`
	Color          *string `json:"color"`
	StartAt        *string `json:"start_at"`
	EndAt          *string `json:"end_at"`
	TimeZone       *string `json:"time_zone"`
	AllDay         *bool   `json:"all_day"`
	Status         *string `json:"status"`
	RecurrenceRule *string `json:"recurrence_rule"`
	GoalID         *int64  `json:"goal_id"`
	// CalendarID anger vilken kalender händelsen ska ligga i. Nil behåller
	// händelsens nuvarande kalender, 0 tar bort den.
	CalendarID *int64 `json:"calendar_id"`
}

// CreateCalendarOccurrenceRequest creates an override for one occurrence in a series.
type CreateCalendarOccurrenceRequest struct {
	OriginalStartAt string  `json:"original_start_at" binding:"required"`
	StartAt         string  `json:"start_at" binding:"required"`
	EndAt           string  `json:"end_at" binding:"required"`
	Title           string  `json:"title"`
	Description     string  `json:"description"`
	Icon            *string `json:"icon"`
	Color           *string `json:"color"`
	TimeZone        string  `json:"time_zone"`
	AllDay          *bool   `json:"all_day"`
	Status          string  `json:"status"`
	GoalID          *int64  `json:"goal_id"`
	CalendarID      *int64  `json:"calendar_id"`
}

type SplitCalendarEventRequest struct {
	OriginalStartAt string  `json:"original_start_at" binding:"required"`
	StartAt         string  `json:"start_at" binding:"required"`
	EndAt           string  `json:"end_at" binding:"required"`
	RecurrenceRule  string  `json:"recurrence_rule" binding:"required"`
	Title           *string `json:"title"`
	Description     *string `json:"description"`
	Icon            *string `json:"icon"`
	Color           *string `json:"color"`
	TimeZone        *string `json:"time_zone"`
	AllDay          *bool   `json:"all_day"`
	GoalID          *int64  `json:"goal_id"`
	CalendarID      *int64  `json:"calendar_id"`
}
