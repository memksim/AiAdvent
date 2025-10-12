package summary

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"time"
)

const tokenizerUrl = "https://llm.api.cloud.yandex.net/foundationModels/v1/tokenize"
const clientTimeout = 30 * time.Second

type Tokenizer struct {
	IamToken string
	Model    string //model uri
}

type tokenizerRequest struct {
	ModelURI string `json:"modelUri"`
	Text     string `json:"text"`
}

type tokenizerResponse struct {
	Tokens []struct {
		ID      string `json:"id"`
		Text    string `json:"text"`
		Special bool   `json:"special"`
	} `json:"tokens"`
	ModelVersion string `json:"modelVersion"`
}

func (t *Tokenizer) GetTokensCount(text string) int {
	if t.IamToken == "" || t.Model == "" {
		log.Println("[Tokenizer.GetTokensCount] iamToken or model is empty")
		return 0
	}

	// Создаем запрос
	reqBody := tokenizerRequest{
		ModelURI: t.Model,
		Text:     text,
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		log.Printf("[Tokenizer.GetTokensCount] Error marshaling request: %v", err)
		return 0
	}

	// Отправляем запрос
	httpClient := &http.Client{Timeout: clientTimeout}
	req, err := http.NewRequest(http.MethodPost, tokenizerUrl, bytes.NewReader(jsonBody))
	if err != nil {
		log.Printf("[Tokenizer.GetTokensCount] Error creating request: %v", err)
		return 0
	}

	req.Header.Set("Authorization", "Bearer "+t.IamToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := httpClient.Do(req)
	if err != nil {
		log.Printf("[Tokenizer.GetTokensCount] Error making request: %v", err)
		return 0
	}
	defer func(Body io.ReadCloser) {
		if cerr := Body.Close(); cerr != nil {
			log.Printf("[Tokenizer.GetTokensCount] Body.Close(): %v", cerr)
		}
	}(resp.Body)

	log.Printf("[Tokenizer.GetTokensCount] HTTP status: %d %s", resp.StatusCode, resp.Status)

	if resp.StatusCode != http.StatusOK {
		log.Printf("[Tokenizer.GetTokensCount] Request failed with status: %d", resp.StatusCode)
		return 0
	}

	rawResp, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("[Tokenizer.GetTokensCount] Error reading response: %v", err)
		return 0
	}

	var tokenResp tokenizerResponse
	if err := json.Unmarshal(rawResp, &tokenResp); err != nil {
		log.Printf("[Tokenizer.GetTokensCount] Error decoding response: %v", err)
		return 0
	}

	tokenCount := len(tokenResp.Tokens)
	log.Printf("[Tokenizer.GetTokensCount] Token count: %d", tokenCount)

	return tokenCount
}
