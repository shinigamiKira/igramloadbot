package main

import (
	"insta-scraper/components"
	"insta-scraper/config"
	"log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func main() {
	// Load config
	cfg := config.LoadConfig("config/secrets.json")

	// Create bot
	bot, err := tgbotapi.NewBotAPI(cfg.BotToken)
	if err != nil {
		log.Panic(err)
	}

	bot.Debug = true
	log.Printf("Authorized on account %s", bot.Self.UserName)

	// Set up update config
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	u.Offset = 0

	// Get initial updates to clear queue
	if _, err := bot.GetUpdates(u); err != nil {
		log.Printf("Warning: failed to clear update queue: %v", err)
	}

	// Create updates channel
	updates := bot.GetUpdatesChan(u)

	// Initialize handlers
	handler := components.NewHandler(bot, cfg)

	// Handle updates
	for update := range updates {
		handler.HandleUpdate(update)
	}
}
