// Package coder driver en kodändring från förslag till verifierat resultat.
//
// Flödet är: agenten föreslår en ändring, en människa godkänner den,
// ändringen utförs i en isolerad workspace och resultatet verifieras med
// projektets egna tester. Först när människan sagt ja ändras något.
package coder

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"plan-agent/internal/executor"
	"plan-agent/internal/planclient"
	"plan-agent/internal/profile"
	"plan-agent/internal/proposer"
	"plan-agent/internal/runusage"
	"plan-agent/internal/workspace"
)

// ProposerWithUsage kan ta emot en recorder för tokenanvändning.
//
// Det är ett valfritt interface - en proposer som inte kan samla
// användning fungerar ändå.
type ProposerWithUsage interface {
	WithUsage(recorder *runusage.Recorder)
}

// Proposer föreslår kodändringar.
type Proposer interface {
	Propose(ctx context.Context, ws *workspace.Workspace, goalTitle, goalDescription string) (*proposer.Proposal, error)
	FixTest(ctx context.Context, ws *workspace.Workspace, goalTitle, goalDescription string, proposal *proposer.Proposal, failure string) (string, error)
	Revise(ctx context.Context, ws *workspace.Workspace, goalTitle, goalDescription string, proposal *proposer.Proposal, failure string) (*proposer.Proposal, error)
}

// Coder driver en kodändring genom hela flödet.
type Coder struct {
	planClient *planclient.Client
	proposer   Proposer
	// usage samlar tokenanvändningen för körningen. Den kan vara nil.
	usage *runusage.Recorder
	// repoPath är repot som kodändringen görs i.
	repoPath string
	// maxDiffLines begränsar hur stor del av diffen som sparas i historiken.
	maxDiffLines int
}

// Config konfigurerar codern.
type Config struct {
	// RepoPath är repot som ska ändras.
	RepoPath string
	// MaxDiffLines begränsar diffen som sparas i historiken.
	MaxDiffLines int
	// Usage samlar tokenanvändningen för körningen. Den kan vara nil.
	Usage *runusage.Recorder
}

// Usage returnerar den samlade tokenanvändningen för körningen.
//
// Workern använder den för att rapportera kostnaden till Plan.
func (c *Coder) Usage() runusage.Snapshot {
	if c.usage == nil {
		return runusage.Snapshot{}
	}
	return c.usage.Snapshot()
}

// New skapar en coder.
func New(planClient *planclient.Client, proposer Proposer, cfg Config) *Coder {
	maxLines := cfg.MaxDiffLines
	if maxLines <= 0 {
		maxLines = 200
	}

	coder := &Coder{
		planClient:   planClient,
		proposer:     proposer,
		repoPath:     cfg.RepoPath,
		maxDiffLines: maxLines,
	}

	// Proposern samlar användningen när den kan det.
	if aware, ok := proposer.(ProposerWithUsage); ok && cfg.Usage != nil {
		aware.WithUsage(cfg.Usage)
		coder.usage = cfg.Usage
	}

	return coder
}

// Process tar hand om ett mål som är redo för kodändring.
//
// Körningen måste vara godkänd av en människa innan något ändras. Är den
// inte godkänd avbryts körningen utan att något skrivs.
func (c *Coder) Process(ctx context.Context, run planclient.AgentRun, goal planclient.Goal) error {
	log.Printf("Coder processing goal #%d: %s", goal.ID, goal.Title)

	// Profilen avgör om körningen får röra kod. Det är en teknisk gräns -
	// inte något modellen kan övertala.
	policy, err := profile.PolicyFor(run.AgentProfile)
	if err != nil {
		return fmt.Errorf("failed to resolve agent profile: %w", err)
	}
	if !policy.MayChangeCode {
		log.Printf("Profile %s may not change code - skipping goal #%d", policy.Name, goal.ID)
		return nil
	}

	approvals, err := c.planClient.GetAgentApprovals(ctx, run.ID)
	if err != nil {
		return fmt.Errorf("failed to check approvals: %w", err)
	}

	approved := findApproved(approvals)
	if approved != nil {
		return c.execute(ctx, run, goal, approved)
	}

	// Ett förslag som redan väntar på beslut ska inte ersättas av ett nytt.
	if findPending(approvals) != nil {
		log.Printf("Goal #%d already has a proposal waiting for a decision", goal.ID)
		return nil
	}

	// Ett avslaget förslag ska inte köras om - det är människan som
	// avgör om agenten får försöka igen.
	if findRejected(approvals) != nil {
		log.Printf("Goal #%d has a rejected proposal - waiting for a human decision", goal.ID)
		return nil
	}

	return c.propose(ctx, run, goal)
}

