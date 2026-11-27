package repository

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"plan/api/internal/database"
	"plan/api/internal/models"
)

// TestGoalPhaseFactsIntegration verifierar att fasmaskinens fakta hämtas
// rätt ur databasen - godkänd definition, delmål och verifiering.
func TestGoalPhaseFactsIntegration(t *testing.T) {
	if os.Getenv("DATABASE_URL") == "" {
		t.Skip("no DATABASE_URL")
	}
	ctx := context.Background()
	if err := database.Connect(ctx); err != nil {
		t.Fatal(err)
	}
	defer database.Close()

	goals := NewGoalRepository()
	phaseRepo := NewGoalPhaseRepository()
	verificationRepo := NewGoalVerificationRepository()

	userID := testUserID(t)

	goal, err := goals.CreateGoal(ctx, userID, fmt.Sprintf("[test] phase facts %d", time.Now().UnixNano()), "lucide:check-square")
	if err != nil {
		t.Fatal("create test goal:", err)
	}
	defer func() {
		if err := goals.DeleteGoal(ctx, goal.ID); err != nil {
			t.Errorf("delete test goal: %v", err)
		}
	}()

	// Ett nytt mål har ingen godkänd definition och ingen verifiering.
	hasDefinition, err := phaseRepo.HasApprovedDefinition(ctx, goal.ID)
	if err != nil {
		t.Fatal("check approved definition:", err)
	}
	if hasDefinition {
		t.Error("ett nytt mål ska inte ha en godkänd definition")
	}

	passed, err := verificationRepo.HasPassedVerification(ctx, goal.ID)
	if err != nil {
		t.Fatal("check verification:", err)
	}
	if passed {
		t.Error("ett nytt mål ska inte ha en godkänd verifiering")
	}

	// Ett mål utan delmål: inget planerat, men inget olöst heller.
	hasPlanned, allResolved, err := phaseRepo.ChildFacts(ctx, goal.ID)
	if err != nil {
		t.Fatal("child facts:", err)
	}
	if hasPlanned {
		t.Error("ett mål utan delmål ska inte ha planerade delmål")
	}
	if !allResolved {
		t.Error("ett mål utan delmål ska räknas som löst")
	}

	// Verifieringsplanen sätts, ett bevis lämnas in, verifieringen körs och
	// en människa godkänner omdömet. Då ska målet räknas som verifierat.
	plan, err := verificationRepo.Upsert(ctx, goal.ID, models.VerificationTypeEvidence, "ett bevis", models.VerificationStrengthWeak)
	if err != nil {
		t.Fatal("upsert verification:", err)
	}

	if _, err := verificationRepo.AddEvidence(ctx, goal.ID, &plan.ID, "text", "lampan är bytt", "user"); err != nil {
		t.Fatal("add evidence:", err)
	}

	result, err := verificationRepo.CreateResult(ctx, goal.ID, &plan.ID, models.VerificationOutcomePassed, "")
	if err != nil {
		t.Fatal("create result:", err)
	}

	// Innan en människa godkänner omdömet räknas det inte som verifierat.
	passed, err = verificationRepo.HasPassedVerification(ctx, goal.ID)
	if err != nil {
		t.Fatal("check verification:", err)
	}
	if passed {
		t.Error("ett omdöme som inte är godkänt ska inte räknas som verifierat")
	}

	if _, err := verificationRepo.ApproveResult(ctx, result.ID, userID); err != nil {
		t.Fatal("approve result:", err)
	}

	passed, err = verificationRepo.HasPassedVerification(ctx, goal.ID)
	if err != nil {
		t.Fatal("check verification:", err)
	}
	if !passed {
		t.Error("ett godkänt omdöme ska räknas som verifierat")
	}

	// En ändrad plan gör tidigare omdömen ogiltiga - de gäller ju den
	// gamla planen.
	if _, err := verificationRepo.Upsert(ctx, goal.ID, models.VerificationTypeChecklist, "3 av 4", models.VerificationStrengthStrong); err != nil {
		t.Fatal("update verification:", err)
	}

	passed, err = verificationRepo.HasPassedVerification(ctx, goal.ID)
	if err != nil {
		t.Fatal("check verification:", err)
	}
	if passed {
		t.Error("en ändrad plan ska ogiltigförklara tidigare omdömen")
	}
}

