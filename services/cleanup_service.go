package services

import (
	"log"
	"os"
	"sync"
	"time"
)

// FileInfo tracks downloaded files and their deletion time
type FileInfo struct {
	Path       string
	DeleteTime time.Time
}

// CleanupService manages automatic file deletion
type CleanupService struct {
	files     map[string]*FileInfo
	mutex     sync.RWMutex
	stopChan  chan bool
	isRunning bool
}

// NewCleanupService creates a new cleanup service
func NewCleanupService() *CleanupService {
	return &CleanupService{
		files:    make(map[string]*FileInfo),
		stopChan: make(chan bool),
	}
}

// Start begins the cleanup goroutine
func (cs *CleanupService) Start() {
	cs.mutex.Lock()
	defer cs.mutex.Unlock()

	if cs.isRunning {
		return
	}

	cs.isRunning = true
	go cs.cleanupRoutine()
	log.Println("Cleanup service started - files will be deleted after 1 hour")
}

// Stop stops the cleanup goroutine
func (cs *CleanupService) Stop() {
	cs.mutex.Lock()
	defer cs.mutex.Unlock()

	if !cs.isRunning {
		return
	}

	cs.isRunning = false
	cs.stopChan <- true
	log.Println("Cleanup service stopped")
}

// RegisterFile registers a file for automatic deletion after 1 hour
func (cs *CleanupService) RegisterFile(filePath string) {
	cs.mutex.Lock()
	defer cs.mutex.Unlock()

	// Calculate deletion time (1 hour from now)
	deleteTime := time.Now().Add(1 * time.Hour)

	cs.files[filePath] = &FileInfo{
		Path:       filePath,
		DeleteTime: deleteTime,
	}

	log.Printf("Registered file for deletion: %s (will be deleted at %s)", filePath, deleteTime.Format("15:04:05"))
}

// UnregisterFile removes a file from the cleanup list
func (cs *CleanupService) UnregisterFile(filePath string) {
	cs.mutex.Lock()
	defer cs.mutex.Unlock()

	delete(cs.files, filePath)
	log.Printf("Unregistered file from cleanup: %s", filePath)
}

// cleanupRoutine runs in the background and deletes expired files
func (cs *CleanupService) cleanupRoutine() {
	ticker := time.NewTicker(5 * time.Minute) // Check every 5 minutes
	defer ticker.Stop()

	for {
		select {
		case <-cs.stopChan:
			return
		case <-ticker.C:
			cs.cleanupExpiredFiles()
		}
	}
}

// cleanupExpiredFiles removes files that have passed their deletion time
func (cs *CleanupService) cleanupExpiredFiles() {
	cs.mutex.Lock()
	defer cs.mutex.Unlock()

	now := time.Now()
	var filesToDelete []string

	// Find expired files
	for filePath, fileInfo := range cs.files {
		if now.After(fileInfo.DeleteTime) {
			filesToDelete = append(filesToDelete, filePath)
		}
	}

	// Delete expired files
	for _, filePath := range filesToDelete {
		if err := os.Remove(filePath); err != nil {
			log.Printf("Failed to delete expired file %s: %v", filePath, err)
		} else {
			log.Printf("Deleted expired file: %s", filePath)
		}
		delete(cs.files, filePath)
	}
}

// GetRegisteredFiles returns a list of currently registered files
func (cs *CleanupService) GetRegisteredFiles() []string {
	cs.mutex.RLock()
	defer cs.mutex.RUnlock()

	var files []string
	for filePath := range cs.files {
		files = append(files, filePath)
	}
	return files
}

// GetFileCount returns the number of registered files
func (cs *CleanupService) GetFileCount() int {
	cs.mutex.RLock()
	defer cs.mutex.RUnlock()

	return len(cs.files)
}
