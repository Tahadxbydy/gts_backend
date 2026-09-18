package services

import (
	"log"
	"time"
)

const (
	// cleanupInterval defines how often the ticker fires.
	cleanupInterval = 10 * time.Minute
	// fileMaxAge is the maximum age a stored audio file may reach before deletion.
	fileMaxAge = 1 * time.Hour
)

// StartCleanupWorker launches a background goroutine that periodically scans the
// in-memory store and removes audio files that are older than fileMaxAge.
// The goroutine exits cleanly when the provided done channel is closed.
func StartCleanupWorker(svc *ExtractionService, done <-chan struct{}) {
	ticker := time.NewTicker(cleanupInterval)
	go func() {
		defer ticker.Stop()
		log.Printf("[cleanup] worker started — interval=%s, max_age=%s", cleanupInterval, fileMaxAge)

		for {
			select {
			case <-done:
				log.Println("[cleanup] worker stopped")
				return

			case t := <-ticker.C:
				log.Printf("[cleanup] tick at %s — scanning for expired files", t.Format(time.RFC3339))
				runCleanup(svc)
			}
		}
	}()
}

// runCleanup iterates all known metadata entries and deletes those that have exceeded fileMaxAge.
func runCleanup(svc *ExtractionService) {
	entries := svc.AllMetadata()
	cutoff := time.Now().UTC().Add(-fileMaxAge)
	deleted := 0

	for _, meta := range entries {
		if meta.CreatedAt.Before(cutoff) {
			log.Printf("[cleanup] deleting expired file: id=%s file=%s age=%s",
				meta.ID, meta.FilePath, time.Since(meta.CreatedAt).Round(time.Second))
			svc.Delete(meta.ID)
			deleted++
		}
	}

	log.Printf("[cleanup] sweep complete — deleted=%d remaining=%d", deleted, len(entries)-deleted)
}
