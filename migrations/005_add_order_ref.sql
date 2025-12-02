-- Migration: 005_add_order_ref.sql
ALTER TABLE orders ADD COLUMN order_ref TEXT;
-- In a real scenario, we would backfill existing rows, but for now we leave them null or empty.
-- CREATE UNIQUE INDEX idx_orders_ref ON orders(order_ref);
