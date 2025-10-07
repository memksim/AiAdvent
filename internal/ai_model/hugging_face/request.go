package hugging_face

type request struct {
	Messages []message `json:"messages"`
	Model    string    `json:"model"`
}
