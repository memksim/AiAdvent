package bot

import (
	"adventBot/internal/db/chat"
	"adventBot/internal/timezone"
	"context"
	"fmt"
	"github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
	"time"
)

type LocationHandler struct {
	TzAPI      timezone.ApiTimezone
	Repository chat.Repository
}

func NewLocationHandler(tz timezone.ApiTimezone, r chat.Repository) *LocationHandler {
	return &LocationHandler{TzAPI: tz, Repository: r}
}

func (h *LocationHandler) Handle(ctx context.Context, b *tgbotapi.BotAPI, upd *tgbotapi.Update) {
	if upd == nil || upd.Message == nil {
		log.Println("[LocationHandler.Handle] nil update/message")
		return
	}

	chatID := upd.Message.Chat.ID
	tz := "UTC"

	if upd.Message.Location == nil {
		log.Println("[LocationHandler.Handle] no location in message, fallback to UTC")
	} else {
		loc := upd.Message.Location
		gotTZ, err := h.TzAPI.Lookup(ctx, loc.Latitude, loc.Longitude)
		if err != nil || gotTZ == "" {
			log.Printf("[LocationHandler.Handle] timezone lookup failed (lat=%.6f lon=%.6f), fallback to UTC: %v",
				loc.Latitude, loc.Longitude, err)
		} else {
			tz = gotTZ
		}
	}

	if err := h.Repository.Upsert(ctx, chatID, tz); err != nil {
		log.Printf("[LocationHandler.Handle] upsert tz failed chatID=%d tz=%s err=%v", chatID, tz, err)
	} else {
		log.Printf("[LocationHandler.Handle] upsert tz OK chatID=%d tz=%s", chatID, tz)
	}

	locTZ, err := time.LoadLocation(tz)
	if err != nil {
		log.Printf("[LocationHandler.Handle] LoadLocation(%s) error=%v, fallback to UTC", tz, err)
		locTZ = time.UTC
	}
	nowLocal := time.Now().In(locTZ)

	msg := fmt.Sprintf(
		"Похоже, ваш часовой пояс: %s\nЛокальная дата и время: %s",
		tz, nowLocal.Format("Mon, 02 Jan 2006 15:04:05"),
	)

	if err := sendWithMenu(ctx, b, h.Repository, chatID, msg); err != nil {
		log.Println("[LocationHandler.Handle] SendWithMenu:", err)
	}
}
