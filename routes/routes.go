package routes

import (
	"net/http"

	"audio_scrapper/handlers"
)

// SetupRoutes configures all the routes for the application
func SetupRoutes(audioHandler *handlers.AudioHandler, statusHandler *handlers.StatusHandler) {
	// Audio extraction endpoint
	http.HandleFunc("/extract-audio", audioHandler.HandleExtractAudio)

	// Status endpoint
	http.HandleFunc("/status", statusHandler.HandleStatus)

	// Static file serving for audio files
	http.Handle("/audio/", http.StripPrefix("/audio/", http.FileServer(http.Dir("output"))))
}
