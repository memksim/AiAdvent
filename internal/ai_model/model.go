package ai_model

import (
	"adventBot/internal/db/message"
	"context"
)

type InputForm struct {
	History []message.Message `json:"messages_history"`
}

type AiModel interface {
	AskGpt(ctx context.Context, chatId int64, inputForm InputForm, isCot bool) (reply string)
	AskWithTemperature(text string, temperature float64) (reply string, tmp float64)
	GetUserRole() Role
}
