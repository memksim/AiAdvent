package sqlite

type Chat struct {
	ChatID   int64  `db:"chat_id"`
	TimeZone string `db:"time_zone"`
}
