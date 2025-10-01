package yandex

type request struct {
	ModelURI          string `json:"modelUri"`
	CompletionOptions struct {
		Stream      bool    `json:"stream"`
		Temperature float64 `json:"temperature"`
		MaxTokens   int     `json:"maxTokens"`
	} `json:"completionOptions"`
	Messages []message `json:"messages"`
}
