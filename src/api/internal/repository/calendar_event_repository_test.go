package repository

import (
	"testing"
	"time"

	"plan/api/internal/models"
)

func TestExpandCalendarEvent(t *testing.T) {
	start := time.Date(2026, 1, 5, 9, 0, 0, 0, time.UTC)
	rule := "FREQ=WEEKLY;COUNT=3"
	event := models.CalendarEvent{
		ID:             7,
		Title:          "Veckomöte",
		StartAt:        start,
		EndAt:          start.Add(time.Hour),
		RecurrenceRule: &rule,
	}

	events, err := expandCalendarEvent(event, start, start.Add(15*24*time.Hour))
	if err != nil {
		t.Fatalf("expandCalendarEvent() error = %v", err)
	}
	if len(events) != 3 {
		t.Fatalf("expandCalendarEvent() returned %d events, want 3", len(events))
	}
	if !events[1].StartAt.Equal(start.Add(7 * 24 * time.Hour)) {
		t.Fatalf("second occurrence starts at %s, want one week later", events[1].StartAt)
	}
	if events[0].OccurrenceID == "" || events[0].OccurrenceID == events[1].OccurrenceID {
		t.Fatal("occurrences should have distinct occurrence IDs")
	}
}

func TestResolveIcon(t *testing.T) {
	stringPtr := func(value string) *string { return &value }

	tests := []struct {
		name       string
		parentIcon string
		requested  *string
		wantNil    bool
		wantIcon   string
	}{
		{
			name:       "ingen ikon och ingen förälderikon",
			parentIcon: "",
			requested:  nil,
			wantNil:    true,
		},
		{
			name:       "ingen ikon ärver förälderns ikon",
			parentIcon: "🏃",
			requested:  nil,
			wantIcon:   "🏃",
		},
		{
			name:       "en tom ikon tömmer ikonen",
			parentIcon: "🏃",
			requested:  stringPtr(""),
			wantIcon:   "",
		},
		{
			name:       "en egen ikon används",
			parentIcon: "🏃",
			requested:  stringPtr("heroicons:star"),
			wantIcon:   "heroicons:star",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			icon := resolveIcon(test.parentIcon, test.requested)

			if test.wantNil {
				if icon != nil {
					t.Fatalf("resolveIcon() = %q, want nil", *icon)
				}
				return
			}

			if icon == nil {
				t.Fatal("resolveIcon() = nil, want an icon")
			}
			if *icon != test.wantIcon {
				t.Fatalf("resolveIcon() = %q, want %q", *icon, test.wantIcon)
			}
		})
	}
}
