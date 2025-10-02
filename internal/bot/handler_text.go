package bot

import (
	"adventBot/internal/ai_model"
	"adventBot/internal/ai_model/parser"
	"adventBot/internal/db"
	"context"
	"encoding/json"
	"fmt"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"log"
)

type TextHandler struct {
	Model      ai_model.AiModel
	Repository db.Repository
}

func NewTextHandler(model ai_model.AiModel, r db.Repository) *TextHandler {
	return &TextHandler{Model: model, Repository: r}
}

func (h *TextHandler) Handle(ctx context.Context, b *bot.Bot, update *models.Update) {
	if update == nil || update.Message == nil || update.Message.Text == "" {
		return
	}

	chatID := update.Message.Chat.ID
	userText := update.Message.Text

	tz, found, err := h.Repository.GetById(ctx, chatID)
	if err != nil {
		log.Printf("[TextHandler.Handle] GetById error chatID=%d err=%v", chatID, err)
		found = false
	}
	if !found {
		kb := &models.ReplyKeyboardMarkup{
			Keyboard:       [][]models.KeyboardButton{{{Text: "/start"}}},
			ResizeKeyboard: true,
		}
		if _, err := b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID:      chatID,
			Text:        "Давай начнём сначала. Нажми /start, чтобы настроить часовой пояс 🚀",
			ReplyMarkup: kb,
		}); err != nil {
			log.Println("[TextHandler.Handle] SendMessage (ask /start):", err)
		}
		return
	}

	tsSec := update.Message.Date
	type inputFmt struct {
		UserText  string `json:"userText"`
		TimeZone  string `json:"timeZone"`
		Timestamp int    `json:"timestamp"`
	}
	payload := map[string]any{
		"input_format": inputFmt{
			UserText:  userText,
			TimeZone:  tz,
			Timestamp: tsSec,
		},
	}

	body, err := json.Marshal(payload)
	if err != nil {
		log.Printf("[TextHandler.Handle] json marshal error: %v", err)
		body = []byte(userText)
	}

	raw := h.Model.AskGpt(string(body))
	parsed, err := parser.ParseStrict(raw)
	if err != nil {
		log.Printf("[TextHandler] parse failed: %v; raw=%s", err, raw)
		_ = sendWithMenu(ctx, b, h.Repository, chatID, "Не смог распарсить ответ. Попробуй ещё раз сформулировать задачу.")
		return
	}

	// тут у тебя уже строгое значение по схеме
	reply := fmt.Sprintf("Задача: %s\nДата/время: %s\nМесто: %s", parsed.Task, parsed.DateTime, parsed.Location)
	_ = sendWithMenu(ctx, b, h.Repository, chatID, reply)

}
