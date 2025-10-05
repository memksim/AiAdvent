package yandex

type mode string

const (
	modeFinal mode = "final"
	modeAsk   mode = "ask"
)

type property string

const (
	propTask     property = "task"
	propDateTime property = "dateTime"
	propLocation property = "location"
)

// Response — универсальный ответ модели (final или ask)
type response struct {
	Mode mode `json:"mode"` // "final" | "ask"

	// final
	Task     string `json:"task,omitempty"`
	DateTime string `json:"dateTime,omitempty"`
	Location string `json:"location,omitempty"`

	// ask
	Question string   `json:"question,omitempty"`
	Property property `json:"property,omitempty"` // "task" | "dateTime" | "location"
}

type yaResponse struct {
	Result struct {
		Alternatives []struct {
			Message struct {
				Role string `json:"role"`
				Text string `json:"text"`
			} `json:"message"`
			Status string `json:"status"`
		} `json:"alternatives"`
	} `json:"result"`
}
