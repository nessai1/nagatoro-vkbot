package storage

import (
	"bytes"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
)

type Chat struct {
	ID       int
	ThreadID string
}

var ErrNotFound = fmt.Errorf("chat not found")

type Storage interface {
	RemoveChat(ctx context.Context, chatID int) error
	CreateChat(ctx context.Context, chatID int, threadID string) (Chat, error)
	GetChat(ctx context.Context, chatID int) (Chat, error)
}

func CreateStorage() (Storage, error) {
	return initSQLite()
}

type SQLiteStorage struct {
	db *sql.DB
}

func initSQLite() (*SQLiteStorage, error) {
	db, err := sql.Open("sqlite3", "storage.db")
	if err != nil {
		return nil, fmt.Errorf("cannot open sqlite3 database: %w", err)
	}

	f, err := os.Open("dbinit.sql")
	if err != nil {
		return nil, fmt.Errorf("cannot open dbinit.sql file: %w", err)
	}
	defer f.Close()

	b := bytes.Buffer{}
	_, err = b.ReadFrom(f)
	if err != nil {
		return nil, fmt.Errorf("cannot read dbinit.sql file: %w", err)
	}

	_, err = db.Exec(b.String())
	if err != nil {
		return nil, fmt.Errorf("cannot execute dbinit.sql file for init: %w", err)
	}

	return &SQLiteStorage{db: db}, nil
}

func (s *SQLiteStorage) RemoveChat(ctx context.Context, chatID int) error {
	_, err := s.db.ExecContext(ctx, "DELETE FROM chat_thread WHERE chat_id = ?", chatID)
	if err != nil {
		return fmt.Errorf("cannot delete chat: %w", err)
	}

	return nil
}

func (s *SQLiteStorage) CreateChat(ctx context.Context, chatID int, threadID string) (Chat, error) {
	_, err := s.db.ExecContext(ctx, "INSERT INTO chat_thread (chat_id, thread_id) VALUES (?, ?)", chatID, threadID)
	if err != nil {
		return Chat{}, fmt.Errorf("cannot insert chat: %w", err)
	}

	return Chat{ID: chatID, ThreadID: threadID}, nil
}

func (s *SQLiteStorage) GetChat(ctx context.Context, chatID int) (Chat, error) {
	row := s.db.QueryRowContext(ctx, "SELECT thread_id FROM chat_thread WHERE chat_id = ?", chatID)

	var threadID string
	err := row.Scan(&threadID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Chat{}, ErrNotFound
		}
		return Chat{}, fmt.Errorf("cannot select chat: %w", err)
	}

	return Chat{ID: chatID, ThreadID: threadID}, nil
}
