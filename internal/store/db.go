package store

import (
	"database/sql"
	"log/slog"

	_ "modernc.org/sqlite" // Pure Go SQLite driver
)

type Store struct {
	DB *sql.DB
}

func NewStore(dataSourceName string) (*Store, error) {
	db, err := sql.Open("sqlite", dataSourceName)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	return &Store{DB: db}, nil
}

func (s *Store) InitSchema() error {
	query := `
	CREATE TABLE IF NOT EXISTS items (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		title TEXT NOT NULL,
		description TEXT,
		price REAL,
		delivery_time TEXT,
		image_url TEXT,
		status TEXT DEFAULT 'available',
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		username TEXT NOT NULL UNIQUE,
		password TEXT NOT NULL
	);
	`
	_, err := s.DB.Exec(query)
	if err != nil {
		slog.Error("Error creating schema", "error", err)
		return err
	}
	return nil
}