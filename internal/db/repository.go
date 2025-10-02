package db

import "context"

type Repository interface {
	Init() error
	Close() error
	Upsert(ctx context.Context, chatID int64, tz string) error
	GetById(ctx context.Context, chatID int64) (tz string, found bool, err error)
	DeleteById(ctx context.Context, chatID int64) (bool, error)
}
