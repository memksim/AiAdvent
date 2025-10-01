package main

import (
	"adventBot/internal/ai_model/yandex"
	internalbot "adventBot/internal/bot"
	"adventBot/internal/config"
	"context"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"log"
	"os"
	"os/signal"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	model := yandex.NewAiModelYandex(cfg.ApiKey, cfg.FolderId)

	b, err := bot.New(cfg.BotToken, bot.WithDefaultHandler(func(ctx context.Context, b *bot.Bot, u *models.Update) {}))
	if err != nil {
		log.Fatal(err)
	}

	cmd := internalbot.NewCommandHandler()
	txt := internalbot.NewTextHandler(model)

	b.RegisterHandler(bot.HandlerTypeMessageText, "/start", bot.MatchTypeExact, cmd.Handle)
	b.RegisterHandler(bot.HandlerTypeMessageText, "", bot.MatchTypePrefix, txt.Handle)

	b.Start(ctx)
}
