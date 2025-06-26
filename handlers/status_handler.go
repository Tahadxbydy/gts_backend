package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"audio_scrapper/services"
)

// StatusHandler handles status and monitoring requests
type StatusHandler struct {
	cleanupService *services.CleanupService
}

// NewStatusHandler creates a new status handler
func NewStatusHandler(cleanupService *services.CleanupService) *StatusHandler {
	return &StatusHandler{
		cleanupService: cleanupService,
	}
}

// HandleStatus handles GET requests to check service status
func (h *StatusHandler) HandleStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Only GET method is allowed", http.StatusMethodNotAllowed)
		return
	}

	registeredFiles := h.cleanupService.GetRegisteredFiles()
	fileCount := h.cleanupService.GetFileCount()

	status := map[string]interface{}{
		"status":           "running",
		"timestamp":        time.Now().Format(time.RFC3339),
		"registered_files": fileCount,
		"files":            registeredFiles,
		"cleanup_interval": "5 minutes",
		"file_lifetime":    "1 hour",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(status)
}
