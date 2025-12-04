package store

import (
	"github.com/alextreichler/crochetbyjuliette/internal/models"
)

func (s *Store) GetPublicItems() ([]models.Item, error) {
	// Exclude archived items
	query := `SELECT id, title, description, price, delivery_time, image_url, COALESCE(status, 'available') as status, created_at 
	          FROM items 
	          WHERE status != 'archived' OR status IS NULL 
	          ORDER BY created_at DESC`
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

func (s *Store) GetItemByID(id int) (*models.Item, error) {
	query := `SELECT id, title, description, price, delivery_time, image_url, COALESCE(status, 'available') as status, created_at FROM items WHERE id = ?`
	var i models.Item
	err := s.DB.QueryRow(query, id).Scan(&i.ID, &i.Title, &i.Description, &i.Price, &i.DeliveryTime, &i.ImageURL, &i.Status, &i.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &i, nil
}

func (s *Store) UpdateItem(item *models.Item) error {
	query := `
		UPDATE items 
		SET title = ?, description = ?, price = ?, delivery_time = ?, status = ?
		WHERE id = ?
	`
	_, err := s.DB.Exec(query, item.Title, item.Description, item.Price, item.DeliveryTime, item.Status, item.ID)
	return err
}

func (s *Store) UpdateItemImage(id int, imageURL string) error {
	query := `UPDATE items SET image_url = ? WHERE id = ?`
	_, err := s.DB.Exec(query, imageURL, id)
	return err
}
