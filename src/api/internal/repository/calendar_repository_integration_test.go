package repository

import (
	"context"
	"os"
	"testing"
	"time"

	"plan/api/internal/database"
	"plan/api/internal/models"
)

// TestCalendarFlowIntegration kontrollerar hela kedjan mellan kalendrar
// och händelser: en händelse läggs i en kalender, en återkommande förekomst
// ärver kalendern, kalendern kan tas bort och händelsen lever vidare utan
// kalender.
//
// Testet raderar bara de rader det själv skapar - användarens övriga data
// lämnas orörd.
func TestCalendarFlowIntegration(t *testing.T) {
	if os.Getenv("DATABASE_URL") == "" {
		t.Skip("no DATABASE_URL")
	}
	ctx := context.Background()
	if err := database.Connect(ctx); err != nil {
		t.Fatal(err)
	}
	defer database.Close()

	calendars := NewCalendarRepository()
	events := NewCalendarEventRepository()

	var userID int64
	if err := database.GetPool().QueryRow(ctx, `SELECT id FROM users LIMIT 1`).Scan(&userID); err != nil {
		t.Skip("no users")
	}

	start := time.Now().Add(time.Hour)
	end := start.Add(time.Hour)

	// 1. Skapa en kalender
	calendar, err := calendars.Create(ctx, userID, models.CreateCalendarRequest{
		Name:  "[test] Kalender",
		Color: strPtr("#ef4444"),
	})
	if err != nil {
		t.Fatal("create calendar:", err)
	}
	defer func() { _, _ = database.GetPool().Exec(ctx, `DELETE FROM calendars WHERE id = $1`, calendar.ID) }()

	if calendar.Color != "#ef4444" {
		t.Fatalf("calendar color = %q, want #ef4444", calendar.Color)
	}

	// 2. En händelse i kalendern
	created, err := events.Create(ctx, userID, models.CreateCalendarEventRequest{
		Title:      "[test] Händelse",
		StartAt:    start.Format(time.RFC3339),
		EndAt:      end.Format(time.RFC3339),
		CalendarID: &calendar.ID,
	}, start, end)
	if err != nil {
		t.Fatal("create event:", err)
	}
	defer func() { _, _ = database.GetPool().Exec(ctx, `DELETE FROM calendar_events WHERE id = $1`, created.ID) }()

	if created.CalendarID == nil || *created.CalendarID != calendar.ID {
		t.Fatalf("event calendar = %v, want %d", created.CalendarID, calendar.ID)
	}

	// 3. Flytta händelsen till en annan kalender
	other, err := calendars.Create(ctx, userID, models.CreateCalendarRequest{Name: "[test] Annan"})
	if err != nil {
		t.Fatal("create second calendar:", err)
	}
	defer func() { _, _ = database.GetPool().Exec(ctx, `DELETE FROM calendars WHERE id = $1`, other.ID) }()

	moved, err := events.Update(ctx, userID, created.ID, models.UpdateCalendarEventRequest{
		CalendarID: &other.ID,
	}, nil, nil)
	if err != nil {
		t.Fatal("move event:", err)
	}
	if moved.CalendarID == nil || *moved.CalendarID != other.ID {
		t.Fatalf("moved event calendar = %v, want %d", moved.CalendarID, other.ID)
	}

	// 4. Ta bort kalenderkopplingen (0 = ingen kalender)
	detached, err := events.Update(ctx, userID, created.ID, models.UpdateCalendarEventRequest{
		CalendarID: int64Ptr(0),
	}, nil, nil)
	if err != nil {
		t.Fatal("detach event:", err)
	}
	if detached.CalendarID != nil {
		t.Fatalf("detached event calendar = %v, want nil", *detached.CalendarID)
	}

	// 5. En återkommande serie ärver kalendern i sina förekomster
	rule := "FREQ=WEEKLY;COUNT=2"
	series, err := events.Create(ctx, userID, models.CreateCalendarEventRequest{
		Title:          "[test] Serie",
		StartAt:        start.Format(time.RFC3339),
		EndAt:          end.Format(time.RFC3339),
		RecurrenceRule: &rule,
		CalendarID:     &other.ID,
	}, start, end)
	if err != nil {
		t.Fatal("create series:", err)
	}
	defer func() {
		_, _ = database.GetPool().Exec(ctx, `DELETE FROM calendar_events WHERE id = $1 OR parent_event_id = $1`, series.ID)
	}()

	occStart := start.Add(7 * 24 * time.Hour)
	occEnd := occStart.Add(time.Hour)
	override, err := events.CreateOverride(ctx, userID, series.ID, models.CreateCalendarOccurrenceRequest{
		OriginalStartAt: start.Format(time.RFC3339),
		StartAt:         occStart.Format(time.RFC3339),
		EndAt:           occEnd.Format(time.RFC3339),
	}, start, occStart, occEnd)
	if err != nil {
		t.Fatal("override:", err)
	}
	if override.CalendarID == nil || *override.CalendarID != other.ID {
		t.Fatalf("override calendar = %v, want inherited %d", override.CalendarID, other.ID)
	}

	// 6. När kalendern tas bort behålls händelserna, men utan kalender
	if _, err := calendars.Delete(ctx, userID, other.ID); err != nil {
		t.Fatal("delete calendar:", err)
	}
	remaining, err := events.Get(ctx, userID, series.ID)
	if err != nil {
		t.Fatal("get series:", err)
	}
	if remaining == nil {
		t.Fatal("series should survive the calendar being deleted")
	}
	if remaining.CalendarID != nil {
		t.Fatalf("series calendar = %v, want nil after delete", *remaining.CalendarID)
	}
}

