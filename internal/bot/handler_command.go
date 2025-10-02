package bot

import (
	"adventBot/internal/db"
	"context"
	"fmt"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"log"
)

type CommandHandler struct {
	Repository db.Repository
}

func NewCommandHandler(r db.Repository) *CommandHandler {
	return &CommandHandler{Repository: r}
}

func (h *CommandHandler) Handle(ctx context.Context, b *bot.Bot, update *models.Update) {
	if update == nil || update.Message == nil {
		return
	}
	chatID := update.Message.Chat.ID

	tz, found, err := h.Repository.GetById(ctx, chatID)
	if err != nil {
		log.Printf("[CommandHandler.Handle] GetById error chatID=%d err=%v", chatID, err)
		found = false
	}

	if found {
		msg := fmt.Sprintf(
			"Мы уже знакомы! \nТвой текущий часовой пояс: %s.\n\n"+
				"Если хочешь начать заново — отправь /reset.\n"+
				"Можешь также поделиться геопозицией, чтобы обновить часовой пояс.",
			tz,
		)

		if err := sendWithMenu(ctx, b, h.Repository, chatID, msg); err != nil {
			log.Println("[CommandHandler.Handle] SendWithMenu:", err)
		}
		return
	}

	intro := "Привет! Я помогу с ведением твоего списка дел. " +
		"Нажми «Поделиться локацией», чтобы я корректно определил твой часовой пояс."

	if err := sendWithMenu(ctx, b, h.Repository, chatID, intro); err != nil {
		log.Println("[CommandHandler.Handle] SendWithMenu:", err)
	}

	locKb := &models.ReplyKeyboardMarkup{
		Keyboard:        [][]models.KeyboardButton{{{Text: "Поделиться локацией", RequestLocation: true}}},
		ResizeKeyboard:  true,
		OneTimeKeyboard: true,
	}
	if _, err := b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:      chatID,
		Text:        "Нажми «Поделиться локацией», чтобы определить часовой пояс.",
		ReplyMarkup: locKb,
	}); err != nil {
		log.Println("[CommandHandler.Handle] SendMessage:", err)
	}
}
