// Package services contains the business-logic layer for gts.
package services

import (
	"bytes"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/tahakhan/gts_backend/models"
)

const (
	storageDir         = "./storage"
	defaultFormat      = "mp3"
	maxConcurrency     = 5 // maximum simultaneous yt-dlp processes
	defaultCookiesPath = "./cookies.txt"
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

// cleanURL removes playlist query parameters so yt-dlp strictly processes a single video.
func cleanURL(rawURL string) string {
	u, err := url.Parse(rawURL)
	if err != nil {
		return rawURL
	}
	q := u.Query()
	q.Del("list")
	q.Del("index")
	u.RawQuery = q.Encode()
	return u.String()
}

// resolveCookies returns the flags for yt-dlp cookies if found on disk or environment.
func resolveCookies() []string {
	// 1. Check raw cookie content environment variable
	if content := os.Getenv("YOUTUBE_COOKIES_CONTENT"); strings.TrimSpace(content) != "" {
		tmpPath := filepath.Join(os.TempDir(), "yt_cookies.txt")
		if err := os.WriteFile(tmpPath, []byte(content), 0o600); err == nil {
			return []string{"--cookies", tmpPath}
		}
	}

	// 2. Check explicitly provided cookie path or default local path
	cookiesPath := os.Getenv("./cookies.txt")
	if cookiesPath == "" {
		cookiesPath = defaultCookiesPath
	}

	if _, err := os.Stat(cookiesPath); err == nil {
		return []string{"--cookies", cookiesPath}
	}

	return nil
}

// getBypassArgs constructs arguments to bypass YouTube bot detection.
// getBypassArgs dynamically constructs yt-dlp flags for PO token provider and cookies.
// getBypassArgs dynamically constructs yt-dlp flags for JS runtimes, PO token provider, and cookies.
func getBypassArgs() []string {
	var args []string

	// 1. Force yt-dlp to use node or deno JS runtimes
	args = append(args, "--js-runtimes", "node,deno")

	// 2. Attach PO Token Provider URL if environment variable is present
	poProviderURL := os.Getenv("PO_TOKEN_PROVIDER_URL")
	if poProviderURL != "" {
		// Format: youtube:po_token=web+https://bgutil-ytdlp-pot-provider-aurp.onrender.com/token
		args = append(args, "--extractor-args", fmt.Sprintf("youtube:po_token=web+%s", poProviderURL))
	} else {
		// Fallback player client selection
		args = append(args, "--extractor-args", "youtube:player_client=web,ios")
	}

	// 3. Attach Cookies if present
	if cookieFlags := resolveCookies(); len(cookieFlags) > 0 {
		args = append(args, cookieFlags...)
	}

	return args
}

// Extract downloads audio from the given YouTube URL using yt-dlp.
// It blocks until yt-dlp finishes (or fails) and returns populated AudioMetadata.
func (s *ExtractionService) Extract(req *models.AudioRequest) (*models.AudioMetadata, error) {
	s.semaphore <- struct{}{}
	defer func() { <-s.semaphore }()

	format := req.Format
	if format == "" {
		format = defaultFormat
	}

	targetURL := cleanURL(req.URL)
	id := uuid.New().String()
	outputTemplate := filepath.Join(storageDir, id+".%(ext)s")

	bypassArgs := getBypassArgs()

	// Step 1: Fetch Video Title First (Fast Metadata Call)
	titleArgs := append([]string{"--no-playlist"}, bypassArgs...)
	titleArgs = append(titleArgs, "--print", "%(title)s", targetURL)

	titleCmd := exec.Command("yt-dlp", titleArgs...)
	var titleOut bytes.Buffer
	titleCmd.Stdout = &titleOut
	if err := titleCmd.Run(); err != nil {
		titleOut.WriteString(id)
	}
	title := strings.TrimSpace(titleOut.String())
	if title == "" {
		title = id
	}

	// Step 2: Download & Convert Audio File
	downloadArgs := []string{
		"--no-playlist",
		"-x",
		"--audio-format", format,
		"--audio-quality", "0",
	}
	downloadArgs = append(downloadArgs, bypassArgs...)
	downloadArgs = append(downloadArgs, "-o", outputTemplate, targetURL)

	cmd := exec.Command("yt-dlp", downloadArgs...)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("yt-dlp extraction failed: %s", stderr.String())
	}

	// Step 3: Resolve Actual File Path from Storage
	filePath, err := resolveFilePath(storageDir, id, format)
	if err != nil {
		return nil, err
	}

	info, err := os.Stat(filePath)
	if err != nil {
		return nil, fmt.Errorf("could not stat output file: %w", err)
	}

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

// resolveFilePath finds the actual file yt-dlp wrote by searching the storage directory.
func resolveFilePath(dir, id, format string) (string, error) {
	exact := filepath.Join(dir, id+"."+format)
	if _, err := os.Stat(exact); err == nil {
		return exact, nil
	}

	matches, err := filepath.Glob(filepath.Join(dir, id+".*"))
	if err != nil {
		return "", fmt.Errorf("glob error: %w", err)
	}
	if len(matches) == 0 {
		return "", fmt.Errorf("no output file found for ID %s", id)
	}
	return matches[0], nil
}

// extractTitle derives a human-readable title from the file name.
func extractTitle(filePath string) string {
	base := filepath.Base(filePath)
	ext := filepath.Ext(base)
	return base[:len(base)-len(ext)]
}
