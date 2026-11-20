package repository

import (
	"context"
	"os"
	"testing"
	"time"

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

// TestGoalStartDateIntegration verifierar att ett startdatum kan sättas,
// läsas tillbaka och rensas - och att ett mål med ett framtida startdatum
// inte dyker upp i prioriteringslistan.
func TestGoalStartDateIntegration(t *testing.T) {
	if os.Getenv("DATABASE_URL") == "" {
		t.Skip("no DATABASE_URL")
	}
	ctx := context.Background()
	if err := database.Connect(ctx); err != nil {
		t.Fatal(err)
	}
	defer database.Close()

	repo := NewGoalRepository()

	// Målet måste tillhöra en användare och vara ojort, annars hamnar det
	// inte i prioriteringslistan oavsett startdatum.
	var goalID, userID int64
	err := database.GetPool().QueryRow(ctx, `
		SELECT g.id, ug.user_id
		FROM goals g
		INNER JOIN user_goals ug ON ug.goal_id = g.id
		WHERE g.id <> 1 AND g.finished IS NULL
		LIMIT 1`).Scan(&goalID, &userID)
	if err != nil {
		t.Skip("no unfinished goals")
	}

	// Städa alltid tillbaka det datum som målet hade innan testet, så
	// att testet inte ändrar användardata.
	original := models.NullableDueDate{Set: true, Value: nil}
	{
		fresh, err := repo.GetGoalByID(ctx, goalID)
		if err != nil {
			t.Fatal("get goal:", err)
		}
		if fresh.StartDate != nil {
			value := fresh.StartDate.Format(models.DueDateFormat)
			original = models.NullableDueDate{Set: true, Value: &value}
		}
	}
	defer func() {
		_, _ = repo.UpdateGoal(ctx, goalID, models.UpdateGoalRequest{StartDate: original})
	}()

	tomorrow := time.Now().UTC().AddDate(0, 0, 1).Format(models.DueDateFormat)
	updated, err := repo.UpdateGoal(ctx, goalID, models.UpdateGoalRequest{
		StartDate: models.NullableDueDate{Set: true, Value: &tomorrow},
	})
	if err != nil {
		t.Fatal("set start date:", err)
	}
	if updated.StartDate == nil {
		t.Fatal("start_date = nil, want tomorrow")
	}
	if got := updated.StartDate.Format(models.DueDateFormat); got != tomorrow {
		t.Fatalf("start_date = %q, want %q", got, tomorrow)
	}

	fetched, err := repo.GetGoalByID(ctx, goalID)
	if err != nil {
		t.Fatal("get goal:", err)
	}
	if fetched.StartDate == nil {
		t.Fatal("start_date = nil after read, want tomorrow")
	}
	if got := fetched.StartDate.Format(models.DueDateFormat); got != tomorrow {
		t.Fatalf("start_date = %q after read, want %q", got, tomorrow)
	}

	// Ett mål med framtida startdatum får inte dyka upp i prioriteringslistan.
	prioritized, err := repo.GetPrioritizedGoals(ctx, userID)
	if err != nil {
		t.Fatal("get prioritized goals:", err)
	}
	for _, g := range prioritized {
		if g.ID == goalID {
			t.Fatal("målet med framtida startdatum dyckte upp i prioriteringslistan")
		}
	}

	// Ett startdatum i det förflutna släpper igenom målet.
	yesterday := time.Now().UTC().AddDate(0, 0, -1).Format(models.DueDateFormat)
	if _, err := repo.UpdateGoal(ctx, goalID, models.UpdateGoalRequest{
		StartDate: models.NullableDueDate{Set: true, Value: &yesterday},
	}); err != nil {
		t.Fatal("set past start date:", err)
	}

	prioritized, err = repo.GetPrioritizedGoals(ctx, userID)
	if err != nil {
		t.Fatal("get prioritized goals:", err)
	}
	found := false
	for _, g := range prioritized {
		if g.ID == goalID {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("målet med förflutet startdatum saknades i prioriteringslistan")
	}

	// Rensa datumet
	cleared, err := repo.UpdateGoal(ctx, goalID, models.UpdateGoalRequest{
		StartDate: models.NullableDueDate{Set: true, Value: nil},
	})
	if err != nil {
		t.Fatal("clear start date:", err)
	}
	if cleared.StartDate != nil {
		t.Fatalf("start_date = %v after clear, want nil", cleared.StartDate)
	}

	// Ett ogiltigt datum ska avvisas
	invalid := "2027-13-45"
	if _, err := repo.UpdateGoal(ctx, goalID, models.UpdateGoalRequest{
		StartDate: models.NullableDueDate{Set: true, Value: &invalid},
	}); err == nil {
		t.Fatal("väntade sig ett fel för ett ogiltigt datum")
	}
}
