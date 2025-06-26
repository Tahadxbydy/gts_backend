package handlers

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"

	"audio_scrapper/models"
	"audio_scrapper/services"
)

// AudioHandler handles audio extraction requests
type AudioHandler struct {
	audioService   *services.AudioService
	cleanupService *services.CleanupService
}

// NewAudioHandler creates a new audio handler
func NewAudioHandler(audioService *services.AudioService, cleanupService *services.CleanupService) *AudioHandler {
	return &AudioHandler{
		audioService:   audioService,
		cleanupService: cleanupService,
	}
}

// HandleExtractAudio handles POST requests to extract audio from a URL
func (h *AudioHandler) HandleExtractAudio(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST method is allowed", http.StatusMethodNotAllowed)
		return
	}

	var payload models.RequestPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if payload.URL == "" {
		http.Error(w, "Missing 'url' field", http.StatusBadRequest)
		return
	}

	// Generate filename with video title
	filename := generateAudioFilename()
	outputPath := filepath.Join("output", filename)

	// Scrape audio and get video title
	videoTitle, err := h.audioService.ScrapeAudio(payload.URL, outputPath)
	if err != nil {
		http.Error(w, "Failed to process audio: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Use video title for the final filename
	finalFilename := videoTitle + ".mp3"
	finalOutputPath := filepath.Join("output", finalFilename)

	// Rename the file to use the video title
	if err := h.renameFile(outputPath, finalOutputPath); err != nil {
		http.Error(w, "Failed to rename audio file: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Register the file for automatic cleanup after 1 hour
	h.cleanupService.RegisterFile(finalOutputPath)

	// Respond with public URL and video title
	audioURL := "/audio/" + finalFilename
	response := map[string]string{
		"audio_url":   audioURL,
		"video_title": videoTitle,
		"filename":    finalFilename,
		"expires_in":  "1 hour",
	}
	json.NewEncoder(w).Encode(response)
}

// generateAudioFilename generates a temporary filename for audio files
func generateAudioFilename() string {
	return "temp_audio.mp3"
}

// renameFile renames a file from oldPath to newPath
func (h *AudioHandler) renameFile(oldPath, newPath string) error {
	return os.Rename(oldPath, newPath)
}
