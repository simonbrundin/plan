package handlers

import (
	"context"
	"log"

	"plan/agent/internal/planclient"
)

// runRecorder skriver händelser till en AgentRun i Plan.
//
// Den interaktiva målchatten hör till samma körning som agentworkern startade,
// så historiken visar hela arbetet - inte bara definitionen.
type runRecorder struct {
	client *planclient.Client
	runID  int64
}

// newRunRecorder hittar den senaste körningen för ett mål.
//
// Om målet saknar körning returneras ingen recorder - historiken är då
// tom, vilket är bättre än att skapa en körning bakom ryggen på chatten.
func newRunRecorder(ctx context.Context, client *planclient.Client, goalID int64) *runRecorder {
	if client == nil {
		return nil
	}

	runs, err := client.GetAgentRuns(ctx, &goalID)
	if err != nil {
		log.Printf("Failed to find agent run for goal #%d: %v", goalID, err)
		return nil
	}
	if len(runs) == 0 {
		return nil
	}

	return &runRecorder{client: client, runID: runs[0].ID}
}

// Record skriver en händelse. Fel ignoreras - historiken får inte
// stoppa chatten.
func (r *runRecorder) Record(ctx context.Context, eventType, phase string, payload map[string]any) {
	if r == nil {
		return
	}

	err := r.client.CreateAgentRunEvent(ctx, r.runID, planclient.CreateAgentRunEventRequest{
		Type:    eventType,
		Phase:   phase,
		Payload: payload,
	})
	if err != nil {
		log.Printf("Failed to record %s event for agent run %d: %v", eventType, r.runID, err)
	}
}
