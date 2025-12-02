package store

import (
	"database/sql"

	"github.com/alextreichler/crochetbyjuliette/internal/models"
)

func (s *Store) GetUserByUsername(username string) (*models.User, error) {
	query := `SELECT id, username, password FROM users WHERE username = ?`
	row := s.DB.QueryRow(query, username)

	var user models.User
	if err := row.Scan(&user.ID, &user.Username, &user.Password); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

// CreateUser is mainly for seeding the initial admin
func (s *Store) CreateUser(username, hashedPassword string) error {
	query := `INSERT INTO users (username, password) VALUES (?, ?)`
	_, err := s.DB.Exec(query, username, hashedPassword)
	return err
}
