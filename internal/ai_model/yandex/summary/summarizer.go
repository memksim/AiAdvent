package summary

import (
	"adventBot/internal/ai_model"
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"sync"
)

const summarizerUrl = "https://llm.api.cloud.yandex.net/foundationModels/v1/completion"

const modelTemperature = 0.3

type Summarizer struct {
	MaxPromptTokens  int // Максимальное количество токенов промпта на вход
	MaxHistoryTokens int // Максимальное количество токенов истории переписки на вход
	MaxOutputTokens  int // Максимальное количество токенов на выход
	IamToken         string
	Model            string
	Tokenizer        Tokenizer
	PromptRule       string // Правило для суммаризации system промпта
	HistoryRule      string // Правило для суммаризации истории
}

type summarizerRequest struct {
	ModelURI string `json:"modelUri"`
	Messages []struct {
		Role string `json:"role"`
		Text string `json:"text"`
	} `json:"messages"`
	CompletionOptions struct {
		Stream      bool    `json:"stream"`
		Temperature float64 `json:"temperature"`
		MaxTokens   int     `json:"maxTokens"`
	} `json:"completionOptions"`
}

type summarizerResponse struct {
	Result struct {
		Alternatives []struct {
			Message struct {
				Role string `json:"role"`
				Text string `json:"text"`
			} `json:"message"`
			Status string `json:"status"`
		} `json:"alternatives"`
		Usage struct {
			InputTextTokens  string `json:"inputTextTokens"`
			CompletionTokens string `json:"completionTokens"`
			TotalTokens      string `json:"totalTokens"`
		} `json:"usage"`
		ModelVersion string `json:"modelVersion"`
	} `json:"result"`
}

func NewSummarizer(
	maxPrompt int,
	maxHistory int,
	maxOutput int,
	iamToken string,
	modelUri string,
) *Summarizer {
	return &Summarizer{
		MaxPromptTokens:  maxPrompt,
		MaxHistoryTokens: maxHistory,
		MaxOutputTokens:  maxOutput,
		IamToken:         iamToken,
		Model:            modelUri,
		PromptRule:       ai_model.MustReadFile("./internal/ai_model/yandex/summary/system_summarizer_rule.txt"),
		HistoryRule:      ai_model.MustReadFile("./internal/ai_model/yandex/summary/history_summarizer_rule.txt"),
		Tokenizer: Tokenizer{
			IamToken: iamToken,
			Model:    modelUri,
		},
	}
}

func (s *Summarizer) Summarize(sys string, h []string) (system string, history []string) {
	system = sys
	h = nil

	var wg sync.WaitGroup

	if systemTokens := s.Tokenizer.GetTokensCount(sys); systemTokens > s.MaxPromptTokens {
		log.Printf("[Summarizer.Summarize] sys tokens: %d, max: %d", systemTokens, s.MaxPromptTokens)
		wg.Add(1)
		go func() {
			defer wg.Done()
			system = s.summarizeRule(sys)
			log.Println("[Summarizer.Summarize] summarized system: ", system)
		}()
	}

	js := struct {
		History []string `json:"messages_history"`
	}{
		History: h,
	}

	input, err := json.Marshal(js)
	if err == nil {
		if historyTokens := s.Tokenizer.GetTokensCount(string(input)); historyTokens > s.MaxHistoryTokens {
			log.Printf("[Summarizer.Summarize] history tokens: %d, max: %d", historyTokens, s.MaxHistoryTokens)
			wg.Add(1)
			go func() {
				defer wg.Done()
				history = s.summarizeUser(h, string(input))
				log.Println("[Summarizer.Summarize] summarized history: ", history)
			}()
		}
	}

	wg.Wait()

	return
}

