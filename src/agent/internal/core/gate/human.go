package gate

import (
	"context"
	"fmt"
	"strings"
	"time"

	"plan-agent/internal/core/checkpoint"
)

// HumanGate requires human approval to pass
type HumanGate struct {
	viewerURL     string
	approvalCh    chan *HumanDecision
	timeout       time.Duration
	onReviewReady func(cp *checkpoint.Checkpoint)
	onFeedback    func(cp *checkpoint.Checkpoint, feedback string)
}

// HumanDecision represents a human's decision on a checkpoint
type HumanDecision struct {
	CheckpointID string
	Approved     bool
	Feedback     string
	Changes      []ChangeRequest
}

// ChangeRequest represents a requested change
type ChangeRequest struct {
	Description string
	Priority    string
}

// HumanGateOption is a functional option for HumanGate
type HumanGateOption func(*HumanGate)

func WithViewerURL(url string) HumanGateOption {
	return func(g *HumanGate) { g.viewerURL = url }
}

func WithTimeout(timeout time.Duration) HumanGateOption {
	return func(g *HumanGate) { g.timeout = timeout }
}

func WithReviewCallback(cb func(*checkpoint.Checkpoint)) HumanGateOption {
	return func(g *HumanGate) { g.onReviewReady = cb }
}

func WithFeedbackCallback(cb func(*checkpoint.Checkpoint, string)) HumanGateOption {
	return func(g *HumanGate) { g.onFeedback = cb }
}

// NewHumanGate creates a new human gate
func NewHumanGate(opts ...HumanGateOption) *HumanGate {
	g := &HumanGate{
		approvalCh: make(chan *HumanDecision, 1),
		timeout:    30 * time.Minute,
	}
	for _, opt := range opts {
		opt(g)
	}
	return g
}

// Gate interface implementation
func (g *HumanGate) Type() GateType { return GateTypeHuman }
func (g *HumanGate) Name() string   { return "Human Review Gate" }
func (g *HumanGate) Required() bool { return true }

func (g *HumanGate) Description() string {
	return "Requires human approval after testing the feature manually"
}

// Run executes the human gate check (blocking)
func (g *HumanGate) Run(ctx context.Context, cp *checkpoint.Checkpoint) (*checkpoint.GateResult, error) {
	result := checkpoint.NewGateResult(string(GateTypeHuman))
	result.Status = checkpoint.GateStatusRunning

	if g.onReviewReady != nil {
		g.onReviewReady(cp)
	}

	select {
	case decision := <-g.approvalCh:
		return g.handleDecision(cp, decision, result)
	case <-ctx.Done():
		result.Fail("Human review cancelled", nil)
		return result, ctx.Err()
	case <-time.After(g.timeout):
		result.Fail("Human review timed out", nil)
		return result, fmt.Errorf("timeout waiting for human review")
	}
}

// RunAsync starts the human review but returns immediately
func (g *HumanGate) RunAsync(cp *checkpoint.Checkpoint) {
	if g.onReviewReady != nil {
		g.onReviewReady(cp)
	}
}

// AwaitDecision waits for human decision with timeout
func (g *HumanGate) AwaitDecision(ctx context.Context, cp *checkpoint.Checkpoint, timeout time.Duration) (*checkpoint.GateResult, error) {
	result := checkpoint.NewGateResult(string(GateTypeHuman))
	result.Status = checkpoint.GateStatusRunning

	select {
	case decision := <-g.approvalCh:
		return g.handleDecision(cp, decision, result)
	case <-ctx.Done():
		result.Fail("Human review cancelled", nil)
		return result, ctx.Err()
	case <-time.After(timeout):
		result.Fail("Human review timed out", nil)
		return result, fmt.Errorf("timeout waiting for human review")
	}
}

// SubmitDecision submits a human decision
func (g *HumanGate) SubmitDecision(decision *HumanDecision) {
	select {
	case g.approvalCh <- decision:
	default:
	}
}

func (g *HumanGate) handleDecision(cp *checkpoint.Checkpoint, decision *HumanDecision, result *checkpoint.GateResult) (*checkpoint.GateResult, error) {
	if decision.Approved {
		result.Pass("Human approved")
		return result, nil
	}

	if g.onFeedback != nil {
		g.onFeedback(cp, decision.Feedback)
	}

	result.Fail("Human requested changes", nil)
	result.Summary = decision.Feedback
	return result, nil
}

// SimpleHumanGate is a simpler version that blocks on stdin
type SimpleHumanGate struct{}

func NewSimpleHumanGate() *SimpleHumanGate { return &SimpleHumanGate{} }

func (g *SimpleHumanGate) Type() GateType      { return GateTypeHuman }
func (g *SimpleHumanGate) Name() string        { return "Human Review Gate" }
func (g *SimpleHumanGate) Required() bool      { return true }
func (g *SimpleHumanGate) Description() string { return "Requires human approval" }

func (g *SimpleHumanGate) Run(ctx context.Context, cp *checkpoint.Checkpoint) (*checkpoint.GateResult, error) {
	result := checkpoint.NewGateResult(string(GateTypeHuman))
	result.Status = checkpoint.GateStatusRunning

	sep := strings.Repeat("=", 60)
	fmt.Printf("\n%s\nHUMAN REVIEW REQUIRED\n%s\n\n", sep, sep)
	fmt.Printf("Checkpoint: %s\nDescription: %s\n\nTasks:\n", cp.Name, cp.Description)
	for i, task := range cp.Tasks {
		fmt.Printf("  %d. %s\n", i+1, task)
	}
	fmt.Printf("\n%s\nPlease test the feature and decide:\n  [y/Y] - Approve\n  [n/N] - Request changes\n  [s/S] - Skip\n> ", strings.Repeat("-", 60))

	return result, nil
}