// propose tar fram ett förslag och lämnar det till en människa.
//
// Förslaget verifieras först med sitt eget testkommando. Det ger två
// saker: modellen får chans att rätta ett trasigt förslag, och människan
// ser om förslaget faktiskt håller när det lämnas över.
func (c *Coder) propose(ctx context.Context, run planclient.AgentRun, goal planclient.Goal) error {
	ws, err := workspace.New(c.repoPath)
	if err != nil {
		return fmt.Errorf("failed to create workspace: %w", err)
	}
	defer ws.Cleanup()

	proposal, err := c.proposer.Propose(ctx, ws, goal.Title, goal.Description)
	if err != nil {
		c.fail(ctx, run.ID, err)
		return fmt.Errorf("failed to propose change: %w", err)
	}

	verification := c.verifyProposal(ctx, proposal)
	if !verification.passed {
		log.Printf("Proposal for goal #%d failed its own verification - asking the model to fix the test", goal.ID)

		// Först provas ett nytt testkommando. Innehållet behålls - det är
		// oftast testet som är för strikt.
		if command, err := c.proposer.FixTest(ctx, ws, goal.Title, goal.Description, proposal, verification.output); err != nil {
			log.Printf("Failed to get a new test command for goal #%d: %v", goal.ID, err)
		} else if command != "" {
			proposal.TestCommand = command
			verification = c.verifyProposal(ctx, proposal)
		}
	}

	if !verification.passed {
		log.Printf("Proposal for goal #%d still fails - asking the model to revise the whole proposal", goal.ID)

		revised, err := c.proposer.Revise(ctx, ws, goal.Title, goal.Description, proposal, verification.output)
		if err != nil {
			log.Printf("Failed to revise proposal for goal #%d: %v", goal.ID, err)
		} else {
			proposal = revised
			verification = c.verifyProposal(ctx, proposal)
		}
	}

	// Förslaget sparas i godkännandet, så att exakt det som människan
	// godkänner är det som appliceras senare.
	payload := proposalPayload(proposal)
	payload["pre_verification"] = map[string]any{
		"passed": verification.passed,
		"output": truncate(verification.output, c.maxDiffLines),
	}

	_, err = c.planClient.CreateAgentApproval(ctx, run.ID, planclient.CreateAgentApprovalRequest{
		Title:            proposal.Title,
		Description:      proposalDescription(proposal, verification),
		Consequence:      describeConsequence(proposal),
		VerificationPlan: proposal.TestCommand,
		Payload:          payload,
	})
	if err != nil {
		c.fail(ctx, run.ID, err)
		return fmt.Errorf("failed to create approval: %w", err)
	}

	c.record(ctx, run.ID, planclient.EventApprovalRequested, map[string]any{
		"title":             proposal.Title,
		"changed_files":     changedFiles(proposal),
		"verification_plan": proposal.TestCommand,
		"pre_verified":      verification.passed,
	})

	// Körningen väntar på en människa - statusen säger det tydligt.
	status := "blocked"
	summary := "Förslag inväntar mänskligt godkännande: " + proposal.Title
	if _, err := c.planClient.UpdateAgentRun(ctx, run.ID, planclient.UpdateAgentRunRequest{
		Status:  &status,
		Summary: &summary,
	}); err != nil {
		return fmt.Errorf("failed to mark run as blocked: %w", err)
	}

	log.Printf("Goal #%d has a proposal waiting for approval: %s", goal.ID, proposal.Title)
	return nil
}

