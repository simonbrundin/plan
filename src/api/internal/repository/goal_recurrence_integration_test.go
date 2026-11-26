package repository

import (
	"context"
	"os"
	"testing"
	"time"

	"plan/api/internal/database"
	"plan/api/internal/models"
	"plan/api/internal/recurrence"
)

// TestGoalRecurrenceIntegration verifierar att ett återkommande mål får en
// första förekomst, att förekomster kan slutföras idempotent och att
// avbockning återställer förekomsten.
func TestGoalRecurrenceIntegration(t *testing.T) {
	if os.Getenv("DATABASE_URL") == "" {
		t.Skip("no DATABASE_URL")
	}
	ctx := context.Background()
	if err := database.Connect(ctx); err != nil {
		t.Fatal(err)
	}
	defer database.Close()

	repo := NewGoalRepository()

	var userID int64
	if err := database.GetPool().QueryRow(ctx, `SELECT id FROM users ORDER BY id LIMIT 1`).Scan(&userID); err != nil {
		t.Skip("no users")
	}

	goal, err := repo.CreateGoal(ctx, userID, "[test] återkommande mål", "heroicons:star")
	if err != nil {
		t.Fatal("create goal:", err)
	}
	defer func() {
		if err := repo.DeleteGoal(ctx, goal.ID); err != nil {
			t.Log("cleanup goal:", err)
		}
	}()

	anchor := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)

	// Sätt regeln - första förekomsten ska vara samma dag som anchor.
	updated, err := repo.SetRecurrence(ctx, goal.ID, "FREQ=DAILY;INTERVAL=2", recurrence.ModeSchedule, anchor)
	if err != nil {
		t.Fatal("set recurrence:", err)
	}
	if updated.RecurrenceRule == nil || *updated.RecurrenceRule != "FREQ=DAILY;INTERVAL=2" {
		t.Fatalf("recurrence_rule = %v, want FREQ=DAILY;INTERVAL=2", updated.RecurrenceRule)
	}
	if updated.RecurrenceNext == nil {
		t.Fatal("recurrence_next = nil, want 2026-01-01")
	}
	if got := updated.RecurrenceNext.Format(models.DueDateFormat); got != "2026-01-01" {
		t.Fatalf("recurrence_next = %q, want 2026-01-01", got)
	}
	if updated.RecurrenceCount != 0 {
		t.Fatalf("recurrence_count = %d, want 0", updated.RecurrenceCount)
	}

	// Slutför förekomsten - nästa förekomst ska bli 2026-01-03.
	completed, err := repo.CompleteOccurrence(ctx, goal.ID, nil, anchor)
	if err != nil {
		t.Fatal("complete occurrence:", err)
	}
	if completed.RecurrenceCount != 1 {
		t.Fatalf("recurrence_count = %d, want 1", completed.RecurrenceCount)
	}
	if completed.RecurrenceNext == nil {
		t.Fatal("recurrence_next = nil after completion")
	}
	if got := completed.RecurrenceNext.Format(models.DueDateFormat); got != "2026-01-03" {
		t.Fatalf("recurrence_next = %q, want 2026-01-03", got)
	}

	// Samma förekomst igen ska inte räknas två gånger.
	again, err := repo.CompleteOccurrence(ctx, goal.ID, &anchor, anchor)
	if err != nil {
		t.Fatal("complete occurrence twice:", err)
	}
	if again.RecurrenceCount != 1 {
		t.Fatalf("recurrence_count = %d after duplicate, want 1", again.RecurrenceCount)
	}
	if again.RecurrenceNext == nil {
		t.Fatal("recurrence_next = nil after duplicate")
	}
	if got := again.RecurrenceNext.Format(models.DueDateFormat); got != "2026-01-03" {
		t.Fatalf("recurrence_next = %q after duplicate, want 2026-01-03", got)
	}

	// Avbockning ska ta bort raden och återställa förekomsten.
	undone, err := repo.UndoOccurrence(ctx, goal.ID)
	if err != nil {
		t.Fatal("undo occurrence:", err)
	}
	if undone.RecurrenceCount != 0 {
		t.Fatalf("recurrence_count = %d after undo, want 0", undone.RecurrenceCount)
	}
	if undone.RecurrenceNext == nil {
		t.Fatal("recurrence_next = nil after undo")
	}
	if got := undone.RecurrenceNext.Format(models.DueDateFormat); got != "2026-01-01" {
		t.Fatalf("recurrence_next = %q after undo, want 2026-01-01", got)
	}

	// En ogiltig regel ska avvisas utan att målets regel ändras.
	if _, err := repo.SetRecurrence(ctx, goal.ID, "FREQ=DAILY;INTERVAL=0", recurrence.ModeSchedule, anchor); err == nil {
		t.Fatal("väntade sig ett fel för en ogiltig regel")
	}
	fetched, err := repo.GetGoalByID(ctx, goal.ID)
	if err != nil {
		t.Fatal("get goal:", err)
	}
	if fetched.RecurrenceRule == nil || *fetched.RecurrenceRule != "FREQ=DAILY;INTERVAL=2" {
		t.Fatalf("recurrence_rule = %v efter ogiltig regel, want oförändrad", fetched.RecurrenceRule)
	}

	// Ta bort upprepningen - då ska målet bli ett vanligt mål igen.
	if err := repo.ClearRecurrence(ctx, goal.ID); err != nil {
		t.Fatal("clear recurrence:", err)
	}
	cleared, err := repo.GetGoalByID(ctx, goal.ID)
	if err != nil {
		t.Fatal("get goal after clear:", err)
	}
	if cleared.RecurrenceRule != nil {
		t.Fatalf("recurrence_rule = %v after clear, want nil", cleared.RecurrenceRule)
	}
	if cleared.RecurrenceCount != 0 {
		t.Fatalf("recurrence_count = %d after clear, want 0", cleared.RecurrenceCount)
	}
}
