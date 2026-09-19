// Package models defines the core data structures used across the gts (Get That Song) backend.
package models

import "time"

// AudioRequest is the incoming payload for the /api/v1/extract endpoint.
// The client supplies the YouTube URL and an optional desired output format.
type AudioRequest struct {
	URL    string `json:"url" validate:"required,url"`
	Format string `json:"format,omitempty"` // "mp3" (default) | "m4a" | "opus"
}

// ExtractionResponse is returned after a successful extraction job is queued / completed.
type ExtractionResponse struct {
	ID          string    `json:"id"`           // UUID used as the download token
	Title       string    `json:"title"`        // Track / video title reported by yt-dlp
	Format      string    `json:"format"`       // Audio format that was extracted
	DownloadURL string    `json:"download_url"` // Relative URL the client can hit to stream the file
	CreatedAt   time.Time `json:"created_at"`
}

// AudioMetadata stores information about a stored audio file on disk.
// It is kept in memory (or can be persisted to a DB in a later iteration).
type AudioMetadata struct {
	ID        string    `json:"id"`
	Title     string    `json:"title"`
	Format    string    `json:"format"`
	FilePath  string    `json:"file_path"`  // Absolute path on disk
	FileSize  int64     `json:"file_size"`  // Bytes
	CreatedAt time.Time `json:"created_at"` // Used by the cleanup ticker
}
