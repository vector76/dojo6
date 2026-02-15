package models

import (
	"database/sql"
	"testing"
)

// seedClassDeps creates the prerequisite data for attendance tests:
// a user, a class type, and a class. Returns (userID, classID).
func seedClassDeps(t *testing.T, db *sql.DB) (int64, int64) {
	t.Helper()

	// Create a user (instructor).
	res, err := db.Exec(`
		INSERT INTO users (name, email, phone, role, password_hash)
		VALUES ('Instructor', 'instructor@example.com', '555-0001', 'instructor', 'hashed')`)
	if err != nil {
		t.Fatalf("seed instructor: %v", err)
	}
	instructorID, _ := res.LastInsertId()

	// Create a class type.
	res, err = db.Exec(`
		INSERT INTO class_types (name, description) VALUES ('Yoga', 'Basic yoga class')`)
	if err != nil {
		t.Fatalf("seed class_type: %v", err)
	}
	classTypeID, _ := res.LastInsertId()

	// Create a class.
	res, err = db.Exec(`
		INSERT INTO classes (class_type_id, instructor_id, start_time, duration_minutes, capacity)
		VALUES (?, ?, '2026-02-15 10:00:00', 60, 20)`, classTypeID, instructorID)
	if err != nil {
		t.Fatalf("seed class: %v", err)
	}
	classID, _ := res.LastInsertId()

	// Create a student user.
	res, err = db.Exec(`
		INSERT INTO users (name, email, phone, role, password_hash)
		VALUES ('Student', 'student@example.com', '555-0002', 'user', 'hashed')`)
	if err != nil {
		t.Fatalf("seed student: %v", err)
	}
	studentID, _ := res.LastInsertId()

	return studentID, classID
}

func TestAttendanceRepository_Create(t *testing.T) {
	db := setupTestDB(t)
	repo := NewAttendanceRepository(db)
	studentID, classID := seedClassDeps(t, db)

	tests := []struct {
		name    string
		att     Attendance
		wantErr bool
	}{
		{
			name: "valid check-in",
			att:  Attendance{ClassID: classID, UserID: studentID},
		},
		{
			name:    "duplicate check-in",
			att:     Attendance{ClassID: classID, UserID: studentID},
			wantErr: true,
		},
		{
			name:    "non-existent class",
			att:     Attendance{ClassID: 9999, UserID: studentID},
			wantErr: true,
		},
		{
			name:    "non-existent user",
			att:     Attendance{ClassID: classID, UserID: 9999},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := tt.att
			err := repo.Create(&a)
			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if a.ID == 0 {
				t.Error("expected non-zero ID")
			}
			if a.CheckedInAt.IsZero() {
				t.Error("expected CheckedInAt to be set")
			}
		})
	}
}

func TestAttendanceRepository_GetByID(t *testing.T) {
	db := setupTestDB(t)
	repo := NewAttendanceRepository(db)
	studentID, classID := seedClassDeps(t, db)

	created := &Attendance{ClassID: classID, UserID: studentID}
	if err := repo.Create(created); err != nil {
		t.Fatalf("create: %v", err)
	}

	tests := []struct {
		name    string
		id      int64
		wantErr bool
	}{
		{name: "existing record", id: created.ID},
		{name: "non-existent record", id: 9999, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var a Attendance
			err := repo.GetByID(tt.id, &a)
			if tt.wantErr {
				if err != sql.ErrNoRows {
					t.Errorf("expected ErrNoRows, got %v", err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if a.ClassID != classID {
				t.Errorf("expected classID %d, got %d", classID, a.ClassID)
			}
			if a.UserID != studentID {
				t.Errorf("expected userID %d, got %d", studentID, a.UserID)
			}
		})
	}
}

func TestAttendanceRepository_ListByClass(t *testing.T) {
	db := setupTestDB(t)
	repo := NewAttendanceRepository(db)
	studentID, classID := seedClassDeps(t, db)

	// Add a second student.
	res, err := db.Exec(`
		INSERT INTO users (name, email, phone, role, password_hash)
		VALUES ('Student2', 'student2@example.com', '555-0003', 'user', 'hashed')`)
	if err != nil {
		t.Fatalf("seed student2: %v", err)
	}
	student2ID, _ := res.LastInsertId()

	a1 := &Attendance{ClassID: classID, UserID: studentID}
	a2 := &Attendance{ClassID: classID, UserID: student2ID}
	if err := repo.Create(a1); err != nil {
		t.Fatalf("create a1: %v", err)
	}
	if err := repo.Create(a2); err != nil {
		t.Fatalf("create a2: %v", err)
	}

	records, err := repo.ListByClass(classID)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(records) != 2 {
		t.Fatalf("expected 2 records, got %d", len(records))
	}

	// Empty class should return nil slice.
	empty, err := repo.ListByClass(9999)
	if err != nil {
		t.Fatalf("list empty: %v", err)
	}
	if len(empty) != 0 {
		t.Errorf("expected 0 records for non-existent class, got %d", len(empty))
	}
}

func TestAttendanceRepository_ListByUser(t *testing.T) {
	db := setupTestDB(t)
	repo := NewAttendanceRepository(db)
	studentID, classID := seedClassDeps(t, db)

	// Create a second class.
	res, err := db.Exec(`
		INSERT INTO classes (class_type_id, instructor_id, start_time, duration_minutes, capacity)
		VALUES (1, 1, '2026-02-16 10:00:00', 60, 20)`)
	if err != nil {
		t.Fatalf("seed class2: %v", err)
	}
	class2ID, _ := res.LastInsertId()

	a1 := &Attendance{ClassID: classID, UserID: studentID}
	a2 := &Attendance{ClassID: class2ID, UserID: studentID}
	if err := repo.Create(a1); err != nil {
		t.Fatalf("create a1: %v", err)
	}
	if err := repo.Create(a2); err != nil {
		t.Fatalf("create a2: %v", err)
	}

	records, err := repo.ListByUser(studentID)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(records) != 2 {
		t.Fatalf("expected 2 records, got %d", len(records))
	}

	// Non-existent user should return empty.
	empty, err := repo.ListByUser(9999)
	if err != nil {
		t.Fatalf("list empty: %v", err)
	}
	if len(empty) != 0 {
		t.Errorf("expected 0 records for non-existent user, got %d", len(empty))
	}
}
