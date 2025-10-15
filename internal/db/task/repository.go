package task

import "context"

type Repository interface {
	Init() error
	GetToday(chatID int64, dateTime string) ([]Task, error)
	GetAll(chatID int64) ([]Task, error)
	Upsert(ctx context.Context, task Task) error
	Delete(ctx context.Context, task Task) error
	CloseConnection() error
}
