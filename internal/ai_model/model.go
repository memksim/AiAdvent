package ai_model

type AiModel interface {
	AskGpt(text string) string
}
