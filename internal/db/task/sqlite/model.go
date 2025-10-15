package sqlite

type Task struct {
	ChatID   int64  `db:"chat_id"`
	Task     string `db:"task"`
	Location string `db:"location"`
	DateTime string `db:"date_time"`
}
