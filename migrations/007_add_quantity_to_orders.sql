-- Migration: 007_add_quantity_to_orders.sql
ALTER TABLE orders ADD COLUMN quantity INTEGER DEFAULT 1;
