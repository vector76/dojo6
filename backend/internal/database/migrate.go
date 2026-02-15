package database

import (
	"database/sql"
	"fmt"
	"io/fs"
	"path"
	"sort"
	"strings"
)

// Migrate runs all numbered SQL migration files from the given filesystem
// that have not yet been applied. Migration files must be named with a
// numeric prefix followed by an underscore (e.g. "001_create_tables.sql").
//
// Applied migrations are tracked in a `schema_migrations` table. Each
// migration runs inside a transaction.
func Migrate(db *sql.DB, migrations fs.FS) error {
	if err := ensureMigrationsTable(db); err != nil {
		return err
	}

	applied, err := appliedMigrations(db)
	if err != nil {
		return err
	}

	files, err := listMigrationFiles(migrations)
	if err != nil {
		return err
	}

	for _, f := range files {
		if applied[f] {
			continue
		}
		if err := runMigration(db, migrations, f); err != nil {
			return fmt.Errorf("migration %s: %w", f, err)
		}
	}

	return nil
}

func ensureMigrationsTable(db *sql.DB) error {
	_, err := db.Exec(`CREATE TABLE IF NOT EXISTS schema_migrations (
		filename TEXT PRIMARY KEY,
		applied_at DATETIME DEFAULT CURRENT_TIMESTAMP
	)`)
	if err != nil {
		return fmt.Errorf("create schema_migrations table: %w", err)
	}
	return nil
}

func appliedMigrations(db *sql.DB) (map[string]bool, error) {
	rows, err := db.Query("SELECT filename FROM schema_migrations")
	if err != nil {
		return nil, fmt.Errorf("query applied migrations: %w", err)
	}
	defer rows.Close()

	applied := make(map[string]bool)
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, fmt.Errorf("scan migration row: %w", err)
		}
		applied[name] = true
	}
	return applied, rows.Err()
}

func listMigrationFiles(migrations fs.FS) ([]string, error) {
	entries, err := fs.ReadDir(migrations, ".")
	if err != nil {
		return nil, fmt.Errorf("read migration directory: %w", err)
	}

	var files []string
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		if strings.HasSuffix(e.Name(), ".sql") {
			files = append(files, e.Name())
		}
	}
	sort.Strings(files)
	return files, nil
}

func runMigration(db *sql.DB, migrations fs.FS, filename string) error {
	content, err := fs.ReadFile(migrations, path.Join(".", filename))
	if err != nil {
		return fmt.Errorf("read file: %w", err)
	}

	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback()

	if _, err := tx.Exec(string(content)); err != nil {
		return fmt.Errorf("execute SQL: %w", err)
	}

	if _, err := tx.Exec("INSERT INTO schema_migrations (filename) VALUES (?)", filename); err != nil {
		return fmt.Errorf("record migration: %w", err)
	}

	return tx.Commit()
}
