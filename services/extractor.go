// Package services contains the business-logic layer for gts.
package services

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/tahakhan/gts_backend/models"
)

const (
	storageDir     = "./storage"
	defaultFormat  = "mp3"
	maxConcurrency = 5 // maximum simultaneous yt-dlp processes
)

// ExtractionService handles audio extraction from YouTube URLs via yt-dlp.
type ExtractionService struct {
	mu        sync.RWMutex
	store     map[string]*models.AudioMetadata // in-memory index keyed by ID
	semaphore chan struct{}                    // limits concurrent extractions
}

// NewExtractionService constructs and returns a ready-to-use ExtractionService.
// It creates the storage directory if it does not already exist.
func NewExtractionService() (*ExtractionService, error) {
	if err := os.MkdirAll(storageDir, 0o755); err != nil {
		return nil, fmt.Errorf("failed to create storage directory: %w", err)
	}

	return &ExtractionService{
		store:     make(map[string]*models.AudioMetadata),
		semaphore: make(chan struct{}, maxConcurrency),
	}, nil
}

// Extract downloads audio from the given YouTube URL using yt-dlp.
// It blocks until yt-dlp finishes (or fails) and returns populated AudioMetadata.
func (s *ExtractionService) Extract(req *models.AudioRequest) (*models.AudioMetadata, error) {
	// Acquire a concurrency slot.
	s.semaphore <- struct{}{}
	defer func() { <-s.semaphore }()

	format := req.Format
	if format == "" {
		format = defaultFormat
	}

	id := uuid.New().String()
	outputTemplate := filepath.Join(storageDir, id+".%(ext)s")

	// Build yt-dlp arguments:
	//   -x                 : extract audio only
	//   --audio-format     : convert to the requested format
	//   --audio-quality 0  : best quality
	//   --no-playlist      : ignore playlists, download only the single video
	//   --print-json       : print JSON metadata to stdout (we ignore it here but useful for debugging)
	//   -o <template>      : output filename template
	args := []string{
		"-x",
		"--audio-format", format,
		"--audio-quality", "0",
		"--no-playlist",
		"--ffmpeg-location", "/usr/local/bin", // common homebrew path; adjust if needed
		"-o", outputTemplate,
		req.URL,
	}

	cmd := exec.Command("yt-dlp", args...)
	cmd.Stdout = os.Stdout // pipe yt-dlp output to server logs
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("yt-dlp extraction failed: %w", err)
	}

	// Resolve the actual file path (yt-dlp may alter the extension).
	filePath, err := resolveFilePath(storageDir, id, format)
	if err != nil {
		return nil, err
	}

	// Stat the file to get its size.
	info, err := os.Stat(filePath)
	if err != nil {
		return nil, fmt.Errorf("could not stat output file: %w", err)
	}

	// Extract the title from the filename as a fallback.
	title := extractTitle(filePath)

	meta := &models.AudioMetadata{
		ID:        id,
		Title:     title,
		Format:    format,
		FilePath:  filePath,
		FileSize:  info.Size(),
		CreatedAt: time.Now().UTC(),
	}

	s.mu.Lock()
	s.store[id] = meta
	s.mu.Unlock()

	return meta, nil
}

// GetByID returns the AudioMetadata for a given ID, or false if not found.
func (s *ExtractionService) GetByID(id string) (*models.AudioMetadata, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	meta, ok := s.store[id]
	return meta, ok
}

// Delete removes a file from disk and from the in-memory index.
func (s *ExtractionService) Delete(id string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	meta, ok := s.store[id]
	if !ok {
		return
	}

	_ = os.Remove(meta.FilePath)
	delete(s.store, id)
}

// AllMetadata returns a snapshot of all stored metadata entries.
func (s *ExtractionService) AllMetadata() []*models.AudioMetadata {
	s.mu.RLock()
	defer s.mu.RUnlock()

	list := make([]*models.AudioMetadata, 0, len(s.store))
	for _, m := range s.store {
		list = append(list, m)
	}
	return list
}

// resolveFilePath finds the actual file yt-dlp wrote by globbing the storage dir.
func resolveFilePath(dir, id, format string) (string, error) {
	// First try the exact expected path.
	exact := filepath.Join(dir, id+"."+format)
	if _, err := os.Stat(exact); err == nil {
		return exact, nil
	}

	// Fall back to a glob in case yt-dlp used a different extension.
	matches, err := filepath.Glob(filepath.Join(dir, id+".*"))
	if err != nil {
		return "", fmt.Errorf("glob error: %w", err)
	}
	if len(matches) == 0 {
		return "", fmt.Errorf("no output file found for ID %s", id)
	}
	return matches[0], nil
}

// extractTitle derives a human-readable title from the file name (sans extension and UUID).
// yt-dlp names files after the video title, so this is a no-op in most cases.
// Since we use an ID-based template, the title will just be the ID unless overridden later.
func extractTitle(filePath string) string {
	base := filepath.Base(filePath)
	ext := filepath.Ext(base)
	return base[:len(base)-len(ext)]
}
