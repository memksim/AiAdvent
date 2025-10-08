package main

import (
	"adventBot/internal/ai_model/yandex"
	internalbot "adventBot/internal/bot"
	"adventBot/internal/config"
	chat "adventBot/internal/db/chat"
	chat_sqlite "adventBot/internal/db/chat/sqlite"
	msg "adventBot/internal/db/message"
	msg_sqlite "adventBot/internal/db/message/sqlite"
	"adventBot/internal/timezone/geonames"
	"context"
	"database/sql"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	_ "github.com/mattn/go-sqlite3"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"
)

var (
	cmd internalbot.Handler
	res internalbot.Handler
	txt internalbot.Handler
	loc internalbot.Handler
	tmp internalbot.Handler

	chatRepository chat.Repository
	msgRepository  msg.Repository
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

	// --- region repository ---
	chatRepository = chat_sqlite.NewRepositorySQlite(db)
	if chatRepository.Init() != nil {
		log.Fatal("Cannot initialize chat repository: ", err, cfg.DbPath)
	}
	msgRepository = msg_sqlite.NewRepositorySQlite(db)
	if msgRepository.Init() != nil {
		log.Fatal("Cannot initialize message repository: ", err, cfg.DbPath)
	}
	defer func() {
		cancel()

		if err := chatRepository.CloseConnection(); err != nil {
			log.Println(err)
		}
		if err := msgRepository.CloseConnection(); err != nil {
			log.Println(err)
		}
	}()

	// --- model ---
	model := yandex.NewAiModelYandex(&cfg, cfg.FolderId, msgRepository)

	// --- handlers ---
	cmd = internalbot.NewCommandHandler(chatRepository)
	res = internalbot.NewResetHandler(chatRepository)
	txt = internalbot.NewTextHandler(model, chatRepository, msgRepository)
	loc = internalbot.NewLocationHandler(timeZone, chatRepository)
	tmp = internalbot.NewTemperatureHandler(model)

	// --- bot ---
	b, err := bot.New(cfg.BotToken, bot.WithDefaultHandler(func(ctx context.Context, b *bot.Bot, u *models.Update) {}))
	if err != nil {
		log.Fatal(err)
	}
	b.RegisterHandler(bot.HandlerTypeMessageText, "/start", bot.MatchTypeExact, cmd.Handle)
	b.RegisterHandler(bot.HandlerTypeMessageText, "/restart", bot.MatchTypeExact, res.Handle)
	b.RegisterHandler(bot.HandlerTypeMessageText, "", bot.MatchTypePrefix, handleText)
	b.Start(ctx)
}

func handleText(ctx context.Context, b *bot.Bot, update *models.Update) {
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
