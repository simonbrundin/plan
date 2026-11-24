package planclient

// Händelsetyper som agenten skriver till Plan.
//
// Namnen matchar de i Plan-API:t (models.AgentRunEvent*) så att historiken
// är läsbar oavsett varifrån den kommer ifrån.
const (
	// EventRunStarted - körningen har startat.
	EventRunStarted = "run_started"
	// EventDefinitionCreated - agenten har definierat målet.
	EventDefinitionCreated = "definition_created"
	// EventPhaseChanged - körningen har bytt fas.
	EventPhaseChanged = "phase_changed"
	// EventToolCall - ett verktyg anropades.
	EventToolCall = "tool_call"
	// EventToolResult - ett verktyg gav ett resultat.
	EventToolResult = "tool_result"
	// EventSubgoalCreated - agenten skapade ett undermål.
	EventSubgoalCreated = "subgoal_created"
	// EventApprovalRequested - agenten begär ett mänskligt beslut.
	EventApprovalRequested = "approval_requested"
	// EventApprovalDecided - en människa fattade beslutet.
	EventApprovalDecided = "approval_decided"
	// EventVerification - en verifiering utfördes.
	EventVerification = "verification"
	// EventArtifact - en artefakt skapades, till exempel en diff.
	EventArtifact = "artifact"
	// EventError - något gick fel under körningen.
	EventError = "error"
	// EventRunFailed - körningen misslyckades.
	EventRunFailed = "run_failed"
	// EventRunCompleted - körningen var klar.
	EventRunCompleted = "run_completed"
)
