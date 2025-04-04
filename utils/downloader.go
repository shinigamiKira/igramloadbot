package utils

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
)

type DownloadResult struct {
	URL       string
	Thumbnail string
	IsVideo   bool
}

type MediaDownloader struct{}

func NewMediaDownloader() *MediaDownloader {
	return &MediaDownloader{}
}

func (d *MediaDownloader) DownloadMedia(url string) (*DownloadResult, error) {
	// Get URL and thumbnail using yt-dlp
	args := []string{
		"--get-url",
		"--get-thumbnail",
		"--format", "best",
		url,
	}

	cmd := exec.Command("yt-dlp", args...)
	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		return nil, fmt.Errorf("yt-dlp error: %v\n%s", err, stderr.String())
	}

	output := strings.Split(strings.TrimSpace(out.String()), "\n")
	if len(output) < 2 {
		return nil, fmt.Errorf("invalid yt-dlp output")
	}

	return &DownloadResult{
		URL:       output[0],
		Thumbnail: output[1],
		IsVideo:   strings.Contains(output[0], ".mp4"),
	}, nil
}
