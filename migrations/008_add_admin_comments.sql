-- Migration: 008_add_admin_comments.sql
ALTER TABLE orders ADD COLUMN admin_comments TEXT DEFAULT '';
