// Package agentworker tar sig an de mål som är markerade för agentarbete.
//
// Arbetsflödet är avsiktligt litet: agenten specificerar målet, skriver
// den egna definitionen tillbaka i Plan och sätter målets status till "plan".
// Resten - planering, utförande, verifiering - är senare faser.
package agentworker

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	"plan/agent/internal/definer"
	"plan/agent/internal/goaldefinition"
	"plan/agent/internal/planclient"
	"plan/agent/internal/profile"
	"plan/agent/internal/runusage"
)

// Faser och tillstånd speglar agent_runs-tabellen i Plan.
const (
	PhaseDefine    = "define"
	PhasePlan      = "plan"
	PhaseExecute   = "execute"
	PhaseVerify    = "verify"
	PhaseCompleted = "completed"

	StatusRunning   = "running"
	StatusSucceeded = "succeeded"
	StatusFailed    = "failed"
	// StatusWaitingApproval betyder att körningen väntar på en människa.
	StatusWaitingApproval = "waiting_approval"
	// StatusQueued betyder att körningen väntar på att få ett Job.
	StatusQueued = "queued"
	// StatusCancelled betyder att en människa avbröt körningen.
	StatusCancelled = planclient.RunStatusCancelled
)

// ErrFinished betyder att målet är färdigt och workern kan avslutas.
var ErrFinished = errors.New("goal finished")

// phaseBlockedError betyder att fasmaskinen stoppade en övergång.
//
// Det är inte ett fel i agenten - det är svaret på att villkoret för
// övergången inte är uppfyllt. Agenten ska då stanna och låta en människa
// bestämma, i stället för att gå vidare på egen hand.
type phaseBlockedError struct {
	// Reason är skälkoden, till exempel "definition_not_approved".
	Reason string
	// Message förklarar skälet.
	Message string
}

// Error beskriver varför bytet stoppades.
func (e *phaseBlockedError) Error() string {
	return fmt.Sprintf("fasbytet stoppades (%s): %s", e.Reason, e.Message)
}

// TriggerAgent beskriver att körningen startades av agenten själv.
const TriggerAgent = "agent"

// Agentprofiler - vilken sorts agent som utför körningen.
const (
	ProfileCoding = profile.Coding
	ProfileLife   = profile.Life
)

// Worker pollar Plan efter agentmål och definierar dem.
type GoalDefiner interface {
	Define(ctx context.Context, input definer.Input) (*definer.Definition, error)
}

// CodeRunner driver kodändringar för ett mål.
type CodeRunner interface {
	Process(ctx context.Context, run planclient.AgentRun, goal planclient.Goal) error
}

// UsageReporter kan berätta hur mycket tokens en körning har använt.
//
// Det är ett valfritt interface - en runner som inte kan rapportera
// användning ska ändå fungera.
type UsageReporter interface {
	Usage() runusage.Snapshot
}

type Worker struct {
	planClient   *planclient.Client
	definer      GoalDefiner
	codeRunner   CodeRunner
	pollInterval time.Duration
	// provider, model och agentID beskriver vem som kör och med vad.
	// Plan behöver dem för att prissätta körningen.
	provider string
	model    string
	agentID  string
	// onlyGoalId begränsar workern till ett mål. Noll betyder alla.
	onlyGoalId int64
	// initiator beskriver vem eller vad som startade körningen.
	initiator string
	// profile är vilken sorts agent som kör - coding eller life.
	profile string
	// stopWhenFinished gör att workern slutar när målet är färdigt.
	stopWhenFinished bool
}

// Config konfigurerar workern
type Config struct {
	// PollInterval är hur ofta Plan läses. Noll ger en minut.
	PollInterval time.Duration
	// OnlyGoalId begränsar workern till ett enda mål - används för att
	// köra ett uppdrag manuellt.
	OnlyGoalId int64
	// Initiator beskriver vem eller vad som startade körningen, t.ex.
	// "worker:agent-worker-goal-502". Tom sträng ger en generell agent.
	Initiator string
	// Profile är vilken sorts agent som kör - coding eller life.
	// Tom sträng ger coding.
	Profile string
	// CodeRunner driver kodändringar. Nil betyder att workern bara
	// definierar mål och inte utför ändringar.
	CodeRunner CodeRunner
	// StopWhenFinished gör att workern slutar när målet är färdigt.
	//
	// Det används när workern körs i ett Kubernetes Job - jobbet ska avslutas
	// när uppgiften är klar, inte köra i evighet.
	StopWhenFinished bool
	// Provider och Model beskriver vilken modell agenten använder. Plan
	// behöver dem för att räkna fram kostnaden.
	Provider string
	Model    string
	// AgentID identifierar agenten. Den gör att kostnader går att
	// filtrera per agent.
	AgentID string
}

