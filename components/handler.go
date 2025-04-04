package components

import (
	"fmt"
	"strings"

	"insta-scraper/config"
	"insta-scraper/utils"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Handler struct {
	bot        *tgbotapi.BotAPI
	cfg        *config.Config
	downloader *utils.MediaDownloader
}

func NewHandler(bot *tgbotapi.BotAPI, cfg *config.Config) *Handler {
	return &Handler{
		bot:        bot,
		cfg:        cfg,
		downloader: utils.NewMediaDownloader(),
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

	// Show processing message
	processingMsg := tgbotapi.NewMessage(msg.Chat.ID, "⏳ Processing your request...")
	h.bot.Send(processingMsg)

	// Download media
	result, err := h.downloader.DownloadMedia(msg.Text)
	if err != nil {
		errorMsg := tgbotapi.NewMessage(msg.Chat.ID, "❌ Failed to process URL")
		h.bot.Send(errorMsg)
		return
	}

	// Send result
	msgText := fmt.Sprintf("[%s](%s)", "Download link", result.URL)
	reply := tgbotapi.NewMessage(msg.Chat.ID, msgText)
	reply.ParseMode = "Markdown"
	h.bot.Send(reply)
}

func (h *Handler) handleInlineQuery(query *tgbotapi.InlineQuery) {
	if !strings.HasPrefix(query.Query, "http") {
		return
	}

	result, err := h.downloader.DownloadMedia(query.Query)
	if err != nil {
		return
	}

	var results []interface{}
	if result.IsVideo {
		video := tgbotapi.NewInlineQueryResultVideo("1", result.URL)
		video.ThumbURL = result.Thumbnail
		video.Title = "Video Download"
		video.MimeType = "video/mp4" // Explicitly set MIME type
		results = append(results, video)
	} else {
		photo := tgbotapi.NewInlineQueryResultPhoto("2", result.Thumbnail)
		photo.Title = "Photo Download"
		results = append(results, photo)
	}

	inlineConf := tgbotapi.InlineConfig{
		InlineQueryID: query.ID,
		Results:       results,
	}

	h.bot.Send(inlineConf)
}
