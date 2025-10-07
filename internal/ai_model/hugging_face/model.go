package hugging_face

import (
	"adventBot/internal/ai_model"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

type Model string

var (
	Zai_org_glm_4dot6      Model = "zai-org/GLM-4.6:novita"                        //357B params
	Qwen_3                 Model = "Qwen/Qwen3-1.7B:featherless-ai"                //2.03B params
	DeepSeek_R1            Model = "deepseek-ai/DeepSeek-R1:fireworks-ai"          //685B params
	Meta_llama_llama_3dot1 Model = "meta-llama/Llama-3.1-8B-Instruct:fireworks-ai" //8.03B params
)

func (m Model) GetModelName() string {
	switch m {
	case Zai_org_glm_4dot6:
		return "zai-org/GLM-4.6:novita"
	case Qwen_3:
		return "Qwen/Qwen3-1.7B:featherless-ai"
	case DeepSeek_R1:
		return "deepseek-ai/DeepSeek-R1:fireworks-ai"
	case Meta_llama_llama_3dot1:
		return "meta-llama/Llama-3.1-8B-Instruct:fireworks-ai"
	// Zai_org_glm_4dot6
	default:
		return "zai-org/GLM-4.6:novita"

	}
}

type hgRole string

var user hgRole = "user"

func (r hgRole) GetValue() string {
	return "user"
}

const url = "https://router.huggingface.co/v1/chat/completions"
const clientTimeout = 60 * time.Second
const failureRequestReply = "Не удалось выполнить запрос. Повторите позже"
const unableTempReply = "Данная модель не поддерживает настройку температуры."

type HuggingFace struct {
	Token string
	Model Model
}

func NewHuggingFaceModel(token string, model Model) *HuggingFace {
	return &HuggingFace{token, model}
}

func (h *HuggingFace) AskGpt(_ context.Context, _ int64, inputForm ai_model.InputForm) (reply string) {
	if h.Token == "" {
		log.Println("[HuggingFace.AskGpt] empty token")
		return failureRequestReply
	}

	log.Printf("[HuggingFace.AskGpt model: %s] input form: %v ", h.Model, inputForm)

	body, err := json.Marshal(request{
		Messages: getMessages(inputForm),
		Model:    h.Model.GetModelName(),
	})
	if err != nil {
		log.Println("[HuggingFace.AskGpt] marshal error:", err)
		return failureRequestReply
	}

	client := &http.Client{Timeout: clientTimeout}

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(body))
	if err != nil {
		log.Println("[HuggingFace.AskGpt] create request error:", err)
		return failureRequestReply
	}
	req.Header.Set("Authorization", "Bearer "+h.Token)
	req.Header.Set("Content-Type", "application/json")

	start := time.Now()

	resp, err := client.Do(req)
	if err != nil {
		log.Println("[HuggingFace.AskGpt] request error:", err)
		return failureRequestReply
	}
	defer func(Body io.ReadCloser) {
		if cerr := Body.Close(); cerr != nil {
			log.Println("[HuggingFace.AskGpt] Body.Close():", cerr)
		}
	}(resp.Body)

	log.Printf("[HuggingFace.AskGpt] HTTP status: %d %s", resp.StatusCode, resp.Status)

	rawResp, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Println("[HuggingFace.AskGpt] Error reading body:", err)
		return failureRequestReply
	}
	log.Printf("[HuggingFace.AskGpt] RAW response:\n%s", string(rawResp))

	if !isRequestSuccessful(resp.StatusCode) {
		log.Println("[HuggingFace.AskGpt] Request failed:", resp.StatusCode)
		return failureRequestReply
	}

	duration := time.Since(start).Seconds()

	var hgResp response
	if err := json.Unmarshal(rawResp, &hgResp); err != nil {
		log.Println("[HuggingFace.AskGpt] decode HuggingFace response:", err)
		return failureRequestReply
	}
	if len(hgResp.Choices) == 0 {
		log.Println("[HuggingFace.AskGpt] no alternatives in response")
		return failureRequestReply
	}

	return fmt.Sprintf("Модель: %s\nВремя выполнения:%v\nВсего токенов затрачено:%v\n%s", h.Model.GetModelName(), duration, hgResp.Usage.TotalTokens, hgResp.Choices[0].Message.Content)
}

func (h *HuggingFace) AskWithTemperature(text string, temperature float64) (reply string, tmp float64) {
	return unableTempReply, temperature
}

func (h *HuggingFace) GetUserRole() ai_model.Role {
	return user
}

func isRequestSuccessful(status int) bool {
	return status >= 200 && status < 300
}

func getMessages(form ai_model.InputForm) []message {
	return []message{
		{
			Role:    user.GetValue(),
			Content: form.History[0].Message,
		},
	}
}