// TestVerificationRejectIntegration verifierar att ett avslaget omdöme
// inte räknas som verifierat, och att det ogiltigförklarar de godkända
// omdömen som fanns innan.
//
// Avslaget är ett beslut av en människa. Ett mål som en människa har
// sagt "det här stämmer inte" ska inte kunna räknas som verifierat.
func TestVerificationRejectIntegration(t *testing.T) {
	if os.Getenv("DATABASE_URL") == "" {
		t.Skip("no DATABASE_URL")
	}
	ctx := context.Background()
	if err := database.Connect(ctx); err != nil {
		t.Fatal(err)
	}
	defer database.Close()

	goals := NewGoalRepository()
	verificationRepo := NewGoalVerificationRepository()

	userID := testUserID(t)

	goal, err := goals.CreateGoal(ctx, userID, fmt.Sprintf("[test] reject %d", time.Now().UnixNano()), "lucide:check-square")
	if err != nil {
		t.Fatal("create test goal:", err)
	}
	defer func() {
		if err := goals.DeleteGoal(ctx, goal.ID); err != nil {
			t.Errorf("delete test goal: %v", err)
		}
	}()

	plan, err := verificationRepo.Upsert(ctx, goal.ID, models.VerificationTypeEvidence, "ett bevis", models.VerificationStrengthWeak)
	if err != nil {
		t.Fatal("upsert verification:", err)
	}

	// Ett godkänt omdöme - målet räknas som verifierat.
	godkant, err := verificationRepo.CreateResult(ctx, goal.ID, &plan.ID, models.VerificationOutcomePassed, "")
	if err != nil {
		t.Fatal("create result:", err)
	}
	if _, err := verificationRepo.ApproveResult(ctx, godkant.ID, userID); err != nil {
		t.Fatal("approve result:", err)
	}

	passed, err := verificationRepo.HasPassedVerification(ctx, goal.ID)
	if err != nil {
		t.Fatal("check verification:", err)
	}
	if !passed {
		t.Fatal("ett godkänt omdöme ska räknas som verifierat")
	}

	// Ett nytt omdöme som en människa avslår.
	avslaget, err := verificationRepo.CreateResult(ctx, goal.ID, &plan.ID, models.VerificationOutcomePassed, "")
	if err != nil {
		t.Fatal("create result:", err)
	}

	result, err := verificationRepo.RejectResult(ctx, avslaget.ID, userID)
	if err != nil {
		t.Fatal("reject result:", err)
	}
	if result == nil {
		t.Fatal("avslaget omdömet ska finnas")
	}
	if result.Outcome != models.VerificationOutcomeRejected {
		t.Errorf("utfallet ska vara %q, fick %q", models.VerificationOutcomeRejected, result.Outcome)
	}
	if result.DecidedBy == nil || result.DecidedAt == nil {
		t.Error("ett avslaget omdöme ska vara beslutat av en människa")
	}

	passed, err = verificationRepo.HasPassedVerification(ctx, goal.ID)
	if err != nil {
		t.Fatal("check verification:", err)
	}
	if passed {
		t.Error("ett avslaget omdöme ska inte räknas som verifierat")
	}
}

// TestGoalPhaseEventIntegration verifierar att fasloggen skrivs.
func TestGoalPhaseEventIntegration(t *testing.T) {
	if os.Getenv("DATABASE_URL") == "" {
		t.Skip("no DATABASE_URL")
	}
	ctx := context.Background()
	if err := database.Connect(ctx); err != nil {
		t.Fatal(err)
	}
	defer database.Close()

	goals := NewGoalRepository()
	phaseRepo := NewGoalPhaseRepository()

	userID := testUserID(t)

	goal, err := goals.CreateGoal(ctx, userID, fmt.Sprintf("[test] phase event %d", time.Now().UnixNano()), "lucide:check-square")
	if err != nil {
		t.Fatal("create test goal:", err)
	}
	defer func() {
		if err := goals.DeleteGoal(ctx, goal.ID); err != nil {
			t.Errorf("delete test goal: %v", err)
		}
	}()

	if err := phaseRepo.RecordEvent(ctx, goal.ID, "define", "plan", "agent", false, ""); err != nil {
		t.Fatal("record event:", err)
	}

	var count int
	err = database.GetPool().QueryRow(ctx, `SELECT count(*) FROM goal_phase_events WHERE goal_id = $1`, goal.ID).Scan(&count)
	if err != nil {
		t.Fatal("count events:", err)
	}
	if count != 1 {
		t.Errorf("fasloggen har %d rader, vill ha 1", count)
	}

	if err := phaseRepo.SetCompletionKind(ctx, goal.ID, "deklarerat"); err != nil {
		t.Fatal("set completion kind:", err)
	}

	kind, err := phaseRepo.CompletionKind(ctx, goal.ID)
	if err != nil {
		t.Fatal("get completion kind:", err)
	}
	if kind != "deklarerat" {
		t.Errorf("avslutandet = %q, vill ha %q", kind, "deklarerat")
	}
}

// testUserID ger en användare att arbeta med, eller hoppar över testet.
func testUserID(t *testing.T) int64 {
	t.Helper()

	var userID int64
	err := database.GetPool().QueryRow(context.Background(), `SELECT id FROM users ORDER BY id LIMIT 1`).Scan(&userID)
	if err != nil {
		t.Skip("no users in database")
	}

	return userID
}
