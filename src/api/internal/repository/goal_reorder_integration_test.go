package repository

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"plan/api/internal/database"
)

// TestReorderPriorityGoalsIntegration verifierar att en ny ordning på
// prioriteringslistan sparas i databasen och att listan läses tillbaka i
// samma ordning - det är det som gör att en omsortering överlever en
// sidladdning.
func TestReorderPriorityGoalsIntegration(t *testing.T) {
	if os.Getenv("DATABASE_URL") == "" {
		t.Skip("no DATABASE_URL")
	}
	ctx := context.Background()
	if err := database.Connect(ctx); err != nil {
		t.Fatal(err)
	}
	defer database.Close()

	repo := NewGoalRepository()
	pool := database.GetPool()

	// Testen använder en egen användare så att den inte röddar någon
	// riktig användardata.
	sub := fmt.Sprintf("reorder-test-%d", time.Now().UnixNano()%1e12)
	var userID int64
	if err := pool.QueryRow(ctx, `
		INSERT INTO users (sub, email) VALUES ($1, $2) RETURNING id`,
		sub, sub+"@test.local").Scan(&userID); err != nil {
		t.Fatal("create user:", err)
	}
	defer func() {
		_, _ = pool.Exec(ctx, `DELETE FROM users WHERE id = $1`, userID)
	}()

	titles := []string{"Ett", "Två", "Tre", "Fyra", "Fem"}
	ids := make([]int64, 0, len(titles))
	for _, title := range titles {
		var id int64
		if err := pool.QueryRow(ctx, `
			INSERT INTO goals (title) VALUES ($1) RETURNING id`, title).Scan(&id); err != nil {
			t.Fatal("create goal:", err)
		}
		if _, err := pool.Exec(ctx, `
			INSERT INTO user_goals (user_id, goal_id) VALUES ($1, $2)`, userID, id); err != nil {
			t.Fatal("link goal:", err)
		}
		ids = append(ids, id)
	}
	defer func() {
		_, _ = pool.Exec(ctx, `DELETE FROM goals WHERE id = ANY($1)`, ids)
	}()

	// Vänd ordningen och spara den.
	reversed := make([]int64, len(ids))
	for i, id := range ids {
		reversed[len(ids)-1-i] = id
	}

	if err := repo.ReorderPriorityGoals(ctx, userID, reversed); err != nil {
		t.Fatal("reorder:", err)
	}

	// Läs listan igen - samma ordning ska komma tillbaka.
	prioritized, err := repo.GetPrioritizedGoals(ctx, userID)
	if err != nil {
		t.Fatal("get prioritized goals:", err)
	}
	if len(prioritized) != len(ids) {
		t.Fatalf("antal mål = %d, vill ha %d", len(prioritized), len(ids))
	}
	for i, goal := range prioritized {
		if goal.ID != reversed[i] {
			t.Fatalf("mål %d har ordningen %d, vill ha %d", goal.ID, i, reversed[i])
		}
	}

	// En ny omsortering ska bara flytta det valda målet - resten ska stå
	// kvar i samma ordning.
	moved := append([]int64{}, reversed...)
	moved[0], moved[1] = moved[1], moved[0]
	if err := repo.ReorderPriorityGoals(ctx, userID, moved); err != nil {
		t.Fatal("reorder again:", err)
	}

	prioritized, err = repo.GetPrioritizedGoals(ctx, userID)
	if err != nil {
		t.Fatal("get prioritized goals:", err)
	}
	for i, goal := range prioritized {
		if goal.ID != moved[i] {
			t.Fatalf("mål %d har ordningen %d, vill ha %d", goal.ID, i, moved[i])
		}
	}
}
