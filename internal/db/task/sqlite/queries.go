package sqlite

const createTableQuery = `
CREATE TABLE IF NOT EXISTS tasks (
	chat_id INTEGER NOT NULL,
	task TEXT NOT NULL,
	location TEXT NOT NULL,
	date_time TEXT NOT NULL,
	PRIMARY KEY (chat_id, date_time)
);
`

const getTodayQuery = `
SELECT chat_id, task, location, date_time
FROM tasks
WHERE chat_id = ? AND date(date_time) = date(?);
`

const getTodayTasksQuery = `
SELECT chat_id, task, location, date_time
FROM tasks
WHERE chat_id = ? AND date(date_time) = date(?)
ORDER BY date_time;
`

const getTasksQuery = `
SELECT chat_id, task, location, date_time
FROM tasks
WHERE chat_id = ?
ORDER BY date_time;
`

const upsertQuery = `
INSERT OR REPLACE INTO tasks (chat_id, task, location, date_time)
VALUES (?, ?, ?, ?);
`

const deleteQuery = `
DELETE FROM tasks
WHERE chat_id = ? AND date_time = ?;
`
