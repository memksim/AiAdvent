package bot

import (
	"adventBot/internal/ai_model"
	"context"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"log"
)

type TextHandler struct {
	Model ai_model.AiModel
}

func NewTextHandler(model ai_model.AiModel) *TextHandler {
	return &TextHandler{Model: model}
}

func (h *TextHandler) Handle(ctx context.Context, b *bot.Bot, update *models.Update) {
	userText := update.Message.Text

	answer := h.Model.AskGpt(userText)

	_, err := b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   answer,
	})
	if err != nil {
		log.Println("[TextHandler.Start] SendMessage:", err)
	}
}