// New creates a worker that defines the goals marked for agent work
func New(planClient *planclient.Client, goalDefiner GoalDefiner, cfg Config) *Worker {
	interval := cfg.PollInterval
	if interval == 0 {
		interval = time.Minute
	}

	// Profilen bestämmer vad workern får göra - den är en teknisk
	// gräns, inte en instruktion till modellen.
	agentProfile, err := profile.PolicyFor(cfg.Profile)
	if err != nil {
		log.Printf("Invalid profile %q, using %s", cfg.Profile, profile.Coding)
		agentProfile = profile.CodingPolicy()
	}

	// En profil som inte får röra kod får inte köra codern.
	codeRunner := cfg.CodeRunner
	if !agentProfile.MayChangeCode {
		codeRunner = nil
	}

	return &Worker{
		planClient:       planClient,
		definer:          goalDefiner,
		codeRunner:       codeRunner,
		pollInterval:     interval,
		onlyGoalId:       cfg.OnlyGoalId,
		initiator:        cfg.Initiator,
		profile:          agentProfile.Name,
		stopWhenFinished: cfg.StopWhenFinished,
		provider:         cfg.Provider,
		model:            cfg.Model,
		agentID:          cfg.AgentID,
	}
}

// Run starts the worker (blocks until the context is cancelled)
func (w *Worker) Run(ctx context.Context) {
	log.Printf("Agent worker starting, polling every %s for goals with agent=true", w.pollInterval)

	ticker := time.NewTicker(w.pollInterval)
	defer ticker.Stop()

	if err := w.RunOnce(ctx); err != nil {
		if errors.Is(err, ErrFinished) {
			log.Println("Agent worker finished the goal and stops")
			return
		}
		log.Printf("Agent worker run failed: %v", err)
	}

	for {
		select {
		case <-ctx.Done():
			log.Println("Agent worker stopping")
			return
		case <-ticker.C:
			err := w.RunOnce(ctx)
			if err == nil {
				continue
			}
			if errors.Is(err, ErrFinished) {
				log.Println("Agent worker finished the goal and stops")
				return
			}
			log.Printf("Agent worker run failed: %v", err)
		}
	}
}

