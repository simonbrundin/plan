// @title Agent API
// @version 1.0
// @description API for the Software Agent system.\n\n## Resources\n\n- [README](https://github.com/simonbrundin/agents#readme)\n- [Roadmap](https://github.com/simonbrundin/agents/blob/main/ROADMAP.md)
// @description Contact: Simon Brundin
// @host localhost:8080
// @BasePath /
// @schemes http
// @externalDocs.description Project README
// @externalDocs.url https://github.com/simonbrundin/agents#readme
package api

import (
	"bytes"
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	openapiui "github.com/PeterTakahashi/gin-openapi/openapiui"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"plan-agent/internal/agentworker"
	"plan-agent/internal/api/db"
	"plan-agent/internal/api/docs"
	"plan-agent/internal/api/handlers"
	"plan-agent/internal/api/middleware"
	"plan-agent/internal/approvals"
	"plan-agent/internal/core/orchestrator"
	"plan-agent/internal/core/provider"
	"plan-agent/internal/core/workflow"
	"plan-agent/internal/definer"
	"plan-agent/internal/jobcontroller"
	"plan-agent/internal/jobrunner"
	"plan-agent/internal/planclient"
	"plan-agent/internal/scheduler"
)

// Orchestratorns tillstånd.
//
// Initieringsfelet sparas här så att det syns i /health och /status. Ett
// fel som bara finns i loggen är lätt att missa - och just det felet gör
// att gate-endpointern aldrig registreras.
var (
	orchestratorReady   bool
	orchestratorInitErr error
)

// readSwaggerSpec returns the embedded swagger spec
func readSwaggerSpec() ([]byte, error) {
	return docs.SwaggerSpec, nil
}

// withHost modifies the swagger spec to use the correct host
func withHost(specJSON []byte, host string) ([]byte, error) {
	var spec map[string]interface{}
	if err := json.Unmarshal(specJSON, &spec); err != nil {
		return nil, err
	}
	// If host is empty, remove the host field (uses relative URLs)
	if host == "" {
		delete(spec, "host")
	} else {
		spec["host"] = host
	}
	return json.Marshal(spec)
}

