package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"plan-api/internal/database"
	"plan-api/internal/handlers"
	"plan-api/internal/middleware"

	openapiui "github.com/PeterTakahashi/gin-openapi/openapiui"
	"github.com/gin-gonic/gin"

	_ "plan-api/docs"
)

// @title Goals Management API
// @version 1.0
// @description API for managing goals with parent-child relations and dependencies
// @termsOfService http://swagger.io/terms/

// @contact.name Simon
// @contact.url https://github.com/simonreming

// @license.name MIT
// @license.url https://opensource.org/licenses/MIT

// @host localhost:8080
// @BasePath /
// @schemes http https

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a token. For development, use "user_<id>" format (e.g., "user_1")

var dbConnected bool

func getDocsPath() string {
	if path := os.Getenv("DOCS_PATH"); path != "" {
		return path
	}
	return "./docs/swagger.json"
}

func readSwaggerSpec() ([]byte, error) {
	return os.ReadFile(getDocsPath())
}

func withHost(specJSON []byte, host string) ([]byte, error) {
	var spec map[string]interface{}
	if err := json.Unmarshal(specJSON, &spec); err != nil {
		return nil, err
	}
	if host == "" {
		delete(spec, "host")
	} else {
		spec["host"] = host
	}
	return json.Marshal(spec)
}

