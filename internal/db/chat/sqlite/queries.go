package sqlite

import "fmt"

const (
	tableChat = "chat"

	colChatID   = "chat_id"
	colTimeZone = "time_zone"
)

var createTable = fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s (
  %s INTEGER PRIMARY KEY,
  %s TEXT NOT NULL
);`, tableChat, colChatID, colTimeZone)

var upsert = fmt.Sprintf(`
INSERT INTO %s (%s, %s)
VALUES (?, ?)
ON CONFLICT(%s) DO UPDATE SET
  %s = excluded.%s;
`, tableChat,
	colChatID, colTimeZone,
	colChatID,
	colTimeZone, colTimeZone,
)

var selectByChatId = fmt.Sprintf(`SELECT %s FROM %s WHERE %s = ?;`,
	colTimeZone, tableChat, colChatID)

var deleteByChatId = fmt.Sprintf(`DELETE FROM %s WHERE %s = ?;`,
	tableChat, colChatID)
