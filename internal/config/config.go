package config

import (
	"fmt"
	"github.com/joho/godotenv"
	"os"
)

type Config struct {
	BotToken     string
	ApiKey       string
	FolderId     string
	GeonamesUser string
	DbPath       string
	RulePath     string
	RulePathCot  string
}

func Load() (c Config, err error) {
	err = godotenv.Load()
	if err != nil {
		return Config{}, err
	}

	c = Config{
		BotToken:     os.Getenv("TELEGRAM_BOT_TOKEN"),
		ApiKey:       os.Getenv("YC_API_KEY"),
		FolderId:     os.Getenv("YC_FOLDER_ID"),
		GeonamesUser: os.Getenv("GEONAMES_USER"),
		DbPath:       os.Getenv("DB_PATH"),
		RulePath:     os.Getenv("RULE_PATH"),
		RulePathCot:  os.Getenv("RULE_PATH_COT"),
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
