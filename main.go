package main

import (
	"adventBot/internal/ai_model"
	"adventBot/internal/ai_model/yandex"
	summary "adventBot/internal/ai_model/yandex/summary/tasks"
	internalbot "adventBot/internal/bot"
	"adventBot/internal/config"
	chat "adventBot/internal/db/chat"
	chat_sqlite "adventBot/internal/db/chat/sqlite"
	msg "adventBot/internal/db/message"
	msg_sqlite "adventBot/internal/db/message/sqlite"
	task "adventBot/internal/db/task"
	task_sqlite "adventBot/internal/db/task/sqlite"
	"adventBot/internal/service"
	"adventBot/internal/timezone/geonames"
	"context"
	"database/sql"
	"github.com/go-telegram-bot-api/telegram-bot-api/v5"
	_ "github.com/mattn/go-sqlite3"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"
)

var (
	cmd     internalbot.Handler
	res     internalbot.Handler
	txt     internalbot.Handler
	loc     internalbot.Handler
	today   internalbot.Handler
	tasks   internalbot.Handler
	trigger internalbot.Handler
	//TODO tmp internalbot.Handler

	model      ai_model.AiModel
	summarizer *summary.SummarizerTask

	chatRepository chat.Repository
	msgRepository  msg.Repository
	taskRepository task.Repository

	manager *service.SchedulerManager
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)

	// --- config ---
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	// --- timezone ---
	client := http.Client{Timeout: time.Second * 60}
	timeZone := geonames.NewApiGeonames(cfg.GeonamesUser, &client)

	// --- db ---
	db, err := sql.Open("sqlite3", cfg.DbPath)
	if err != nil {
		log.Fatal(err)
	}

	// --- repository ---
	chatRepository = chat_sqlite.NewRepositorySQlite(db)
	if chatRepository.Init() != nil {
		log.Fatal("Cannot initialize chat repository: ", err, cfg.DbPath)
	}

	msgRepository = msg_sqlite.NewRepositorySQlite(db)
	if msgRepository.Init() != nil {
		log.Fatal("Cannot initialize message repository: ", err, cfg.DbPath)
	}

	taskRepository = task_sqlite.NewRepositorySQlite(db)
	if taskRepository.Init() != nil {
		log.Fatal("Cannot initialize task repository: ", err, cfg.DbPath)
	}

	defer func() {
		cancel()

		if err := chatRepository.CloseConnection(); err != nil {
			log.Println(err)
		}
		if err := msgRepository.CloseConnection(); err != nil {
			log.Println(err)
		}
		if err := taskRepository.CloseConnection(); err != nil {
			log.Println(err)
		}
	}()

	// --- model ---
	model = yandex.NewAiModelYandex(&cfg, cfg.FolderId, msgRepository, taskRepository)
	summarizer = summary.NewSummarizerTask(cfg.ApiKey, cfg.FolderId)

	//--- schedule ---
	manager = service.NewSchedulerManager(taskRepository, summarizer)
	defer func() {
		manager.Shutdown()
	}()

	// --- handlers ---
	cmd = internalbot.NewCommandHandler(chatRepository, manager)
	res = internalbot.NewResetHandler(chatRepository, manager)
	txt = internalbot.NewTextHandler(model, chatRepository, msgRepository)
	loc = internalbot.NewLocationHandler(timeZone, chatRepository)
	//TODO tmp = internalbot.NewTemperatureHandler(model)
	today = internalbot.NewTodayHandler(taskRepository)
	tasks = internalbot.NewTasksHandler(taskRepository)
	trigger = internalbot.NewTriggerHandler(manager)

	// --- bot ---
	botAPI, err := tgbotapi.NewBotAPI(cfg.BotToken)
	if err != nil {
		log.Fatal(err)
	}

	botAPI.Debug = true

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := botAPI.GetUpdatesChan(u)

	for update := range updates {
		if update.Message == nil {
			continue
		}

		if update.Message.IsCommand() {
			switch update.Message.Command() {
			case "start":
				cmd.Handle(ctx, botAPI, &update)
			case "restart":
				res.Handle(ctx, botAPI, &update)
			case "today":
				today.Handle(ctx, botAPI, &update)
			case "tasks":
				tasks.Handle(ctx, botAPI, &update)
			case "trigger":
				trigger.Handle(ctx, botAPI, &update)
			default:
				handleText(ctx, botAPI, &update)
			}
		} else {
			handleText(ctx, botAPI, &update)
		}
	}
}

func handleText(ctx context.Context, b *tgbotapi.BotAPI, update *tgbotapi.Update) {
	if update.Message == nil {
		return
	}
	if update.Message.Location != nil {
		loc.Handle(ctx, b, update)
		return
	}
	if update.Message.Text != "" {
		txt.Handle(ctx, b, update)
		return
	}
}
