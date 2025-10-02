package bot

import (
	"adventBot/internal/db"
	"context"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

func sendWithMenu(ctx context.Context, b *bot.Bot, r db.Repository, chatID int64, text string) error {
	kb := buildMainKeyboard(ctx, r, chatID)
	_, err := b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:      chatID,
		Text:        text,
		ReplyMarkup: kb,
	})
	return err
}

func buildMainKeyboard(ctx context.Context, r db.Repository, chatID int64) *models.ReplyKeyboardMarkup {
	_, found, _ := r.GetById(ctx, chatID)

	btn := "/start"
	if found {
		btn = "/restart"
	}

	return &models.ReplyKeyboardMarkup{
		Keyboard: [][]models.KeyboardButton{
			{{Text: btn}},
		},
		ResizeKeyboard:  true,
		OneTimeKeyboard: false,
		Selective:       false,
	}
}
