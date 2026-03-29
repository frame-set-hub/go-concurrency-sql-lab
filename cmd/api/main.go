package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go-concurrency-sql-lab/internal/analytics"
	"go-concurrency-sql-lab/internal/db"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, relying on OS vars")
	}

	// Create a root context that we can cancel on SIGINT (Ctrl+C)
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	log.Println("Connecting to Database...")
	pool, err := db.NewPool(ctx)
	if err != nil {
		log.Fatalf("Failed to initialize database pool: %v", err)
	}
	defer pool.Close()

	// Initialize the Analytics Service
	analyticsSvc := analytics.NewAnalyticsService(pool)
	
	// Phase 5: Start Background Job (updates cache every 30 seconds)
	// Passing our root `ctx` so it gracefully stops when Ctrl+C is pressed.
	go analyticsSvc.StartBackgroundJob(ctx, 30*time.Second)

	// Phase 5: Build REST API (Gin)
	r := setupRouter(analyticsSvc)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: r,
	}

	// Start Server entirely in a Goroutine so we can listen to shutdown signals
	go func() {
		log.Printf("🌐 Server starting on port %s...", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server Listen error: %s\n", err)
		}
	}()

	// Wait for interrupt signal to gracefully shut down the server
	<-ctx.Done()
	log.Println("🛑 Shutting down server gracefully...")

	// Graceful shutdown with 5 seconds timeout
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("✅ Server exiting")
}

func setupRouter(svc *analytics.AnalyticsService) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()

	api := r.Group("/api")
	{
		// GET /api/analytics
		// Serving data instantly from cache populated by background goroutine
		api.GET("/analytics", func(c *gin.Context) {
			dash := svc.GetCachedDashboard()
			if dash == nil {
				c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Analytics data is still compiling in the background, please try again."})
				return
			}
			c.JSON(http.StatusOK, dash)
		})
	}

	return r
}
