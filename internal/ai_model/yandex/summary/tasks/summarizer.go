package tasks

import (
	"adventBot/internal/ai_model"
	"adventBot/internal/ai_model/yandex"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
)

const summarizerUrl = "https://llm.api.cloud.yandex.net/foundationModels/v1/completion"
const modelTemperature = 0.7

type SummarizerTask struct {
	IamToken string
	FolderId string
}

func NewSummarizerTask(iamToken string, folderId string) *SummarizerTask {
	return &SummarizerTask{IamToken: iamToken, FolderId: folderId}
}

type SummarizerRequest struct {
	ModelURI          string `json:"modelUri"`
	CompletionOptions struct {
		Stream      bool    `json:"stream"`
		Temperature float64 `json:"temperature"`
		MaxTokens   int     `json:"maxTokens"`
	} `json:"completionOptions"`
	Messages []yandex.MessageYandexGpt `json:"messages"`
}

type SummarizerResponse struct {
	Result struct {
		Alternatives []struct {
			Message struct {
				Role string `json:"role"`
				Text string `json:"text"`
			} `json:"message"`
			Status string `json:"status"`
		} `json:"alternatives"`
	} `json:"result"`
}

func (t *SummarizerTask) Summarize(text string) string {
	system := yandex.MessageYandexGpt{
		Role: "system",
		Text: ai_model.MustReadFile("internal/ai_model/yandex/summary/tasks/rule.txt"),
	}
	user := yandex.MessageYandexGpt{
		Role: "user",
		Text: text,
	}
	requestBody := SummarizerRequest{
		ModelURI: fmt.Sprintf("gpt://%s/yandexgpt-lite", t.FolderId),
		CompletionOptions: struct {
			Stream      bool    `json:"stream"`
			Temperature float64 `json:"temperature"`
			MaxTokens   int     `json:"maxTokens"`
		}{
			Stream:      false,
			Temperature: modelTemperature,
			MaxTokens:   2000,
		},
		Messages: []yandex.MessageYandexGpt{system, user},
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return fmt.Sprintf("[SummarizerTask.Summarize] Error marshaling request: %v", err)
	}

	req, err := http.NewRequest("POST", summarizerUrl, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Sprintf("[SummarizerTask.Summarize] Error creating request: %v", err)
	}

	t.prepareHttpRequest(req)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Sprintf("[SummarizerTask.Summarize] Error making request: %v", err)
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.Printf("[SummarizerTask.Summarize] Error closing response body: %v", err)
		}
	}(resp.Body)

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Sprintf("[SummarizerTask.Summarize] Error reading response: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Sprintf("API error: %d - %s", resp.StatusCode, string(body))
	}

	var response SummarizerResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return fmt.Sprintf("[SummarizerTask.Summarize] Error unmarshaling response: %v", err)
	}

	if len(response.Result.Alternatives) == 0 {
		return "No alternatives in response"
	}

	return response.Result.Alternatives[0].Message.Text
}

func (s *SummarizerTask) prepareHttpRequest(req *http.Request) {
	req.Header.Set("Authorization", "Api-Key "+s.IamToken)
	req.Header.Set("Content-Type", "application/json")
}
