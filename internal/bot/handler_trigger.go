package bot

import (
	"adventBot/internal/service"
	"context"
	"github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
)

type TriggerHandler struct {
	manager *service.SchedulerManager
}

func NewTriggerHandler(m *service.SchedulerManager) *TriggerHandler {
	return &TriggerHandler{m}
}

func (h *TriggerHandler) Handle(_ context.Context, b *tgbotapi.BotAPI, update *tgbotapi.Update) {
	if update == nil || update.Message == nil {
		return
	}

	chatID := update.Message.Chat.ID
	if s := h.manager.GetScheduler(chatID); s != nil {
		s.ProcessNow()
	} else {
		scheduleMsg := tgbotapi.NewMessage(chatID, "Scheduler не найден")
		if _, err := b.Send(scheduleMsg); err != nil {
			log.Println("[TriggerHandler.Handle] SendMessage:", err)
		}
	}
}
