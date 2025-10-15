package bot

import (
	"adventBot/internal/db/task"
	"adventBot/internal/utils"
	"context"
	"fmt"
	"github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
)

type TodayHandler struct {
	taskRepo task.Repository
}

func NewTodayHandler(r task.Repository) *TodayHandler { return &TodayHandler{r} }

func (h *TodayHandler) Handle(_ context.Context, b *tgbotapi.BotAPI, update *tgbotapi.Update) {
	if update == nil || update.Message == nil {
		return
	}

	chatID := update.Message.Chat.ID

	tasks, err := h.taskRepo.GetToday(chatID, utils.GetToday())
	if err != nil {
		log.Println("[TodayHandler.Handle] error getting tasks: ", err)
	}

	if len(tasks) == 0 {
		msg := tgbotapi.NewMessage(chatID, "На сегодня не осталось задач")
		if _, err := b.Send(msg); err != nil {
			log.Printf("[TodayHandler.Handle] Error sending message: %v", err)
		}
	} else {
		msg := tgbotapi.NewMessage(chatID, "Грядущие задачи на сегодня: ")
		if _, err := b.Send(msg); err != nil {
			log.Printf("[TodayHandler.Handle] Error sending message: %v", err)
		}
	}

	for _, t := range tasks {
		msg := tgbotapi.NewMessage(chatID, fmt.Sprintf("Задача: %s\nДата: %s\nЛокация: %s", t.Task, t.DateTime, t.Location))
		if _, err := b.Send(msg); err != nil {
			log.Printf("[TodayHandler.Handle] Error sending message: %v", err)
		}
	}
}