// RunOnce processes the agent goals once
func (w *Worker) RunOnce(ctx context.Context) error {
	goals, err := w.planClient.GetAgentGoals(ctx)
	if err != nil {
		return fmt.Errorf("failed to fetch agent goals: %w", err)
	}

	defineStatusID, err := w.planClient.GetStatusIDByName(ctx, PhaseDefine)
	if err != nil {
		return fmt.Errorf("failed to resolve define phase: %w", err)
	}

	runs, err := w.planClient.GetAgentRuns(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to fetch agent runs: %w", err)
	}

	// Körningar per mål - den senaste används.
	latestRun := make(map[int64]planclient.AgentRun, len(runs))
	runCount := make(map[int64]int, len(runs))
	for _, run := range runs {
		existing, found := latestRun[run.GoalID]
		if !found || run.ID > existing.ID {
			latestRun[run.GoalID] = run
		}
		runCount[run.GoalID]++
	}

	// En köad körning är en uttrycklig begäran om att få köras. Workern
	// tar över den i stället för att skapa en egen.
	queuedRun := findQueuedRun(runs)

	// Mål som redan behandlats - en köad körning får inte hanteras två
	// gånger.
	handled := make(map[int64]bool, len(goals))

	planStatusID, err := w.planClient.GetStatusIDByName(ctx, PhasePlan)
	if err != nil {
		return fmt.Errorf("failed to resolve plan phase: %w", err)
	}

	for _, goal := range goals {
		if !goal.Agent {
			continue
		}
		if w.onlyGoalId != 0 && goal.ID != w.onlyGoalId {
			continue
		}

		run, hasRun := latestRun[goal.ID]

		// En avbruten körning ska inte återupptas - en människa har sagt
		// stopp. Jobbet avslutas i stället för att börja om.
		if hasRun && run.Status == StatusCancelled {
			log.Printf("Run %d for goal #%d was cancelled - the worker stops", run.ID, goal.ID)
			return ErrFinished
		}

		// En köad körning är en uttrycklig begäran om att få köras. Den
		// tas över och arbetet utförs i den körningen.
		if queued, found := queuedRun[goal.ID]; found {
			handled[goal.ID] = true
			if err := w.runQueued(ctx, goal, queued, defineStatusID); err != nil {
				log.Printf("Failed to run queued run %d: %v", queued.ID, err)
			}
			continue
		}

		// Ett mål som väntar på ett mänskligt beslut ska återupptas så
		// snart beslutet är fattat.
		//
		// I definitionfasen är det förslaget om målets utseende som
		// väntar. Lämnar agenten över till kodern först när definitionen
		// är godkänd och applicerad.
		if hasRun && run.Status == StatusWaitingApproval {
			if goal.StatusID == defineStatusID {
				resumed, err := w.resumeDefinition(ctx, run, goal)
				if err != nil {
					log.Printf("Failed to resume the definition of goal #%d: %v", goal.ID, err)
				}
				if resumed {
					handled[goal.ID] = true
					continue
				}
			}

			if w.codeRunner != nil {
				if err := w.codeRunner.Process(ctx, run, goal); err != nil {
					log.Printf("Failed to resume goal #%d: %v", goal.ID, err)
				}
			}
			continue
		}

		// En pågående körning lämnas ifred.
		if hasRun && run.Status == StatusRunning {
			continue
		}

		// Ett mål i definitionfasen specificeras först.
		if goal.StatusID == defineStatusID {
			if err := w.processGoal(ctx, goal, hasRun, run); err != nil {
				log.Printf("Failed to process goal #%d: %v", goal.ID, err)
			}
			continue
		}

		// Ett definierat mål lämnas över till codern. Det sker bara en gång
		// per mål - efter definitionen finns exakt en körning, och när
		// kodkörningen finns är det två.
		if w.codeRunner != nil && goal.StatusID == planStatusID && runCount[goal.ID] < 2 {
			if err := w.startCodeRun(ctx, goal, false, planclient.AgentRun{}); err != nil {
				log.Printf("Failed to start code run for goal #%d: %v", goal.ID, err)
			}
			continue
		}

		// Ett färdigt mål avslutar workern - ett Kubernetes Job ska inte
		// köra i evighet.
		if w.stopWhenFinished && w.isFinished(goal, run, hasRun) {
			return ErrFinished
		}
	}

	// En köad körning är en uttrycklig begäran och ska inte glömmas bara
	// för att målet inte är ett agentmål. Utan det här skulle jobbet köra
	// i evighet på ett mål som aldrig behandlas.
	for goalID, queued := range queuedRun {
		if handled[goalID] {
			continue
		}
		if w.onlyGoalId != 0 && goalID != w.onlyGoalId {
			continue
		}

		goal, err := w.planClient.GetGoal(ctx, goalID)
		if err != nil {
			log.Printf("Failed to fetch goal #%d for queued run %d: %v", goalID, queued.ID, err)
			continue
		}

		if err := w.runQueued(ctx, *goal, queued, defineStatusID); err != nil {
			log.Printf("Failed to run queued run %d: %v", queued.ID, err)
		}
	}

	// Ett jobb finns för att utföra en körning. När körningen är klar är
	// jobbet klart - det ska inte stå och vänta på en fas som aldrig
	// kommer, som när en livsprofil lämnar målet i planfasen.
	if w.stopWhenFinished && w.onlyGoalId != 0 {
		done, err := w.goalRunDone(ctx, w.onlyGoalId)
		if err != nil {
			log.Printf("Failed to check whether goal #%d is finished: %v", w.onlyGoalId, err)
		} else if done {
			return ErrFinished
		}
	}

	return nil
}

// goalRunDone berättar om den senaste körningen för målet är klar.
func (w *Worker) goalRunDone(ctx context.Context, goalID int64) (bool, error) {
	runs, err := w.planClient.GetAgentRuns(ctx, nil)
	if err != nil {
		return false, fmt.Errorf("failed to fetch agent runs: %w", err)
	}

	var latest planclient.AgentRun
	found := false
	for _, run := range runs {
		if run.GoalID != goalID {
			continue
		}
		if !found || run.ID > latest.ID {
			latest = run
			found = true
		}
	}

	if !found {
		return false, nil
	}

	return latest.Status == StatusSucceeded, nil
}

