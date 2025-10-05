package sqlite

type Message struct {
	ChatID    int64  `db:"chat_id"`
	Role      string `db:"role"`
	Message   string `db:"message"`
	Timestamp int    `db:"timestamp"`
}
