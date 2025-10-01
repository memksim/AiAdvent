package bot

import (
	"context"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

type Handler interface {
	Handle(ctx context.Context, b *bot.Bot, update *models.Update)
}