// runQueued tar över en köad körning och utför arbetet i den.
func (w *Worker) runQueued(ctx context.Context, goal planclient.Goal, queued planclient.AgentRun, defineStatusID int) error {
	if err := w.adoptRun(ctx, queued); err != nil {
		return err
	}

	if goal.StatusID == defineStatusID {
		return w.processGoal(ctx, goal, true, queued)
	}

	// En profil som inte får röra kod har ingen kodfas - körningen
	// avslutas utan ändringar.
	if w.codeRunner == nil {
		return w.closeRunWithoutChanges(ctx, goal, queued)
	}

	return w.startCodeRun(ctx, goal, true, queued)
}

// findQueuedRun ger de köade körningarna, per mål.
//
// En köad körning är en uttrycklig begäran om att få köras - jobbcontrollern
// skapar ett Kubernetes Job för den.
func findQueuedRun(runs []planclient.AgentRun) map[int64]planclient.AgentRun {
	queued := make(map[int64]planclient.AgentRun)
	for _, run := range runs {
		if run.Status != StatusQueued {
			continue
		}
		existing, found := queued[run.GoalID]
		if !found || run.ID > existing.ID {
			queued[run.GoalID] = run
		}
	}
	return queued
}

// adoptRun tar över en köad körning och markerar den som pågående.
func (w *Worker) adoptRun(ctx context.Context, run planclient.AgentRun) error {
	status := StatusRunning
	summary := "Körningen har startat"

	if _, err := w.planClient.UpdateAgentRun(ctx, run.ID, planclient.UpdateAgentRunRequest{
		Status:  &status,
		Summary: &summary,
	}); err != nil {
		return fmt.Errorf("failed to adopt run %d: %w", run.ID, err)
	}

	log.Printf("Adopted queued agent run %d for goal #%d", run.ID, run.GoalID)
	return nil
}

// isFinished berättar om målet är färdigt och ingen körning pågår.
func (w *Worker) isFinished(goal planclient.Goal, run planclient.AgentRun, hasRun bool) bool {
	if !hasRun {
		return false
	}
	// En misslyckad körning är inte färdig - jobbet ska inte avslutas
	// innan en människa eller en ny körning tagit ställning till felet.
	if run.Status != StatusSucceeded {
		return false
	}

	// En profil som får röra kod flyttar målet till verify när ändringen
	// är verifierad. Då är jobbet klart.
	verifyStatusID, err := w.planClient.GetStatusIDByName(context.Background(), PhaseVerify)
	if err == nil && goal.StatusID == verifyStatusID {
		return true
	}

	// En profil som inte får röra kod lämnar målet i plan - men körningen
	// är klar, och jobbet ska inte stå och vänta på en fas som aldrig
	// kommer.
	return w.codeRunner == nil
}

// closeRunWithoutChanges avslutar en körning som inte får göra
// kodändringar.
//
// Profilen säger vad som är tillåtet - körningen ska inte lämnas öppen
// bara för att profilen förbjuder det steget.
func (w *Worker) closeRunWithoutChanges(ctx context.Context, goal planclient.Goal, run planclient.AgentRun) error {
	status := StatusSucceeded
	summary := fmt.Sprintf("Profilen %s får inte ändra kod - körningen avslutades utan ändringar", w.profile)
	finishedAt := time.Now().UTC()

	if _, err := w.planClient.UpdateAgentRun(ctx, run.ID, planclient.UpdateAgentRunRequest{
		Status:     &status,
		Summary:    &summary,
		Result:     &summary,
		FinishedAt: &finishedAt,
	}); err != nil {
		return fmt.Errorf("failed to close agent run: %w", err)
	}

	w.recordEvent(ctx, run.ID, planclient.CreateAgentRunEventRequest{
		Type:  planclient.EventRunStarted,
		Phase: run.Phase,
		Payload: map[string]any{
			"goal_id":       goal.ID,
			"trigger":       run.Trigger,
			"initiator":     w.initiator,
			"agent_profile": w.profile,
		},
	})

	w.recordEvent(ctx, run.ID, planclient.CreateAgentRunEventRequest{
		Type:    planclient.EventRunCompleted,
		Payload: map[string]any{"reason": "profile does not allow code changes", "profile": w.profile},
	})

	w.reportUsage(ctx, run.ID)
	return nil
}

