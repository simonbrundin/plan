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
}

// Körningens tillstånd - de ägs av Plan, agenten läser dem.
const (
	RunStatusQueued    = "queued"
	RunStatusRunning   = "running"
	RunStatusBlocked   = "blocked"
	RunStatusDone      = "done"
	RunStatusFailed    = "failed"
	RunStatusCancelled = "cancelled"
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
type UpdateGoalRequest struct {
	Title       *string `json:"title,omitempty"`
	Description *string `json:"description,omitempty"`
	Icon        *string `json:"icon,omitempty"`
	StatusID    *int    `json:"status_id,omitempty"`
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
	err := c.do(ctx, http.MethodGet, "/api/v1/goals", nil, &goals)
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
	err := c.do(ctx, http.MethodGet, fmt.Sprintf("/api/v1/goals/%d", id), nil, &result)
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
	err := c.do(ctx, http.MethodPatch, fmt.Sprintf("/api/v1/goals/%d", id), req, &result)
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

	return c.do(ctx, http.MethodPatch, fmt.Sprintf("/api/v1/goals/%d/status", id), req, nil)
}

// GetStatuses returns all available statuses
func (c *Client) GetStatuses(ctx context.Context) ([]Status, error) {
	var statuses []Status
	err := c.do(ctx, http.MethodGet, "/api/v1/statuses", nil, &statuses)
	if err != nil {
		return nil, err
	}
	return statuses, nil
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
	if err := c.do(ctx, http.MethodGet, "/api/v1/goals?agent=true", nil, &goals); err != nil {
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
	if err := c.do(ctx, http.MethodPost, "/api/v1/goals", CreateGoalRequest{
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

	if err := c.do(ctx, http.MethodPost, "/api/v1/goals/relations", CreateRelationRequest{
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
	if err := c.do(ctx, http.MethodPost, "/api/v1/agent-runs", req, &run); err != nil {
		return nil, err
	}
	return &run, nil
}

// RecordAgentRunUsage reports the token usage of an agent run.
//
// Only the usage is reported - Plan calculates the cost from its versioned
// price list, so every run uses the same prices.
func (c *Client) RecordAgentRunUsage(ctx context.Context, id int64, req AgentRunUsageRequest) error {
	return c.do(ctx, http.MethodPost, fmt.Sprintf("/api/v1/agent-runs/%d/usage", id), req, nil)
}

// CreateAgentRunEvent appends a structured event to an AgentRun.
func (c *Client) CreateAgentRunEvent(ctx context.Context, id int64, req CreateAgentRunEventRequest) error {
	return c.do(ctx, http.MethodPost, fmt.Sprintf("/api/v1/agent-runs/%d/events", id), req, nil)
}

// UpdateAgentRun updates phase, status or summary of an agent run
func (c *Client) UpdateAgentRun(ctx context.Context, id int64, req UpdateAgentRunRequest) (*AgentRun, error) {
	var run AgentRun
	if err := c.do(ctx, http.MethodPatch, fmt.Sprintf("/api/v1/agent-runs/%d", id), req, &run); err != nil {
		return nil, err
	}
	return &run, nil
}

// GetAgentRuns returns the agent runs registered in Plan
func (c *Client) GetAgentRuns(ctx context.Context, goalID *int64) ([]AgentRun, error) {
	path := "/api/v1/agent-runs"
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
	if err := c.do(ctx, http.MethodPost, fmt.Sprintf("/api/v1/agent-runs/%d/approvals", runID), req, &approval); err != nil {
		return nil, err
	}
	return &approval, nil
}

// GetAgentApprovals returnerar godkännandena för en körning.
func (c *Client) GetAgentApprovals(ctx context.Context, runID int64) ([]AgentApproval, error) {
	var approvals []AgentApproval
	if err := c.do(ctx, http.MethodGet, fmt.Sprintf("/api/v1/agent-runs/%d/approvals", runID), nil, &approvals); err != nil {
		return nil, err
	}
	return approvals, nil
}
