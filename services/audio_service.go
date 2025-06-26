package services

import (
	"os"
	"os/exec"
	"regexp"
	"strings"
)

// AudioService handles audio scraping operations
type AudioService struct{}

// NewAudioService creates a new audio service
func NewAudioService() *AudioService {
	return &AudioService{}
}

// ScrapeAudio downloads a video and extracts audio from it
// Returns the audio path and video title
func (s *AudioService) ScrapeAudio(videoURL, audioPath string) (string, error) {
	tmpVideo := "temp_video.mp4"

	// Get video title first
	videoTitle, err := s.getVideoTitle(videoURL)
	if err != nil {
		return "", err
	}

	// Download the video
	if err := s.downloadVideo(videoURL, tmpVideo); err != nil {
		return "", err
	}

	// Extract audio from the video
	if err := s.extractAudio(tmpVideo, audioPath); err != nil {
		// Clean up temp file even if extraction fails
		_ = os.Remove(tmpVideo)
		return "", err
	}

	// Clean up temporary video file
	_ = os.Remove(tmpVideo)
	return videoTitle, nil
}

// getVideoTitle extracts the title of the video from YouTube
func (s *AudioService) getVideoTitle(url string) (string, error) {
	cmd := exec.Command("yt-dlp", "--get-title", "--no-playlist", url)
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}

	title := strings.TrimSpace(string(output))
	// Clean the title to make it safe for filenames
	return s.sanitizeFilename(title), nil
}

// sanitizeFilename removes or replaces characters that are not safe for filenames
func (s *AudioService) sanitizeFilename(filename string) string {
	// Remove or replace invalid characters
	invalidChars := regexp.MustCompile(`[<>:"/\\|?*]`)
	filename = invalidChars.ReplaceAllString(filename, "_")

	// Remove extra spaces and replace with single underscore
	spaces := regexp.MustCompile(`\s+`)
	filename = spaces.ReplaceAllString(filename, "_")

	// Remove leading/trailing underscores
	filename = strings.Trim(filename, "_")

	// Limit length to avoid filesystem issues
	if len(filename) > 100 {
		filename = filename[:100]
	}

	return filename
}

// downloadVideo downloads a video from the given URL
func (s *AudioService) downloadVideo(url, output string) error {
	cmd := exec.Command("yt-dlp", "-o", output, "-f", "bestaudio+bestvideo", url)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// extractAudio extracts audio from a video file
func (s *AudioService) extractAudio(videoPath, audioPath string) error {
	cmd := exec.Command("ffmpeg", "-i", videoPath, "-vn", "-acodec", "libmp3lame", "-ab", "192k", audioPath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