// startCodeRun skapar en körning för ett definierat mål och lämnar över
// till codern.
func (w *Worker) startCodeRun(ctx context.Context, goal planclient.Goal, hasRun bool, existing planclient.AgentRun) error {
	run := &existing
	if !hasRun {
		created, err := w.planClient.CreateAgentRun(ctx, planclient.CreateAgentRunRequest{
			GoalID:       goal.ID,
			Phase:        PhasePlan,
			Status:       StatusRunning,
			AgentID:      w.agentID,
			Provider:     w.provider,
			Model:        w.model,
			Trigger:      TriggerAgent,
			Initiator:    w.initiator,
			AgentProfile: w.profile,
			Summary:      "Agenten tar fram ett kodförslag",
		})
		if err != nil {
			return fmt.Errorf("failed to create agent run: %w", err)
		}
		run = created
	}

	w.recordEvent(ctx, run.ID, planclient.CreateAgentRunEventRequest{
		Type:  planclient.EventRunStarted,
		Phase: PhasePlan,
		Payload: map[string]any{
			"goal_id":       goal.ID,
			"trigger":       TriggerAgent,
			"initiator":     w.initiator,
			"agent_profile": w.profile,
		},
	})

	if err := w.codeRunner.Process(ctx, *run, goal); err != nil {
		return err
	}
	return nil
}

// processGoal defines one goal and leaves the result to a human.
//
// Definitionen skickas inte direkt till målet. Den lämnas som ett
// godkännande, och det är exakt de värden människan godkänner som sedan
// skrivs till målet.
//
// Om det redan finns en körning för målet - en köad körning som workern
// tagit över - används den i stället för att skapa en ny.
func (w *Worker) processGoal(ctx context.Context, goal planclient.Goal, hasRun bool, existing planclient.AgentRun) error {
	log.Printf("Agent worker processing goal #%d: %s", goal.ID, goal.Title)

	run := &existing
	if !hasRun {
		created, err := w.planClient.CreateAgentRun(ctx, planclient.CreateAgentRunRequest{
			GoalID:       goal.ID,
			Phase:        PhaseDefine,
			AgentID:      w.agentID,
			Provider:     w.provider,
			Model:        w.model,
			Status:       StatusRunning,
			Trigger:      TriggerAgent,
			Initiator:    w.initiator,
			AgentProfile: w.profile,
			Summary:      "Agenten specificerar målet",
		})
		if err != nil {
			return fmt.Errorf("failed to create agent run: %w", err)
		}
		run = created
	}
	w.recordEvent(ctx, run.ID, planclient.CreateAgentRunEventRequest{
		Type:  planclient.EventRunStarted,
		Phase: PhaseDefine,
		Payload: map[string]any{
			"goal_id":       goal.ID,
			"trigger":       TriggerAgent,
			"initiator":     w.initiator,
			"agent_profile": w.profile,
		},
	})

	// Platserna behövs för att agenten ska kunna föreslå platser på
	// målet - och för att den inte ska hitta på platser som inte finns.
	places := w.places(ctx)

	definition, err := w.definer.Define(ctx, definer.Input{
		Title:            goal.Title,
		Description:      goal.Description,
		Icon:             goal.Icon,
		Now:              goal.Now,
		Eternal:          goal.Eternal,
		Agent:            goal.Agent,
		Due:              value(goal.Due),
		StartDate:        value(goal.StartDate),
		EstimatedMinutes: minutes(goal.EstimatedMinutes),
		RecurrenceRule:   value(goal.RecurrenceRule),
		Tags:             goal.Tags,
		Places:           names(goal.Places),
	})
	if err != nil {
		w.failRun(ctx, run.ID, err)
		return fmt.Errorf("failed to define goal: %w", err)
	}
	w.recordEvent(ctx, run.ID, planclient.CreateAgentRunEventRequest{
		Type:  planclient.EventDefinitionCreated,
		Phase: PhaseDefine,
	})

	proposal := goaldefinition.Build(goaldefinition.Context{
		Goal:         goal,
		Definition:   *definition,
		Places:       places,
		CurrentPhase: w.phaseName(ctx, goal.StatusID),
		NextPhase:    PhasePlan,
	})

	// Utan data att föreslå finns inget att godkänna - målet går bara
	// vidare till nästa fas.
	if !proposal.HasDataFields() {
		log.Printf("Goal #%d has nothing new to propose - moving it to planning", goal.ID)
		return w.finishDefinition(ctx, goal, *run)
	}

	if _, err := w.planClient.CreateAgentApproval(ctx, run.ID, planclient.CreateAgentApprovalRequest{
		Title:            proposalTitle(proposal, goal),
		Description:      proposal.Summary(),
		Consequence:      proposal.Consequence(),
		VerificationPlan: "Förslaget är validerat mot Plans regler för datum, ikon, uppskattad tid och upprepning.",
		Payload:          proposal.Payload(),
	}); err != nil {
		w.failRun(ctx, run.ID, err)
		return fmt.Errorf("failed to create approval: %w", err)
	}

	w.recordEvent(ctx, run.ID, planclient.CreateAgentRunEventRequest{
		Type:  planclient.EventApprovalRequested,
		Phase: PhaseDefine,
		Payload: map[string]any{
			"title":   proposalTitle(proposal, goal),
			"kind":    goaldefinition.Kind,
			"changes": len(proposal.Fields),
		},
	})

	// Körningen väntar på en människa - statusen säger det tydligt.
	status := StatusWaitingApproval
	summary := "Definitionen inväntar mänskligt godkännande"
	if _, err := w.planClient.UpdateAgentRun(ctx, run.ID, planclient.UpdateAgentRunRequest{
		Status:  &status,
		Summary: &summary,
	}); err != nil {
		return fmt.Errorf("failed to mark run as waiting for approval: %w", err)
	}

	log.Printf("Goal #%d has a definition waiting for approval", goal.ID)
	return nil
}

