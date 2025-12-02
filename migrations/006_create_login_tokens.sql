-- Migration: 006_create_login_tokens.sql
CREATE TABLE IF NOT EXISTS login_tokens (
    token TEXT PRIMARY KEY,
    email TEXT NOT NULL,
    expires_at DATETIME NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
