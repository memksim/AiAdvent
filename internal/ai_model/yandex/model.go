package yandex

import (
	"adventBot/internal/ai_model"
	dbmessage "adventBot/internal/db/message"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"
)

const url = "https://llm.api.cloud.yandex.net/foundationModels/v1/completion"
const clientTimeout = 60 * time.Second
const failureRequestReply = "Не удалось выполнить запрос. Повторите позже"
const modelTemperature = 0.2
const modelMaxTokens = 2000

type messages []message

type AiModelYandex struct {
	ApiKey     string
	FolderID   string
	system     message
	Repository dbmessage.Repository
}

func NewAiModelYandex(apiKey, rulePath string, folderId string, r dbmessage.Repository) *AiModelYandex {
	return &AiModelYandex{
		ApiKey:   apiKey,
		FolderID: folderId,
		system: message{
			Role: "system",
			Text: ai_model.MustReadFile(rulePath),
		},
		Repository: r,
	}
}

func (a *AiModelYandex) GetUserRole() ai_model.Role {
	return &user
}

func (a *AiModelYandex) AskGpt(ctx context.Context, chatId int64, inputForm ai_model.InputForm) string {

	log.Println("[AiModelYandex.AskGpt] input form: ", inputForm)

	if !a.checkAuthorizationInfo() {
		log.Println("[AiModelYandex.AskGpt] ApiKey or FolderID is empty")
		return failureRequestReply
	}

	reqBody, err := a.prepareModelRequest(inputForm)
	if err != nil {
		log.Println("[AiModelYandex.prepareModelRequest] Failed to prepare model request", err)
		return failureRequestReply
	}

	log.Printf("[AiModelYandex.AskGpt] REQUEST body:\n%s", string(reqBody))

	httpClient := &http.Client{Timeout: clientTimeout}
	req, _ := http.NewRequest(http.MethodPost, url, bytes.NewReader(reqBody))
	a.prepareHttpRequest(req)

	resp, err := httpClient.Do(req)
	if err != nil {
		log.Println("[AiModelYandex.AskGpt] Error while making request:", err)
		return failureRequestReply
	}
	defer func(Body io.ReadCloser) {
		if cerr := Body.Close(); cerr != nil {
			log.Println("[AiModelYandex.AskGpt] Body.Close():", cerr)
		}
	}(resp.Body)

	log.Printf("[AiModelYandex.AskGpt] HTTP status: %d %s", resp.StatusCode, resp.Status)

	rawResp, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Println("[AiModelYandex.AskGpt] Error reading body:", err)
		return failureRequestReply
	}
	log.Printf("[AiModelYandex.AskGpt] RAW response:\n%s", string(rawResp))

	if !isRequestSuccessful(resp.StatusCode) {
		log.Println("[AiModelYandex.AskGpt] Request failed:", resp.StatusCode)
		return failureRequestReply
	}

	var yr yaResponse
	if err := json.Unmarshal(rawResp, &yr); err != nil {
		log.Println("[AiModelYandex.AskGpt] decode yandex response:", err)
		return failureRequestReply
	}
	if len(yr.Result.Alternatives) == 0 {
		log.Println("[AiModelYandex.AskGpt] no alternatives in response")
		return failureRequestReply
	}
	modelText := yr.Result.Alternatives[0].Message.Text
	modelText = stripCodeFence(modelText)
	if strings.TrimSpace(modelText) == "" {
		log.Println("[AiModelYandex.AskGpt] empty text in alternative")
		return failureRequestReply
	}

	var parsed response
	if err := json.Unmarshal([]byte(modelText), &parsed); err != nil {
		log.Printf("[AiModelYandex.AskGpt] cannot parse model JSON: %v; text=%s", err, modelText)
		return failureRequestReply
	}

	switch parsed.Mode {
	case modeAsk:

		last := inputForm.History[0]
		err := a.Repository.Upsert(ctx, chatId, last.Role, last.Message, last.Timestamp)
		if err != nil {
			log.Println("[AiModelYandex.AskGpt] Repository.Upsert error:", err)
		}

		if parsed.Question == "" {
			log.Println("[AiModelYandex.AskGpt] ask without question")
			return failureRequestReply
		}
		return parsed.Question

	case modeFinal:
		if _, err := time.Parse(time.RFC3339, parsed.DateTime); err != nil {
			log.Println("[AiModelYandex.AskGpt] invalid date time")
			//TODO отправлять повторный запрос с нужной датой
		}

		_, err := a.Repository.DeleteById(ctx, chatId)
		if err != nil {
			log.Println("[AiModelYandex.AskGpt] failed to delete chat history:", err)
		}

		return fmt.Sprintf(
			"Задача: %s\nДата/время: %s\nМесто: %s",
			parsed.Task, parsed.DateTime, parsed.Location,
		)

	default:
		log.Printf("[AiModelYandex.AskGpt] unknown mode: %s; raw=%s", parsed.Mode, modelText)
		return failureRequestReply
	}
}

func (a *AiModelYandex) checkAuthorizationInfo() bool {
	return a.ApiKey != "" && a.FolderID != ""
}

func (a *AiModelYandex) prepareModelRequest(form ai_model.InputForm) ([]byte, error) {
	dst := make(messages, 0, len(form.History))
	dst = append(dst, a.system)
	for _, m := range form.History {
		dst = append(dst, mapToInternal(m))
	}

	r := request{
		ModelURI: fmt.Sprintf("gpt://%s/yandexgpt-lite", a.FolderID),
		Messages: dst.filterEmpty(),
	}

	r.CompletionOptions.Stream = false
	r.CompletionOptions.Temperature = modelTemperature
	r.CompletionOptions.MaxTokens = modelMaxTokens

	req, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		log.Println("[AiModelYandex.prepareModelRequest] Error while encoding request:", err)
		return nil, err
	}

	return req, nil
}

func (a *AiModelYandex) prepareHttpRequest(req *http.Request) {
	req.Header.Set("Authorization", "Api-Key "+a.ApiKey)
	req.Header.Set("Content-Type", "application/json")
}

func isRequestSuccessful(status int) bool {
	return status >= 200 && status < 300
}

func mapToInternal(src dbmessage.Message) message {
	text := src.Message
	if src.TimeZone != "" {
		text += fmt.Sprintf(" (timeZone: %s)", src.TimeZone)
	}
	if src.Timestamp != 0 {
		text += fmt.Sprintf(" [timestamp: %d]", src.Timestamp)
	}

	return message{
		Role: src.Role,
		Text: text,
	}
}

func (m messages) filterEmpty() messages {
	var out messages
	for _, msg := range m {
		if strings.TrimSpace(msg.Text) == "" {
			continue
		}
		out = append(out, msg)
	}
	return out
}

func stripCodeFence(s string) string {
	s = strings.TrimSpace(s)
	if strings.HasPrefix(s, "```") {
		// срежем первую строку ``` или ```json
		if i := strings.IndexByte(s, '\n'); i >= 0 {
			s = s[i+1:]
		} else {
			return ""
		}
	}
	s = strings.TrimSpace(s)
	// срежем закрывающие ```
	if strings.HasSuffix(s, "```") {
		s = strings.TrimSuffix(s, "```")
	}
	return strings.TrimSpace(s)
}
