package utils

import (
	"os"
	"path/filepath"
)

// EnsureDirectoryExists creates a directory if it doesn't exist
func EnsureDirectoryExists(dirPath string) error {
	if _, err := os.Stat(dirPath); os.IsNotExist(err) {
		return os.MkdirAll(dirPath, 0755)
	}
	return nil
}

// CleanupTempFile removes a temporary file and ignores errors
func CleanupTempFile(filePath string) {
	_ = os.Remove(filePath)
}

// GetFileExtension returns the file extension from a path
func GetFileExtension(filePath string) string {
	return filepath.Ext(filePath)
}

// IsValidAudioFile checks if a file has a valid audio extension
func IsValidAudioFile(filePath string) bool {
	ext := GetFileExtension(filePath)
	validExtensions := []string{".mp3", ".wav", ".flac", ".aac", ".ogg"}

	for _, validExt := range validExtensions {
		if ext == validExt {
			return true
		}
	}
	return false
}
