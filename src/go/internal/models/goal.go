package models

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"time"
)

// DueDateFormat är formatet som due-datum skickas över API:t ("YYYY-MM-DD").
const DueDateFormat = "2006-01-02"

// ErrInvalidDueDate betyder att ett due-datum inte kunde tolkas som "YYYY-MM-DD".
var ErrInvalidDueDate = errors.New("invalid due date")

// DueDate är det kalenderdatum då ett mål ska vara klart.
//
// Det är ett eget datum utan klockslag, till skillnad från started/finished
// som är tidsstämplar. Egen typ gör att datumet inte flyttas när tidszonen
// ändras - "2026-11-20" är alltid den 20 november, oavsett var klienten står.
type DueDate struct {
	time.Time
}

// MarshalJSON serialiserar datumet som "YYYY-MM-DD".
func (d DueDate) MarshalJSON() ([]byte, error) {
	return json.Marshal(d.Format(DueDateFormat))
}

// UnmarshalJSON tolkar "YYYY-MM-DD" (eller null).
func (d *DueDate) UnmarshalJSON(data []byte) error {
	if bytes.Equal(bytes.TrimSpace(data), []byte("null")) {
		d.Time = time.Time{}
		return nil
	}

	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}

	t, err := time.Parse(DueDateFormat, s)
	if err != nil {
		return fmt.Errorf("%w: %q", ErrInvalidDueDate, s)
	}
	d.Time = t

	return nil
}

// Goal represents a goal in the system
type Goal struct {
	ID    int64  `json:"id" example:"1"`
	Title string `json:"title" example:"My Goal"`
	// Description är en längre fritext om målet. Används när titeln
	// inte räcker för att beskriva vad som ska uppnås.
	Description string    `json:"description" example:"En längre beskrivning av målet"`
	Icon        string    `json:"icon" example:"heroicons:star"`
	StatusID    int       `json:"status_id" example:"1"`
	Created     time.Time `json:"created"`
	// Started och Finished skickas alltid ut, även när de är null.
	// Med omitempty skulle fältet försvinna ur JSON-svaret och klienten
	// fick undefined i stället för null - vilket gjorde att filter som
	// "started !== null" felaktigt matchade mål utan startdatum.
	Started  *time.Time `json:"started"`
	Finished *time.Time `json:"finished"`
	// Due är det datum målet ska vara klart. Skickas alltid ut, även när
	// det är null - samma resonemang som för started/finished.
	Due  *DueDate `json:"due"`
	Inbox int     `json:"inbox" example:"0"`
	// Eternal marks a goal that never ends (t.ex. "kroppen"). Ett evigt mål
	// kan inte slutföras och markeras alltid som pågående.
	Eternal bool `json:"eternal" example:"false"`
	// HasParent anger om målet har en förälder (det är child i goal_relations).
	// Endast grundmål (utan förälder) listas på /goals.
	HasParent bool `json:"has_parent" example:"false"`
	// Order är målets placering bland grundmål. Grundmål sorteras på /goals
	// efter detta värde (stigande), och det sätts när målen flyttas med
	// piltangenterna.
	Order int `json:"order" example:"0"`
	// Places är de platser där målet kan slutföras. Ett mål kan ha flera
	// platser - då räcker det att vara på någon av dem (OR-semantik).
	// Skickas alltid ut, även när listan är tom.
	Places []Place `json:"places"`
	// RecurrenceRule är upprepningsregeln för ett återkommande mål, som
	// en RRULE-sträng. Null = målet är inte återkommande.
	RecurrenceRule *string `json:"recurrence_rule"`
	// RecurrenceMode är schemaläggningsstrategin: "schedule" (håll
	// schemat) eller "push" (skjuta fram nästa förekomst).
	RecurrenceMode *string `json:"recurrence_mode"`
	// RecurrenceNext är datumet för nästa förekomst som ska slutföras.
	// Null = serien är slut och målet har inget mer att göra.
	RecurrenceNext *DueDate `json:"recurrence_next"`
	// RecurrenceCount är antalet slutförda förekomster.
	RecurrenceCount int `json:"recurrence_count"`
}

