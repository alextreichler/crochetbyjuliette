-- Migration: 002_add_status_column.sql
ALTER TABLE items ADD COLUMN status TEXT DEFAULT 'available';
