-- Migration: 003_create_orders_table.sql
CREATE TABLE IF NOT EXISTS orders (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    item_id INTEGER NOT NULL,
    customer_name TEXT NOT NULL,
    customer_email TEXT NOT NULL,
    customer_address TEXT,
    status TEXT DEFAULT 'Ordered', -- Ordered, In Progress, Completed, Needs Shipping, Shipped, Delivered
    notes TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY(item_id) REFERENCES items(id)
);
