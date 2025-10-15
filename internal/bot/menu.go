package bot

import (
	"adventBot/internal/db/chat"
	"context"
	"github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func sendWithMenu(ctx context.Context, b *tgbotapi.BotAPI, r chat.Repository, chatID int64, text string) error {
	kb := buildMainKeyboard(ctx, r, chatID)
	msg := tgbotapi.NewMessage(chatID, text)
	msg.ReplyMarkup = kb
	_, err := b.Send(msg)
	return err
}

func sendWithStart(_ context.Context, b *tgbotapi.BotAPI, chatID int64, text string) error {
	kb := tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("/start"),
		),
	)
	kb.ResizeKeyboard = true
	kb.OneTimeKeyboard = false

	msg := tgbotapi.NewMessage(chatID, text)
	msg.ReplyMarkup = kb
	_, err := b.Send(msg)
	return err
}

func buildMainKeyboard(ctx context.Context, r chat.Repository, chatID int64) tgbotapi.ReplyKeyboardMarkup {
	_, found, _ := r.GetById(ctx, chatID)

	if found {
		keyboard := tgbotapi.NewReplyKeyboard(
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton("/today"),
				tgbotapi.NewKeyboardButton("/tasks"),
			),
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton("/trigger"),
				tgbotapi.NewKeyboardButton("/restart"),
			),
		)
		keyboard.ResizeKeyboard = true
		keyboard.OneTimeKeyboard = false
		return keyboard
	}

	keyboard := tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("/start"),
		),
	)
	keyboard.ResizeKeyboard = true
	keyboard.OneTimeKeyboard = false

	return keyboard
}