// Run starts the API server
func Run() {
	// Connect to database
	pool, err := db.Connect()
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	db.SetPool(pool)
	log.Println("Connected to database")

	// Initiera orchestratorn.
	//
	// StateStoren ska överleva en omstart när databasen finns, så att en
	// checkpointplan inte försvinner bara för att podden startades om.
	// Utan databas faller den tillbaka på minneslagring.
	var stateStore orchestrator.StateStore
	if pool != nil {
		stateStore = orchestrator.NewCheckpointStore(pool)
	}

	o, err := workflow.InitOrchestrator(workflow.Config{StateStore: stateStore})
	if err != nil {
		// Initieringsfelet ska vara synligt - inte bara i loggen. Utan
		// orchestratorn saknas gate-endpointern, och det ska gå att se
		// att de saknas.
		orchestratorInitErr = err
		log.Printf("Warning: Failed to initialize orchestrator: %v", err)
		o = nil
	} else {
		orchestratorReady = true
		log.Println("Orchestrator initialized")
	}

	// Initialize Plan API client (if configured)
	var planClient *planclient.Client
	if os.Getenv("PLAN_API_KEY") != "" {
		planClient, err = planclient.New()
		if err != nil {
			log.Printf("Warning: Failed to create Plan client: %v", err)
		} else {
			log.Println("Plan API client created")
		}
	}

	// Godkännandestören finns alltid - den kan användas utan Plan-koppling.
	approvalStore, err := approvals.New()
	if err != nil {
		log.Printf("Warning: Failed to create approval store: %v", err)
		approvalStore = nil
	} else {
		log.Printf("Approval store initialized at: %s/pending-approvals", os.Getenv("AGENT_DATA_DIR"))
	}

	// Create the approval handler - godkännanden tas i Plan, inte i Slack.
	approvalHandler := handlers.NewApprovalHandler(approvalStore, planClient)

	// Start goal scheduler (if all components are configured)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if planClient != nil && approvalStore != nil {
		go startGoalScheduler(ctx, planClient, approvalStore)
	}

	r := gin.Default()

	// CORS middleware
	config := cors.DefaultConfig()
	config.AllowAllOrigins = true
	config.AllowHeaders = []string{"Origin", "Content-Type", "Accept", "Authorization"}
	r.Use(cors.New(config))

	// API documentation with Scalar UI
	// Use /openapi.json to avoid conflict with /docs/*any wildcard
	r.GET("/docs/*any", openapiui.WrapHandler(openapiui.Config{
		SpecURL: "/openapi.json", // Points to our endpoint
		Title:   "Agent API",
		Theme:   "dark",
	}))

	// Serve OpenAPI spec at /openapi.json (not /docs/ to avoid wildcard conflict)
	r.GET("/openapi.json", func(c *gin.Context) {
		spec, err := readSwaggerSpec()
		if err != nil {
			c.String(http.StatusInternalServerError, "Failed to read spec: %v", err)
			return
		}
		// Remove host to use relative URLs (Scalar will use its own origin)
		spec, err = withHost(spec, "")
		if err != nil {
			c.String(http.StatusInternalServerError, "Failed to modify spec: %v", err)
			return
		}
		c.DataFromReader(http.StatusOK, int64(len(spec)), "application/json", bytes.NewReader(spec), nil)
	})

	// Create workflow handler
	var workflowHandler *handlers.WorkflowHandler
	if o != nil {
		workflowHandler = handlers.NewWorkflowHandler(o)
	}

	// API v1
	//
	// Nyckeln sätts i AGENT_API_KEY. Utan den är tjänsten öppen, och det
	// ska inte behöva sägas till i loggen - det ska vara ett medvetet val.
	if os.Getenv("AGENT_API_KEY") == "" {
		log.Println("AGENT_API_KEY is not set - the agent API accepts calls without a key")
	}

	v1 := r.Group("/api/v1")
	v1.Use(middleware.RequireAPIKey())
	{
		// Health & Status
		v1.GET("/health", healthHandler)
		v1.GET("/status", statusHandler)

		// Workflow (only if initialized)
		if workflowHandler != nil {
			v1.POST("/workflow", workflowHandler.StartWorkflow)
			v1.GET("/workflow/plans", workflowHandler.ListPlans)
			v1.GET("/workflow/:id", workflowHandler.GetPlan)
			v1.POST("/workflow/:id/approve", workflowHandler.ApprovePlan)
			v1.POST("/workflow/:id/reject", workflowHandler.RejectPlan)
			v1.GET("/workflow/current/checkpoint", workflowHandler.GetCurrentCheckpoint)
			v1.GET("/workflow/current/progress", workflowHandler.GetCheckpointProgress)
			v1.POST("/workflow/decision", workflowHandler.SubmitHumanDecision)
		}

		// Goals
		v1.GET("/goals", handlers.ListGoals)
		v1.GET("/goals/stats", handlers.GetGoalStats)
		v1.GET("/goals/:id", handlers.GetGoal)
		v1.PATCH("/goals/:id", handlers.UpdateGoal)
		v1.POST("/goals/start-agent", handlers.StartAgentForGoal)

		// Conversations
		v1.GET("/conversations", handlers.ListConversations)
		v1.POST("/conversations", handlers.CreateConversation)
		v1.GET("/conversations/:id", handlers.GetConversation)
		v1.PATCH("/conversations/:id", handlers.UpdateConversation)

		// Messages
		v1.GET("/conversations/:id/messages", handlers.GetMessages)
		v1.POST("/conversations/:id/messages", handlers.CreateMessage)
		v1.GET("/conversations/:id/stream", handlers.StreamMessages)

		// API Keys
		v1.GET("/api-keys", handlers.ListAPIKeys)
		v1.POST("/api-keys", handlers.SetAPIKey)
		v1.DELETE("/api-keys/:provider", handlers.DeleteAPIKey)

		// Godkännanden - besluten tas i Plan
		v1.GET("/approvals", approvalHandler.ListApprovals)
		v1.GET("/approvals/:goal_id", approvalHandler.GetApproval)
		v1.POST("/approvals/:goal_id/decision", approvalHandler.DecideApproval)

		// Docs (README/Roadmap som markdown)
		v1.GET("/readme", handlers.GetReadme)
		v1.GET("/roadmap", handlers.GetRoadmap)

		// WebSocket för realtids-chatt
		v1.GET("/ws/chat", handlers.NewWebSocketHandler().HandleChat)

		// Verktygskörning (avbryt)
		v1.POST("/tools/cancel", handlers.CancelTool)
		v1.POST("/tools/cancel-all", handlers.CancelAllTools)
		v1.GET("/tools/status", handlers.ToolStatus)
	}

	// Health endpoint
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    "ok",
			"service":   "agent-api",
			"version":   os.Getenv("APP_VERSION"),
			"timestamp": time.Now().UTC().Format(time.RFC3339),
		})
	})

	// Handle graceful shutdown
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		<-sigChan
		log.Println("Shutting down...")
		cancel()
	}()

	// Porten kan sättas med PORT - annars används 8080.
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server starting on :%s", port)
	log.Printf("Scalar Docs: http://localhost:%s/docs (OpenAPI spec: /openapi.json)", port)

	switch mode := os.Getenv("AGENT_GOAL_MODE"); mode {
	case "worker":
		go startAgentWorker(ctx, planClient)
	case "", "approval":
		go startGoalScheduler(ctx, planClient, approvalStore)
	default:
		log.Printf("Unknown AGENT_GOAL_MODE %q; starting approval scheduler", mode)
		go startGoalScheduler(ctx, planClient, approvalStore)
	}

	// Jobbcontrollern startar Kubernetes Jobs för mål som behöver en
	// agentkörning. Den kan stängas av när jobb inte ska köras.
	if os.Getenv("AGENT_JOB_CONTROLLER_ENABLED") == "true" {
		go startJobController(ctx, planClient)
	}

	log.Fatal(r.Run(":" + port))
}