// execute utför den godkända ändringen och verifierar den.
//
// Ändringen kommer från godkännandet - inte från en ny modellkörning.
// Därför är det garanterat att det som appliceras är det en människa
// godkände.
func (c *Coder) execute(ctx context.Context, run planclient.AgentRun, goal planclient.Goal, approval *planclient.AgentApproval) error {
	changes, testCommand, err := changesFromApproval(approval)
	if err != nil {
		c.fail(ctx, run.ID, err)
		return fmt.Errorf("failed to read approved change: %w", err)
	}

	// Körningen går från väntan till utförande.
	if err := c.startExecuting(ctx, run, goal); err != nil {
		return err
	}

	exec, err := executor.New(c.repoPath)
	if err != nil {
		c.fail(ctx, run.ID, err)
		return fmt.Errorf("failed to create executor: %w", err)
	}
	defer exec.Cleanup()

	result, err := exec.Execute(ctx, executor.Request{
		Changes:     changes,
		TestCommand: testCommand,
	})
	if err != nil {
		c.fail(ctx, run.ID, err)
		return fmt.Errorf("failed to execute change: %w", err)
	}

	c.record(ctx, run.ID, planclient.EventArtifact, map[string]any{
		"kind":          "diff",
		"changed_files": result.ChangedFiles,
		"diff":          result.DiffPreview(c.maxDiffLines),
	})

	c.record(ctx, run.ID, planclient.EventVerification, map[string]any{
		"tests_passed": result.TestsPassed,
		"test_command": testCommand,
		"test_output":  truncate(result.TestOutput, c.maxDiffLines),
		"test_error":   result.TestError,
	})

	if !result.TestsPassed {
		c.fail(ctx, run.ID, fmt.Errorf("testerna gick inte igenom"))
		return fmt.Errorf("verification failed for goal #%d", goal.ID)
	}

	// Den godkända ändringen landar på en egen branch - huvudgrenen rörs
	// aldrig, en människa väljer om den ska mergeas.
	//
	// Branchen bär körningens nummer, så varje försök får sin egen
	// branch att granska. Utan det skulle ett nytt försök krocka med
	// den förra branchen i repot.
	branch := fmt.Sprintf("agent/goal-%d-run-%d", goal.ID, run.ID)
	if err := exec.Commit(branch, commitMessage(goal, result)); err != nil {
		if !errors.Is(err, workspace.ErrNotPublished) {
			c.fail(ctx, run.ID, err)
			return fmt.Errorf("failed to publish the change: %w", err)
		}

		// Ändringen är gjord och verifierad men kom inte ut till repot.
		// Det är inte ett fel i ändringen - det är en sak som behöver
		// nämnas tydligt i historiken.
		log.Printf("Goal #%d was verified but could not be published: %v", goal.ID, err)
		c.record(ctx, run.ID, planclient.EventArtifact, map[string]any{
			"kind":    "branch",
			"branch":  branch,
			"publish": "failed",
			"error":   err.Error(),
		})
	} else {
		result.Branch = branch
		c.record(ctx, run.ID, planclient.EventArtifact, map[string]any{
			"kind":   "branch",
			"branch": branch,
		})
	}

	status := "done"
	phase := "verify"
	resultText := fmt.Sprintf("Ändringen är gjord på branchen %s och testerna gick igenom", result.Branch)
	if result.Branch == "" {
		resultText = "Ändringen är gjord och testerna gick igenom, men branchen kunde inte publiceras"
	}
	finishedAt := time.Now().UTC()
	if _, err := c.planClient.UpdateAgentRun(ctx, run.ID, planclient.UpdateAgentRunRequest{
		Status:     &status,
		Phase:      &phase,
		Summary:    &resultText,
		Result:     &resultText,
		FinishedAt: &finishedAt,
	}); err != nil {
		return fmt.Errorf("failed to update agent run: %w", err)
	}

	// Målet går till verifiering - en människa avgör om det är klart.
	if err := c.planClient.UpdateGoalPhase(ctx, goal.ID, phase); err != nil {
		return fmt.Errorf("failed to move goal to verification: %w", err)
	}

	c.record(ctx, run.ID, planclient.EventRunCompleted, map[string]any{
		"changed_files": result.ChangedFiles,
		"branch":        result.Branch,
	})

	c.reportUsage(ctx, run.ID)

	log.Printf("Goal #%d executed and verified", goal.ID)
	return nil
}

// startExecuting markerar att körningen börjar utföra den godkända
// ändringen.
func (c *Coder) startExecuting(ctx context.Context, run planclient.AgentRun, goal planclient.Goal) error {
	phase := "execute"
	status := "running"
	summary := "Agenten utför den godkända ändringen"

	if _, err := c.planClient.UpdateAgentRun(ctx, run.ID, planclient.UpdateAgentRunRequest{
		Phase:   &phase,
		Status:  &status,
		Summary: &summary,
	}); err != nil {
		return fmt.Errorf("failed to start execution: %w", err)
	}

	if err := c.planClient.UpdateGoalPhase(ctx, goal.ID, phase); err != nil {
		return fmt.Errorf("failed to move goal to execution: %w", err)
	}

	c.record(ctx, run.ID, planclient.EventPhaseChanged, map[string]any{
		"phase":  phase,
		"status": status,
	})
	return nil
}

