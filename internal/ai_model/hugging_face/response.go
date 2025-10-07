package hugging_face

type response struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	Model   string `json:"model"`

	Choices []struct {
		Index   int `json:"index"`
		Message struct {
			Role             string      `json:"role"`
			Content          string      `json:"content"`
			Refusal          interface{} `json:"refusal"`
			Annotations      interface{} `json:"annotations"`
			Audio            interface{} `json:"audio"`
			FunctionCall     interface{} `json:"function_call"`
			ToolCalls        []any       `json:"tool_calls"`
			ReasoningContent interface{} `json:"reasoning_content"`
		} `json:"message"`
		Logprobs     interface{} `json:"logprobs"`
		FinishReason string      `json:"finish_reason"`
		StopReason   interface{} `json:"stop_reason"`
	} `json:"choices"`

	ServiceTier       interface{} `json:"service_tier"`
	SystemFingerprint interface{} `json:"system_fingerprint"`

	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`

	Error *struct {
		Message string `json:"message"`
	} `json:"error,omitempty"`
}