// Status represents a goal status
type Status struct {
	ID      int64  `json:"id" example:"1"`
	Name    string `json:"name" example:"define"`
	Label   string `json:"label" example:"Defining"`
	Created string `json:"created,omitempty"`
}

// StatusUpdate represents a status change history entry
type StatusUpdate struct {
	ID             int64  `json:"id"`
	GoalID         int64  `json:"goal_id"`
	FromStatusID   *int64 `json:"from_status_id,omitempty"`
	ToStatusID     int64  `json:"to_status_id"`
	ChangedAt      string `json:"changed_at"`
	FromStatusName string `json:"from_status_name,omitempty"`
	ToStatusName   string `json:"to_status_name,omitempty"`
}

// ChildRelation represents a child goal relation
type ChildRelation struct {
	ChildID int64 `json:"child_id" example:"2"`
	Order   int   `json:"order" example:"0"`
	Weight  int   `json:"weight" example:"10"`
}

// ParentRelation represents a parent goal relation
type ParentRelation struct {
	GoalByParentID Goal `json:"goalByParentId"`
}

// GoalWithRelations combines a goal with its relations
type GoalWithRelations struct {
	Goal
	ChildRelations  []ChildRelation `json:"childRelations"`
	ParentRelations []Goal          `json:"parentRelations"`
}

// GoalRelation binds a child goal to a parent goal.
//
// Order och Weight hör till relationen, inte till målet - samma mål kan
// ligga under olika föräldrar med olika ordning.
type GoalRelation struct {
	ParentID int64 `json:"parent_id"`
	ChildID  int64 `json:"child_id"`
	Order    int   `json:"order"`
	Weight   int   `json:"weight"`
}

// GoalGraphResponse är en användares hela målgraph: målen, relationerna
// mellan dem och beroendena mellan dem.
//
// Klienten hämtar grafen en gång och räknar själv ut barn, föräldrar och
// beroenden - i stället för att fråga servern om varje målsida.
type GoalGraphResponse struct {
	Goals        []Goal           `json:"goals"`
	Relations    []GoalRelation   `json:"relations"`
	Dependencies []GoalDependency `json:"dependencies"`
}

// GoalDependency represents a dependency between goals
type GoalDependency struct {
	ID          int64     `json:"id" example:"1"`
	GoalID      int64     `json:"goal_id" example:"3"`
	DependsOnID int64     `json:"depends_on_id" example:"2"`
	Created     time.Time `json:"created"`
}

// GetGoalResponse is the full response for getting a single goal
type GetGoalResponse struct {
	Goal          GoalWithRelations `json:"goal"`
	Children      []ChildWithDeps   `json:"children"`
	Parents       []Goal            `json:"parents"`
	AllGoals      []Goal            `json:"allGoals"`
	Dependencies  []GoalDependency  `json:"dependencies"`
	DependsOn     []Goal            `json:"dependsOn"`
	Blocking      []Goal            `json:"blocking"`
	StatusHistory []StatusUpdate    `json:"statusHistory"`
}

// ChildWithDeps extends ChildRelation with dependency info and full goal data
type ChildWithDeps struct {
	Goal
	Order     int    `json:"order"`
	Weight    int    `json:"weight"`
	DependsOn []Goal `json:"dependsOn"`
	Blocking  []Goal `json:"blocking"`
}

// CreateGoalRequest is the request body for creating a goal
type CreateGoalRequest struct {
	Title    string `json:"title" binding:"required" example:"New Goal"`
	Icon     string `json:"icon" example:"heroicons:star"`
	StatusID int    `json:"status_id" example:"1"`
	// RecurrenceRule gör målet återkommande direkt ("FREQ=DAILY;INTERVAL=2").
	RecurrenceRule *string `json:"recurrence_rule" example:"FREQ=DAILY;INTERVAL=2"`
	// RecurrenceMode är schemaläggningsstrategin: "schedule" eller "push".
	// Utan den används "schedule".
	RecurrenceMode *string `json:"recurrence_mode" example:"schedule"`
	// StartDate är det datum serien startar på ("YYYY-MM-DD"). Utan den
	// används dagens datum.
	StartDate *string `json:"start_date" example:"2026-01-05"`
}

