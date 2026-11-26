package planclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

// Client talks to the Plan API
type Client struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
}

// Goal represents a goal from Plan API
type Goal struct {
	ID          int64      `json:"id"`
	Title       string     `json:"title"`
	Description string     `json:"description,omitempty"`
	Icon        string     `json:"icon,omitempty"`
	StatusID    int        `json:"status_id"`
	Created     time.Time  `json:"created"`
	Started     *time.Time `json:"started,omitempty"`
	Finished    *time.Time `json:"finished,omitempty"`
	Inbox       int        `json:"inbox,omitempty"`
	// Agent är true om målet är markerat för en agent. Bara sådana mål
	// får agenten ta sig an.
	Agent bool `json:"agent,omitempty"`
	// Now betyder att målet ska göras nu.
	Now bool `json:"now,omitempty"`
	// Eternal betyder att målet aldrig slutar.
	Eternal bool `json:"eternal,omitempty"`
	// Due är det datum målet ska vara klart ("YYYY-MM-DD").
	Due *string `json:"due,omitempty"`
	// StartDate är det datum målet får dyka upp i prioriteringslistan.
	StartDate *string `json:"start_date,omitempty"`
	// EstimatedMinutes är den uppskattade tidsåtgången, i minuter.
	EstimatedMinutes *int `json:"estimated_minutes,omitempty"`
	// RecurrenceRule är upprepningsregeln (RRULE). Nil = inte återkommande.
	RecurrenceRule *string `json:"recurrence_rule,omitempty"`
	// RecurrenceMode är schemaläggningsstrategin: "schedule" eller "push".
	RecurrenceMode *string `json:"recurrence_mode,omitempty"`
	// Tags är målets taggar.
	Tags []string `json:"tags,omitempty"`
	// Places är platserna där målet kan slutföras.
	Places []Place `json:"places,omitempty"`
}

// Place är en plats i Plan - till exempel "Hem" eller "Gymmet".
type Place struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
	Icon string `json:"icon,omitempty"`
}

// Körningens tillstånd - de ägs av Plan, agenten läser dem.
const (
	RunStatusQueued          = "queued"
	RunStatusRunning         = "running"
	RunStatusWaitingApproval = "waiting_approval"
	RunStatusSucceeded       = "succeeded"
	RunStatusFailed          = "failed"
	RunStatusCancelled       = "cancelled"
)

