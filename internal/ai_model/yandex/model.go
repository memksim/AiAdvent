package yandex

import (
	"adventBot/internal/ai_model"
	"adventBot/internal/config"
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
const modelTemperature = 0.1
const modelMaxTokens = 4000

type messages []message

type AiModelYandex struct {
	ApiKey     string
	FolderID   string
	system     message
	systemCoT  message
	Repository dbmessage.Repository
	Finalizer  *FinalizerModel
}

func NewAiModelYandex(cfg *config.Config, folderId string, r dbmessage.Repository) *AiModelYandex {
	cotRulePath := cfg.RulePathCot

	return &AiModelYandex{
		ApiKey:   cfg.ApiKey,
		FolderID: folderId,
		system: message{
			Role: "system",
			Text: ai_model.MustReadFile(cfg.RulePath),
		},
		systemCoT: message{
			Role: "system",
			Text: ai_model.MustReadFile(cotRulePath),
		},
		Repository: r,
		Finalizer:  NewFinalizerModel(cfg, folderId, "yandexgpt-5-lite/latest"),
	}
}

func (a *AiModelYandex) GetUserRole() ai_model.Role {
	return &user
}

func (a *AiModelYandex) AskGpt(ctx context.Context, chatId int64, inputForm ai_model.InputForm, isCot bool) string {

	log.Println("[AiModelYandex.AskGpt] input form: ", inputForm)

	if !a.checkAuthorizationInfo() {
		log.Println("[AiModelYandex.AskGpt] ApiKey or FolderID is empty")
		return failureRequestReply
	}

	reqBody, err := a.prepareModelRequest(inputForm, isCot)
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
	log.Printf("response:\n%s", yr)
	modelText := yr.Result.Alternatives[0].Message.Text
	modelText = stripCodeFence(modelText)
	if strings.TrimSpace(modelText) == "" {
		log.Println("[AiModelYandex.AskGpt] empty text in alternative")
		return failureRequestReply
	}

	var parsed response
	if err := json.Unmarshal([]byte(modelText), &parsed); err != nil {
		log.Printf("[AiModelYandex.AskGpt] cannot parse model JSON: %v; text=%s", err, modelText)
		// Если не удалось распарсить JSON, возможно модель вернула обычный текст
		// В этом случае возвращаем текст как есть для режима ask
		return fmt.Sprintf("%s\n\n📱 Модель: %s", modelText, yr.Result.ModelVersion)
	}

	switch parsed.Mode {
	case modeAsk:
		last := inputForm.History[0]
		log.Println("[AiModelYandex.AskGpt] last history:", last)

		if parsed.Question == "" {
			log.Println("[AiModelYandex.AskGpt] ask without question")
			return failureRequestReply
		}
		if dberr := a.Repository.Upsert(ctx, chatId, last.Role, last.Message, last.Timestamp); dberr != nil {
			log.Println("[AiModelYandex.AskGpt] Repository.Upsert user error:", err)
		}

		currTime := int(time.Now().UnixMilli()) / 1000

		// Формируем ответ с рассуждениями, если они есть
		responseText := parsed.Question
		if parsed.Reasoning != "" {
			responseText = fmt.Sprintf("%s\n\n%s", parsed.Reasoning, parsed.Question)
		}

		// Добавляем информацию о версии модели
		responseText = fmt.Sprintf("%s\n\n📱 Модель: %s", responseText, yr.Result.ModelVersion)

		if dberr := a.Repository.Upsert(ctx, chatId, model.GetValue(), responseText, currTime); dberr != nil {
			log.Println("[AiModelYandex.AskGpt] Repository.Upsert assistant error:", err)
		}

		return responseText

	case modeFinal:
		if _, err := time.Parse(time.RFC3339, parsed.DateTime); err != nil {
			log.Println("[AiModelYandex.AskGpt] invalid date time")
			//TODO отправлять повторный запрос с нужной датой
		}

		_, err := a.Repository.DeleteById(ctx, chatId)
		if err != nil {
			log.Println("[AiModelYandex.AskGpt] failed to delete chat history:", err)
		}

		// Создаем JSON с final ответом для передачи в финализатор
		finalJson, err := json.Marshal(parsed)
		if err != nil {
			log.Printf("[AiModelYandex.AskGpt] failed to marshal final response: %v", err)
			return failureRequestReply
		}

		// Используем финализатор для форматирования ответа
		finalizedText := a.Finalizer.Finalize(string(finalJson))
		if finalizedText == "Не удалось обработать ответ модели" {
			// Если финализатор не сработал, возвращаем стандартный формат
			responseText := fmt.Sprintf(
				"Задача: %s\nДата/время: %s\nМесто: %s",
				parsed.Task, parsed.DateTime, parsed.Location,
			)
			if parsed.Reasoning != "" {
				responseText = fmt.Sprintf("%s\n\n%s", parsed.Reasoning, responseText)
			}
			responseText = fmt.Sprintf("%s\n\n📱 Модель: %s", responseText, yr.Result.ModelVersion)
			return responseText
		}

		return finalizedText

	default:
		log.Printf("[AiModelYandex.AskGpt] unknown mode: %s; raw=%s", parsed.Mode, modelText)
		return failureRequestReply
	}
}

func (a *AiModelYandex) AskWithTemperature(text string, temperature float64) (reply string, tmp float64) {
	tmp = temperature
	if tmp < 0 {
		tmp = modelTemperature
	}

	log.Printf("[AiModelYandex.AskWithTemperature] start request %v, %v", text, tmp)

	r := request{
		ModelURI: fmt.Sprintf("gpt://%s/yandexgpt-5-pro/latest", a.FolderID),
		Messages: []message{{Role: "user", Text: text}},
	}

	r.CompletionOptions.Stream = false
	r.CompletionOptions.Temperature = tmp
	r.CompletionOptions.MaxTokens = modelMaxTokens

	b, _ := json.Marshal(r)

	httpClient := &http.Client{Timeout: clientTimeout}
	req, _ := http.NewRequest(http.MethodPost, url, bytes.NewReader(b))
	a.prepareHttpRequest(req)

	resp, err := httpClient.Do(req)
	if err != nil {
		log.Println("[AiModelYandex.AskGpt] Error making request:", err)
		return failureRequestReply, tmp
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.Println("[AiModelYandex.AskGpt] Body.Close(): ", err)
		}
	}(resp.Body)

	if !isRequestSuccessful(resp.StatusCode) {
		log.Println("[AiModelYandex.AskGpt] Request failed: ", resp.StatusCode)
		return failureRequestReply, tmp
	}

	result := struct {
		Result struct {
			Alternatives []struct {
				Message struct {
					Role string `json:"role"`
					Text string `json:"text"`
				} `json:"message"`
			} `json:"alternatives"`
		} `json:"result"`
	}{}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		log.Println("[AiModelYandex.AskGpt] Error while decoding response: ", err)
		return failureRequestReply, tmp
	}

	if len(result.Result.Alternatives) == 0 {
		log.Println("[AiModelYandex.AskGpt] No alternatives found")
		return failureRequestReply, tmp
	}

	log.Printf("[AiModelYandex.AskGpt] result: %+v", result)

	return result.Result.Alternatives[0].Message.Text, tmp
}

// --- private ---

func (a *AiModelYandex) checkAuthorizationInfo() bool {
	return a.ApiKey != "" && a.FolderID != ""
}

func (a *AiModelYandex) prepareModelRequest(form ai_model.InputForm, isCot bool) ([]byte, error) {
	dst := make(messages, 0, len(form.History))

	if isCot {
		dst = append(dst, a.systemCoT)
	} else {
		dst = append(dst, a.system)
	}

	for _, m := range form.History {
		dst = append(dst, mapToInternal(m))
	}

	r := request{
		ModelURI: fmt.Sprintf("gpt://%s/yandexgpt/rc", a.FolderID),
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
