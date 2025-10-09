package yandex

import (
	"adventBot/internal/ai_model"
	"adventBot/internal/config"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
)

type ModelVersion string

type FinalizerModel struct {
	modelVersion ModelVersion
	ApiKey       string
	FolderID     string
	ruleText     string
}

type finalizerRequest struct {
	ModelURI          string    `json:"modelUri"`
	Messages          []message `json:"messages"`
	CompletionOptions struct {
		Stream      bool    `json:"stream"`
		Temperature float64 `json:"temperature"`
		MaxTokens   int     `json:"maxTokens"`
	} `json:"completionOptions"`
}

type finalizerResponse struct {
	Mode      string `json:"mode"`
	Message   string `json:"message"`
	Reasoning string `json:"reasoning"`
}

type finalizerYaResponse struct {
	Result struct {
		Alternatives []struct {
			Message struct {
				Role string `json:"role"`
				Text string `json:"text"`
			} `json:"message"`
			Status string `json:"status"`
		} `json:"alternatives"`
		ModelVersion string `json:"modelVersion"`
	} `json:"result"`
}

func NewFinalizerModel(cfg *config.Config, folderId string, modelEndpoint string) *FinalizerModel {
	return &FinalizerModel{
		modelVersion: ModelVersion(modelEndpoint),
		ApiKey:       cfg.ApiKey,
		FolderID:     folderId,
		ruleText:     ai_model.MustReadFile(cfg.RulePathFinalizer),
	}
}

func (f *FinalizerModel) Finalize(rawJson string) string {
	log.Printf("[FinalizerModel.Finalize] processing raw JSON: %s", rawJson)

	if f.ApiKey == "" || f.FolderID == "" {
		log.Println("[FinalizerModel.Finalize] ApiKey or FolderID is empty")
		return "Не удалось обработать ответ модели"
	}

	// Проверяем, что это действительно final ответ
	var finalResp response
	if err := json.Unmarshal([]byte(rawJson), &finalResp); err != nil {
		log.Printf("[FinalizerModel.Finalize] cannot parse final JSON: %v", err)
		return "Не удалось обработать ответ модели"
	}

	if finalResp.Mode != modeFinal {
		log.Printf("[FinalizerModel.Finalize] expected final mode, got: %s", finalResp.Mode)
		return "Не удалось обработать ответ модели"
	}

	// Проверяем наличие всех обязательных полей
	if finalResp.Task == "" || finalResp.DateTime == "" {
		log.Println("[FinalizerModel.Finalize] missing required fields in final response")
		return "Не удалось обработать ответ модели"
	}

	// Формируем запрос к модели для финализации
	reqBody, err := f.prepareFinalizerRequest(rawJson)
	if err != nil {
		log.Printf("[FinalizerModel.Finalize] Failed to prepare finalizer request: %v", err)
		return "Не удалось обработать ответ модели"
	}

	log.Printf("[FinalizerModel.Finalize] REQUEST body:\n%s", string(reqBody))

	httpClient := &http.Client{Timeout: clientTimeout}
	req, _ := http.NewRequest(http.MethodPost, url, bytes.NewReader(reqBody))
	f.prepareHttpRequest(req)

	resp, err := httpClient.Do(req)
	if err != nil {
		log.Printf("[FinalizerModel.Finalize] Error while making request: %v", err)
		return "Не удалось обработать ответ модели"
	}
	defer func(Body io.ReadCloser) {
		if cerr := Body.Close(); cerr != nil {
			log.Printf("[FinalizerModel.Finalize] Body.Close(): %v", cerr)
		}
	}(resp.Body)

	log.Printf("[FinalizerModel.Finalize] HTTP status: %d %s", resp.StatusCode, resp.Status)

	rawResp, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("[FinalizerModel.Finalize] Error reading body: %v", err)
		return "Не удалось обработать ответ модели"
	}
	log.Printf("[FinalizerModel.Finalize] RAW response:\n%s", string(rawResp))

	if !isRequestSuccessful(resp.StatusCode) {
		log.Printf("[FinalizerModel.Finalize] Request failed: %d", resp.StatusCode)
		return "Не удалось обработать ответ модели"
	}

	var yr finalizerYaResponse
	if err := json.Unmarshal(rawResp, &yr); err != nil {
		log.Printf("[FinalizerModel.Finalize] decode yandex response: %v", err)
		return "Не удалось обработать ответ модели"
	}

	if len(yr.Result.Alternatives) == 0 {
		log.Println("[FinalizerModel.Finalize] no alternatives in response")
		return "Не удалось обработать ответ модели"
	}

	modelText := yr.Result.Alternatives[0].Message.Text
	modelText = stripCodeFence(modelText)

	if strings.TrimSpace(modelText) == "" {
		log.Println("[FinalizerModel.Finalize] empty text in alternative")
		return "Не удалось обработать ответ модели"
	}

	var parsed finalizerResponse
	if err := json.Unmarshal([]byte(modelText), &parsed); err != nil {
		log.Printf("[FinalizerModel.Finalize] cannot parse finalizer JSON: %v; text=%s", err, modelText)
		return "Не удалось обработать ответ модели"
	}

	if parsed.Mode != "finalized" || parsed.Message == "" {
		log.Printf("[FinalizerModel.Finalize] invalid finalizer response: %+v", parsed)
		return "Не удалось обработать ответ модели"
	}

	// Формируем ответ с рассуждениями, если они есть
	responseText := parsed.Message
	if parsed.Reasoning != "" {
		responseText = fmt.Sprintf("%s\n\n%s", parsed.Reasoning, responseText)
	}

	// Добавляем информацию о версии модели
	responseText = fmt.Sprintf("%s\n\n📱 Модель: %s", responseText, yr.Result.ModelVersion)

	return responseText
}

func (f *FinalizerModel) prepareFinalizerRequest(rawJson string) ([]byte, error) {
	// Создаем системное сообщение с правилами
	systemMsg := message{
		Role: "system",
		Text: f.ruleText,
	}

	// Создаем пользовательское сообщение с final ответом
	userMsg := message{
		Role: "user",
		Text: fmt.Sprintf("final_response: %s", rawJson),
	}

	r := finalizerRequest{
		ModelURI: fmt.Sprintf("gpt://%s/%v", f.FolderID, f.modelVersion),
		Messages: []message{systemMsg, userMsg},
	}

	r.CompletionOptions.Stream = false
	r.CompletionOptions.Temperature = 0.1
	r.CompletionOptions.MaxTokens = 1000

	req, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		log.Printf("[FinalizerModel.prepareFinalizerRequest] Error while encoding request: %v", err)
		return nil, err
	}

	return req, nil
}

func (f *FinalizerModel) prepareHttpRequest(req *http.Request) {
	req.Header.Set("Authorization", "Api-Key "+f.ApiKey)
	req.Header.Set("Content-Type", "application/json")
}
