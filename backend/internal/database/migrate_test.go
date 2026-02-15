package database

import (
	"database/sql"
	"testing"
	"testing/fstest"
)

func openTestDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := Open(":memory:")
	if err != nil {
		t.Fatalf("failed to open test database: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return db
}

func TestMigrate_AppliesMigrations(t *testing.T) {
	db := openTestDB(t)

	migrations := fstest.MapFS{
		"001_create_foo.sql": &fstest.MapFile{
			Data: []byte("CREATE TABLE foo (id INTEGER PRIMARY KEY, name TEXT);"),
		},
	}

	if err := Migrate(db, migrations); err != nil {
		t.Fatalf("Migrate failed: %v", err)
	}

	// Verify table was created by inserting a row.
	if _, err := db.Exec("INSERT INTO foo (name) VALUES ('bar')"); err != nil {
		t.Fatalf("expected foo table to exist: %v", err)
	}
}

func TestMigrate_RecordsAppliedMigrations(t *testing.T) {
	db := openTestDB(t)

	migrations := fstest.MapFS{
		"001_first.sql": &fstest.MapFile{
			Data: []byte("CREATE TABLE first (id INTEGER PRIMARY KEY);"),
		},
	}

	if err := Migrate(db, migrations); err != nil {
		t.Fatalf("Migrate failed: %v", err)
	}

	var count int
	if err := db.QueryRow("SELECT COUNT(*) FROM schema_migrations").Scan(&count); err != nil {
		t.Fatalf("query schema_migrations: %v", err)
	}
	if count != 1 {
		t.Errorf("expected 1 applied migration, got %d", count)
	}
}

func TestMigrate_SkipsAlreadyApplied(t *testing.T) {
	db := openTestDB(t)

	migrations := fstest.MapFS{
		"001_create_foo.sql": &fstest.MapFile{
			Data: []byte("CREATE TABLE foo (id INTEGER PRIMARY KEY);"),
		},
	}

	// Run twice — second run should be a no-op.
	if err := Migrate(db, migrations); err != nil {
		t.Fatalf("first Migrate failed: %v", err)
	}
	if err := Migrate(db, migrations); err != nil {
		t.Fatalf("second Migrate failed (should be idempotent): %v", err)
	}
}

func TestMigrate_RunsInOrder(t *testing.T) {
	db := openTestDB(t)

	migrations := fstest.MapFS{
		"002_insert.sql": &fstest.MapFile{
			Data: []byte("INSERT INTO ordered (val) VALUES ('second');"),
		},
		"001_create.sql": &fstest.MapFile{
			Data: []byte("CREATE TABLE ordered (id INTEGER PRIMARY KEY AUTOINCREMENT, val TEXT);"),
		},
	}

	if err := Migrate(db, migrations); err != nil {
		t.Fatalf("Migrate failed: %v", err)
	}

	var val string
	if err := db.QueryRow("SELECT val FROM ordered").Scan(&val); err != nil {
		t.Fatalf("query ordered table: %v", err)
	}
	if val != "second" {
		t.Errorf("expected 'second', got '%s'", val)
	}
}

func TestMigrate_AppliesOnlyNewMigrations(t *testing.T) {
	db := openTestDB(t)

	// First run with one migration.
	first := fstest.MapFS{
		"001_create.sql": &fstest.MapFile{
			Data: []byte("CREATE TABLE incremental (id INTEGER PRIMARY KEY);"),
		},
	}
	if err := Migrate(db, first); err != nil {
		t.Fatalf("first Migrate failed: %v", err)
	}

	// Second run with two migrations — only the new one should run.
	both := fstest.MapFS{
		"001_create.sql": &fstest.MapFile{
			Data: []byte("CREATE TABLE incremental (id INTEGER PRIMARY KEY);"),
		},
		"002_add_column.sql": &fstest.MapFile{
			Data: []byte("ALTER TABLE incremental ADD COLUMN name TEXT;"),
		},
	}
	if err := Migrate(db, both); err != nil {
		t.Fatalf("second Migrate failed: %v", err)
	}

	// Verify the new column exists.
	if _, err := db.Exec("INSERT INTO incremental (name) VALUES ('test')"); err != nil {
		t.Fatalf("expected name column to exist: %v", err)
	}
}

func TestMigrate_RollsBackOnError(t *testing.T) {
	db := openTestDB(t)

	migrations := fstest.MapFS{
		"001_bad.sql": &fstest.MapFile{
			Data: []byte("THIS IS NOT VALID SQL;"),
		},
	}

	err := Migrate(db, migrations)
	if err == nil {
		t.Fatal("expected Migrate to return error for invalid SQL")
	}

	// Verify the failed migration was not recorded.
	var count int
	if err := db.QueryRow("SELECT COUNT(*) FROM schema_migrations").Scan(&count); err != nil {
		t.Fatalf("query schema_migrations: %v", err)
	}
	if count != 0 {
		t.Errorf("expected 0 applied migrations after failure, got %d", count)
	}
}

func TestMigrate_RealMigrations(t *testing.T) {
	db := openTestDB(t)

	if err := Migrate(db, MigrationsFS); err != nil {
		t.Fatalf("Migrate with real migrations failed: %v", err)
	}

	// Verify all expected tables exist by querying sqlite_master.
	expectedTables := []string{"users", "class_types", "classes", "attendance", "payments"}
	for _, table := range expectedTables {
		var name string
		err := db.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name=?", table).Scan(&name)
		if err != nil {
			t.Errorf("expected table %q to exist: %v", table, err)
		}
	}
}

func TestMigrate_ForeignKeysEnforced(t *testing.T) {
	db := openTestDB(t)

	if err := Migrate(db, MigrationsFS); err != nil {
		t.Fatalf("Migrate failed: %v", err)
	}

	// Attempt to insert a class with a non-existent class_type_id.
	_, err := db.Exec(`INSERT INTO classes (class_type_id, instructor_id, start_time, duration_minutes, capacity)
		VALUES (999, 999, '2026-01-01 10:00:00', 60, 20)`)
	if err == nil {
		t.Error("expected foreign key violation, got nil")
	}
}

func TestOpen_SetsPragmas(t *testing.T) {
	db := openTestDB(t)

	var journalMode string
	if err := db.QueryRow("PRAGMA journal_mode").Scan(&journalMode); err != nil {
		t.Fatalf("query journal_mode: %v", err)
	}
	// In-memory databases use "memory" journal mode, but for file-based DBs
	// it would be "wal". We just verify the pragma was accepted.
	if journalMode != "memory" && journalMode != "wal" {
		t.Errorf("unexpected journal_mode: %s", journalMode)
	}

	var fkEnabled int
	if err := db.QueryRow("PRAGMA foreign_keys").Scan(&fkEnabled); err != nil {
		t.Fatalf("query foreign_keys: %v", err)
	}
	if fkEnabled != 1 {
		t.Errorf("expected foreign_keys=1, got %d", fkEnabled)
	}
}
