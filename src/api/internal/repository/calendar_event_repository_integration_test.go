package repository

import (
	"context"
	"os"
	"testing"
	"time"

	"plan/api/internal/database"
	"plan/api/internal/models"
)

func TestIconFlowIntegration(t *testing.T) {
	if os.Getenv("DATABASE_URL") == "" {
		t.Skip("no DATABASE_URL")
	}
	ctx := context.Background()
	if err := database.Connect(ctx); err != nil {
		t.Fatal(err)
	}
	defer database.Close()

	repo := NewCalendarEventRepository()

	// Hitta en användare
	var userID int64
	if err := database.GetPool().QueryRow(ctx, `SELECT id FROM users LIMIT 1`).Scan(&userID); err != nil {
		t.Skip("no users")
	}

	start := time.Now().Add(time.Hour)
	end := start.Add(time.Hour)

	// 1. Skapa med emoji
	// Endast testens egna rader får raderas - aldrig hela användarens kalender.
	icon := "🏃"
	created, err := repo.Create(ctx, userID, models.CreateCalendarEventRequest{
		Title: "[test] Löpning", Icon: &icon, StartAt: start.Format(time.RFC3339), EndAt: end.Format(time.RFC3339),
	}, start, end)
	if err != nil {
		t.Fatal("create:", err)
	}
	defer func() {
		_, _ = database.GetPool().Exec(ctx, `DELETE FROM calendar_events WHERE id = $1 OR parent_event_id = $1`, created.ID)
	}()
	t.Log("created icon:", created.Icon)
	if created.Icon != "🏃" {
		t.Fatalf("created icon = %q", created.Icon)
	}

	// 2. Uppdatera till Iconify-namn
	newIcon := "heroicons:star"
	updated, err := repo.Update(ctx, userID, created.ID, models.UpdateCalendarEventRequest{Icon: &newIcon}, nil, nil)
	if err != nil {
		t.Fatal("update:", err)
	}
	t.Log("updated icon:", updated.Icon)
	if updated.Icon != "heroicons:star" {
		t.Fatalf("updated icon = %q", updated.Icon)
	}

	// 3. PATCH utan icon -> behålls
	noIcon := (*string)(nil)
	kept, err := repo.Update(ctx, userID, created.ID, models.UpdateCalendarEventRequest{Icon: noIcon}, nil, nil)
	if err != nil {
		t.Fatal("keep:", err)
	}
	t.Log("kept icon:", kept.Icon)
	if kept.Icon != "heroicons:star" {
		t.Fatalf("kept icon = %q", kept.Icon)
	}

	// 4. PATCH med tom icon -> töms
	empty := ""
	cleared, err := repo.Update(ctx, userID, created.ID, models.UpdateCalendarEventRequest{Icon: &empty}, nil, nil)
	if err != nil {
		t.Fatal("clear:", err)
	}
	t.Log("cleared icon:", cleared.Icon)
	if cleared.Icon != "" {
		t.Fatalf("cleared icon = %q", cleared.Icon)
	}

	// 5. Återkommande serie + override som ärver ikonen
	seriesIcon := "🎉"
	rule := "FREQ=WEEKLY;COUNT=3"
	series, err := repo.Create(ctx, userID, models.CreateCalendarEventRequest{
		Title: "[test] Fest", Icon: &seriesIcon, StartAt: start.Format(time.RFC3339), EndAt: end.Format(time.RFC3339),
		RecurrenceRule: &rule,
	}, start, end)
	if err != nil {
		t.Fatal("series:", err)
	}
	defer func() {
		_, _ = database.GetPool().Exec(ctx, `DELETE FROM calendar_events WHERE id = $1 OR parent_event_id = $1`, series.ID)
	}()

	original := start
	occStart := start.Add(7 * 24 * time.Hour)
	occEnd := occStart.Add(time.Hour)
	override, err := repo.CreateOverride(ctx, userID, series.ID, models.CreateCalendarOccurrenceRequest{
		OriginalStartAt: original.Format(time.RFC3339),
		StartAt:         occStart.Format(time.RFC3339),
		EndAt:           occEnd.Format(time.RFC3339),
	}, original, occStart, occEnd)
	if err != nil {
		t.Fatal("override:", err)
	}
	t.Log("override icon (ska ärva 🎉):", override.Icon)
	if override.Icon != "🎉" {
		t.Fatalf("override icon = %q, want inherited", override.Icon)
	}

	// 6. Split - ny serie ärver ikonen
	split, err := repo.SplitSeries(ctx, userID, series.ID, models.SplitCalendarEventRequest{
		OriginalStartAt: original.Format(time.RFC3339),
		StartAt:         occStart.Format(time.RFC3339),
		EndAt:           occEnd.Format(time.RFC3339),
		RecurrenceRule:  "FREQ=WEEKLY;COUNT=2",
	}, original, occStart, occEnd)
	if err != nil {
		t.Fatal("split:", err)
	}
	defer func() {
		_, _ = database.GetPool().Exec(ctx, `DELETE FROM calendar_events WHERE id = $1 OR parent_event_id = $1`, split.ID)
	}()
	t.Log("split icon (ska ärva 🎉):", split.Icon)
	if split.Icon != "🎉" {
		t.Fatalf("split icon = %q, want inherited", split.Icon)
	}
}
