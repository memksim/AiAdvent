package message

type Message struct {
	Role      string `json:"role"`
	Message   string `json:"message"`
	TimeZone  string `json:"timeZone"`
	Timestamp int    `json:"timestamp"`
}
