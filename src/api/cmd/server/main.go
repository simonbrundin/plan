package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"plan-api/internal/database"
	"plan-api/internal/handlers"
	"plan-api/internal/middleware"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

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

	// Swagger UI
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// API routes with authentication
	api := r.Group("/api")
	api.Use(middleware.AuthMiddleware())
	api.Use(func(c *gin.Context) {
		c.Set("dbConnected", dbConnected)
		c.Next()
	})
	{
		goalHandler := handlers.NewGoalHandler(dbConnected)

		// Goals endpoints
		api.GET("/goals", goalHandler.GetGoals)
		api.GET("/goals/prioritized", goalHandler.GetPrioritizedGoals)
		api.POST("/goals", goalHandler.CreateGoal)
		api.GET("/goals/:id", goalHandler.GetGoal)
		api.PATCH("/goals/:id", goalHandler.UpdateGoal)
		api.DELETE("/goals/:id", goalHandler.DeleteGoal)

		// Status endpoints
		api.GET("/statuses", goalHandler.GetStatuses)
		api.GET("/goals/:id/status", goalHandler.GetGoalStatus)
		api.PATCH("/goals/:id/status", goalHandler.UpdateGoalStatus)
		api.GET("/goals/:id/status/history", goalHandler.GetStatusHistory)

		// Relations endpoints
		api.POST("/goals/relations", goalHandler.CreateRelation)
		api.PATCH("/goals/relations", goalHandler.UpdateRelation)
		api.DELETE("/goals/relations", goalHandler.DeleteRelation)

		// Dependencies endpoints
		api.POST("/goals/dependencies", goalHandler.CreateDependency)
		api.DELETE("/goals/dependencies", goalHandler.DeleteDependency)
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
		log.Printf("Swagger UI available at http://localhost:%s/swagger/index.html", port)
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