// startJobController startar controllern som skapar Kubernetes Jobs.
//
// Controllern är kopplad mellan Plan och klustret: Plan äger målen,
// klustret utför arbetet. Den behöver rättigheter att skapa jobb - inga
// andra.
func startJobController(ctx context.Context, planClient *planclient.Client) {
	if planClient == nil {
		log.Println("Plan client is nil, job controller not started")
		return
	}

	inCluster, err := rest.InClusterConfig()
	if err != nil {
		log.Printf("Failed to read in-cluster config: %v", err)
		log.Println("Job controller not started")
		return
	}

	client, err := kubernetes.NewForConfig(inCluster)
	if err != nil {
		log.Printf("Failed to create Kubernetes client: %v", err)
		log.Println("Job controller not started")
		return
	}

	runner := jobrunner.New(client, jobConfigFromEnv())

	controller := jobcontroller.New(planClient, runner, jobcontroller.Config{
		PollInterval:      jobPollIntervalFromEnv(),
		Namespace:         os.Getenv("AGENT_JOB_NAMESPACE"),
		MaxConcurrentJobs: jobMaxConcurrentJobsFromEnv(),
	})

	controller.Run(ctx)
}

// jobConfigFromEnv bygger jobbkonfigurationen från miljön.
func jobConfigFromEnv() jobrunner.Config {
	config := jobrunner.DefaultConfig()

	if value := os.Getenv("AGENT_JOB_NAMESPACE"); value != "" {
		config.Namespace = value
	}
	if value := os.Getenv("AGENT_JOB_IMAGE"); value != "" {
		config.Image = value
	}
	if value := os.Getenv("AGENT_JOB_SERVICE_ACCOUNT"); value != "" {
		config.ServiceAccount = value
	}
	if value := os.Getenv("AGENT_JOB_REPO_URL"); value != "" {
		config.RepoURL = value
	}
	if value := os.Getenv("AGENT_JOB_REPO_BRANCH"); value != "" {
		config.RepoBranch = value
	}
	if value := os.Getenv("AGENT_JOB_PLAN_API_URL"); value != "" {
		config.PlanAPIURL = value
	}
	if value := os.Getenv("AGENT_JOB_API_KEY_SECRET"); value != "" {
		config.APIKeySecretName = value
	}
	if value := os.Getenv("AGENT_JOB_LLM_SECRET"); value != "" {
		config.LLMSecretName = value
	}
	if value := os.Getenv("AGENT_JOB_GIT_SECRET"); value != "" {
		config.GitSecretName = value
	}
	if value := os.Getenv("AGENT_JOB_CPU_LIMIT"); value != "" {
		config.CPULimit = value
	}
	if value := os.Getenv("AGENT_JOB_MEMORY_LIMIT"); value != "" {
		config.MemoryLimit = value
	}
	if value := os.Getenv("AGENT_JOB_DEADLINE_SECONDS"); value != "" {
		if seconds, err := strconv.ParseInt(value, 10, 64); err == nil {
			config.ActiveDeadlineSeconds = seconds
		}
	}

	return config
}

// jobMaxConcurrentJobsFromEnv läser den globala jobbkapaciteten.
func jobMaxConcurrentJobsFromEnv() int {
	raw := os.Getenv("AGENT_JOB_MAX_CONCURRENT")
	if raw == "" {
		return 4
	}

	value, err := strconv.Atoi(raw)
	if err != nil || value < 1 {
		log.Printf("Invalid AGENT_JOB_MAX_CONCURRENT %q, using 4", raw)
		return 4
	}
	return value
}