// NullableTime is a timestamp field in an update request.
//
// Det skiljer på tre tillstånd:
//   - fältet saknas i request-bodyn (Set == false) → lämnas oförändrat
//   - fältet finns med ett värde (Value != nil) → sätts till tidsstämpeln
//   - fältet finns som null (Set == true, Value == nil) → tidsstämpeln rensas
//
// En vanlig *string räcker inte, eftersom JSON null och ett utelämnat fält
// är samma sak för en pekare.
type NullableTime struct {
	Set   bool
	Value *string
}

// UnmarshalJSON parses a timestamp that is either a RFC3339-sträng eller null.
func (n *NullableTime) UnmarshalJSON(data []byte) error {
	n.Set = true
	if bytes.Equal(bytes.TrimSpace(data), []byte("null")) {
		n.Value = nil
		return nil
	}

	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	n.Value = &s
	return nil
}

// UpdateGoalRequest is the request body for updating a goal
type UpdateGoalRequest struct {
	Title *string `json:"title,omitempty"`
	// Description sätter eller rensar målets beskrivning.
	// nil = lämnas oförändrat.
	Description *string      `json:"description,omitempty"`
	Icon        *string      `json:"icon,omitempty"`
	StatusID    *int         `json:"status_id,omitempty"`
	Started     NullableTime `json:"started,omitempty"`
	Finished    NullableTime `json:"finished,omitempty"`
	// Due sätter eller rensar målets due-datum ("YYYY-MM-DD" eller null).
	Due NullableDueDate `json:"due,omitempty"`
	// Eternal sätter eller tar bort evigt-flaggan på ett mål.
	// nil = lämnas oförändrat, true = evigt mål, false = vanligt mål.
	Eternal *bool `json:"eternal,omitempty"`
	// RecurrenceRule sätter eller tar bort målets upprepningsregel.
	// Utan fältet lämnas regeln oförändrad, null tar bort den.
	RecurrenceRule NullableString `json:"recurrence_rule,omitempty"`
	// RecurrenceMode sätter schemaläggningsstrategin ("schedule" eller "push").
	RecurrenceMode NullableString `json:"recurrence_mode,omitempty"`
	// StartDate är det datum serien startar på ("YYYY-MM-DD").
	StartDate *string `json:"start_date,omitempty"`
}

// FinishesGoal returnerar true om requesten försöker markera målet som klart,
// dvs. sätter en finished-tidsstämpel. Att rensa finished (null) räknas
// inte som att slutföra målet.
func (r UpdateGoalRequest) FinishesGoal() bool {
	return r.Finished.Set && r.Finished.Value != nil && *r.Finished.Value != ""
}

// NullableDueDate sätter eller rensar målets due-datum.
//
// Det skiljer på tre tillstånd:
//   - fältet saknas i request-bodyn (Set == false) → lämnas oförändrat
//   - fältet finns med ett värde (Value != nil) → sätts till datumet
//   - fältet finns som null (Set == true, Value == nil) → datumet rensas
type NullableDueDate struct {
	Set   bool
	Value *string
}

// UnmarshalJSON parses a due date that is either a "YYYY-MM-DD"-sträng eller null.
func (n *NullableDueDate) UnmarshalJSON(data []byte) error {
	n.Set = true
	if bytes.Equal(bytes.TrimSpace(data), []byte("null")) {
		n.Value = nil
		return nil
	}

	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	n.Value = &s

	return nil
}