func main() {
	// Connect to database
	ctx := context.Background()
	if err := database.Connect(ctx); err != nil {
		log.Printf("Warning: Could not connect to database: %v", err)
		log.Println("Running in mock mode without database access")
		dbConnected = false
	} else {
		log.Println("Connected to database")
		defer database.Close()
		dbConnected = true

		// The schema is applied declaratively (DDL only), so the seed data
		// has to be ensured by the application itself.
		if err := database.EnsureSeedData(ctx); err != nil {
			log.Printf("Warning: Could not seed data: %v", err)
		} else {
			log.Println("Seed data verified")
		}
	}

	// Set Gin mode
	if os.Getenv("GIN_MODE") == "" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Create Gin router
	r := gin.Default()

	// Apply middleware
	r.Use(middleware.CORSMiddleware())

	// Health check endpoint (no auth required)
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// API documentation with Scalar UI
	r.GET("/docs/*any", openapiui.WrapHandler(openapiui.Config{
		SpecURL: "/docs/openapi.json",
		SpecProvider: func() ([]byte, error) {
			spec, err := readSwaggerSpec()
			if err != nil {
				return nil, err
			}
			return withHost(spec, os.Getenv("API_HOST"))
		},
		Title: "Goals Management API",
		Theme: "dark",
	}))

	// Auth routes (no authentication required)
	authHandler := handlers.NewAuthHandler(dbConnected)
	r.GET("/api/v1/auth/login", authHandler.Login)
	r.GET("/api/v1/auth/callback", authHandler.Callback)
	r.POST("/api/v1/auth/dev-login", authHandler.DevLogin)
	r.GET("/api/v1/auth/session", authHandler.Session)
	r.POST("/api/v1/auth/logout", authHandler.Logout)

	// API routes with authentication
	api := r.Group("/api/v1")
	api.Use(middleware.APIKeyAuth())     // Try API key first
	api.Use(middleware.AuthMiddleware()) // Fall back to session auth
	api.Use(func(c *gin.Context) {
		c.Set("dbConnected", dbConnected)
		c.Next()
	})
	{
		goalHandler := handlers.NewGoalHandler(dbConnected)
		calendarEventHandler := handlers.NewCalendarEventHandler(dbConnected)
		calendarHandler := handlers.NewCalendarHandler(dbConnected)
		placeHandler := handlers.NewPlaceHandler(dbConnected)
		tagHandler := handlers.NewTagHandler(dbConnected)

		// Calendar endpoints
		api.GET("/calendars", calendarHandler.List)
		api.POST("/calendars", calendarHandler.Create)
		api.PATCH("/calendars/:id", calendarHandler.Update)
		api.DELETE("/calendars/:id", calendarHandler.Delete)

		// Calendar event endpoints
		api.GET("/calendar-events", calendarEventHandler.List)
		api.POST("/calendar-events", calendarEventHandler.Create)
		api.GET("/calendar-events/:id", calendarEventHandler.Get)
		api.PATCH("/calendar-events/:id", calendarEventHandler.Update)
		api.POST("/calendar-events/:id/occurrences", calendarEventHandler.CreateOccurrenceOverride)
		api.POST("/calendar-events/:id/split", calendarEventHandler.SplitSeries)
		api.DELETE("/calendar-events/:id", calendarEventHandler.Delete)

		// Goals endpoints
		api.GET("/goals", goalHandler.GetGoals)
		api.GET("/goals/prioritized", goalHandler.GetPrioritizedGoals)
		api.GET("/goals/graph", goalHandler.GetGoalGraph)
		api.POST("/goals", goalHandler.CreateGoal)
		api.GET("/goals/:id", goalHandler.GetGoal)
		api.PATCH("/goals/:id", goalHandler.UpdateGoal)
		api.DELETE("/goals/:id", goalHandler.DeleteGoal)
		api.POST("/goals/:id/completions", goalHandler.CompleteGoalOccurrence)
		api.DELETE("/goals/:id/completions", goalHandler.UndoGoalOccurrence)
		api.PATCH("/goals/reorder", goalHandler.ReorderGoals)
		api.PATCH("/goals/priority/reorder", goalHandler.ReorderPriorityGoals)

		// Status endpoints
		api.GET("/statuses", goalHandler.GetStatuses)
		api.GET("/goals/:id/status", goalHandler.GetGoalStatus)
		api.PATCH("/goals/:id/status", goalHandler.UpdateGoalStatus)
		api.GET("/goals/:id/status/history", goalHandler.GetStatusHistory)

		// Relations endpoints
		api.POST("/goals/relations", goalHandler.CreateRelation)
		api.PATCH("/goals/relations", goalHandler.UpdateRelation)
		api.PATCH("/goals/relations/weights", goalHandler.UpdateRelationWeights)
		api.DELETE("/goals/relations", goalHandler.DeleteRelation)
		api.PATCH("/goals/relations/reorder", goalHandler.ReorderRelations)

		// Dependencies endpoints
		api.POST("/goals/dependencies", goalHandler.CreateDependency)
		api.DELETE("/goals/dependencies", goalHandler.DeleteDependency)

		// Places endpoints
		api.GET("/places", placeHandler.List)
		api.POST("/places", placeHandler.Create)
		api.GET("/places/nearby", placeHandler.Nearby)
		api.PATCH("/places/:id", placeHandler.Update)
		api.DELETE("/places/:id", placeHandler.Delete)

		// Places for goals and calendar events
		api.GET("/goals/:id/places", placeHandler.GetGoalPlaces)
		api.PUT("/goals/:id/places", placeHandler.SetGoalPlaces)
		api.GET("/calendar-events/:id/places", placeHandler.GetEventPlaces)
		api.PUT("/calendar-events/:id/places", placeHandler.SetEventPlaces)

		// The user's "I am here" place
		api.GET("/me/current-place", placeHandler.GetCurrentPlace)
		api.PUT("/me/current-place", placeHandler.SetCurrentPlace)

		// Tags endpoints
		api.GET("/tags", tagHandler.List)

		// Tags for goals
		api.GET("/goals/:id/tags", tagHandler.GetGoalTags)
		api.PUT("/goals/:id/tags", tagHandler.SetGoalTags)
		api.DELETE("/goals/:id/tags/:name", tagHandler.DeleteGoalTag)

		// API Keys endpoints
		apiKeyHandler := handlers.NewAPIKeyHandler()
		api.GET("/api-keys", apiKeyHandler.ListAPIKeys)
		api.POST("/api-keys", apiKeyHandler.CreateAPIKey)
		api.DELETE("/api-keys", apiKeyHandler.DeleteAPIKey)
	}

	// Get port from environment or default to 8080
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Create server
	srv := &http.Server{
		Addr:    ":" + port,
		Handler: r,
	}

	// Graceful shutdown
	go func() {
		log.Printf("Starting server on port %s", port)
		log.Printf("Scalar UI available at http://localhost:%s/docs/", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}

	log.Println("Server exited")
}
