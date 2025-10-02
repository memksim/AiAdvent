package main

import (
	"adventBot/internal/ai_model/yandex"
	internalbot "adventBot/internal/bot"
	"adventBot/internal/config"
	"adventBot/internal/db/sqlite"
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
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	db, err := sql.Open("sqlite3", cfg.DbPath)
	if err != nil {
		log.Fatal(err)
	}
	repository := sqlite.NewRepositorySQlite(db)
	if repository.Init() != nil {
		log.Fatal("Cannot initialize repository: ", err, cfg.DbPath)
	}
	defer func(repository *sqlite.RepositorySQlite) {
		err := repository.Close()
		if err != nil {
			log.Println(err)
		}
	}(repository)

	model := yandex.NewAiModelYandex(cfg.ApiKey, cfg.RulePath, cfg.FolderId)

	b, err := bot.New(cfg.BotToken, bot.WithDefaultHandler(func(ctx context.Context, b *bot.Bot, u *models.Update) {}))
	if err != nil {
		log.Fatal(err)
	}

	client := http.Client{Timeout: time.Second * 60}
	timeZone := geonames.NewApiGeonames(cfg.GeonamesUser, &client)

	cmd = internalbot.NewCommandHandler(repository)
	res = internalbot.NewResetHandler(repository)
	txt = internalbot.NewTextHandler(model, repository)
	loc = internalbot.NewLocationHandler(timeZone, repository)

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
