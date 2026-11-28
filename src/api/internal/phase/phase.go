// Package phase innehåller fasmaskinen för mål.
//
// Maskinen är ren logik: den tar fakta om målet och svarar om en
// fasövergång är tillåten. Här finns varken databas eller HTTP, så reglerna
// går att testa utan en databas - och samma regler gäller i API:t, i
// agenten och i gränssnittet.
//
// Tanken är att en människa och en agent inte ska behöva olika regler. De
// har olika rättigheter: en människa får avsluta ett mål direkt, en agent
// får inte det - den får föreslå, och en människa bestämmer.
package phase

import "fmt"

// Faserna, i ordning.
const (
	Define    = "define"
	Plan      = "plan"
	Execute   = "execute"
	Verify    = "verify"
	Completed = "completed"
)

// Order är fasernas ordning.
var Order = []string{Define, Plan, Execute, Verify, Completed}

// Sättet ett mål avslutades på.
const (
	// CompletionDeclared betyder att en människa avslutade målet utan att
	// en verifiering gått igenom.
	CompletionDeclared = "deklarerat"
	// CompletionVerified betyder att målet avslutades med en verifiering
	// som gått igenom.
	CompletionVerified = "verifierat"
)

// Skälen till att en övergång stoppas. De är stabila strängar - klienten
// kan visa dem och en framtida agent kan svara på dem.
const (
	// ReasonUnknownPhase betyder att någon fas inte finns.
	ReasonUnknownPhase = "unknown_phase"
	// ReasonDefinitionNotApproved betyder att målet saknar en godkänd
	// agentdefinition.
	ReasonDefinitionNotApproved = "definition_not_approved"
	// ReasonMissingPlannedChildren betyder att målet saknar delmål att
	// arbeta med.
	ReasonMissingPlannedChildren = "missing_planned_children"
	// ReasonUnresolvedChildren betyder att ett delmål saknar resultat.
	ReasonUnresolvedChildren = "unresolved_children"
	// ReasonVerificationNotPassed betyder att verifieringen inte har
	// gått igenom.
	ReasonVerificationNotPassed = "verification_not_passed"
	// ReasonBackwards betyder att fasen skulle gå bakåt.
	ReasonBackwards = "backwards_transition"
	// ReasonPhaseSkipped betyder att en agent försökte hoppa över en fas.
	// Faserna är kontraktet mellan agenten och Plan - agenten får inte
	// ta genvägar.
	ReasonPhaseSkipped = "phase_skipped"
	// ReasonApprovalRequired betyder att övergången kräver ett
	// mänskligt beslut först.
	ReasonApprovalRequired = "approval_required"
)

// Facts beskriver vad som är sant om ett mål just nu.
//
// Fälten fylls i av API:t innan maskinen får svara. De är medvetet
// enkla - maskinen ska kunna resonera om dem utan att veta något om
// databasen.
type Facts struct {
	// HasApprovedDefinition betyder att en människa har godkänt en
	// agentdefinition för målet.
	HasApprovedDefinition bool
	// HasPlannedChildren betyder att målet har minst ett delmål med
	// datum eller uppskattad tid.
	HasPlannedChildren bool
	// AllChildrenResolved betyder att varje delmål har ett resultat -
	// avslutat, uppskjutet eller hoppat.
	AllChildrenResolved bool
	// HasVerificationPlan betyder att målet har en verifieringsplan.
	HasVerificationPlan bool
	// VerificationPassed betyder att verifieringen gått igenom och
	// godkänts av en människa.
	VerificationPassed bool
	// IsHuman betyder att övergången begärs av en människa.
	IsHuman bool
	// OverrideReason gör att en människa tvingar övergången. Orsaken
	// loggas som en avvikelse.
	OverrideReason string
}

// Result är maskinens svar.
type Result struct {
	// Allowed säger om övergången får ske.
	Allowed bool
	// Code är en stabil skälkod när övergången stoppas.
	Code string
	// Message förklarar skälet på svenska.
	Message string
	// CompletionKind säger hur målet avslutades - tomt för andra
	// övergångar.
	CompletionKind string
	// Forced säger att människan tvingade övergången, antingen för att
	// gå bakåt eller för att hoppa över faser.
	Forced bool
	// SkippedPhases är antalet faser som hoppades över.
	SkippedPhases int
}