func (s *Summarizer) summarizeRule(prompt string) string {
	if s.IamToken == "" || s.Model == "" {
		log.Println("[Summarizer.Summarize] IamToken or Model is empty")
		return prompt
	}

	// Создаем запрос для суммаризации
	reqBody := summarizerRequest{
		ModelURI: s.Model,
		Messages: []struct {
			Role string `json:"role"`
			Text string `json:"text"`
		}{
			{Role: "system", Text: s.PromptRule},
			{Role: "user", Text: prompt},
		},
	}

	reqBody.CompletionOptions.Stream = false
	reqBody.CompletionOptions.Temperature = modelTemperature
	reqBody.CompletionOptions.MaxTokens = s.MaxOutputTokens

	jsonBody, err := json.MarshalIndent(reqBody, "", "  ")
	if err != nil {
		log.Printf("[Summarizer.Summarize] Error marshaling request: %v", err)
		return prompt
	}

	log.Printf("[Summarizer.Summarize] REQUEST body:\n%s", string(jsonBody))

	// Отправляем запрос
	httpClient := &http.Client{Timeout: clientTimeout}
	req, err := http.NewRequest(http.MethodPost, summarizerUrl, bytes.NewReader(jsonBody))
	if err != nil {
		log.Printf("[Summarizer.Summarize] Error creating request: %v", err)
		return prompt
	}

	s.prepareHttpRequest(req)

	resp, err := httpClient.Do(req)
	if err != nil {
		log.Printf("[Summarizer.Summarize] Error making request: %v", err)
		return prompt
	}
	defer func(Body io.ReadCloser) {
		if cerr := Body.Close(); cerr != nil {
			log.Printf("[Summarizer.Summarize] Body.Close(): %v", cerr)
		}
	}(resp.Body)

	log.Printf("[Summarizer.Summarize] HTTP status: %d %s", resp.StatusCode, resp.Status)

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		log.Printf("[Summarizer.Summarize] Request failed with status: %d", resp.StatusCode)
		return prompt
	}

	rawResp, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("[Summarizer.Summarize] Error reading response: %v", err)
		return prompt
	}

	log.Printf("[Summarizer.Summarize] RAW response:\n%s", string(rawResp))

	var sumResp summarizerResponse
	if err := json.Unmarshal(rawResp, &sumResp); err != nil {
		log.Printf("[Summarizer.Summarize] Error decoding response: %v", err)
		return prompt
	}

	if len(sumResp.Result.Alternatives) == 0 {
		log.Println("[Summarizer.Summarize] No alternatives in response")
		return prompt
	}

	result := sumResp.Result.Alternatives[0].Message.Text
	if result == "" {
		log.Println("[Summarizer.Summarize] Empty result")
		return prompt
	}

	log.Printf("[Summarizer.Summarize] Tokens used: %s input, %s output",
		sumResp.Result.Usage.InputTextTokens, sumResp.Result.Usage.CompletionTokens)

	return result
}

func (s *Summarizer) summarizeUser(raw []string, text string) (result []string) {
	if s.IamToken == "" || s.Model == "" {
		log.Println("[Summarizer.Summarize] IamToken or Model is empty")
		return nil
	}

	// Создаем запрос для суммаризации
	reqBody := summarizerRequest{
		ModelURI: s.Model,
		Messages: []struct {
			Role string `json:"role"`
			Text string `json:"text"`
		}{
			{Role: "system", Text: s.HistoryRule},
			{Role: "user", Text: text},
		},
	}

	reqBody.CompletionOptions.Stream = false
	reqBody.CompletionOptions.Temperature = modelTemperature
	reqBody.CompletionOptions.MaxTokens = s.MaxOutputTokens

	jsonBody, err := json.MarshalIndent(reqBody, "", "  ")
	if err != nil {
		log.Printf("[Summarizer.Summarize] Error marshaling request: %v", err)
		return nil
	}

	log.Printf("[Summarizer.Summarize] REQUEST body:\n%s", string(jsonBody))

	// Отправляем запрос
	httpClient := &http.Client{Timeout: clientTimeout}
	req, err := http.NewRequest(http.MethodPost, summarizerUrl, bytes.NewReader(jsonBody))
	if err != nil {
		log.Printf("[Summarizer.Summarize] Error creating request: %v", err)
		return nil
	}

	s.prepareHttpRequest(req)

	resp, err := httpClient.Do(req)
	if err != nil {
		log.Printf("[Summarizer.Summarize] Error making request: %v", err)
		return nil
	}
	defer func(Body io.ReadCloser) {
		if cerr := Body.Close(); cerr != nil {
			log.Printf("[Summarizer.Summarize] Body.Close(): %v", cerr)
		}
	}(resp.Body)

	log.Printf("[Summarizer.Summarize] HTTP status: %d %s", resp.StatusCode, resp.Status)

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		log.Printf("[Summarizer.Summarize] Request failed with status: %d", resp.StatusCode)
		return nil
	}

	rawResp, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("[Summarizer.Summarize] Error reading response: %v", err)
		return nil
	}

	log.Printf("[Summarizer.Summarize] RAW response:\n%s", string(rawResp))

	var sumResp summarizerResponse
	if err := json.Unmarshal(rawResp, &sumResp); err != nil {
		log.Printf("[Summarizer.Summarize] Error decoding response: %v", err)
		return nil
	}

	if len(sumResp.Result.Alternatives) == 0 {
		log.Println("[Summarizer.Summarize] No alternatives in response")
		return nil
	}

	result = append(result, sumResp.Result.Alternatives[0].Message.Text)
	if result[0] == "" {
		log.Println("[Summarizer.Summarize] Empty result")
		return nil
	}

	log.Printf("[Summarizer.Summarize] Tokens used: %s input, %s output",
		sumResp.Result.Usage.InputTextTokens, sumResp.Result.Usage.CompletionTokens)

	return result
}

func (s *Summarizer) prepareHttpRequest(req *http.Request) {
	req.Header.Set("Authorization", "Api-Key "+s.IamToken)
	req.Header.Set("Content-Type", "application/json")
}
