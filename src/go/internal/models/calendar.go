package models

import "time"

// Calendar är en användares egen kalender. Användaren kan skapa hur många
// kalendrar som helst, till exempel en jobb- och en fritidskalender, och
// varje kalender har en egen färg.
type Calendar struct {
	ID     int64  `json:"id"`
	UserID int64  `json:"user_id"`
	Name   string `json:"name"`
	Color  string `json:"color"`
	Icon   string `json:"icon,omitempty"`
	// IsDefault gör kalendern till användarens standardkalender. Nya
	// händelser hamnar där om ingen kalender väljs uttryckligen.
	IsDefault bool      `json:"is_default"`
	SortOrder int       `json:"sort_order"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type CreateCalendarRequest struct {
	Name      string  `json:"name" binding:"required"`
	Color     *string `json:"color"`
	Icon      *string `json:"icon"`
	SortOrder *int    `json:"sort_order"`
	// IsDefault sätts när kalendern ska vara standardkalender.
	IsDefault *bool `json:"is_default"`
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
	// IsDefault har tre-state-semantik: nil betyder "behåll nuvarande
	// värde", true "gör denna till standardkalender" och false "avmarkera
	// den som standardkalender".
	IsDefault *bool `json:"is_default"`
}