// resumeDefinition hanterar ett förslag om målets utseende.
//
// Den returnerar true om förslaget är agentens - då ska koderaren inte
// köras förrän definitionen är applicerad.
func (w *Worker) resumeDefinition(ctx context.Context, run planclient.AgentRun, goal planclient.Goal) (bool, error) {
	approvals, err := w.planClient.GetAgentApprovals(ctx, run.ID)
	if err != nil {
		return false, fmt.Errorf("failed to check approvals: %w", err)
	}

	approval := findDefinitionApproval(approvals)
	if approval == nil {
		return false, nil
	}

	switch approval.Status {
	case "approved":
		proposal, err := goaldefinition.FromPayload(approval.Payload)
		if err != nil {
			w.failRun(ctx, run.ID, err)
			return true, fmt.Errorf("failed to read the approved definition: %w", err)
		}

		// Exakt de godkända värdena appliceras - inget genereras om.
		if err := proposal.Apply(ctx, w.planClient, goal.ID); err != nil {
			w.failRun(ctx, run.ID, err)
			return true, fmt.Errorf("failed to apply the approved definition: %w", err)
		}

		log.Printf("Goal #%d got its approved definition", goal.ID)

		// Åtgärdsprofilen styr vad agenten får göra i resten av
		// körningen. Uten profil i förslaget gäller den tekniska
		// profilen - så att äldre förslag fungerar som de gjorde för.
		policy, err := w.policyFor(actionProfile(proposal))
		if err != nil {
			w.failRun(ctx, run.ID, err)
			return true, fmt.Errorf("failed to resolve the action profile: %w", err)
		}

		// Med en coder igång fortsätter samma körning - men med det
		// uppdaterade målet, så förslaget bygger på den nya texten.
		// Kodaren körs bara om profilen får röra kod.
		if w.codeRunner != nil && policy.MayChangeCode {
			updated, err := w.planClient.GetGoal(ctx, goal.ID)
			if err != nil {
				w.failRun(ctx, run.ID, err)
				return true, fmt.Errorf("failed to fetch the updated goal: %w", err)
			}
			if err := w.codeRunner.Process(ctx, run, *updated); err != nil {
				return true, fmt.Errorf("failed to hand over to the coder: %w", err)
			}
			return true, nil
		}

		if err := w.finishDefinition(ctx, goal, run); err != nil {
			return false, fmt.Errorf("failed to finish the definition: %w", err)
		}
		return true, nil
	case "rejected":
		// Ett avslaget förslag körs inte om - det är människan som
		// avgör om agenten får försöka igen.
		log.Printf("Goal #%d has a rejected definition - waiting for a human decision", goal.ID)
		return true, nil

	default:
		// Förslaget väntar fortfarande på beslut.
		return true, nil
	}
}

