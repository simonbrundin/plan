package repository

import (
	"context"
	"os"
	"testing"

	"plan-api/internal/database"
	"plan-api/internal/models"
)

// testUserAndGoal hämtar en användare och ett mål som tillhör användaren.
// Skips om det inte finns några.
func testUserAndGoal(ctx context.Context, t *testing.T) (int64, int64) {
	t.Helper()

	var userID int64
	err := database.GetPool().QueryRow(ctx, `SELECT id FROM users ORDER BY id LIMIT 1`).Scan(&userID)
	if err != nil {
		t.Skip("no users")
	}

	var goalID int64
	err = database.GetPool().QueryRow(ctx,
		`SELECT goal_id FROM user_goals WHERE user_id = $1 ORDER BY goal_id LIMIT 1`, userID).Scan(&goalID)
	if err != nil {
		t.Skip("no goals for user")
	}

	return userID, goalID
}

// TestPlaceRepositoryIntegration verifierar att en plats kan skapas,
// kopplas till ett mål, läsas tillbaka och raderas.
func TestPlaceRepositoryIntegration(t *testing.T) {
	if os.Getenv("DATABASE_URL") == "" {
		t.Skip("no DATABASE_URL")
	}
	ctx := context.Background()
	if err := database.Connect(ctx); err != nil {
		t.Fatal(err)
	}
	defer database.Close()

	userID, goalID := testUserAndGoal(ctx, t)
	repo := NewPlaceRepository()

	// Skapa platsen
	created, err := repo.Create(ctx, userID, models.CreatePlaceRequest{
		Name:         "[test] plats",
		Latitude:     59.3293,
		Longitude:    18.0686,
		RadiusMeters: intPtr(100),
	})
	if err != nil {
		t.Fatal("create place:", err)
	}
	defer func() {
		if _, err := repo.Delete(ctx, userID, created.ID); err != nil {
			t.Log("cleanup place:", err)
		}
	}()

	if created.Name != "[test] plats" {
		t.Fatalf("name = %q, want %q", created.Name, "[test] plats")
	}
	if created.RadiusMeters != 100 {
		t.Fatalf("radius = %d, want 100", created.RadiusMeters)
	}

	// Läs tillbaka
	fetched, err := repo.Get(ctx, userID, created.ID)
	if err != nil {
		t.Fatal("get place:", err)
	}
	if fetched == nil {
		t.Fatal("place = nil, want the created place")
	}

	// Koppla till målet
	places, err := repo.SetGoalPlaces(ctx, userID, goalID, []int64{created.ID})
	if err != nil {
		t.Fatal("set goal places:", err)
	}
	if len(places) != 1 || places[0].ID != created.ID {
		t.Fatalf("goal places = %v, want the created place", places)
	}

	byGoal, err := repo.GetPlacesForGoals(ctx, []int64{goalID})
	if err != nil {
		t.Fatal("get places for goals:", err)
	}
	if len(byGoal[goalID]) != 1 {
		t.Fatalf("places for goal = %v, want 1 place", byGoal[goalID])
	}

	// Ta bort kopplingen igen
	places, err = repo.SetGoalPlaces(ctx, userID, goalID, []int64{})
	if err != nil {
		t.Fatal("clear goal places:", err)
	}
	if len(places) != 0 {
		t.Fatalf("goal places = %v, want none", places)
	}
}

// TestPlaceNearbyIntegration verifierar att en plats matchas mot en
// position inom radien, men inte utanför.
func TestPlaceNearbyIntegration(t *testing.T) {
	if os.Getenv("DATABASE_URL") == "" {
		t.Skip("no DATABASE_URL")
	}
	ctx := context.Background()
	if err := database.Connect(ctx); err != nil {
		t.Fatal(err)
	}
	defer database.Close()

	userID, _ := testUserAndGoal(ctx, t)
	repo := NewPlaceRepository()

	created, err := repo.Create(ctx, userID, models.CreatePlaceRequest{
		Name:         "[test] när plats",
		Latitude:     59.3293,
		Longitude:    18.0686,
		RadiusMeters: intPtr(100),
	})
	if err != nil {
		t.Fatal("create place:", err)
	}
	defer func() {
		if _, err := repo.Delete(ctx, userID, created.ID); err != nil {
			t.Log("cleanup place:", err)
		}
	}()

	nearby, err := repo.GetNearbyPlaces(ctx, userID, 59.3293, 18.0686)
	if err != nil {
		t.Fatal("nearby places:", err)
	}
	if !containsPlace(nearby, created.ID) {
		t.Fatalf("nearby = %v, want the created place", nearby)
	}

	far, err := repo.GetNearbyPlaces(ctx, userID, 59.3473, 18.0686)
	if err != nil {
		t.Fatal("nearby places:", err)
	}
	if containsPlace(far, created.ID) {
		t.Fatalf("nearby = %v, want the created place to be excluded", far)
	}
}

// TestCurrentPlaceIntegration verifierar att "jag är här"-platsen kan
// sättas och rensas.
func TestCurrentPlaceIntegration(t *testing.T) {
	if os.Getenv("DATABASE_URL") == "" {
		t.Skip("no DATABASE_URL")
	}
	ctx := context.Background()
	if err := database.Connect(ctx); err != nil {
		t.Fatal(err)
	}
	defer database.Close()

	userID, _ := testUserAndGoal(ctx, t)
	repo := NewPlaceRepository()

	// Spara den nuvarande platsen så att testet inte rörar användarens data.
	original, err := repo.GetCurrentPlace(ctx, userID)
	if err != nil {
		t.Fatal("get current place:", err)
	}
	defer func() {
		var originalID *int64
		if original != nil {
			id := original.ID
			originalID = &id
		}
		if err := repo.SetCurrentPlace(ctx, userID, originalID); err != nil {
			t.Log("restore current place:", err)
		}
	}()

	created, err := repo.Create(ctx, userID, models.CreatePlaceRequest{
		Name:         "[test] vald plats",
		Latitude:     59.3293,
		Longitude:    18.0686,
		RadiusMeters: intPtr(100),
	})
	if err != nil {
		t.Fatal("create place:", err)
	}
	defer func() {
		if _, err := repo.Delete(ctx, userID, created.ID); err != nil {
			t.Log("cleanup place:", err)
		}
	}()

	if err := repo.SetCurrentPlace(ctx, userID, &created.ID); err != nil {
		t.Fatal("set current place:", err)
	}

	current, err := repo.GetCurrentPlace(ctx, userID)
	if err != nil {
		t.Fatal("get current place:", err)
	}
	if current == nil || current.ID != created.ID {
		t.Fatalf("current place = %v, want the created place", current)
	}

	if err := repo.SetCurrentPlace(ctx, userID, nil); err != nil {
		t.Fatal("clear current place:", err)
	}
	current, err = repo.GetCurrentPlace(ctx, userID)
	if err != nil {
		t.Fatal("get current place:", err)
	}
	if current != nil {
		t.Fatalf("current place = %v, want none", current)
	}
}

func containsPlace(places []models.Place, id int64) bool {
	for _, p := range places {
		if p.ID == id {
			return true
		}
	}
	return false
}

func intPtr(v int) *int {
	return &v
}
