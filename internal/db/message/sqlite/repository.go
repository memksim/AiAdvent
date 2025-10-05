package sqlite

import (
	msg "adventBot/internal/db/message"
	"context"
	"database/sql"
	"fmt"
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
		log.Println("[message/RepositorySQlite.Init] failed to create table:", err)
		return err
	}
	log.Println("[message/RepositorySQlite.Init] table created or already exists")
	return nil
}

func (r *RepositorySQlite) CloseConnection() error {
	log.Println("[message/RepositorySQlite.Close] closing db connection")
	return r.db.Close()
}

func (r *RepositorySQlite) Upsert(ctx context.Context, chatID int64, role string, text string, timestamp int) error {
	_, err := r.db.ExecContext(ctx, upsert, chatID, role, text, timestamp)
	if err != nil {
		log.Printf("[message/RepositorySQlite.Upsert] chatID=%d role=%s text=%s timestamp=%d err=%v", chatID, role, text, timestamp, err)
		return err
	}
	log.Printf("[message/RepositorySQlite.Upsert] success chatID=%d role=%s text=%s timestamp=%d", chatID, role, text, timestamp)
	return nil
}

func (r *RepositorySQlite) GetById(ctx context.Context, chatID int64, tz string) (messages []msg.Message, found bool, err error) {
	rows, err := r.db.QueryContext(ctx, selectByChatId, chatID)
	if err != nil {
		log.Printf("[message/RepositorySQlite.GetById] selectByChatId err=%v", err)
		return nil, false, fmt.Errorf("select messages by chat_id: %w", err)
	}
	defer func(rows *sql.Rows) {
		err := rows.Close()
		if err != nil {
			log.Println("[message/RepositorySQlite.GetById] failed to close rows:", err)
		}
	}(rows)

	for rows.Next() {
		var m Message
		if err := rows.Scan(&m.ChatID, &m.Role, &m.Message, &m.Timestamp); err != nil {
			log.Printf("[message/RepositorySQlite.GetById] failed to scan rows:%v", err)
			return nil, false, fmt.Errorf("scan message row: %w", err)
		}
		message := msg.Message{
			Role:      m.Role,
			Message:   m.Message,
			TimeZone:  tz,
			Timestamp: m.Timestamp,
		}
		if len(message.Role) > 0 && len(message.Message) > 0 {
			messages = append(messages, message)
		}
	}

	if err := rows.Err(); err != nil {
		log.Printf("[message/RepositorySQlite.GetById] failed to scan rows:%v", err)
		return nil, false, fmt.Errorf("iterate rows: %w", err)
	}

	found = len(messages) > 0
	
	log.Printf("[message/RepositorySQlite.GetById] found %d messages: %v", len(messages), messages)
	return messages, found, nil
}

// TODO убедиться что все сообщения стираются
func (r *RepositorySQlite) DeleteById(ctx context.Context, chatID int64) (bool, error) {
	res, err := r.db.ExecContext(ctx, deleteByChatId, chatID)
	if err != nil {
		log.Printf("[message/RepositorySQlite.DeleteById] chatID=%d error=%v", chatID, err)
		return false, err
	}

	rows, err := res.RowsAffected()
	if err != nil {
		log.Printf("[message/RepositorySQlite.DeleteById] chatID=%d error getting RowsAffected=%v", chatID, err)
		return false, err
	}

	if rows > 0 {
		log.Printf("[message/RepositorySQlite.DeleteById] success deleted chatID=%d", chatID)
		return true, nil
	}

	log.Printf("[message/RepositorySQlite.DeleteById] no record found chatID=%d", chatID)
	return false, nil
}
