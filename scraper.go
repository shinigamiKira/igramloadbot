package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Config struct {
	BotToken string `json:"bot_token"`
	ChatID   int64  `json:"chat_id"`
}

var (
	config      Config
	downloadDir = "downloads"
)

func main() {
	// Load config
	loadConfig("secrets.json")

	// Create bot
	bot, err := tgbotapi.NewBotAPI(config.BotToken)
	if err != nil {
		log.Panic(err)
	}

	bot.Debug = true
	log.Printf("Authorized on account %s", bot.Self.UserName)

	// Set up update config with offset
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	u.Offset = 0 // Explicitly set offset

	// Get initial updates to clear the queue
	if _, err := bot.GetUpdates(u); err != nil {
		log.Printf("Warning: failed to clear update queue: %v", err)
	}

	// Create buffered updates channel
	updates := bot.GetUpdatesChan(u)

	// Handle updates with recovery
	for {
		select {
		case update, ok := <-updates:
			if !ok {
				log.Println("Updates channel closed, reconnecting...")
				time.Sleep(5 * time.Second)
				updates = bot.GetUpdatesChan(u)
				continue
			}

			if update.Message != nil {
				handleMessage(bot, update.Message)
			} else if update.InlineQuery != nil {
				handleInlineQuery(bot, update.InlineQuery)
			}
		}
	}
}

func loadConfig(filename string) {
	file, err := os.ReadFile(filename)
	if err != nil {
		log.Panic(err)
	}

	err = json.Unmarshal(file, &config)
	if err != nil {
		log.Panic(err)
	}
}

func handleMessage(bot *tgbotapi.BotAPI, msg *tgbotapi.Message) {
	if !strings.HasPrefix(msg.Text, "http") {
		reply := tgbotapi.NewMessage(msg.Chat.ID, "Please send a valid URL starting with http:// or https://")
		bot.Send(reply)
		return
	}

	// Send processing message
	processingMsg := tgbotapi.NewMessage(msg.Chat.ID, "⏳ Downloading media, please wait...")
	bot.Send(processingMsg)

	// Download media
	result, err := downloadMedia(msg.Text, msg.From.ID)
	if err != nil {
		errorMsg := tgbotapi.NewMessage(msg.Chat.ID, "❌ Failed to download media. Please check the URL and try again.")
		bot.Send(errorMsg)
		return
	}

	// Send media
	if result.IsVideo {
		video := tgbotapi.NewVideo(msg.Chat.ID, tgbotapi.FilePath(result.Path))
		bot.Send(video)
	} else {
		photo := tgbotapi.NewPhoto(msg.Chat.ID, tgbotapi.FilePath(result.Path))
		bot.Send(photo)
	}
}

func handleInlineQuery(bot *tgbotapi.BotAPI, query *tgbotapi.InlineQuery) {
	if !strings.HasPrefix(query.Query, "http") {
		return
	}

	// Download media with debug logging
	log.Printf("Processing inline query for URL: %s", query.Query)
	result, err := downloadMedia(query.Query, query.From.ID)
	if err != nil {
		log.Printf("Download failed for %s: %v", query.Query, err)
		return
	}
	log.Printf("Downloaded media: %s (isVideo: %v)", result.Path, result.IsVideo)

	// Validate chat ID exists
	if config.ChatID == 0 {
		log.Printf("Invalid chat ID configured")
		return
	}

	// Send media to chat to get file ID
	var fileID string
	if result.IsVideo {
		video := tgbotapi.NewVideo(config.ChatID, tgbotapi.FilePath(result.Path))
		msg, err := bot.Send(video)
		if err != nil {
			return
		}
		fileID = msg.Video.FileID
	} else {
		photo := tgbotapi.NewPhoto(config.ChatID, tgbotapi.FilePath(result.Path))
		msg, err := bot.Send(photo)
		if err != nil {
			log.Printf("Failed to send photo: %v", err)
			return
		}
		if len(msg.Photo) == 0 {
			log.Printf("No photo data in message response")
			return
		}
		fileID = msg.Photo[0].FileID
		log.Printf("Got photo file ID: %s", fileID)
	}

	// Answer inline query
	var results []interface{}
	if result.IsVideo {
		// Extract filename for title
		_, filename := filepath.Split(result.Path)
		videoResult := tgbotapi.NewInlineQueryResultCachedVideo("1", fileID, filename)
		results = append(results, videoResult)
	} else {
		photoResult := tgbotapi.NewInlineQueryResultCachedPhoto("1", fileID)
		results = append(results, photoResult)
	}

	inlineConfig := tgbotapi.InlineConfig{
		InlineQueryID: query.ID,
		Results:       results,
		CacheTime:     0,
	}

	bot.Send(inlineConfig)
}