// finishDefinition avslutar en körning där målet är definierat.
func (w *Worker) finishDefinition(ctx context.Context, goal planclient.Goal, run planclient.AgentRun) error {
	if err := w.syncPhase(ctx, goal.ID, run.ID, PhasePlan); err != nil {
		// Ett stoppat fasbyte är inte ett fel - det är en fråga till
		// människan. Körningen väntar, i stället för att falla.
		var stoppad *phaseBlockedError
		if errors.As(err, &stoppad) {
			w.waitForHuman(ctx, run, stoppad)
			return nil
		}
		w.failRun(ctx, run.ID, err)
		return fmt.Errorf("failed to synchronize plan phase: %w", err)
	}

	w.recordEvent(ctx, run.ID, planclient.CreateAgentRunEventRequest{
		Type:  planclient.EventPhaseChanged,
		Phase: PhasePlan,
		Payload: map[string]any{
			"status": StatusSucceeded,
		},
	})

	status := StatusSucceeded
	summary := goal.Title
	result := "Målet definierat och flyttat till plan"
	finishedAt := time.Now().UTC()
	if _, err := w.planClient.UpdateAgentRun(ctx, run.ID, planclient.UpdateAgentRunRequest{
		Status:     &status,
		Summary:    &summary,
		Result:     &result,
		FinishedAt: &finishedAt,
	}); err != nil {
		return fmt.Errorf("failed to update agent run: %w", err)
	}

	// En avslutad körning ska alltid sluta med run_completed - annars ser
	// historiken ut som om körningen bara upphörde.
	w.recordEvent(ctx, run.ID, planclient.CreateAgentRunEventRequest{
		Type:    planclient.EventRunCompleted,
		Phase:   PhasePlan,
		Payload: map[string]any{"reason": "goal defined and moved to planning"},
	})

	log.Printf("Goal #%d defined and moved to planning", goal.ID)
	return nil
}

// findDefinitionApproval ger det förslag om målets utseende, om det finns.
func findDefinitionApproval(approvals []planclient.AgentApproval) *planclient.AgentApproval {
	for i := range approvals {
		if approvals[i].ApprovalKind() == goaldefinition.Kind {
			return &approvals[i]
		}
	}
	return nil
}

// proposalTitle ger förslagets rubrik - den föreslagna titeln om det
// finns en ny.
func proposalTitle(proposal goaldefinition.Proposal, goal planclient.Goal) string {
	if proposal.Values.Title != nil && *proposal.Values.Title != "" {
		return *proposal.Values.Title
	}
	return goal.Title
}

// places hämtar användarens platser.
//
// Ett fel här ska inte stoppa agenten - ett mål kan definieras utan
// platser också.
func (w *Worker) places(ctx context.Context) []planclient.Place {
	places, err := w.planClient.GetPlaces(ctx)
	if err != nil {
		log.Printf("Failed to fetch places: %v", err)
		return nil
	}
	return places
}

// phaseName ger fasnamnet för ett status-id.
func (w *Worker) phaseName(ctx context.Context, statusID int) string {
	name, err := w.planClient.GetPhaseName(ctx, statusID)
	if err != nil {
		return ""
	}
	return name
}

// value ger värdet i en pekare, eller tomt om pekaren är nil.
func value(pointer *string) string {
	if pointer == nil {
		return ""
	}
	return *pointer
}

// minutes ger värdet i en pekare, eller noll om pekaren är nil.
func minutes(pointer *int) int {
	if pointer == nil {
		return 0
	}
	return *pointer
}

// names listar platsernas namn.
func names(places []planclient.Place) []string {
	result := make([]string, 0, len(places))
	for _, place := range places {
		result = append(result, place.Name)
	}
	return result
}

func (w *Worker) recordEvent(ctx context.Context, runID int64, event planclient.CreateAgentRunEventRequest) {
	if err := w.planClient.CreateAgentRunEvent(ctx, runID, event); err != nil {
		log.Printf("Failed to record event for agent run %d: %v", runID, err)
	}
}

func (w *Worker) syncPhase(ctx context.Context, goalID, runID int64, phase string) error {
	// Övergången går via fasmaskinen i API:t - villkoren kontrolleras
	// där, inte i agenten. Ett stoppat byte ger en skälkod, och då ska
	// agenten inte fortsätta: människan bestämmer.
	if _, err := w.planClient.RequestPhase(ctx, goalID, phase); err != nil {
		if reason := planclient.ReasonCode(err); reason != "" {
			return &phaseBlockedError{Reason: reason, Message: err.Error()}
		}
		return fmt.Errorf("failed to update goal phase: %w", err)
	}
	if _, err := w.planClient.UpdateAgentRun(ctx, runID, planclient.UpdateAgentRunRequest{
		Phase: &phase,
	}); err != nil {
		return fmt.Errorf("failed to update agent run phase: %w", err)
	}
	return nil
}

