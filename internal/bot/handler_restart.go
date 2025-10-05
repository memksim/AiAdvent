package bot

import (
	"adventBot/internal/db/chat"
	"context"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"log"
)

type ResetHandler struct {
	Repository chat.Repository
}

func NewResetHandler(r chat.Repository) *ResetHandler {
	return &ResetHandler{Repository: r}
}

func (h *ResetHandler) Handle(ctx context.Context, b *bot.Bot, upd *models.Update) {
	if upd.Message == nil {
		return
	}
	chatID := upd.Message.Chat.ID

	ok, err := h.Repository.DeleteById(ctx, chatID)
	if err != nil {
		log.Printf("[ResetHandler.Handle] delete error chatID=%d err=%v", chatID, err)
		_, _ = b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: chatID,
			Text:   "Произошла ошибка при сбросе 😔",
		})
		return
	}

	if ok {
		log.Printf("[ResetHandler.Handle] tz cleared chatID=%d", chatID)
		_, _ = b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: chatID,
			Text:   "Данные сброшены. Нажмите /start, чтобы начать заново",
			ReplyMarkup: &models.ReplyKeyboardMarkup{
				Keyboard: [][]models.KeyboardButton{
					{{Text: "/start"}},
				},
				ResizeKeyboard: true,
			},
		})
	} else {
		log.Printf("[ResetHandler.Handle] nothing to delete chatID=%d", chatID)
		_, _ = b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: chatID,
			Text:   "Для этого чата не было сохранённых данных.",
		})
	}
}
