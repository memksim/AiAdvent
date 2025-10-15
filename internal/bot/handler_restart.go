package bot

import (
	"adventBot/internal/db/chat"
	"adventBot/internal/service"
	"context"
	"github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
)

type ResetHandler struct {
	Repository chat.Repository
	manager    *service.SchedulerManager
}

func NewResetHandler(r chat.Repository, m *service.SchedulerManager) *ResetHandler {
	return &ResetHandler{r, m}
}

func (h *ResetHandler) Handle(ctx context.Context, b *tgbotapi.BotAPI, upd *tgbotapi.Update) {
	if upd.Message == nil {
		return
	}
	chatID := upd.Message.Chat.ID

	ok, err := h.Repository.DeleteById(ctx, chatID)
	if err != nil {
		log.Printf("[ResetHandler.Handle] delete error chatID=%d err=%v", chatID, err)
		msg := tgbotapi.NewMessage(chatID, "Произошла ошибка при сбросе 😔")
		_, _ = b.Send(msg)
		return
	}

	if ok {
		log.Printf("[ResetHandler.Handle] tz cleared chatID=%d", chatID)
		keyboard := tgbotapi.NewReplyKeyboard(
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton("/start"),
			),
		)
		keyboard.ResizeKeyboard = true

		h.manager.RemoveScheduler(chatID)

		msg := tgbotapi.NewMessage(chatID, "Данные сброшены, scheduler удален. Нажмите /start, чтобы начать заново")
		msg.ReplyMarkup = keyboard
		_, _ = b.Send(msg)
	} else {
		log.Printf("[ResetHandler.Handle] nothing to delete chatID=%d", chatID)
		msg := tgbotapi.NewMessage(chatID, "Для этого чата не было сохранённых данных.")
		_, _ = b.Send(msg)
	}
}
