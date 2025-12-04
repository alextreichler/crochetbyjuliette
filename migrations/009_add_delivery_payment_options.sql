-- Migration: 009_add_delivery_payment_options.sql
ALTER TABLE orders ADD COLUMN delivery_method TEXT DEFAULT 'shipping'; -- 'shipping' or 'hand_delivered'
ALTER TABLE orders ADD COLUMN payment_method TEXT DEFAULT 'in_person'; -- 'in_person'
