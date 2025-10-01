package config

import (
	"fmt"
	"github.com/joho/godotenv"
	"os"
)

type YandexConfig struct {
	BotToken string
	ApiKey   string
	FolderId string
}

func Load() (c YandexConfig, err error) {
	err = godotenv.Load()
	if err != nil {
		return YandexConfig{}, err
	}

	c = YandexConfig{
		BotToken: os.Getenv("TELEGRAM_BOT_TOKEN"),
		ApiKey:   os.Getenv("YC_API_KEY"),
		FolderId: os.Getenv("YC_FOLDER_ID"),
	}

	if c.BotToken == "" {
		return c, fmt.Errorf("TELEGRAM_BOT_TOKEN is required")
	}
	if c.ApiKey == "" {
		return c, fmt.Errorf("YC_API_KEY is required")
	}
	if c.FolderId == "" {
		return c, fmt.Errorf("YC_FOLDER_ID is required")
	}

	return c, nil
}
