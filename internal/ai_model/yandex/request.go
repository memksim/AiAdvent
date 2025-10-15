package yandex

type yandexRole string

var (
	user  = yandexRole("user")
	model = yandexRole("assistant")
)

func (r *yandexRole) GetValue() string {
	switch *r {
	case user:
		return "user"
	case model:
		return "assistant"
	}
	return ""
}

type request struct {
	ModelURI          string `json:"modelUri"`
	CompletionOptions struct {
		Stream      bool    `json:"stream"`
		Temperature float64 `json:"temperature"`
		MaxTokens   int     `json:"maxTokens"`
	} `json:"completionOptions"`
	Messages []MessageYandexGpt `json:"messages"`
}
