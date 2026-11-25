package planclient

// Sökvägarna agenten använder mot Plan.
//
// De finns i kontraktet (src/contracts/plan-agent.yaml) - ett test
// kontrollerar att de stämmer. Här skrivs parametrar som %d, i kontraktet
// som {id}.
const (
	pathGoals             = "/api/v1/goals"
	pathGoal              = "/api/v1/goals/%d"
	pathGoalStatus        = "/api/v1/goals/%d/status"
	pathGoalRelations     = "/api/v1/goals/relations"
	pathStatuses          = "/api/v1/statuses"
	pathAgentRuns         = "/api/v1/agent-runs"
	pathAgentRun          = "/api/v1/agent-runs/%d"
	pathAgentRunUsage     = "/api/v1/agent-runs/%d/usage"
	pathAgentRunEvents    = "/api/v1/agent-runs/%d/events"
	pathAgentRunApprovals = "/api/v1/agent-runs/%d/approvals"
)

// AllaSökvägar är de sökvägar agenten använder. Testet mot kontraktet
// går igenom listan.
var AllaSökvägar = []string{
	pathGoals,
	pathGoal,
	pathGoalStatus,
	pathGoalRelations,
	pathStatuses,
	pathAgentRuns,
	pathAgentRun,
	pathAgentRunUsage,
	pathAgentRunEvents,
	pathAgentRunApprovals,
}
