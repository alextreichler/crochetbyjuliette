package store

import (
	"github.com/alextreichler/crochetbyjuliette/internal/models"
)

func (s *Store) CreateItem(item *models.Item) error {
	query := `
		INSERT INTO items (title, description, price, delivery_time, image_url, status, created_at)
		VALUES (?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP)
	`
	_, err := s.DB.Exec(query, item.Title, item.Description, item.Price, item.DeliveryTime, item.ImageURL, item.Status)
	return err
}

func (s *Store) GetAllItems() ([]models.Item, error) {
	// Ensure we select status. For migration safety, if column doesn't exist this fails.
	// Ideally we'd use a migration tool.
	query := `SELECT id, title, description, price, delivery_time, image_url, COALESCE(status, 'available') as status, created_at FROM items ORDER BY created_at DESC`
	rows, err := s.DB.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []models.Item
	for rows.Next() {
		var i models.Item
		if err := rows.Scan(&i.ID, &i.Title, &i.Description, &i.Price, &i.DeliveryTime, &i.ImageURL, &i.Status, &i.CreatedAt); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	return items, nil
}

func (s *Store) DeleteItem(id int) error {
	query := `DELETE FROM items WHERE id = ?`
	_, err := s.DB.Exec(query, id)
	return err
}