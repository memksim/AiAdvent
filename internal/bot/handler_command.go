package bot

import (
	"adventBot/internal/db/chat"
	"adventBot/internal/service"
	"context"
	"fmt"
	"github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
)

type CommandHandler struct {
	Repository chat.Repository
	manager    *service.SchedulerManager
}

func NewCommandHandler(r chat.Repository, m *service.SchedulerManager) *CommandHandler {
	return &CommandHandler{r, m}
}

func (h *CommandHandler) Handle(ctx context.Context, b *tgbotapi.BotAPI, update *tgbotapi.Update) {
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

	locKb := tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButtonLocation("Поделиться локацией"),
		),
	)
	locKb.ResizeKeyboard = true
	locKb.OneTimeKeyboard = true

	msg := tgbotapi.NewMessage(chatID, "Нажми «Поделиться локацией», чтобы определить часовой пояс.")
	msg.ReplyMarkup = locKb
	if _, err := b.Send(msg); err != nil {
		log.Println("[CommandHandler.Handle] SendMessage:", err)
	}

	h.manager.AddScheduler(chatID, b)
	scheduleMsg := tgbotapi.NewMessage(chatID, "Для чата создан Scheduler, он будет уведомлять вас раз в 24 часа")
	if _, err := b.Send(scheduleMsg); err != nil {
		log.Println("[CommandHandler.Handle] SendMessage:", err)
	}
}