// NullableString sätter eller rensar en textsträng i en update-request.
//
// Det skiljer på tre tillstånd:
//   - fältet saknas i request-bodyn (Set == false) → lämnas oförändrat
//   - fältet finns med ett värde (Value != nil) → sätts till värdet
//   - fältet finns som null (Set == true, Value == nil) → värdet rensas
type NullableString struct {
	Set   bool
	Value *string
}

// UnmarshalJSON parses a string that is either en textsträng eller null.
func (n *NullableString) UnmarshalJSON(data []byte) error {
	n.Set = true
	if bytes.Equal(bytes.TrimSpace(data), []byte("null")) {
		n.Value = nil
		return nil
	}

	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	n.Value = &s

	return nil
}

// CompleteOccurrenceRequest is the request body for completing an occurrence
// of a recurring goal.
type CompleteOccurrenceRequest struct {
	// OccurrenceDate är den förekomst som slutförs ("YYYY-MM-DD"). Utan
	// det slutförs målets aktuella förekomst. Skickas det en förekomst
	// som redan är slutförd ändras ingenting.
	OccurrenceDate string `json:"occurrence_date" example:"2026-01-01"`
	// CompletedOn är det datum slutförandet skedde ("YYYY-MM-DD"). Utan
	// det används dagens datum.
	CompletedOn string `json:"completed_on" example:"2026-01-05"`
}

// UpdateStatusRequest is the request body for updating goal status
type UpdateStatusRequest struct {
	StatusID int `json:"status_id" binding:"required" example:"2"`
}

// CreateRelationRequest is the request body for creating a relation
type CreateRelationRequest struct {
	ChildID  int64 `json:"childId" binding:"required" example:"2"`
	ParentID int64 `json:"parentId" binding:"required" example:"1"`
	Order    int   `json:"order" example:"0"`
	Weight   int   `json:"weight" example:"10"`
}

// UpdateRelationRequest is the request body for updating a relation
type UpdateRelationRequest struct {
	ChildID  int64 `json:"childId" binding:"required"`
	ParentID int64 `json:"parentId" binding:"required"`
	Order    *int  `json:"order,omitempty"`
	Weight   *int  `json:"weight,omitempty"`
}

// ReorderRelationsRequest is the request body for reordering the children of a goal.
// ChildIDs contains the children in their new order.
type ReorderRelationsRequest struct {
	ParentID int64   `json:"parentId" binding:"required"`
	ChildIDs []int64 `json:"childIds" binding:"required"`
}

// ReorderGoalsRequest is the request body for reordering root goals.
// GoalIDs contains the root goals (goals without a parent) in their new order.
type ReorderGoalsRequest struct {
	GoalIDs []int64 `json:"goalIds" binding:"required"`
}

// ReorderPriorityGoalsRequest is the request body for reordering the priority list.
type ReorderPriorityGoalsRequest struct {
	GoalIDs []int64 `json:"goalIds" binding:"required"`
}

// DeleteRelationRequest is the request body for deleting a relation
type DeleteRelationRequest struct {
	ChildID  int64 `json:"childId" binding:"required"`
	ParentID int64 `json:"parentId" binding:"required"`
}

// CreateDependencyRequest is the request body for creating a dependency
type CreateDependencyRequest struct {
	GoalID      int64 `json:"goalId" binding:"required" example:"3"`
	DependsOnID int64 `json:"dependsOnId" binding:"required" example:"2"`
}

// DeleteDependencyRequest is the request body for deleting a dependency
type DeleteDependencyRequest struct {
	GoalID      int64 `json:"goalId" binding:"required"`
	DependsOnID int64 `json:"dependsOnId" binding:"required"`
}

// PrioritizedGoal represents a goal with priority information
type PrioritizedGoal struct {
	Goal
	Weight        int     `json:"weight" example:"10"`
	ParentTitle   *string `json:"parentTitle,omitempty"`
	ParentID      *int64  `json:"parentId,omitempty"`
	PriorityOrder int     `json:"priorityOrder" example:"0"`
}

// ErrorResponse represents an API error
type ErrorResponse struct {
	Message string `json:"message" example:"Error description"`
}