// AgentRun is an agent run registered in Plan
type AgentRun struct {
	ID             int64     `json:"id"`
	GoalID         int64     `json:"goal_id"`
	Phase          string    `json:"phase"`
	Status         string    `json:"status"`
	Trigger        string    `json:"trigger,omitempty"`
	Initiator      string    `json:"initiator,omitempty"`
	AgentProfile   string    `json:"agent_profile,omitempty"`
	Summary        string    `json:"summary,omitempty"`
	Result         string    `json:"result,omitempty"`
	ArtifactURL    *string   `json:"artifact_url,omitempty"`
	CostUSD        *float64  `json:"cost_usd,omitempty"`
	ConversationID *string   `json:"conversation_id,omitempty"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// CreateAgentRunEventRequest records a structured event in Plan.
type CreateAgentRunEventRequest struct {
	Type    string         `json:"type"`
	Phase   string         `json:"phase,omitempty"`
	Payload map[string]any `json:"payload,omitempty"`
}

// CreateAgentRunRequest creates a new agent run in Plan
type CreateAgentRunRequest struct {
	GoalID         int64   `json:"goal_id"`
	Phase          string  `json:"phase,omitempty"`
	Status         string  `json:"status,omitempty"`
	Trigger        string  `json:"trigger,omitempty"`
	Initiator      string  `json:"initiator,omitempty"`
	AgentProfile   string  `json:"agent_profile,omitempty"`
	Summary        string  `json:"summary,omitempty"`
	Result         string  `json:"result,omitempty"`
	ArtifactURL    *string `json:"artifact_url,omitempty"`
	ConversationID *string `json:"conversation_id,omitempty"`
	// AgentID identifierar agenten. Den används för att filtrera
	// kostnader per agent.
	AgentID string `json:"agent_id,omitempty"`
	// Provider och Model behövs för att Plan ska kunna räkna fram
	// kostnaden ur prislistan.
	Provider     string `json:"provider,omitempty"`
	Model        string `json:"model,omitempty"`
	ModelVersion string `json:"model_version,omitempty"`
}

// AgentRunUsageRequest rapporterar tokenanvändningen för en körning.
type AgentRunUsageRequest struct {
	AgentID         string     `json:"agent_id,omitempty"`
	Provider        string     `json:"provider,omitempty"`
	Model           string     `json:"model,omitempty"`
	ModelVersion    string     `json:"model_version,omitempty"`
	InputTokens     int        `json:"input_tokens"`
	OutputTokens    int        `json:"output_tokens"`
	CacheReadTokens int        `json:"cache_read_tokens"`
	ToolCalls       int        `json:"tool_calls"`
	ToolNames       []string   `json:"tool_names,omitempty"`
	StartedAt       *time.Time `json:"started_at,omitempty"`
	FinishedAt      *time.Time `json:"finished_at,omitempty"`
}

// UpdateAgentRunRequest updates an existing agent run
type UpdateAgentRunRequest struct {
	Phase          *string    `json:"phase,omitempty"`
	Status         *string    `json:"status,omitempty"`
	AgentProfile   *string    `json:"agent_profile,omitempty"`
	Summary        *string    `json:"summary,omitempty"`
	Result         *string    `json:"result,omitempty"`
	ArtifactURL    *string    `json:"artifact_url,omitempty"`
	CostUSD        *float64   `json:"cost_usd,omitempty"`
	ConversationID *string    `json:"conversation_id,omitempty"`
	FinishedAt     *time.Time `json:"finished_at,omitempty"`
}

// Status represents a goal status
type Status struct {
	ID    int64  `json:"id"`
	Name  string `json:"name"`
	Label string `json:"label"`
}

// UpdateGoalRequest is the request body for updating a goal
//
// Pekare används för att skilja "sätt värdet" från "lämna oförändrat":
// ett fält som inte finns i requesten lämnas som det var.
type UpdateGoalRequest struct {
	Title       *string `json:"title,omitempty"`
	Description *string `json:"description,omitempty"`
	Icon        *string `json:"icon,omitempty"`
	StatusID    *int    `json:"status_id,omitempty"`
	// Now, Eternal och Agent sätter målets operativa flaggor.
	Now     *bool `json:"now,omitempty"`
	Eternal *bool `json:"eternal,omitempty"`
	Agent   *bool `json:"agent,omitempty"`
	// Due och StartDate sätter målets datum ("YYYY-MM-DD").
	Due       *string `json:"due,omitempty"`
	StartDate *string `json:"start_date,omitempty"`
	// EstimatedMinutes sätter den uppskattade tidsåtgången i minuter.
	EstimatedMinutes *int `json:"estimated_minutes,omitempty"`
	// RecurrenceRule sätter upprepningsregeln ("FREQ=WEEKLY;BYDAY=MO").
	RecurrenceRule *string `json:"recurrence_rule,omitempty"`
	// RecurrenceMode sätter schemaläggningsstrategin ("schedule" eller "push").
	RecurrenceMode *string `json:"recurrence_mode,omitempty"`
	// RecurrenceStartDate sätter seriens startdatum ("YYYY-MM-DD").
	RecurrenceStartDate *string `json:"recurrence_start_date,omitempty"`
}

// New creates a new Plan API client
func New() (*Client, error) {
	baseURL := os.Getenv("PLAN_API_URL")
	if baseURL == "" {
		baseURL = "http://localhost:8081"
	}

	apiKey := os.Getenv("PLAN_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("PLAN_API_KEY environment variable is required")
	}

	return &Client{
		baseURL: baseURL,
		apiKey:  apiKey,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}, nil
}

// do makes an authenticated request to the Plan API
func (c *Client) do(ctx context.Context, method, path string, body, result interface{}) error {
	url := c.baseURL + path

	var bodyReader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("failed to marshal request body: %w", err)
		}
		bodyReader = bytes.NewReader(data)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, bodyReader)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	// Plan läser API-nycklar i X-API-Key - Bearer är sessionstoken i Plan
	// och fungerar inte för en agent.
	req.Header.Set("X-API-Key", c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API error %d: %s", resp.StatusCode, string(body))
	}

	if result != nil {
		if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
			return fmt.Errorf("failed to decode response: %w", err)
		}
	}

	return nil
}

// GetGoalsByStatus returns all goals with a specific status
func (c *Client) GetGoalsByStatus(ctx context.Context, statusID int) ([]Goal, error) {
	var goals []Goal
	err := c.do(ctx, http.MethodGet, pathGoals, nil, &goals)
	if err != nil {
		return nil, err
	}

	// Filter by status locally (API doesn't support filtering by status)
	var filtered []Goal
	for _, g := range goals {
		if g.StatusID == statusID {
			filtered = append(filtered, g)
		}
	}

	return filtered, nil
}

// GetGoal returns a single goal by ID
func (c *Client) GetGoal(ctx context.Context, id int64) (*Goal, error) {
	var result struct {
		Goal Goal `json:"goal"`
	}
	err := c.do(ctx, http.MethodGet, fmt.Sprintf(pathGoal, id), nil, &result)
	if err != nil {
		return nil, err
	}
	return &result.Goal, nil
}

// UpdateGoal updates a goal's fields
func (c *Client) UpdateGoal(ctx context.Context, id int64, req UpdateGoalRequest) (*Goal, error) {
	var result struct {
		Goal Goal `json:"goal"`
	}
	err := c.do(ctx, http.MethodPatch, fmt.Sprintf(pathGoal, id), req, &result)
	if err != nil {
		return nil, err
	}
	return &result.Goal, nil
}

// UpdateGoalWithStatus updates title, description, and optionally status
func (c *Client) UpdateGoalWithStatus(ctx context.Context, id int64, title, description *string, statusID int) error {
	req := UpdateGoalRequest{
		Title:       title,
		Description: description,
	}

	// Update goal first
	_, err := c.UpdateGoal(ctx, id, req)
	if err != nil {
		return fmt.Errorf("failed to update goal: %w", err)
	}

	// Then update status
	if statusID > 0 {
		err = c.UpdateGoalStatus(ctx, id, statusID)
		if err != nil {
			return fmt.Errorf("failed to update status: %w", err)
		}
	}

	return nil
}

func (c *Client) UpdateGoalWithPhase(ctx context.Context, id int64, title, description *string, phase string) error {
	if _, err := c.UpdateGoal(ctx, id, UpdateGoalRequest{
		Title:       title,
		Description: description,
	}); err != nil {
		return fmt.Errorf("failed to update goal: %w", err)
	}
	if err := c.UpdateGoalPhase(ctx, id, phase); err != nil {
		return fmt.Errorf("failed to update goal phase: %w", err)
	}
	return nil
}

// UpdateGoalStatus updates a goal's status
//
// Status ändras via en egen endpoint - PATCH mot målet ignorerar
// status_id.
func (c *Client) UpdateGoalStatus(ctx context.Context, id int64, statusID int) error {
	req := struct {
		StatusID int `json:"status_id"`
	}{StatusID: statusID}

	return c.do(ctx, http.MethodPatch, fmt.Sprintf(pathGoalStatus, id), req, nil)
}

// GetStatuses returns all available statuses
func (c *Client) GetStatuses(ctx context.Context) ([]Status, error) {
	var statuses []Status
	err := c.do(ctx, http.MethodGet, pathStatuses, nil, &statuses)
	if err != nil {
		return nil, err
	}
	return statuses, nil
}

// GetPlaces returnerar användarens platser.
//
// Agenten behöver dem för att kunna föreslå platser på ett mål - den
// föreslår bara platser som redan finns.
func (c *Client) GetPlaces(ctx context.Context) ([]Place, error) {
	var places []Place
	if err := c.do(ctx, http.MethodGet, pathPlaces, nil, &places); err != nil {
		return nil, err
	}
	return places, nil
}

// SetGoalTags ersätter målets taggar med de angivna.
//
// API:t skapar taggar som inte finns ännu - så agenten kan föreslå en ny
// tagg. Befintliga taggar ska skickas med, eftersom anropet ersätter
// hela uppsättningen.
func (c *Client) SetGoalTags(ctx context.Context, goalID int64, tags []string) error {
	req := struct {
		Tags []string `json:"tags"`
	}{Tags: tags}

	return c.do(ctx, http.MethodPut, fmt.Sprintf(pathGoalTags, goalID), req, nil)
}

// SetGoalPlaces ersätter målets platser med de angivna.
func (c *Client) SetGoalPlaces(ctx context.Context, goalID int64, placeIDs []int64) error {
	req := struct {
		PlaceIDs []int64 `json:"place_ids"`
	}{PlaceIDs: placeIDs}

	return c.do(ctx, http.MethodPut, fmt.Sprintf(pathGoalPlaces, goalID), req, nil)
}

// Status IDs (from seed data)
const (
	StatusDefine    = 1
	StatusPlan      = 2
	StatusExecute   = 3
	StatusVerify    = 4
	StatusCompleted = 5
)

var goalPhases = map[string]struct{}{
	"define":    {},
	"plan":      {},
	"execute":   {},
	"verify":    {},
	"completed": {},
}

func (c *Client) GetStatusIDByName(ctx context.Context, phase string) (int, error) {
	if _, ok := goalPhases[phase]; !ok {
		return 0, fmt.Errorf("unknown goal phase %q", phase)
	}

	statuses, err := c.GetStatuses(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to get goal statuses: %w", err)
	}
	for _, status := range statuses {
		if status.Name == phase {
			return int(status.ID), nil
		}
	}
	return 0, fmt.Errorf("goal phase %q is not available in Plan", phase)
}

// GetPhaseName ger fasnamnet för ett status-id.
//
// Det är väntet att statusen motsvarar en fas - men ett okänt id ger
// ett fel, inte en tom sträng.
func (c *Client) GetPhaseName(ctx context.Context, statusID int) (string, error) {
	statuses, err := c.GetStatuses(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get goal statuses: %w", err)
	}
	for _, status := range statuses {
		if int(status.ID) == statusID {
			return status.Name, nil
		}
	}
	return "", fmt.Errorf("status %d is not available in Plan", statusID)
}

func (c *Client) UpdateGoalPhase(ctx context.Context, goalID int64, phase string) error {
	statusID, err := c.GetStatusIDByName(ctx, phase)
	if err != nil {
		return err
	}
	return c.UpdateGoalStatus(ctx, goalID, statusID)
}

// GetAgentGoals returns the goals marked for agent work.
//
// Endast markerade mål hämtas - det är agentens enda ingång.
func (c *Client) GetAgentGoals(ctx context.Context) ([]Goal, error) {
	var goals []Goal
	if err := c.do(ctx, http.MethodGet, pathGoals+"?agent=true", nil, &goals); err != nil {
		return nil, err
	}
	return goals, nil
}

// CreateGoalRequest contains the fields needed to create a Plan goal.
type CreateGoalRequest struct {
	Title    string `json:"title"`
	Icon     string `json:"icon,omitempty"`
	StatusID int    `json:"status_id,omitempty"`
}

// CreateRelationRequest connects a child goal to a parent goal.
type CreateRelationRequest struct {
	ChildID  int64 `json:"childId"`
	ParentID int64 `json:"parentId"`
	Order    int   `json:"order"`
	Weight   int   `json:"weight"`
}

// CreateSubgoal creates a goal and attaches it to parentID.
func (c *Client) CreateSubgoal(ctx context.Context, parentID int64, title, description string) (*Goal, error) {
	var goal Goal
	if err := c.do(ctx, http.MethodPost, pathGoals, CreateGoalRequest{
		Title:    title,
		StatusID: StatusDefine,
	}, &goal); err != nil {
		return nil, fmt.Errorf("create subgoal: %w", err)
	}

	if description != "" {
		if _, err := c.UpdateGoal(ctx, goal.ID, UpdateGoalRequest{Description: &description}); err != nil {
			return nil, fmt.Errorf("set subgoal description: %w", err)
		}
	}

	if err := c.do(ctx, http.MethodPost, pathGoalRelations, CreateRelationRequest{
		ChildID:  goal.ID,
		ParentID: parentID,
		Weight:   10,
	}, nil); err != nil {
		return nil, fmt.Errorf("attach subgoal to parent: %w", err)
	}

	return &goal, nil
}

// CreateAgentRun registers that the agent started working on a goal
func (c *Client) CreateAgentRun(ctx context.Context, req CreateAgentRunRequest) (*AgentRun, error) {
	var run AgentRun
	if err := c.do(ctx, http.MethodPost, pathAgentRuns, req, &run); err != nil {
		return nil, err
	}
	return &run, nil
}

// RecordAgentRunUsage reports the token usage of an agent run.
//
// Only the usage is reported - Plan calculates the cost from its versioned
// price list, so every run uses the same prices.
func (c *Client) RecordAgentRunUsage(ctx context.Context, id int64, req AgentRunUsageRequest) error {
	return c.do(ctx, http.MethodPost, fmt.Sprintf(pathAgentRunUsage, id), req, nil)
}

// CreateAgentRunEvent appends a structured event to an AgentRun.
func (c *Client) CreateAgentRunEvent(ctx context.Context, id int64, req CreateAgentRunEventRequest) error {
	return c.do(ctx, http.MethodPost, fmt.Sprintf(pathAgentRunEvents, id), req, nil)
}

// UpdateAgentRun updates phase, status or summary of an agent run
func (c *Client) UpdateAgentRun(ctx context.Context, id int64, req UpdateAgentRunRequest) (*AgentRun, error) {
	var run AgentRun
	if err := c.do(ctx, http.MethodPatch, fmt.Sprintf(pathAgentRun, id), req, &run); err != nil {
		return nil, err
	}
	return &run, nil
}

// GetAgentRuns returns the agent runs registered in Plan
func (c *Client) GetAgentRuns(ctx context.Context, goalID *int64) ([]AgentRun, error) {
	path := pathAgentRuns
	if goalID != nil {
		path = fmt.Sprintf("%s?goal_id=%d", path, *goalID)
	}

	var runs []AgentRun
	if err := c.do(ctx, http.MethodGet, path, nil, &runs); err != nil {
		return nil, err
	}
	return runs, nil
}

// AgentApproval är ett mänskligt beslut i en agentkörning.
type AgentApproval struct {
	ID          int64  `json:"id"`
	AgentRunID  int64  `json:"agent_run_id"`
	GoalID      int64  `json:"goal_id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Status      string `json:"status"`
	Feedback    string `json:"feedback"`
	Consequence string `json:"consequence"`
	// VerificationPlan beskriver hur resultatet ska kontrolleras.
	VerificationPlan string `json:"verification_plan"`
	// Payload innehåller exakt det som föreslogs.
	Payload   map[string]any `json:"payload,omitempty"`
	DecidedAt *time.Time     `json:"decided_at"`
}

