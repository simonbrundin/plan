package models

import "time"

// Calendar är en användares egen kalender. Användaren kan skapa hur många
// kalendrar som helst, till exempel en jobb- och en fritidskalender, och
// varje kalender har en egen färg.
type Calendar struct {
	ID        int64     `json:"id"`
	UserID    int64     `json:"user_id"`
	Name      string    `json:"name"`
	Color     string    `json:"color"`
	Icon      string    `json:"icon,omitempty"`
	SortOrder int       `json:"sort_order"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type CreateCalendarRequest struct {
	Name      string  `json:"name" binding:"required"`
	Color     *string `json:"color"`
	Icon      *string `json:"icon"`
	SortOrder *int    `json:"sort_order"`
}

type UpdateCalendarRequest struct {
	// Name sätts när den skickas. Ett tomt värde är ogiltigt.
	Name *string `json:"name"`
	// Color har samma tre-state-semantik som på händelser: nil betyder
	// "behåll nuvarande värde", en tom sträng "töm" och ett värde "sätt".
	Color *string `json:"color"`
	Icon  *string `json:"icon"`
	// SortOrder sätts när den skickas.
	SortOrder *int `json:"sort_order"`
}