// policyFor ger policyn för en körning.
//
// Åtgärdsprofilen styr när förslaget innehåller en - den bestämmer vad
// agenten får göra. Utan profil behålls den tekniska profilen, så att
// förslag som skapats innan profilen fanns fungerar som de gjorde för.
func (w *Worker) policyFor(action string) (profile.Policy, error) {
	if action == "" {
		return profile.PolicyFor(w.profile)
	}
	return profile.ActionPolicy(action)
}

// actionProfile ger åtgärdsprofilen ur ett godkänt förslag.
//
// Tom sträng betyder att förslaget inte innehöll någon profil.
func actionProfile(proposal goaldefinition.Proposal) string {
	if proposal.Values.ActionProfile == nil {
		return ""
	}
	return *proposal.Values.ActionProfile
}

// reportUsage rapporterar tokenanvändningen för en körning.
//
// Plan räknar fram kostnaden ur sin prislista. Agenten skickar bara
// användningen, så att priset gäller likadant för alla körningar.
//
// Ett fel här ska inte stoppa arbetet - kostnaden får komma i efterhand.
func (w *Worker) reportUsage(ctx context.Context, runID int64) {
	reporter, ok := w.codeRunner.(UsageReporter)
	if !ok {
		return
	}

	snapshot := reporter.Usage()
	startedAt := snapshot.StartedAt.UTC()
	finishedAt := snapshot.FinishedAt.UTC()

	if err := w.planClient.RecordAgentRunUsage(ctx, runID, planclient.AgentRunUsageRequest{
		AgentID:         w.agentID,
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
		return
	}

	log.Printf("Reported usage for run %d: %d input, %d output tokens",
		runID, snapshot.InputTokens, snapshot.OutputTokens)
}

// waitForHuman markerar en körning som väntar på en människa.
//
// Det händer när fasmaskinen stoppar en övergång - till exempel när
// definitionen inte är godkänd, eller när ett delmål saknar resultat.
// Agenten går inte vidare på egen hand; människan bestämmer.
func (w *Worker) waitForHuman(ctx context.Context, run planclient.AgentRun, stoppad *phaseBlockedError) {
	log.Printf("Goal #%d can not move forward yet: %s", run.GoalID, stoppad.Reason)

	w.recordEvent(ctx, run.ID, planclient.CreateAgentRunEventRequest{
		Type:  planclient.EventApprovalRequested,
		Phase: PhaseDefine,
		Payload: map[string]any{
			"reason":  stoppad.Reason,
			"message": stoppad.Message,
			"goal_id": run.GoalID,
		},
	})

	status := StatusWaitingApproval
	summary := "Målet kan inte gå vidare: " + stoppad.Reason
	if _, err := w.planClient.UpdateAgentRun(ctx, run.ID, planclient.UpdateAgentRunRequest{
		Status:  &status,
		Summary: &summary,
	}); err != nil {
		log.Printf("Failed to mark agent run %d as waiting for a human: %v", run.ID, err)
	}
}

// failRun marks a run as failed so the goal does not get stuck
func (w *Worker) failRun(ctx context.Context, runID int64, cause error) {
	status := StatusFailed
	summary := "Agenten kunde inte specificera målet: " + cause.Error()
	result := "Misslyckades: " + cause.Error()
	finishedAt := time.Now().UTC()
	if _, err := w.planClient.UpdateAgentRun(ctx, runID, planclient.UpdateAgentRunRequest{
		Status:     &status,
		Summary:    &summary,
		Result:     &result,
		FinishedAt: &finishedAt,
	}); err != nil {
		log.Printf("Failed to mark agent run %d as failed: %v", runID, err)
	}
	w.recordEvent(ctx, runID, planclient.CreateAgentRunEventRequest{
		Type:  planclient.EventRunFailed,
		Phase: PhaseDefine,
		Payload: map[string]any{
			"error": cause.Error(),
		},
	})

	w.reportUsage(ctx, runID)
}

// PollIntervalFromEnv reads the poll interval from the environment
func PollIntervalFromEnv() time.Duration {
	raw := os.Getenv("AGENT_POLL_INTERVAL")
	if raw == "" {
		return time.Minute
	}

	interval, err := time.ParseDuration(raw)
	if err != nil {
		log.Printf("Invalid AGENT_POLL_INTERVAL %q, using one minute", raw)
		return time.Minute
	}
	return interval
}