// Typer av godkännanden. Typen ligger i godkännandets payload ("kind") så
// att en människa ser vad som föreslogs och att agenten inte blandar
// ihop ett definitionsförslag med ett kodförslag i samma körning.
//
// Typerna deklareras i kontraktet (src/contracts/plan-agent.yaml).
const (
	// ApprovalKindGoalDefinition är ett förslag på hur ett mål ska se ut.
	ApprovalKindGoalDefinition = "goal_definition"
	// ApprovalKindCodeChange är ett förslag på en kodändring.
	ApprovalKindCodeChange = "code_change"
)

// ApprovalKind ger godkännandets typ.
//
// Ett godkännande utan typ är ett äldre förslag och behandlas som ett
// kodförslag - så att förslag som skapats tidigare fortfarande fungerar.
func (a AgentApproval) ApprovalKind() string {
	kind, _ := a.Payload["kind"].(string)
	if kind == "" {
		return ApprovalKindCodeChange
	}
	return kind
}

// CreateAgentApprovalRequest skapar ett godkännande som väntar på beslut.
type CreateAgentApprovalRequest struct {
	Title            string `json:"title"`
	Description      string `json:"description"`
	Consequence      string `json:"consequence"`
	VerificationPlan string `json:"verification_plan"`
	// Payload innehåller exakt det som föreslogs.
	Payload map[string]any `json:"payload,omitempty"`
}

// CreateAgentApproval registrerar ett förslag som väntar på ett mänskligt beslut.
func (c *Client) CreateAgentApproval(ctx context.Context, runID int64, req CreateAgentApprovalRequest) (*AgentApproval, error) {
	var approval AgentApproval
	if err := c.do(ctx, http.MethodPost, fmt.Sprintf(pathAgentRunApprovals, runID), req, &approval); err != nil {
		return nil, err
	}
	return &approval, nil
}

// GetAgentApprovals returnerar godkännandena för en körning.
func (c *Client) GetAgentApprovals(ctx context.Context, runID int64) ([]AgentApproval, error) {
	var approvals []AgentApproval
	if err := c.do(ctx, http.MethodGet, fmt.Sprintf(pathAgentRunApprovals, runID), nil, &approvals); err != nil {
		return nil, err
	}
	return approvals, nil
}
