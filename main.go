// gts — Get That Song
// Entry point: initialises the Echo HTTP server, wires services, and starts background workers.
package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/tahakhan/gts_backend/routes"
	"github.com/tahakhan/gts_backend/services"
)

func main() {
	// ── Service Layer ─────────────────────────────────────────────────────────
	extractor, err := services.NewExtractionService()
	if err != nil {
		log.Fatalf("failed to initialise extraction service: %v", err)
	}

	_, err = services.NewStorageService("./storage")
	if err != nil {
		log.Fatalf("failed to initialise storage service: %v", err)
	}

	// done channel used to signal background workers to stop.
	done := make(chan struct{})
	services.StartCleanupWorker(extractor, done)

	// ── Echo Setup ────────────────────────────────────────────────────────────
	e := echo.New()
	e.HideBanner = false

	// Global middleware
	e.Use(middleware.Logger())  // structured request/response logging
	e.Use(middleware.Recover()) // recover from panics; return 500 instead of crashing
	e.Use(middleware.CORS())    // allow Flutter app on a different origin

	// ── Route Registration ────────────────────────────────────────────────────
	routes.Register(e, extractor)

	// ── Graceful Shutdown ─────────────────────────────────────────────────────
	port := getenv("PORT", "8080")

	// Start server in a goroutine so we can listen for OS signals concurrently.
	go func() {
		log.Printf("🎵 gts server listening on :%s", port)
		if err := e.Start(":" + port); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}()

	// Block until SIGINT or SIGTERM.
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("shutting down — waiting for in-flight requests to complete…")
	close(done) // signal cleanup worker to stop

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := e.Shutdown(ctx); err != nil {
		log.Fatalf("forced shutdown: %v", err)
	}
	log.Println("server exited cleanly")
}

// getenv returns the value of an environment variable or a fallback default.
func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
