package service

import (
	summary "adventBot/internal/ai_model/yandex/summary/tasks"
	"adventBot/internal/db/task"
	"adventBot/internal/utils"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
	"strings"
	"time"
)

type DailyTaskScheduler struct {
	taskRepo   task.Repository
	chatID     int64
	bot        *tgbotapi.BotAPI
	summarizer *summary.SummarizerTask
	ticker     *time.Ticker
	stopCh     chan struct{}
}

func NewDailyTaskScheduler(taskRepo task.Repository, chatID int64, b *tgbotapi.BotAPI, s *summary.SummarizerTask) *DailyTaskScheduler {
	return &DailyTaskScheduler{
		taskRepo:   taskRepo,
		chatID:     chatID,
		bot:        b,
		summarizer: s,
		stopCh:     make(chan struct{}),
	}
}

func (s *DailyTaskScheduler) Start() {
	log.Printf("Starting daily task scheduler for chat ID: %d", s.chatID)

	go s.run()
}

func (s *DailyTaskScheduler) Stop() {
	if s.ticker != nil {
		s.ticker.Stop()
	}
	close(s.stopCh)
	log.Printf("Stopped daily task scheduler for chat ID: %d", s.chatID)
}

func (s *DailyTaskScheduler) run() {
	s.processDailyTasks()

	s.ticker = time.NewTicker(2 * time.Minute) //TODO заменить на 24 часа
	defer s.ticker.Stop()

	for {
		select {
		case <-s.ticker.C:
			s.processDailyTasks()
		case <-s.stopCh:
			return
		}
	}
}

func (s *DailyTaskScheduler) processDailyTasks() {
	today := utils.GetToday()
	log.Printf("Processing daily tasks for chat ID %d, date: %s", s.chatID, today)

	tasks, err := s.taskRepo.GetToday(s.chatID, today)
	if err != nil {
		log.Printf("[DailyTaskScheduler.processDailyTasks] Error retrieving tasks for chat ID %d: %v", s.chatID, err)
		return
	}

	if len(tasks) == 0 {
		log.Printf("No tasks found for chat ID %d on %s", s.chatID, today)
		return
	}

	log.Printf("Found %d tasks for chat ID %d on %s:", len(tasks), s.chatID, today)

	var b strings.Builder
	for _, t := range tasks {
		_, err := fmt.Fprintf(&b, "Задача: %s\nДата: %s\nЛокация: %s\n\n", t.Task, t.DateTime, t.Location)
		if err != nil {
			log.Printf("[DailyTaskScheduler.processDailyTasks] Error writing daily task output: %v", err)
		}
	}
	text := b.String()
	log.Printf(text)

	reply := s.summarizer.Summarize(text)
	log.Printf(reply)

	msg := tgbotapi.NewMessage(s.chatID, reply)
	if _, err := s.bot.Send(msg); err != nil {
		log.Printf("[DailyTaskScheduler.processDailyTasks] Error sending message: %v", err)
	}
}

func (s *DailyTaskScheduler) ProcessNow() {
	s.processDailyTasks()
}
