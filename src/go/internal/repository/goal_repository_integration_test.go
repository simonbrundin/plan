package repository

import (
	"context"
	"os"
	"testing"

	"plan-api/internal/database"
	"plan-api/internal/models"
)

// TestGoalDueDateIntegration verifierar att ett due-datum kan sättas,
// läsas tillbaka och rensas på ett riktigt mål.
func TestGoalDueDateIntegration(t *testing.T) {
	if os.Getenv("DATABASE_URL") == "" {
		t.Skip("no DATABASE_URL")
	}
	ctx := context.Background()
	if err := database.Connect(ctx); err != nil {
		t.Fatal(err)
	}
	defer database.Close()

	repo := NewGoalRepository()

	var goalID int64
	if err := database.GetPool().QueryRow(ctx, `SELECT id FROM goals WHERE id <> 1 LIMIT 1`).Scan(&goalID); err != nil {
		t.Skip("no goals")
	}

	// Sätt ett due-datum
	due := "2026-11-20"
	updated, err := repo.UpdateGoal(ctx, goalID, models.UpdateGoalRequest{
		Due: models.NullableDueDate{Set: true, Value: &due},
	})
	if err != nil {
		t.Fatal("set due:", err)
	}
	if updated.Due == nil {
		t.Fatal("due = nil, want 2026-11-20")
	}
	if got := updated.Due.Format(models.DueDateFormat); got != due {
		t.Fatalf("due = %q, want %q", got, due)
	}

	// Läs tillbaka
	fetched, err := repo.GetGoalByID(ctx, goalID)
	if err != nil {
		t.Fatal("get goal:", err)
	}
	if fetched.Due == nil {
		t.Fatal("due = nil after read, want 2026-11-20")
	}
	if got := fetched.Due.Format(models.DueDateFormat); got != due {
		t.Fatalf("due = %q after read, want %q", got, due)
	}

	// Rensa datumet
	cleared, err := repo.UpdateGoal(ctx, goalID, models.UpdateGoalRequest{
		Due: models.NullableDueDate{Set: true, Value: nil},
	})
	if err != nil {
		t.Fatal("clear due:", err)
	}
	if cleared.Due != nil {
		t.Fatalf("due = %v after clear, want nil", cleared.Due)
	}

	// Ett ogiltigt datum ska avvisas
	invalid := "20-11-2026"
	if _, err := repo.UpdateGoal(ctx, goalID, models.UpdateGoalRequest{
		Due: models.NullableDueDate{Set: true, Value: &invalid},
	}); err == nil {
		t.Fatal("väntade sig ett fel för ett ogiltigt datum")
	}
}
