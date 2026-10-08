package models

import "time"

// Goal represents a goal in the system
type Goal struct {
	ID       int64      `json:"id" example:"1"`
	Title    string     `json:"title" example:"My Goal"`
	Icon     string     `json:"icon" example:"heroicons:star"`
	StatusID int        `json:"status_id" example:"1"`
	Created  time.Time  `json:"created"`
	Started  *time.Time `json:"started,omitempty"`
	Finished *time.Time `json:"finished,omitempty"`
	Inbox    int        `json:"inbox" example:"0"`
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
}

// UpdateGoalRequest is the request body for updating a goal
type UpdateGoalRequest struct {
	Title    *string `json:"title,omitempty"`
	Icon     *string `json:"icon,omitempty"`
	StatusID *int    `json:"status_id,omitempty"`
	Started  *string `json:"started,omitempty"`
	Finished *string `json:"finished,omitempty"`
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
	Weight      int     `json:"weight" example:"10"`
	ParentTitle *string `json:"parentTitle,omitempty"`
	ParentID    *int64  `json:"parentId,omitempty"`
}

// ErrorResponse represents an API error
type ErrorResponse struct {
	Message string `json:"message" example:"Error description"`
}
