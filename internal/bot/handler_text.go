package bot

import (
	"adventBot/internal/ai_model"
	"adventBot/internal/db/chat"
	"adventBot/internal/db/message"
	"context"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"log"
)

type TextHandler struct {
	Model          ai_model.AiModel
	MultUseCase    ai_model.UseCaseMultiple
	ChatRepository chat.Repository
	MsgRepository  message.Repository
}

func NewTextHandler(model ai_model.AiModel, r chat.Repository, m message.Repository, mult ai_model.UseCaseMultiple) *TextHandler {
	return &TextHandler{Model: model, ChatRepository: r, MsgRepository: m, MultUseCase: mult}
}

func (h *TextHandler) Handle(ctx context.Context, b *bot.Bot, update *models.Update) {
	if update == nil || update.Message == nil || update.Message.Text == "" {
		return
	}

	//TODO
	h.askMultiple(ctx, b, update)
}

func (h *TextHandler) askMultiple(ctx context.Context, b *bot.Bot, update *models.Update) {
	chatID := update.Message.Chat.ID

	var msgs = [1]message.Message{
		{"user", update.Message.Text, "UTC", 0},
	}

	input := ai_model.InputForm{
		History: msgs[:],
	}

	log.Println("[TextHandler.askMultiple] input", input)
	h.MultUseCase.AskMultiple(ctx, chatID, input, func(reply string) {
		_, err := b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: chatID,
			Text:   reply,
		})
		if err != nil {
			log.Println("[TemperatureHandler.Handle] Error send message:", err)
		}
	})
}

func (h *TextHandler) askBaseQuestion(ctx context.Context, b *bot.Bot, update *models.Update) {
	chatID := update.Message.Chat.ID

	found, tz := h.getTimeZone(ctx, b, update)
	if !found {
		return
	}

	payload := h.getInput(ctx, update, tz)
	reply := h.Model.AskGpt(ctx, chatID, payload)
	_ = sendWithMenu(ctx, b, h.ChatRepository, chatID, reply)
}

//--- utils ---

func (h *TextHandler) getTimeZone(ctx context.Context, b *bot.Bot, update *models.Update) (found bool, tz string) {
	chatID := update.Message.Chat.ID

	tz, found, err := h.ChatRepository.GetById(ctx, chatID)
	if err != nil {
		log.Printf("[TextHandler.Handle.getTimeZone] GetById error chatID=%d err=%v", chatID, err)
		found = false
	}

	if !found {
		reply := "Не помню твою временную зону. Давай начнём сначала. Нажми /start, чтобы настроить часовой пояс или /restart если мы уже знакомы."
		err := sendWithStart(ctx, b, chatID, reply)
		if err != nil {
			log.Printf("[TextHandler.Handle.getTimeZone] Error sendWithStart chatID=%d err=%v", chatID, err)
			return false, ""
		}
		return false, ""
	}

	return found, tz
}

func (h *TextHandler) getInput(ctx context.Context, update *models.Update, tz string) ai_model.InputForm {
	var messages = make([]message.Message, 0, 3)
	chatID := update.Message.Chat.ID
	role := h.Model.GetUserRole()
	text := update.Message.Text
	ts := update.Message.Date

	current := message.Message{
		Role:      role.GetValue(),
		Message:   text,
		TimeZone:  tz,
		Timestamp: ts,
	}

	prev, found, err := h.MsgRepository.GetById(ctx, chatID, tz)
	if err != nil {
		log.Printf("[TextHandler.Handle.getInput] GetById error chatID=%d err=%v", chatID, err)
	}

	if found {
		for _, msg := range prev {
			messages = append(messages, msg)
		}
	}
	messages = append(messages, current)

	log.Println("[TextHandler.getInput] Get history + new message", messages)
	return ai_model.InputForm{History: messages}
}