// approvedProposal ger det senaste godkända förslaget, om det finns.
func (c *Coder) approvedProposal(ctx context.Context, runID int64) (*planclient.AgentApproval, error) {
	approvals, err := c.planClient.GetAgentApprovals(ctx, runID)
	if err != nil {
		return nil, err
	}

	return findApproved(approvals), nil
}

// findApproved ger det godkända förslaget, om det finns.
func findApproved(approvals []planclient.AgentApproval) *planclient.AgentApproval {
	for i := range approvals {
		if approvals[i].Status == "approved" {
			return &approvals[i]
		}
	}
	return nil
}

// findPending ger det förslag som fortfarande väntar på beslut.
func findPending(approvals []planclient.AgentApproval) *planclient.AgentApproval {
	for i := range approvals {
		if approvals[i].Status == "pending" {
			return &approvals[i]
		}
	}
	return nil
}

// findRejected ger det förslag som en människa har avslagit.
func findRejected(approvals []planclient.AgentApproval) *planclient.AgentApproval {
	for i := range approvals {
		if approvals[i].Status == "rejected" {
			return &approvals[i]
		}
	}
	return nil
}

// proposalPayload gör om ett förslag till något som kan sparas i Plan.
func proposalPayload(proposal *proposer.Proposal) map[string]any {
	changes := make([]map[string]any, 0, len(proposal.Changes))
	for _, change := range proposal.Changes {
		changes = append(changes, map[string]any{
			"path":    change.Path,
			"content": change.Content,
			"reason":  change.Reason,
		})
	}

	return map[string]any{
		"changes":      changes,
		"test_command": proposal.TestCommand,
	}
}

// changesFromApproval läser de godkända ändringarna ur ett godkännande.
func changesFromApproval(approval *planclient.AgentApproval) ([]executor.Change, string, error) {
	if approval == nil {
		return nil, "", fmt.Errorf("godkännandet saknas")
	}
	if approval.Payload == nil {
		return nil, "", fmt.Errorf("godkännandet saknar det föreslagna innehållet")
	}

	rawChanges, ok := approval.Payload["changes"].([]any)
	if !ok || len(rawChanges) == 0 {
		return nil, "", fmt.Errorf("godkännandet innehåller inga ändringar")
	}

	changes := make([]executor.Change, 0, len(rawChanges))
	for _, raw := range rawChanges {
		entry, ok := raw.(map[string]any)
		if !ok {
			return nil, "", fmt.Errorf("en ändring i godkännandet har fel format")
		}

		path, _ := entry["path"].(string)
		if path == "" {
			return nil, "", fmt.Errorf("en ändring i godkännandet saknar sökväg")
		}
		content, _ := entry["content"].(string)
		reason, _ := entry["reason"].(string)

		changes = append(changes, executor.Change{
			Path:    path,
			Content: content,
			Reason:  reason,
		})
	}

	testCommand, _ := approval.Payload["test_command"].(string)
	if testCommand == "" {
		testCommand = approval.VerificationPlan
	}

	return changes, testCommand, nil
}

// record skriver en händelse i körningens historik.
func (c *Coder) record(ctx context.Context, runID int64, eventType string, payload map[string]any) {
	err := c.planClient.CreateAgentRunEvent(ctx, runID, planclient.CreateAgentRunEventRequest{
		Type:    eventType,
		Payload: payload,
	})
	if err != nil {
		log.Printf("Failed to record %s event for agent run %d: %v", eventType, runID, err)
	}
}

