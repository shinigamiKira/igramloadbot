package config

import (
	"encoding/json"
	"os"
)

type Config struct {
	BotToken    string `json:"bot_token"`
	ChatID      int64  `json:"chat_id"`
	BrowserType string `json:"browser_type"` // brave/chrome/edge
	BrowserPath string `json:"browser_path"` // path to executable
}

func LoadConfig(filename string) *Config {
	file, err := os.ReadFile(filename)
	if err != nil {
		panic(err)
	}

	var cfg Config
	if err := json.Unmarshal(file, &cfg); err != nil {
		panic(err)
	}
	return &cfg
}
