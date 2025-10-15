package bot

import (
	"context"
	"github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Handler interface {
	Handle(ctx context.Context, b *tgbotapi.BotAPI, update *tgbotapi.Update)
}
