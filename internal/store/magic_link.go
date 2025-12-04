package store

import (
	"github.com/alextreichler/crochetbyjuliette/internal/models"
)

func (s *Store) GetOrderByToken(token string) (*models.Order, error) {
	// Join with items table to get item details
	query := `
		SELECT o.id, COALESCE(o.order_ref, CAST(o.id AS TEXT)) as order_ref, o.item_id, i.title, i.image_url, COALESCE(o.quantity, 1) as quantity, o.customer_name, o.customer_email, o.customer_address, o.status, o.notes, COALESCE(o.admin_comments, '') as admin_comments, o.magic_token, o.magic_token_expiry, o.created_at 
		FROM orders o
		JOIN items i ON o.item_id = i.id
		WHERE o.magic_token = ?
	`
	row := s.DB.QueryRow(query, token)

	var o models.Order
	if err := row.Scan(&o.ID, &o.OrderRef, &o.ItemID, &o.ItemTitle, &o.ItemImageURL, &o.Quantity, &o.CustomerName, &o.CustomerEmail, &o.CustomerAddress, &o.Status, &o.Notes, &o.AdminComments, &o.MagicToken, &o.MagicTokenExpiry, &o.CreatedAt); err != nil {
		return nil, err
	}
	return &o, nil
}

// Updated to be case-insensitive
func (s *Store) GetOrdersByEmail(email string) ([]models.Order, error) {
	// Select basic info needed for the list
	query := `
		SELECT o.id, COALESCE(o.order_ref, CAST(o.id AS TEXT)) as order_ref, o.item_id, i.title, i.image_url, COALESCE(o.quantity, 1) as quantity, o.status, o.created_at, o.magic_token 
		FROM orders o
		JOIN items i ON o.item_id = i.id
		WHERE LOWER(o.customer_email) = LOWER(?)
		ORDER BY o.created_at DESC
	`
	rows, err := s.DB.Query(query, email)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orders []models.Order
	for rows.Next() {
		var o models.Order
		// Note: scanning into partial struct (fields not in query will be zero-value)
		if err := rows.Scan(&o.ID, &o.OrderRef, &o.ItemID, &o.ItemTitle, &o.ItemImageURL, &o.Quantity, &o.Status, &o.CreatedAt, &o.MagicToken); err != nil {
			return nil, err
		}
		orders = append(orders, o)
	}
	return orders, nil
}

func (s *Store) UpdateOrderToken(id int, token string) error {
	query := `UPDATE orders SET magic_token = ?, magic_token_expiry = datetime('now', '+30 days') WHERE id = ?`
	_, err := s.DB.Exec(query, token, id)
	return err
}

// New Login Token Logic

func (s *Store) CreateLoginToken(email, token string) error {
	// Expires in 1 hour
	query := `INSERT INTO login_tokens (token, email, expires_at) VALUES (?, ?, datetime('now', '+1 hour'))`
	_, err := s.DB.Exec(query, token, email)
	return err
}

func (s *Store) GetEmailByLoginToken(token string) (string, error) {
	var email string
	// Check if token exists and is not expired
	query := `SELECT email FROM login_tokens WHERE token = ? AND expires_at > datetime('now')`
	err := s.DB.QueryRow(query, token).Scan(&email)
	if err != nil {
		return "", err
	}
	return email, nil
}
