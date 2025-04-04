package utils

import (
	"bytes"
	"fmt"
	"log"
	"os/exec"
	"strings"

	"insta-scraper/pkg/scraper"
)

type DownloadResult struct {
	URLs       []string
	Thumbnail  string
	IsVideo    bool
	PrimaryURL string
}

type MediaDownloader struct{}

func NewMediaDownloader() *MediaDownloader {
	return &MediaDownloader{}
}

func (d *MediaDownloader) DownloadMedia(url string, browserScraper scraper.BrowserScraper) (*DownloadResult, error) {
	// For Instagram posts, try browser first
	if strings.Contains(url, "instagram.com/p/") && browserScraper != nil {
		log.Printf("Using browser scraper for Instagram post: %s", url)
		urls, err := browserScraper.Scrape(url)
		if err != nil {
			log.Printf("Browser scrape failed: %v", err)
		} else if len(urls) > 0 {
			isVideo := false
			for _, u := range urls {
				if strings.Contains(u, ".mp4") {
					isVideo = true
					break
				}
			}
			return &DownloadResult{
				URLs:       urls,
				PrimaryURL: urls[0],
				IsVideo:    isVideo,
			}, nil
		}
	}

	// Fall back to yt-dlp
	log.Printf("Trying yt-dlp for: %s", url)
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
	if err == nil {
		output := strings.Split(strings.TrimSpace(out.String()), "\n")
		if len(output) >= 2 {
			return &DownloadResult{
				PrimaryURL: output[0],
				Thumbnail:  output[1],
				IsVideo:    strings.Contains(output[0], ".mp4"),
				URLs:       []string{output[0]},
			}, nil
		}
	}

	// Fallback to browser scraper if available
	if browserScraper != nil {
		urls, err := browserScraper.Scrape(url)
		if err == nil && len(urls) > 0 {
			isVideo := false
			for _, u := range urls {
				if strings.Contains(u, ".mp4") {
					isVideo = true
					break
				}
			}
			return &DownloadResult{
				URLs:       urls,
				PrimaryURL: urls[0],
				IsVideo:    isVideo,
			}, nil
		}
	}

	return nil, fmt.Errorf("failed to download media: %v\n%s", err, stderr.String())
}
