package sqlite

import "fmt"

const (
	tableName    = "messages"
	colChatId    = "chat_id"
	colRole      = "role"
	colMessage   = "message"
	colTimestamp = "timestamp"
)

var createTable = fmt.Sprintf(`
CREATE TABLE IF NOT EXISTS %s (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  %s INTEGER NOT NULL,
  %s TEXT NOT NULL,
  %s TEXT NOT NULL,
  %s INTEGER NOT NULL,
  UNIQUE (%s, %s, %s)
);`,
	tableName,
	colChatId,
	colRole,
	colMessage,
	colTimestamp,
	colChatId, colTimestamp, colRole,
)

var upsert = fmt.Sprintf(`
INSERT INTO %s (%s, %s, %s, %s)
VALUES (?, ?, ?, ?)
ON CONFLICT(%s, %s, %s) DO UPDATE SET
  %s = excluded.%s;`,
	tableName,
	colChatId, colRole, colMessage, colTimestamp,
	colChatId, colTimestamp, colRole,
	colMessage, colMessage,
)

var selectByChatId = fmt.Sprintf(`
SELECT %s, %s, %s, %s
FROM %s
WHERE %s = ?
ORDER BY %s ASC, rowid ASC;`,
	colChatId, colRole, colMessage, colTimestamp,
	tableName,
	colChatId,
	colTimestamp,
)

var deleteByChatId = fmt.Sprintf(`
DELETE FROM %s
WHERE %s = ?;`,
	tableName,
	colChatId,
)
