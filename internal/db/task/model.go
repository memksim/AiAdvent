package task

type Task struct {
	ChatID   int64  `json:"chat_id"`
	Task     string `json:"task"`
	Location string `json:"location"`
	DateTime string `json:"dateTime"`
}
