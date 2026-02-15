package repository

import (
	"database/sql"
	"testing"

	"dojo6/backend/internal/database"
)

// openTestDB opens an in-memory SQLite database with migrations applied.
func openTestDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := database.Open(":memory:")
	if err != nil {
		t.Fatalf("open test db: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	if err := database.Migrate(db, database.MigrationsFS); err != nil {
		t.Fatalf("migrate test db: %v", err)
	}
	return db
}

// seedInstructor inserts a minimal user row so foreign key constraints on
// classes.instructor_id are satisfied. Returns the user ID.
func seedInstructor(t *testing.T, db *sql.DB) int64 {
	t.Helper()
	res, err := db.Exec(`INSERT INTO users (name, email, phone, role, password_hash)
		VALUES ('Test Instructor', 'instructor@test.com', '555-0100', 'instructor', 'hash')`)
	if err != nil {
		t.Fatalf("seed instructor: %v", err)
	}
	id, _ := res.LastInsertId()
	return id
}

// seedClassType inserts a minimal class_type row. Returns the class type ID.
func seedClassType(t *testing.T, db *sql.DB) int64 {
	t.Helper()
	res, err := db.Exec(`INSERT INTO class_types (name) VALUES ('Karate')`)
	if err != nil {
		t.Fatalf("seed class type: %v", err)
	}
	id, _ := res.LastInsertId()
	return id
}
