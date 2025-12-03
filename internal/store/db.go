package store

import (
	"database/sql"

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