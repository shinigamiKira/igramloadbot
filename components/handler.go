package components

import (
	"fmt"
	"log"
	"strings"

	"insta-scraper/config"
	"insta-scraper/pkg/scraper"
	"insta-scraper/utils"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Handler struct {
	bot            *tgbotapi.BotAPI
	cfg            *config.Config
	browserScraper scraper.BrowserScraper
	downloader     *utils.MediaDownloader
}

func NewHandler(bot *tgbotapi.BotAPI, cfg *config.Config) *Handler {
	var browserScraper scraper.BrowserScraper
	if cfg.BrowserPath != "" && cfg.BrowserType != "" {
		browserScraper = NewBrowserScraper(cfg.BrowserType, cfg.BrowserPath)
	}

	return &Handler{
		bot:            bot,
		cfg:            cfg,
		downloader:     utils.NewMediaDownloader(),
		browserScraper: browserScraper,
	}
}

func (h *Handler) HandleUpdate(update tgbotapi.Update) {
	if update.Message != nil {
		h.handleMessage(update.Message)
	} else if update.InlineQuery != nil {
		h.handleInlineQuery(update.InlineQuery)
	}
}

func (h *Handler) handleMessage(msg *tgbotapi.Message) {
	if !strings.HasPrefix(msg.Text, "http") {
		reply := tgbotapi.NewMessage(msg.Chat.ID, "Please send a valid URL starting with http:// or https://")
		h.bot.Send(reply)
		return
	}

	processingMsg := tgbotapi.NewMessage(msg.Chat.ID, "⏳ Processing your request...")
	h.bot.Send(processingMsg)

	result, err := h.downloader.DownloadMedia(msg.Text, h.browserScraper)
	if err != nil {
		errorMsg := tgbotapi.NewMessage(msg.Chat.ID, "❌ Failed to process URL")
		h.bot.Send(errorMsg)
		return
	}

	msgText := fmt.Sprintf("[%s](%s)", "link", result.PrimaryURL)
	reply := tgbotapi.NewMessage(msg.Chat.ID, msgText)
	reply.ParseMode = "Markdown"
	h.bot.Send(reply)
}

func (h *Handler) handleInlineQuery(query *tgbotapi.InlineQuery) {
	if !strings.HasPrefix(query.Query, "http") {
		return
	}

	processingResult := tgbotapi.NewInlineQueryResultArticle("processing", "Processing...", "Fetching media...")

	_, err := h.bot.Request(tgbotapi.InlineConfig{
		InlineQueryID: query.ID,
		Results:       []interface{}{processingResult},
		CacheTime:     1,
	})
	if err != nil {
		log.Printf("Failed to send processing message: %v", err)
	}

	go func(q tgbotapi.InlineQuery) {
		result, err := h.downloader.DownloadMedia(q.Query, h.browserScraper)
		if err != nil {
			log.Printf("Failed to download media: %v", err)
			return
		}

		var results []interface{}
		if result.IsVideo {
			video := tgbotapi.NewInlineQueryResultVideo("1", result.PrimaryURL)
			video.ThumbURL = result.Thumbnail
			video.Title = "Video Download"
			video.MimeType = "video/mp4"
			results = append(results, video)
		} else {
			photoURL := result.PrimaryURL
			if len(result.URLs) > 0 {
				photoURL = result.URLs[0]
			}
			photo := tgbotapi.NewInlineQueryResultPhoto("2", photoURL)
			photo.ThumbURL = photoURL
			photo.Title = "Photo Download"
			results = append(results, photo)
		}

		_, err = h.bot.Request(tgbotapi.InlineConfig{
			InlineQueryID: q.ID,
			Results:       results,
			CacheTime:     1,
		})
		if err != nil {
			log.Printf("Failed to send results: %v", err)
		}
	}(*query)
}