// jobPollIntervalFromEnv läser hur ofta controllern läser Plan och klustret.
func jobPollIntervalFromEnv() time.Duration {
	raw := os.Getenv("AGENT_JOB_POLL_INTERVAL")
	if raw == "" {
		return 30 * time.Second
	}

	interval, err := time.ParseDuration(raw)
	if err != nil {
		log.Printf("Invalid AGENT_JOB_POLL_INTERVAL %q, using 30s", raw)
		return 30 * time.Second
	}
	return interval
}

// startAgentWorker starts the worker that picks up goals marked for agent work.
//
// Den här workern är skild från goal schedulern: schedulern definierar
// vanliga mål och frågar efter godkännande i Slack, workern tar sig an de
// mål som är markerade med agent-flaggan.
func startAgentWorker(ctx context.Context, planClient *planclient.Client) {
	if planClient == nil {
		log.Println("Plan client is nil, agent worker not started")
		return
	}

	// Kan stängas av - till exempel när bara chatten ska testas.
	if os.Getenv("AGENT_WORKER_ENABLED") == "false" {
		log.Println("Agent worker disabled (AGENT_WORKER_ENABLED=false)")
		return
	}

	factory := provider.NewFactory()
	llmProvider, model, err := factory.CreateFromEnv()
	if err != nil {
		log.Printf("Failed to create LLM provider: %v", err)
		log.Println("Agent worker not started (set MINIMAX_API_KEY or OPENAI_API_KEY)")
		return
	}
	log.Printf("Agent worker using LLM provider: %s (%s)", model.Provider, model.Name)

	goalDefiner := definer.New(llmProvider, model)
	worker := agentworker.New(planClient, goalDefiner, agentworker.Config{
		PollInterval: agentworker.PollIntervalFromEnv(),
	})

	worker.Run(ctx)
}

// startGoalScheduler initializes and starts the goal polling scheduler
func startGoalScheduler(ctx context.Context, planClient *planclient.Client, approvalStore *approvals.Store) {
	// Kan stängas av - till exempel när bara chatten ska testas.
	if os.Getenv("AGENT_SCHEDULER_ENABLED") == "false" {
		log.Println("Goal scheduler disabled (AGENT_SCHEDULER_ENABLED=false)")
		return
	}

	log.Println("Starting goal scheduler...")

	if planClient == nil {
		log.Println("Plan client is nil, scheduler not started")
		return
	}
	log.Println("Plan API client connected")

	// Create LLM provider for goal definition
	factory := provider.NewFactory()
	llmProvider, model, err := factory.CreateFromEnv()
	if err != nil {
		log.Printf("Failed to create LLM provider: %v", err)
		log.Println("Goal scheduler not started (set MINIMAX_API_KEY or OPENAI_API_KEY)")
		return
	}
	log.Printf("LLM provider created: %s (%s)", model.Provider, model.Name)

	// Create goal definer
	goalDefiner := definer.New(llmProvider, model)

	// Create scheduler
	s := scheduler.New(planClient, approvalStore, goalDefiner, scheduler.Config{
		Initiator: "scheduler:agent-api",
	})

	// Run scheduler
	log.Println("Goal scheduler running...")
	s.Run(ctx)
}

// healthHandler returns the health status
// @Summary Health check
// @Description Returns OK if the service is running
// @Tags health
// @Produce json
// @Success 200 {object} map[string]any
// @Router /api/v1/health [get]
func healthHandler(c *gin.Context) {
	c.JSON(200, gin.H{
		"status":  "ok",
		"service": "agent-api",
		// Orchestratorns tillstånd syns här också. Health svarar 200 även
		// när orchestratorn inte startade - tjänsten lever, men en del
		// av den saknas. Det visas i /status.
		"orchestrator": orchestratorState(),
	})
}

// statusHandler returns detailed service status
// @Summary Service status
// @Description Returns detailed service status
// @Tags health
// @Produce json
// @Success 200 {object} map[string]any
// @Router /api/v1/status [get]
func statusHandler(c *gin.Context) {
	c.JSON(200, gin.H{
		"status":       "healthy",
		"version":      "1.0.0",
		"uptime":       "running",
		"orchestrator": orchestratorState(),
	})
}

// orchestratorState beskriver orchestratorns tillstånd.
//
// Ett initieringsfel ska vara synligt utan att tjänsten dör - en del av
// API:t saknas, men resten fungerar.
func orchestratorState() gin.H {
	if orchestratorInitErr != nil {
		return gin.H{
			"initialized": false,
			"error":       orchestratorInitErr.Error(),
		}
	}
	return gin.H{
		"initialized": orchestratorReady,
	}
}
