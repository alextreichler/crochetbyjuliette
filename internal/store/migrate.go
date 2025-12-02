package store

import (
	"database/sql"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// Migrate runs all .sql files in the migrations directory in order
func (s *Store) Migrate(migrationsDir string) error {
	// 1. Create migrations table if not exists to track applied migrations
	_, err := s.DB.Exec(`CREATE TABLE IF NOT EXISTS schema_migrations (
		version TEXT PRIMARY KEY,
		applied_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);`)
	if err != nil {
		return fmt.Errorf("failed to create schema_migrations table: %w", err)
	}

	// 2. Read migration files
	files, err := os.ReadDir(migrationsDir)
	if err != nil {
		return fmt.Errorf("failed to read migrations directory: %w", err)
	}

	var migrationFiles []string
	for _, f := range files {
		if !f.IsDir() && strings.HasSuffix(f.Name(), ".sql") {
			migrationFiles = append(migrationFiles, f.Name())
		}
	}
	sort.Strings(migrationFiles) // Ensure order 001, 002, ...

	// 3. Apply new migrations
	for _, file := range migrationFiles {
		if isApplied(s.DB, file) {
			slog.Info("Skipping already applied migration", "file", file)
			continue
		}

		slog.Info("Applying migration", "file", file)
		content, err := os.ReadFile(filepath.Join(migrationsDir, file))
		if err != nil {
			return fmt.Errorf("failed to read migration file %s: %w", file, err)
		}

		// Execute migration in a transaction
		tx, err := s.DB.Begin()
		if err != nil {
			return err
		}

		if _, err := tx.Exec(string(content)); err != nil {
			// If it's a known ignorable error (e.g., column already exists),
			// log a warning, rollback the current transaction, but proceed to record it as applied.
			if strings.Contains(err.Error(), "duplicate column name") {
				slog.Warn("Column likely already exists, marking as applied", "file", file)
				tx.Rollback() // Rollback the failed ALTER, but we still want to record this version
			} else {
				tx.Rollback() // Rollback for other, actual errors
				return fmt.Errorf("failed to execute migration %s: %w", file, err)
			}
		} else {
			// If Exec succeeds, commit the transaction
			if err := tx.Commit(); err != nil {
				return err
			}
		}

		// Always record the migration version after attempting to execute it,
		// whether it was fully successful or skipped due to existing column.
		// Use a separate, non-transactional Exec to avoid issues with previous transaction state.
		if _, err := s.DB.Exec(`INSERT INTO schema_migrations (version) VALUES (?)`, file); err != nil {
			return fmt.Errorf("failed to record migration %s: %w", file, err)
		}
	}

	return nil
}

func isApplied(db *sql.DB, version string) bool {
	var exists int
	err := db.QueryRow(`SELECT 1 FROM schema_migrations WHERE version = ?`, version).Scan(&exists)
	return err == nil
}
