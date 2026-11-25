package handlers

import (
	"net/http"

	"plan-api/internal/models"

	"github.com/gin-gonic/gin"
)

// AgentRunLifecycleHandler beskriver agentkörningens livscykel.
//
// Gränssnittet och andra klienter ska inte hålla egna listor över faser
// och tillstånd - de läser dem härifrån. Så slipper en status byta namn
// utan att klienterna lämnas med gamla värden.
type AgentRunLifecycleHandler struct{}

// NewAgentRunLifecycleHandler skapar en ny livscykel-handler.
func NewAgentRunLifecycleHandler() *AgentRunLifecycleHandler {
	return &AgentRunLifecycleHandler{}
}

// AgentRunLifecycle beskriver faser, tillstånd och övergångar.
type AgentRunLifecycle struct {
	// Phases är faser i den ordning de sker.
	Phases []string `json:"phases" example:"define,plan,execute,verify,completed"`
	// Status är de tillstånd en körning kan ha.
	Status []string `json:"status" example:"queued,running,blocked,done,failed,cancelled"`
	// Terminal är de tillstånd en körning inte kan lämna.
	Terminal []string `json:"terminal" example:"done,failed,cancelled"`
	// NeedsHuman är de tillstånd där en människa måste agera.
	NeedsHuman []string `json:"needs_human" example:"blocked"`
	// Transitions visar vilka tillstånd ett tillstånd får övergå till.
	Transitions map[string][]string `json:"transitions"`
}

// GetAgentRunLifecycle godoc
// @Summary Get agent run lifecycle
// @Description The canonical phases, statuses and allowed transitions for agent runs
// @Tags agent-runs
// @Produce json
// @Security BearerAuth
// @Success 200 {object} handlers.AgentRunLifecycle
// @Failure 401 {object} models.ErrorResponse
// @Router /api/v1/agent-run-lifecycle [get]
func (h *AgentRunLifecycleHandler) GetAgentRunLifecycle(c *gin.Context) {
	terminal := []string{}
	needsHuman := []string{}

	for _, status := range models.AgentRunStatuses {
		if models.IsTerminalAgentRunStatus(status) {
			terminal = append(terminal, status)
		}
		if models.AgentRunStatusNeedsHuman(status) {
			needsHuman = append(needsHuman, status)
		}
	}

	c.JSON(http.StatusOK, AgentRunLifecycle{
		Phases:      models.AgentRunPhases,
		Status:      models.AgentRunStatuses,
		Terminal:    terminal,
		NeedsHuman:  needsHuman,
		Transitions: models.AgentRunStatusTransitions(),
	})
}
