package message

import "context"

type Repository interface {
	Init() error
	CloseConnection() error
	Upsert(ctx context.Context, chatID int64, role string, text string, timestamp int) error
	GetById(ctx context.Context, chatID int64, tz string) (messages []Message, found bool, err error)
	DeleteById(ctx context.Context, chatID int64) (bool, error)
}
