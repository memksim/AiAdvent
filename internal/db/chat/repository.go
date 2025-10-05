package chat

import "context"

type Repository interface {
	Init() error
	CloseConnection() error
	Upsert(ctx context.Context, chatID int64, tz string) error
	GetById(ctx context.Context, chatID int64) (tz string, found bool, err error)
	DeleteById(ctx context.Context, chatID int64) (bool, error)
}
