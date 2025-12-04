package store

import (
	"github.com/alextreichler/crochetbyjuliette/internal/models"
)

func (s *Store) CreateOrder(order *models.Order) error {
	query := `
		INSERT INTO orders (item_id, order_ref, quantity, customer_name, customer_email, customer_address, delivery_method, payment_method, status, notes, magic_token, magic_token_expiry, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP)
	`
	_, err := s.DB.Exec(query, order.ItemID, order.OrderRef, order.Quantity, order.CustomerName, order.CustomerEmail, order.CustomerAddress, order.DeliveryMethod, order.PaymentMethod, order.Status, order.Notes, order.MagicToken, order.MagicTokenExpiry)
	return err
}

func (s *Store) GetAllOrders(limit, offset int) ([]models.Order, error) {
	query := `
		SELECT o.id, COALESCE(o.order_ref, CAST(o.id AS TEXT)) as order_ref, o.item_id, i.title, i.image_url, COALESCE(o.quantity, 1) as quantity, o.customer_name, o.customer_email, o.customer_address, COALESCE(o.delivery_method, 'shipping'), COALESCE(o.payment_method, 'in_person'), o.status, o.notes, COALESCE(o.admin_comments, '') as admin_comments, o.created_at 
		FROM orders o
		JOIN items i ON o.item_id = i.id
		ORDER BY o.created_at DESC
		LIMIT ? OFFSET ?
	`
	rows, err := s.DB.Query(query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orders []models.Order
	for rows.Next() {
		var o models.Order
		if err := rows.Scan(&o.ID, &o.OrderRef, &o.ItemID, &o.ItemTitle, &o.ItemImageURL, &o.Quantity, &o.CustomerName, &o.CustomerEmail, &o.CustomerAddress, &o.DeliveryMethod, &o.PaymentMethod, &o.Status, &o.Notes, &o.AdminComments, &o.CreatedAt); err != nil {
			return nil, err
		}
		orders = append(orders, o)
	}
	return orders, nil
}

func (s *Store) GetTotalOrdersCount() (int, error) {
	var count int
	err := s.DB.QueryRow("SELECT COUNT(*) FROM orders").Scan(&count)
	if err != nil {
		return 0, err
	}
	return count, nil
}

func (s *Store) UpdateOrderStatus(id int, status string, adminComments string) error {
	query := `UPDATE orders SET status = ?, admin_comments = ? WHERE id = ?`
	_, err := s.DB.Exec(query, status, adminComments, id)
	return err
}

func (s *Store) UpdateOrderDetails(order *models.Order) error {
	query := `UPDATE orders SET quantity = ?, customer_name = ?, customer_email = ?, customer_address = ?, delivery_method = ?, payment_method = ?, notes = ? WHERE id = ?`
	_, err := s.DB.Exec(query, order.Quantity, order.CustomerName, order.CustomerEmail, order.CustomerAddress, order.DeliveryMethod, order.PaymentMethod, order.Notes, order.ID)
	return err
}

func (s *Store) CancelOrder(id int) error {
	query := `UPDATE orders SET status = 'Cancelled' WHERE id = ?`
	_, err := s.DB.Exec(query, id)
	return err
}
