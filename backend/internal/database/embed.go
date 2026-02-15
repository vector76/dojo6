package database

import (
	"embed"
	"io/fs"
)

//go:embed migrations/*.sql
var migrationsEmbed embed.FS

// MigrationsFS is the embedded migration files with the "migrations/" prefix
// stripped so file names appear at the root (e.g. "001_create_tables.sql").
var MigrationsFS, _ = fs.Sub(migrationsEmbed, "migrations")
