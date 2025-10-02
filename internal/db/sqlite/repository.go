package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"log"
)

type RepositorySQlite struct {
	db *sql.DB
}

func NewRepositorySQlite(db *sql.DB) *RepositorySQlite {
	return &RepositorySQlite{db: db}
}

func (r *RepositorySQlite) Init() error {
	_, err := r.db.Exec(createTable)
	if err != nil {
		log.Println("[RepositorySQlite.Init] failed to create table:", err)
		return err
	}
	log.Println("[RepositorySQlite.Init] table created or already exists")
	return nil
}

func (r *RepositorySQlite) Close() error {
	log.Println("[RepositorySQlite.Close] closing db connection")
	return r.db.Close()
}

func (r *RepositorySQlite) Upsert(ctx context.Context, chatID int64, tz string) error {
	_, err := r.db.ExecContext(ctx, upsert, chatID, tz)
	if err != nil {
		log.Printf("[RepositorySQlite.Upsert] chatID=%d tz=%s error=%v", chatID, tz, err)
		return err
	}
	log.Printf("[RepositorySQlite.Upsert] success chatID=%d tz=%s", chatID, tz)
	return nil
}

func (r *RepositorySQlite) GetById(ctx context.Context, chatID int64) (tz string, found bool, err error) {
	row := r.db.QueryRowContext(ctx, selectByChatId, chatID)
	switch err = row.Scan(&tz); {
	case err == nil:
		log.Printf("[RepositorySQlite.GetById] found chatID=%d tz=%s", chatID, tz)
		return tz, true, nil
	case errors.Is(err, sql.ErrNoRows):
		log.Printf("[RepositorySQlite.GetById] not found chatID=%d", chatID)
		return "", false, nil
	default:
		log.Printf("[RepositorySQlite.GetById] error chatID=%d err=%v", chatID, err)
		return "", false, err
	}
}

func (r *RepositorySQlite) DeleteById(ctx context.Context, chatID int64) (bool, error) {
	res, err := r.db.ExecContext(ctx, deleteByChatId, chatID)
	if err != nil {
		log.Printf("[RepositorySQlite.DeleteById] chatID=%d error=%v", chatID, err)
		return false, err
	}

	rows, err := res.RowsAffected()
	if err != nil {
		log.Printf("[RepositorySQlite.DeleteById] chatID=%d error getting RowsAffected=%v", chatID, err)
		return false, err
	}

	if rows > 0 {
		log.Printf("[RepositorySQlite.DeleteById] success deleted chatID=%d", chatID)
		return true, nil
	}

	log.Printf("[RepositorySQlite.DeleteById] no record found chatID=%d", chatID)
	return false, nil
}
