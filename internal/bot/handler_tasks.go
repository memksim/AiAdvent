package bot

import (
	"adventBot/internal/db/task"
	"context"
	"fmt"
	"github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
)

type TasksHandler struct {
	taskRepo task.Repository
}

func NewTasksHandler(r task.Repository) *TasksHandler { return &TasksHandler{r} }

func (h *TasksHandler) Handle(ctx context.Context, b *tgbotapi.BotAPI, update *tgbotapi.Update) {
	if update == nil || update.Message == nil {
		return
	}

	chatID := update.Message.Chat.ID

	tasks, err := h.taskRepo.GetAll(chatID)
	if err != nil {
		log.Println("[TasksHandler.Handle] error getting tasks: ", err)
	}

	if len(tasks) == 0 {
		msg := tgbotapi.NewMessage(chatID, "У вас не осталось задач")
		if _, err := b.Send(msg); err != nil {
			log.Printf("[TodayHandler.Handle] Error sending message: %v", err)
		}
	} else {
		msg := tgbotapi.NewMessage(chatID, "Ваши задачи: ")
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
