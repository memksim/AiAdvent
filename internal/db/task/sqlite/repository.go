package sqlite

import (
	"adventBot/internal/db/task"
	"context"
	"database/sql"
	"log"
)

type RepositorySQlite struct {
	db *sql.DB
}

func NewRepositorySQlite(db *sql.DB) *RepositorySQlite {
	return &RepositorySQlite{db: db}
}

func (r *RepositorySQlite) Init() error {
	_, err := r.db.Exec(createTableQuery)
	return err
}

func (r *RepositorySQlite) GetToday(chatID int64, dateTime string) ([]task.Task, error) {
	rows, err := r.db.Query(getTodayTasksQuery, chatID, dateTime)
	if err != nil {
		return nil, err
	}
	defer func(rows *sql.Rows) {
		err := rows.Close()
		if err != nil {
			log.Fatal("[GetToday] Error closing rows")
		}
	}(rows)

	var tasks []task.Task
	for rows.Next() {
		var t Task
		if err := rows.Scan(
			&chatID,
			&t.Task,
			&t.Location,
			&t.DateTime,
		); err != nil {
			return nil, err
		}

		tasks = append(tasks, task.Task{
			ChatID:   chatID,
			Task:     t.Task,
			Location: t.Location,
			DateTime: t.DateTime,
		})
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return tasks, nil
}

func (r *RepositorySQlite) GetAll(chatID int64) ([]task.Task, error) {
	rows, err := r.db.Query(getTasksQuery, chatID)
	if err != nil {
		return nil, err
	}
	defer func(rows *sql.Rows) {
		err := rows.Close()
		if err != nil {
			log.Fatal("[GetToday] Error closing rows")
		}
	}(rows)

	var tasks []task.Task
	for rows.Next() {
		var t Task
		if err := rows.Scan(
			&chatID,
			&t.Task,
			&t.Location,
			&t.DateTime,
		); err != nil {
			return nil, err
		}

		tasks = append(tasks, task.Task{
			ChatID:   chatID,
			Task:     t.Task,
			Location: t.Location,
			DateTime: t.DateTime,
		})
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return tasks, nil
}

func (r *RepositorySQlite) Upsert(ctx context.Context, task task.Task) error {
	_, err := r.db.ExecContext(ctx, upsertQuery, task.ChatID, task.Task, task.Location, task.DateTime)
	log.Println("upserted task:", task)
	return err
}

func (r *RepositorySQlite) Delete(ctx context.Context, task task.Task) error {
	_, err := r.db.ExecContext(ctx, deleteQuery, task.ChatID, task.DateTime)
	return err
}

func (r *RepositorySQlite) CloseConnection() error {
	return r.db.Close()
}
