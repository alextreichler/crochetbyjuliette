package store

import "database/sql"

type DashboardStats struct {
	TotalItems       int
	TotalOrders      int
	OrdersByStatus   map[string]int
	ItemOrderCounts  []ItemOrderCount
}

type ItemOrderCount struct {
	ItemID    int
	Title     string
	OrderCount int
}

func (s *Store) GetDashboardStats() (*DashboardStats, error) {
	stats := &DashboardStats{
		OrdersByStatus: make(map[string]int),
	}

	// 1. Total Items
	err := s.DB.QueryRow("SELECT COUNT(*) FROM items").Scan(&stats.TotalItems)
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}

	// 2. Total Orders
	err = s.DB.QueryRow("SELECT COUNT(*) FROM orders").Scan(&stats.TotalOrders)
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}

	// 3. Orders by Status
	rows, err := s.DB.Query("SELECT status, COUNT(*) FROM orders GROUP BY status")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var status string
		var count int
		if err := rows.Scan(&status, &count); err != nil {
			return nil, err
		}
		stats.OrdersByStatus[status] = count
	}

	// 4. Orders per Item
	itemRows, err := s.DB.Query(`
		SELECT i.id, i.title, COUNT(o.id) as order_count 
		FROM items i 
		LEFT JOIN orders o ON i.id = o.item_id 
		GROUP BY i.id 
		ORDER BY order_count DESC
	`)
	if err != nil {
		return nil, err
	}
	defer itemRows.Close()
	for itemRows.Next() {
		var ioc ItemOrderCount
		if err := itemRows.Scan(&ioc.ItemID, &ioc.Title, &ioc.OrderCount); err != nil {
			return nil, err
		}
		stats.ItemOrderCounts = append(stats.ItemOrderCounts, ioc)
	}

	return stats, nil
}
