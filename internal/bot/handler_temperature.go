package bot

import (
	"adventBot/internal/ai_model"
	"context"
	"fmt"
	"github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
	"regexp"
	"strconv"
	"strings"
)

type TemperatureHandler struct {
	Model ai_model.AiModel
}

func NewTemperatureHandler(model ai_model.AiModel) *TemperatureHandler {
	return &TemperatureHandler{model}
}

func (h *TemperatureHandler) Handle(ctx context.Context, b *tgbotapi.BotAPI, update *tgbotapi.Update) {
	txt, temp, success := parseMessage(update.Message.Text)

	if !success {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Не удалось выполнить запрос. Повторите позже.")
		_, err := b.Send(msg)
		if err != nil {
			log.Println("[TemperatureHandler.Handle] Error send message]")
		}
		return
	}

	var reply string
	reply, temp = h.Model.AskWithTemperature(txt, temp)

	msg := tgbotapi.NewMessage(update.Message.Chat.ID, fmt.Sprintf("%v\n%s", temp, reply))
	_, err := b.Send(msg)
	if err != nil {
		log.Println("[TemperatureHandler.Handle] Error send message]")
	}
}

func parseMessage(text string) (cleanText string, temp float64, ok bool) {
	re := regexp.MustCompile(`\btemp:([0-9]*\.?[0-9]+)\b`)
	matches := re.FindStringSubmatch(text)

	// Значение температуры
	if len(matches) >= 2 {
		val, err := strconv.ParseFloat(matches[1], 64)
		if err == nil {
			temp = val
			ok = true
		}
	}

	// Убираем temp:... из текста
	cleanText = re.ReplaceAllString(text, "")
	cleanText = strings.TrimSpace(cleanText)

	return
}
