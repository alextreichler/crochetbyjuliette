-- Migration: 004_add_magic_token.sql
ALTER TABLE orders ADD COLUMN magic_token TEXT;
ALTER TABLE orders ADD COLUMN magic_token_expiry DATETIME;
