// Package phaseservice kopplar ihop fasmaskinen med databasen.
//
// Maskinen i package phase är ren logik - den vet ingenting om
// databasen. Den här tjänsten hämtar faktorna, låter maskinen avgöra och
// skriver sedan resultatet. Både fasendpointern och den gamla
// statusendpointern går via den här - annars skulle fasmaskinen kunna
// kringgås.
package phaseservice

import (
	"context"
	"fmt"

	"plan/api/internal/models"
	"plan/api/internal/phase"
	"plan/api/internal/repository"
)

// Service utför fasövergångar.
type Service struct {
	phase        *repository.GoalPhaseRepository
	goals        *repository.GoalRepository
	verification *repository.GoalVerificationRepository
}

// New ger en service med databasen.
func New() *Service {
	return &Service{
		phase:        repository.NewGoalPhaseRepository(),
		goals:        repository.NewGoalRepository(),
		verification: repository.NewGoalVerificationRepository(),
	}
}

// Facts hämtar det maskinen behöver veta om målet.
func (s *Service) Facts(ctx context.Context, goalID int64) (phase.Facts, error) {
	var facts phase.Facts

	hasDefinition, err := s.phase.HasApprovedDefinition(ctx, goalID)
	if err != nil {
		return facts, err
	}
	facts.HasApprovedDefinition = hasDefinition

	hasPlanned, allResolved, err := s.phase.ChildFacts(ctx, goalID)
	if err != nil {
		return facts, err
	}
	facts.HasPlannedChildren = hasPlanned
	facts.AllChildrenResolved = allResolved

	verification, err := s.verification.GetByGoal(ctx, goalID)
	if err != nil {
		return facts, err
	}
	facts.HasVerificationPlan = verification != nil

	passed, err := s.verification.HasPassedVerification(ctx, goalID)
	if err != nil {
		return facts, err
	}
	facts.VerificationPassed = passed

	return facts, nil
}

// Outcome är resultatet av en övergång.
type Outcome struct {
	// Result är maskinens svar.
	Result phase.Result
	// Update är statusändringen som skrevs, om någon skrevs.
	Update *models.StatusUpdate
}

// Apply låter maskinen avgöra övergången och skriver den.
//
// Den aktuella fasen läses ur databasen - klienten kan inte bestämma
// varifrån övergången sker. Övergången skrivs bara om maskinen säger ja.
func (s *Service) Apply(
	ctx context.Context, goalID int64, to string, isHuman bool, reason string,
) (Outcome, error) {
	current, err := s.goals.GetGoalStatus(ctx, goalID)
	if err != nil {
		return Outcome{}, fmt.Errorf("failed to read current phase: %w", err)
	}
	if current == nil {
		return Outcome{}, ErrGoalNotFound
	}

	facts, err := s.Facts(ctx, goalID)
	if err != nil {
		return Outcome{}, err
	}
	facts.IsHuman = isHuman
	facts.OverrideReason = reason

	result := phase.Evaluate(current.Name, to, facts)
	if !result.Allowed {
		return Outcome{Result: result}, nil
	}

	statusID, err := s.phase.StatusIDByName(ctx, to)
	if err != nil {
		return Outcome{}, err
	}

	update, err := s.goals.UpdateGoalStatus(ctx, goalID, statusID)
	if err != nil {
		return Outcome{}, fmt.Errorf("failed to update status: %w", err)
	}

	if result.CompletionKind != "" {
		if err := s.phase.SetCompletionKind(ctx, goalID, result.CompletionKind); err != nil {
			return Outcome{}, err
		}
	}

	actor := "human"
	if !isHuman {
		actor = "agent"
	}
	if err := s.phase.RecordEvent(ctx, goalID, current.Name, to, actor, result.Forced, reason); err != nil {
		return Outcome{}, err
	}

	return Outcome{Result: result, Update: update}, nil
}

// ErrGoalNotFound betyder att målet inte finns.
var ErrGoalNotFound = fmt.Errorf("goal not found")
