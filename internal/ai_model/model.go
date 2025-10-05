package ai_model

import (
	"adventBot/internal/db/message"
	"context"
)

type InputForm struct {
	History []message.Message `json:"messages_history"`
}

type AiModel interface {
	AskGpt(ctx context.Context, chatId int64, inputForm InputForm) (reply string)
	GetUserRole() Role
}