type DownloadResult struct {
	Path    string
	IsVideo bool
}

func downloadMedia(url string, userID int64) (*DownloadResult, error) {
	// Create user directory
	userDir := filepath.Join(downloadDir, fmt.Sprintf("%d", userID))
	os.MkdirAll(userDir, os.ModePerm)
	timestamp := time.Now().Unix()

	// Check if URL is Instagram
	if strings.Contains(url, "instagram.com") {
		// Use yt-dlp for Instagram downloads
		outputTemplate := filepath.Join(userDir, fmt.Sprintf("%d_%%(title)s.%%(ext)s", timestamp))
		cmd := exec.Command("yt-dlp",
			"-o", outputTemplate,
			"--no-playlist",
			"--force-overwrites",
			url,
		)

		var out bytes.Buffer
		var stderr bytes.Buffer
		cmd.Stdout = &out
		cmd.Stderr = &stderr

		err := cmd.Run()
		if err != nil {
			if strings.Contains(stderr.String(), "Private") || strings.Contains(stderr.String(), "Restricted") {
				return nil, fmt.Errorf("this bot doesn't work with private or restricted Instagram posts")
			}
			return nil, fmt.Errorf("failed to download Instagram media: %v\n%s", err, stderr.String())
		}

		// Find downloaded file
		matches, _ := filepath.Glob(filepath.Join(userDir, fmt.Sprintf("%d_*", timestamp)))
		if len(matches) == 0 {
			return nil, fmt.Errorf("no file downloaded from Instagram")
		}

		filePath := matches[0]
		return &DownloadResult{
			Path:    filePath,
			IsVideo: strings.HasSuffix(strings.ToLower(filePath), ".mp4"),
		}, nil
	}

	// Fall back to yt-dlp for non-Instagram URLs
	outputTemplate := filepath.Join(userDir, fmt.Sprintf("%d_%%(title)s.%%(ext)s", timestamp))

	// Try with cookies first if available
	args := []string{
		"-o", outputTemplate,
		"--no-playlist",
		"--extractor-args", "instagram:include_dash_manifest=true;flat=1",
		"--force-overwrites",
		"--retries", "3",
		"--format", "(bestvideo+bestaudio/best)[protocol^=http]",
		"--merge-output-format", "mp4",
		"--convert-thumbnails", "jpg",
		"--verbose",
		"--add-header", "User-Agent:Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
		"--add-header", "Referer:https://www.instagram.com/",
		"--add-header", "Origin:https://www.instagram.com",
	}

	// Only use cookies.txt if exists
	if _, err := os.Stat("cookies.txt"); err == nil {
		args = append(args, "--cookies", "cookies.txt")
	} else {
		log.Printf("Note: cookies.txt not found - some Instagram content may require login")
	}

	args = append(args, url)
	cmd := exec.Command("yt-dlp", args...)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		fullError := fmt.Sprintf("yt-dlp error: %v\n=== Stdout ===\n%s\n=== Stderr ===\n%s",
			err, stdout.String(), stderr.String())
		log.Printf(fullError)
		return nil, fmt.Errorf(fullError)
	}

	// Get downloaded file
	matches, _ := filepath.Glob(filepath.Join(userDir, fmt.Sprintf("%d_*", timestamp)))
	if len(matches) == 0 {
		return nil, fmt.Errorf("no file downloaded")
	}

	filePath := matches[0]
	isVideo := strings.HasSuffix(strings.ToLower(filePath), ".mp4") ||
		strings.HasSuffix(strings.ToLower(filePath), ".mkv") ||
		strings.HasSuffix(strings.ToLower(filePath), ".webm") ||
		strings.HasSuffix(strings.ToLower(filePath), ".mov")

	return &DownloadResult{
		Path:    filePath,
		IsVideo: isVideo,
	}, nil
}