// TestDefaultCalendarIntegration kontrollerar att en användare bara kan ha
// en standardkalender, och att den byts ut när en ny sätts.
func TestDefaultCalendarIntegration(t *testing.T) {
	if os.Getenv("DATABASE_URL") == "" {
		t.Skip("no DATABASE_URL")
	}
	ctx := context.Background()
	if err := database.Connect(ctx); err != nil {
		t.Fatal(err)
	}
	defer database.Close()

	calendars := NewCalendarRepository()

	var userID int64
	if err := database.GetPool().QueryRow(ctx, `SELECT id FROM users LIMIT 1`).Scan(&userID); err != nil {
		t.Skip("no users")
	}

	// Rensa bort standardkalendern från tidigare körningar, så att testet
	// inte påverkas av dem.
	if _, err := database.GetPool().Exec(ctx, `UPDATE calendars SET is_default = FALSE WHERE user_id = $1`, userID); err != nil {
		t.Fatal("reset:", err)
	}

	first, err := calendars.Create(ctx, userID, models.CreateCalendarRequest{Name: "[test] Första"})
	if err != nil {
		t.Fatal("create first:", err)
	}
	defer func() { _, _ = database.GetPool().Exec(ctx, `DELETE FROM calendars WHERE id = $1`, first.ID) }()

	if first.IsDefault {
		t.Fatal("a new calendar should not be default")
	}

	// 1. Sätt den första kalendern som standard
	marked, err := calendars.Update(ctx, userID, first.ID, models.UpdateCalendarRequest{IsDefault: boolPtr(true)})
	if err != nil {
		t.Fatal("mark default:", err)
	}
	if !marked.IsDefault {
		t.Fatal("calendar should be default")
	}

	defaultID, err := calendars.DefaultCalendarID(ctx, userID)
	if err != nil {
		t.Fatal("default id:", err)
	}
	if defaultID == nil || *defaultID != first.ID {
		t.Fatalf("default calendar = %v, want %d", defaultID, first.ID)
	}

	// 2. En ny kalender som standard tar över rollen
	second, err := calendars.Create(ctx, userID, models.CreateCalendarRequest{
		Name:      "[test] Andra",
		IsDefault: boolPtr(true),
	})
	if err != nil {
		t.Fatal("create second:", err)
	}
	defer func() { _, _ = database.GetPool().Exec(ctx, `DELETE FROM calendars WHERE id = $1`, second.ID) }()

	if !second.IsDefault {
		t.Fatal("second calendar should be default")
	}

	listed, err := calendars.List(ctx, userID)
	if err != nil {
		t.Fatal("list:", err)
	}
	defaults := 0
	for _, candidate := range listed {
		if candidate.IsDefault {
			defaults++
		}
	}
	if defaults != 1 {
		t.Fatalf("there should be exactly one default calendar, got %d", defaults)
	}

	defaultID, err = calendars.DefaultCalendarID(ctx, userID)
	if err != nil {
		t.Fatal("default id:", err)
	}
	if defaultID == nil || *defaultID != second.ID {
		t.Fatalf("default calendar = %v, want %d", defaultID, second.ID)
	}

	// 3. Ta bort standardmarkeringen - då finns ingen standardkalender
	cleared, err := calendars.Update(ctx, userID, second.ID, models.UpdateCalendarRequest{IsDefault: boolPtr(false)})
	if err != nil {
		t.Fatal("clear default:", err)
	}
	if cleared.IsDefault {
		t.Fatal("calendar should not be default")
	}

	defaultID, err = calendars.DefaultCalendarID(ctx, userID)
	if err != nil {
		t.Fatal("default id:", err)
	}
	if defaultID != nil {
		t.Fatalf("default calendar = %v, want none", *defaultID)
	}
}

func strPtr(value string) *string { return &value }

func int64Ptr(value int64) *int64 { return &value }

func boolPtr(value bool) *bool { return &value }
