package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"audio_scrapper/handlers"
	"audio_scrapper/routes"
	"audio_scrapper/services"
	"audio_scrapper/utils"
)

func main() {
	// Ensure output directory exists
	if err := utils.EnsureDirectoryExists("output"); err != nil {
		log.Fatal("Failed to create output directory:", err)
	}

	// Initialize services
	audioService := services.NewAudioService()
	cleanupService := services.NewCleanupService()

	// Start the cleanup service
	cleanupService.Start()

	// Initialize handlers
	audioHandler := handlers.NewAudioHandler(audioService, cleanupService)
	statusHandler := handlers.NewStatusHandler(cleanupService)

	// Setup routes
	routes.SetupRoutes(audioHandler, statusHandler)

	// Setup graceful shutdown
	setupGracefulShutdown(cleanupService)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080" // fallback for local testing
	}

	log.Printf("Server started on port %s\n", port)

	log.Println("Audio files will be automatically deleted after 1 hour")
	log.Println("Check /status for cleanup service information")

	log.Fatal(http.ListenAndServe(":"+port, nil))
}

// setupGracefulShutdown handles graceful shutdown of the application
func setupGracefulShutdown(cleanupService *services.CleanupService) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-c
		fmt.Println("\nShutting down gracefully...")
		cleanupService.Stop()
		os.Exit(0)
	}()
}
