package yandex

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

const url = "https://llm.api.cloud.yandex.net/foundationModels/v1/completion"
const clientTimeout = 60 * time.Second
const failureRequestReply = "Не удалось выполнить запрос. Повторите позже"
const modelTemperature = 0.6
const modelMaxTokens = 200

type AiModelYandex struct {
	ApiKey   string
	FolderID string
	msg      message
}

func NewAiModelYandex(apiKey, folderId string) *AiModelYandex {
	return &AiModelYandex{
		ApiKey:   apiKey,
		FolderID: folderId,
	}
}

func (a *AiModelYandex) AskGpt(text string) string {
	if !a.checkAuthorizationInfo() {
		log.Println("[AiModelYandex.AskGpt] ApiKey or FolderID is empty")
		return failureRequestReply
	}

	reqBody := a.prepareModelRequest(text)
	b, _ := json.Marshal(reqBody)

	httpClient := &http.Client{Timeout: clientTimeout}
	req, _ := http.NewRequest(http.MethodPost, url, bytes.NewReader(b))
	a.prepareHttpRequest(req)

	resp, err := httpClient.Do(req)
	if err != nil {
		log.Println("[AiModelYandex.AskGpt] Error while making request: ", err)
		return failureRequestReply
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.Println("[AiModelYandex.AskGpt] Body.Close(): ", err)
		}
	}(resp.Body)

	if !isRequestSuccessful(resp.StatusCode) {
		log.Println("[AiModelYandex.AskGpt] Request failed: ", resp.StatusCode)
		return failureRequestReply
	}

	var r response
	if err := json.NewDecoder(resp.Body).Decode(&r); err != nil {
		log.Println("[AiModelYandex.AskGpt] Error while decoding response: ", err)
		return failureRequestReply
	}

	if len(r.Result.Alternatives) == 0 {
		log.Println("[AiModelYandex.AskGpt] No alternatives found")
		return failureRequestReply
	}

	return r.Result.Alternatives[0].Message.Text
}

func (a *AiModelYandex) checkAuthorizationInfo() bool {
	return a.ApiKey != "" && a.FolderID != ""
}

func (a *AiModelYandex) prepareModelRequest(text string) request {
	r := request{
		ModelURI: fmt.Sprintf("gpt://%s/yandexgpt-lite", a.FolderID),
		Messages: []message{{Role: "user", Text: text}},
	}
	r.CompletionOptions.Stream = false
	r.CompletionOptions.Temperature = modelTemperature
	r.CompletionOptions.MaxTokens = modelMaxTokens
	return r
}

func (a *AiModelYandex) prepareHttpRequest(req *http.Request) {
	req.Header.Set("Authorization", "Api-Key "+a.ApiKey)
	req.Header.Set("Content-Type", "application/json")
}

func isRequestSuccessful(status int) bool {
	return status >= 200 && status < 300
}
