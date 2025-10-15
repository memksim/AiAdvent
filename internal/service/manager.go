package service

import (
	"adventBot/internal/ai_model/yandex/summary/tasks"
	"adventBot/internal/db/task"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"sync"
)

type SchedulerManager struct {
	schedulers map[int64]*DailyTaskScheduler
	summary    *tasks.SummarizerTask
	taskRepo   task.Repository
	mu         sync.RWMutex
}

func NewSchedulerManager(taskRepo task.Repository, s *tasks.SummarizerTask) *SchedulerManager {
	return &SchedulerManager{
		schedulers: make(map[int64]*DailyTaskScheduler),
		summary:    s,
		taskRepo:   taskRepo,
	}
}

func (m *SchedulerManager) AddScheduler(chatID int64, bot *tgbotapi.BotAPI) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.schedulers[chatID]; !exists {
		scheduler := NewDailyTaskScheduler(m.taskRepo, chatID, bot, m.summary)
		m.schedulers[chatID] = scheduler
		scheduler.Start()
	}
}

func (m *SchedulerManager) RemoveScheduler(chatID int64) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if scheduler, exists := m.schedulers[chatID]; exists {
		scheduler.Stop()
		delete(m.schedulers, chatID)
	}
}

func (m *SchedulerManager) GetScheduler(chatID int64) *DailyTaskScheduler {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.schedulers[chatID]
}

func (m *SchedulerManager) ProcessAllNow() {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, scheduler := range m.schedulers {
		scheduler.ProcessNow()
	}
}

func (m *SchedulerManager) Shutdown() {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, scheduler := range m.schedulers {
		scheduler.Stop()
	}
	m.schedulers = make(map[int64]*DailyTaskScheduler)
}
