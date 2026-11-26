package repository

import (
	"context"
	"os"
	"testing"

	"plan-api/internal/database"
	"plan-api/internal/models"
)

// TestGoalEstimateIntegration verifierar att en uppskattad tidsåtgång kan
// sättas, läsas tillbaka och rensas på ett riktigt mål.
func TestGoalEstimateIntegration(t *testing.T) {
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

	// Sätt en uppskattad tidsåtgång
	minutes := 90
	updated, err := repo.UpdateGoal(ctx, goalID, models.UpdateGoalRequest{
		EstimatedMinutes: models.NullableInt{Set: true, Value: &minutes},
	})
	if err != nil {
		t.Fatal("set estimate:", err)
	}
	if updated.EstimatedMinutes == nil || *updated.EstimatedMinutes != 90 {
		t.Fatalf("estimated_minutes = %v, want 90", updated.EstimatedMinutes)
	}

	// Läs tillbaka
	fetched, err := repo.GetGoalByID(ctx, goalID)
	if err != nil {
		t.Fatal("get goal:", err)
	}
	if fetched.EstimatedMinutes == nil || *fetched.EstimatedMinutes != 90 {
		t.Fatalf("estimated_minutes = %v after read, want 90", fetched.EstimatedMinutes)
	}

	// Rensa uppskattningen
	cleared, err := repo.UpdateGoal(ctx, goalID, models.UpdateGoalRequest{
		EstimatedMinutes: models.NullableInt{Set: true, Value: nil},
	})
	if err != nil {
		t.Fatal("clear estimate:", err)
	}
	if cleared.EstimatedMinutes != nil {
		t.Fatalf("estimated_minutes = %v after clear, want nil", cleared.EstimatedMinutes)
	}

	// En negativ uppskattning ska avvisas
	negative := -30
	if _, err := repo.UpdateGoal(ctx, goalID, models.UpdateGoalRequest{
		EstimatedMinutes: models.NullableInt{Set: true, Value: &negative},
	}); err == nil {
		t.Fatal("negative estimate was accepted, want error")
	}
}
