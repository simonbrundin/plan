package models

import "time"

// Verifieringstyper. Typen säger var beviset kommer ifrån.
const (
	// VerificationTypeChecklist är en checklista som användaren bockar av.
	VerificationTypeChecklist = "checklista"
	// VerificationTypeSubgoal betyder att alla undermål måste vara klara.
	VerificationTypeSubgoal = "delmal"
	// VerificationTypeCalendar betyder att en kalenderhändelse kopplad till
	// målet måste ha passerat.
	VerificationTypeCalendar = "kalender"
	// VerificationTypeEvidence betyder att användaren lämnar in ett bevis.
	VerificationTypeEvidence = "bevis"
	// VerificationTypeValue betyder att ett uppmätt värde finns, till exempel
	// vikt eller antal.
	VerificationTypeValue = "matvarde"
	// VerificationTypeCode betyder att kodens läge avgör, till exempel en
	// grön CI-körning.
	VerificationTypeCode = "kod"
)

// Styrkan på en verifiering.
const (
	// VerificationStrengthStrong betyder att beviset är objektivt.
	VerificationStrengthStrong = "stark"
	// VerificationStrengthWeak betyder att beviset är användarens eget
	// intyg.
	VerificationStrengthWeak = "svag"
)

// Utfallet på en verifiering.
const (
	// VerificationOutcomePassed betyder att villkoret är uppfyllt.
	VerificationOutcomePassed = "passed"
	// VerificationOutcomeFailed betyder att villkoret inte är uppfyllt.
	VerificationOutcomeFailed = "failed"
	// VerificationOutcomePending betyder att bevis saknas.
	VerificationOutcomePending = "pending"
)

// GoalVerification är en verifieringsplan för ett mål.
//
// Planen är valfri. Ett mål utan plan fungerar precis som ett mål med en
// plan - skillnaden är att agenten då inte får föreslå att målet är klart
// utan att en människa bestämmer det.
type GoalVerification struct {
	ID        int64      `json:"id" example:"1"`
	GoalID    int64      `json:"goal_id" example:"449"`
	Type      string     `json:"type" example:"checklista"`
	Criterion string     `json:"criterion" example:"3 av 4 moment avbockade"`
	Strength  string     `json:"strength" example:"svag"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	Outcome   string     `json:"outcome,omitempty" example:"passed"`
	DecidedBy *int64     `json:"decided_by,omitempty"`
	DecidedAt *time.Time `json:"decided_at,omitempty"`
}

// GoalEvidence är ett bevis som hör till ett mål.
//
// Beviset sparas i Plan, så det finns kvar även när agenttjänsten är nere.
type GoalEvidence struct {
	ID             int64     `json:"id" example:"1"`
	GoalID         int64     `json:"goal_id" example:"449"`
	VerificationID *int64    `json:"verification_id,omitempty"`
	Type           string    `json:"type" example:"text"`
	Value          string    `json:"value" example:"Köpt 6 glödlampor"`
	Source         string    `json:"source,omitempty" example:"user"`
	CreatedAt      time.Time `json:"created_at"`
}

// GoalVerificationResult är en bedömning av en verifiering.
//
// Resultatet måste godkännas av en människa innan det räknas som
// verifierat - agenten får skapa och revidera resultat, men inte godkänna.
type GoalVerificationResult struct {
	ID             int64      `json:"id" example:"1"`
	GoalID         int64      `json:"goal_id" example:"449"`
	VerificationID *int64     `json:"verification_id,omitempty"`
	Outcome        string     `json:"outcome" example:"passed"`
	Note           string     `json:"note,omitempty"`
	DecidedBy      *int64     `json:"decided_by,omitempty"`
	DecidedAt      *time.Time `json:"decided_at,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
}

// UpdateGoalVerificationRequest är body för att sätta en verifieringsplan.
type UpdateGoalVerificationRequest struct {
	Type      string `json:"type" binding:"required" example:"checklista"`
	Criterion string `json:"criterion" binding:"required" example:"3 av 4 moment avbockade"`
	Strength  string `json:"strength" example:"svag"`
}

// CreateGoalEvidenceRequest är body för att lämna in ett bevis.
type CreateGoalEvidenceRequest struct {
	Type   string `json:"type" binding:"required" example:"text"`
	Value  string `json:"value" binding:"required"`
	Source string `json:"source,omitempty" example:"user"`
}

// PhaseTransitionRequest är body för en fasövergång.
//
// OverrideReason används när en människa tvingar en övergång. Orsaken
// sparas som en avvikelse.
type PhaseTransitionRequest struct {
	To             string `json:"to" binding:"required" example:"plan"`
	OverrideReason string `json:"override_reason,omitempty" example:"vill hinna ikapp"`
}

// CompleteGoalRequest är body för att avsluta ett mål direkt.
type CompleteGoalRequest struct {
	Reason string `json:"reason,omitempty" example:"klart"`
}

// PhaseTransitionResponse svarar på en fasövergång.
type PhaseTransitionResponse struct {
	GoalID         int64  `json:"goal_id" example:"449"`
	From           string `json:"from" example:"define"`
	To             string `json:"to" example:"plan"`
	Forced         bool   `json:"forced" example:"false"`
	CompletionKind string `json:"completion_kind,omitempty" example:"deklarerat"`
}
