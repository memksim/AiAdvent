package bot

import (
	"context"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"log"
)

type CommandHandler struct{}

func NewCommandHandler() *CommandHandler {
	return &CommandHandler{}
}

func (_ *CommandHandler) Handle(ctx context.Context, b *bot.Bot, update *models.Update) {
	_, err := b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   "Привет! Готов начать обсуждение.",
	})
	if err != nil {
		log.Println("[CommandHandler.Start] SendMessage:", err)
	}
}
