package services

import (
	"fmt"
	"os"
)

// StorageService provides helpers for local file system operations that are
// not tightly coupled to the extraction pipeline.
type StorageService struct {
	baseDir string
}

// NewStorageService constructs a StorageService rooted at baseDir.
func NewStorageService(baseDir string) (*StorageService, error) {
	if err := os.MkdirAll(baseDir, 0o755); err != nil {
		return nil, fmt.Errorf("storage: failed to create base directory %q: %w", baseDir, err)
	}
	return &StorageService{baseDir: baseDir}, nil
}

// BaseDir returns the directory where files are stored.
func (s *StorageService) BaseDir() string {
	return s.baseDir
}

// FileExists reports whether a file with the given name exists inside the base directory.
func (s *StorageService) FileExists(filename string) bool {
	path := fmt.Sprintf("%s/%s", s.baseDir, filename)
	_, err := os.Stat(path)
	return err == nil
}

// RemoveFile deletes a file by name from the base directory.
func (s *StorageService) RemoveFile(filename string) error {
	path := fmt.Sprintf("%s/%s", s.baseDir, filename)
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("storage: failed to remove %q: %w", path, err)
	}
	return nil
}

// DiskUsageBytes returns the total bytes consumed by all files in the base directory.
func (s *StorageService) DiskUsageBytes() (int64, error) {
	entries, err := os.ReadDir(s.baseDir)
	if err != nil {
		return 0, fmt.Errorf("storage: failed to read directory: %w", err)
	}

	var total int64
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		info, err := e.Info()
		if err != nil {
			continue
		}
		total += info.Size()
	}
	return total, nil
}
