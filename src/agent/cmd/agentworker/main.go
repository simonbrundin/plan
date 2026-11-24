// Command agentworker kör agent workern utanför API-servern.
//
// Workern pollar Plan efter mål som är markerade för agentarbete och
// specificerar dem. Den kan också köras för ett enda mål:
//
//	AGENT_GOAL_ID=449 go run ./cmd/agentworker
package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"

	"plan-agent/internal/agentworker"
	"plan-agent/internal/coder"
	"plan-agent/internal/core/provider"
	"plan-agent/internal/definer"
	"plan-agent/internal/planclient"
	"plan-agent/internal/profile"
	"plan-agent/internal/proposer"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigChan
		log.Println("Shutting down agent worker...")
		cancel()
	}()

	planClient, err := planclient.New()
	if err != nil {
		log.Fatalf("Failed to create Plan client: %v", err)
	}

	factory := provider.NewFactory()
	llmProvider, model, err := factory.CreateFromEnv()
	if err != nil {
		log.Fatalf("Failed to create LLM provider: %v", err)
	}
	log.Printf("Using LLM provider: %s (%s)", model.Provider, model.Name)

	goalDefiner := definer.New(llmProvider, model)

	agentProfile := profileFromEnv()

	worker := agentworker.New(planClient, goalDefiner, agentworker.Config{
		PollInterval:     agentworker.PollIntervalFromEnv(),
		OnlyGoalId:       goalIDFromEnv(),
		Initiator:        initiatorFromEnv(),
		Profile:          agentProfile,
		CodeRunner:       codeRunnerFromEnv(planClient, llmProvider, model, agentProfile),
		StopWhenFinished: stopWhenFinishedFromEnv(),
	})

	worker.Run(ctx)
}

// codeRunnerFromEnv bygger codern om repot är satt.
//
// Utan AGENT_REPO_PATH gör workern bara definitioner - kodändringar
// kräver ett repo att arbeta i.
func codeRunnerFromEnv(planClient *planclient.Client, llmProvider provider.LLMProvider, model provider.Model, agentProfile string) agentworker.CodeRunner {
	// Profilen avgör om körningen får röra kod - det är en teknisk gräns.
	policy, err := profile.PolicyFor(agentProfile)
	if err != nil {
		log.Printf("Invalid agent profile %q: %v", agentProfile, err)
		return nil
	}
	if !policy.MayChangeCode {
		log.Printf("Profile %s may not change code - the worker only defines goals", policy.Name)
		return nil
	}

	repoPath := os.Getenv("AGENT_REPO_PATH")
	if repoPath == "" {
		log.Println("AGENT_REPO_PATH is not set - the worker only defines goals")
		return nil
	}

	if _, err := os.Stat(repoPath); err != nil {
		log.Printf("AGENT_REPO_PATH %s is not readable: %v", repoPath, err)
		return nil
	}

	log.Printf("Code runner enabled for repository %s", repoPath)
	return coder.New(planClient, proposer.New(llmProvider, model), coder.Config{
		RepoPath: repoPath,
	})
}

// initiatorFromEnv beskriver vem eller vad som startade körningen.
// AGENT_INITIATOR används om den sätts, annars används podens namn.
func initiatorFromEnv() string {
	if raw := os.Getenv("AGENT_INITIATOR"); raw != "" {
		return raw
	}
	if hostname, err := os.Hostname(); err == nil && hostname != "" {
		return "worker:" + hostname
	}
	return "worker:agentworker"
}

// profileFromEnv läser AGENT_PROFILE - coding eller life.
func profileFromEnv() string {
	switch raw := os.Getenv("AGENT_PROFILE"); raw {
	case agentworker.ProfileLife:
		return agentworker.ProfileLife
	case agentworker.ProfileCoding:
		return agentworker.ProfileCoding
	case "":
		return agentworker.ProfileCoding
	default:
		log.Printf("Invalid AGENT_PROFILE %q, using %s", raw, agentworker.ProfileCoding)
		return agentworker.ProfileCoding
	}
}

// stopWhenFinishedFromEnv läser AGENT_STOP_WHEN_FINISHED.
//
// I ett Kubernetes Job ska workern avslutas när uppgiften är klar.
// Ett långlivet jobb som bara väntar är dyrt och svårt att felsöka.
func stopWhenFinishedFromEnv() bool {
	raw := strings.ToLower(os.Getenv("AGENT_STOP_WHEN_FINISHED"))
	return raw == "true" || raw == "1" || raw == "yes"
}

// goalIDFromEnv läser AGENT_GOAL_ID - om den sätts körs workern bara
// för det målet.
func goalIDFromEnv() int64 {
	raw := os.Getenv("AGENT_GOAL_ID")
	if raw == "" {
		return 0
	}

	id, err := strconv.ParseInt(raw, 10, 64)
	if err != nil {
		log.Printf("Invalid AGENT_GOAL_ID %q, running for all agent goals", raw)
		return 0
	}
	return id
}