// Index ger fasens position i ordningen, eller -1 för en okänd fas.
func Index(name string) int {
	for i, phase := range Order {
		if phase == name {
			return i
		}
	}
	return -1
}

// Known berättar om namnet är en fas.
func Known(name string) bool {
	return Index(name) >= 0
}

// Evaluate svarar på om ett mål får gå från fasen from till fasen to.
//
// Reglerna:
//   - En okänd fas stoppas alltid.
//   - En agent får inte gå bakåt, och får inte hoppa över faser.
//   - Varje övergång har ett villkor: en godkänd definition, delmål att
//     arbeta med, avslutade delmål, eller en verifiering som gått igenom.
//   - En människa får alltid avsluta ett mål - från vilken fas som helst.
//     Då blir avslutandet "deklarerat", eller "verifierat" om verifieringen
//     gick igenom.
//   - En agent får inte avsluta ett mål själv. Den får föreslå det, och
//     en människa bestämmer - därför stoppas övergången här.
func Evaluate(from, to string, facts Facts) Result {
	if !Known(from) {
		return Result{Code: ReasonUnknownPhase, Message: fmt.Sprintf("okänd fas %q", from)}
	}
	if !Known(to) {
		return Result{Code: ReasonUnknownPhase, Message: fmt.Sprintf("okänd fas %q", to)}
	}

	fromIdx, toIdx := Index(from), Index(to)

	if fromIdx == toIdx {
		return Result{Allowed: true}
	}

	// Bakåtgång. Bara en människa får öppna upp ett mål igen, och det
	// loggas som en avvikelse.
	if toIdx < fromIdx {
		if !facts.IsHuman {
			return Result{
				Code:    ReasonBackwards,
				Message: "fasen får inte gå bakåt för en agent",
			}
		}
		return Result{Allowed: true, Forced: true}
	}

	// Avslutande. Det är den enda övergången där en människa och en agent
	// har helt olika regler.
	if to == Completed {
		if !facts.IsHuman {
			return Result{
				Code:    ReasonApprovalRequired,
				Message: "en agent får inte avsluta ett mål själv - förslaget ska godkännas av en människa",
			}
		}
		kind := CompletionDeclared
		if facts.HasVerificationPlan && facts.VerificationPassed {
			kind = CompletionVerified
		}
		return Result{Allowed: true, CompletionKind: kind}
	}

	// Framåtgång. Villkoren gäller agenten - en människa bestämmer själv
	// var hennes mål befinner sig, och får hoppa över faser.
	skipped := toIdx - fromIdx - 1

	if !facts.IsHuman {
		if skipped > 0 {
			return Result{
				Code:    ReasonPhaseSkipped,
				Message: "en agent får inte hoppa över faser",
			}
		}
		if blocked, ok := agentPrecondition(to, facts); !ok {
			return blocked
		}
	}

	forced := skipped > 0

	return Result{Allowed: true, Forced: forced, SkippedPhases: skipped}
}

// agentPrecondition ger villkoret för agentens övergång till fasen to.
//
// Ett okänt villkor (en fas utan villkor) släpper igenom.
func agentPrecondition(to string, facts Facts) (Result, bool) {
	switch to {
	case Plan:
		if !facts.HasApprovedDefinition {
			return Result{
				Code:    ReasonDefinitionNotApproved,
				Message: "målet saknar en godkänd definition",
			}, false
		}
	case Execute:
		if !facts.HasPlannedChildren {
			return Result{
				Code:    ReasonMissingPlannedChildren,
				Message: "målet saknar delmål med datum eller uppskattad tid",
			}, false
		}
	case Verify:
		if !facts.AllChildrenResolved {
			return Result{
				Code:    ReasonUnresolvedChildren,
				Message: "ett delmål saknar resultat",
			}, false
		}
	}

	return Result{}, true
}