// fail markerar körningen som misslyckad.
// reportUsage rapporterar tokenanvändningen för körningen till Plan.
//
// Plan räknar fram kostnaden. Ett fel här får inte stoppa körningen -
// kostnaden kan rapporteras i efterhand.
func (c *Coder) reportUsage(ctx context.Context, runID int64) {
	if c.usage == nil {
		return
	}

	snapshot := c.usage.Snapshot()
	startedAt := snapshot.StartedAt.UTC()
	finishedAt := snapshot.FinishedAt.UTC()

	if err := c.planClient.RecordAgentRunUsage(ctx, runID, planclient.AgentRunUsageRequest{
		Provider:        snapshot.Provider,
		Model:           snapshot.Model,
		ModelVersion:    snapshot.ModelVersion,
		InputTokens:     snapshot.InputTokens,
		OutputTokens:    snapshot.OutputTokens,
		CacheReadTokens: snapshot.CacheReadTokens,
		ToolCalls:       snapshot.ToolCalls,
		ToolNames:       snapshot.ToolNames,
		StartedAt:       &startedAt,
		FinishedAt:      &finishedAt,
	}); err != nil {
		log.Printf("Failed to report usage for run %d: %v", runID, err)
	}
}

func (c *Coder) fail(ctx context.Context, runID int64, cause error) {
	status := "failed"
	summary := "Kodändringen misslyckades: " + cause.Error()
	result := "Misslyckades: " + cause.Error()
	finishedAt := time.Now().UTC()

	if _, err := c.planClient.UpdateAgentRun(ctx, runID, planclient.UpdateAgentRunRequest{
		Status:     &status,
		Summary:    &summary,
		Result:     &result,
		FinishedAt: &finishedAt,
	}); err != nil {
		log.Printf("Failed to mark agent run %d as failed: %v", runID, err)
	}

	c.record(ctx, runID, planclient.EventRunFailed, map[string]any{
		"error": cause.Error(),
	})
}

// commitMessage beskriver ändringen i committen.
func commitMessage(goal planclient.Goal, result *executor.Result) string {
	return fmt.Sprintf("%s\n\nÄndringen är gjord av agenten för mål #%d.\nÄndrade filer: %v",
		goal.Title, goal.ID, result.ChangedFiles)
}

// changedFiles listar filerna i ett förslag.
func changedFiles(proposal *proposer.Proposal) []string {
	files := make([]string, 0, len(proposal.Changes))
	for _, change := range proposal.Changes {
		files = append(files, change.Path)
	}
	return files
}

// verification är utfallet av en förslagsverifiering.
type verification struct {
	// passed säger om testkommandot gick igenom.
	passed bool
	// output är utdatan från testkörningen.
	output string
}

// verifyProposal applicerar förslaget i en isolerad kopia och kör dess
// eget testkommando.
//
// Ingenting rapporteras till Plan här - verifieringen är bara till för att
// fånga trasiga förslag innan en människa läser dem.
func (c *Coder) verifyProposal(ctx context.Context, proposal *proposer.Proposal) verification {
	if proposal.TestCommand == "" {
		return verification{passed: true}
	}

	exec, err := executor.New(c.repoPath)
	if err != nil {
		return verification{passed: false, output: err.Error()}
	}
	defer exec.Cleanup()

	result, err := exec.Execute(ctx, executor.Request{
		Changes:     proposal.Changes,
		TestCommand: proposal.TestCommand,
	})
	if err != nil {
		return verification{passed: false, output: err.Error()}
	}

	output := result.TestOutput
	if !result.TestsPassed && result.TestError != "" {
		output = result.TestError + "\n" + output
	}

	return verification{passed: result.TestsPassed, output: output}
}

// proposalDescription beskriver förslaget och hur det verifierades.
func proposalDescription(proposal *proposer.Proposal, result verification) string {
	description := proposal.Description
	if description == "" {
		description = proposal.Title
	}

	if result.passed {
		return description + "\n\nFörslaget har verifierats med sitt eget testkommando."
	}

	return description + "\n\nVarning: förslagets eget testkommando gick inte igenom:\n" + truncate(result.output, 20)
}

// describeConsequence beskriver vad förslaget innebär.
func describeConsequence(proposal *proposer.Proposal) string {
	files := changedFiles(proposal)
	if len(files) == 0 {
		return "Inga filer ändras."
	}
	return fmt.Sprintf("%d fil(er) ändras: %v", len(files), files)
}

// truncate kortar av en lång text till ett antal rader.
func truncate(text string, maxLines int) string {
	if maxLines <= 0 {
		return text
	}

	lines := strings.Split(text, "\n")
	if len(lines) <= maxLines {
		return text
	}

	return strings.Join(lines[:maxLines], "\n") + "\n...(förkortat)"
}
