package models

import (
	"database/sql"
	"time"
)

type Attendance struct {
	ID          int64
	ClassID     int64
	UserID      int64
	CheckedInAt time.Time
}

type AttendanceRepository struct {
	db *sql.DB
}

func NewAttendanceRepository(db *sql.DB) *AttendanceRepository {
	return &AttendanceRepository{db: db}
}

// Create records a check-in. Attendance is append-only.
func (r *AttendanceRepository) Create(a *Attendance) error {
	result, err := r.db.Exec(`
		INSERT INTO attendance (class_id, user_id) VALUES (?, ?)`,
		a.ClassID, a.UserID,
	)
	if err != nil {
		return err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	a.ID = id
	return r.GetByID(a.ID, a)
}

// GetByID fetches a single attendance record.
func (r *AttendanceRepository) GetByID(id int64, a *Attendance) error {
	return r.db.QueryRow(`
		SELECT id, class_id, user_id, checked_in_at
		FROM attendance WHERE id = ?`, id,
	).Scan(&a.ID, &a.ClassID, &a.UserID, &a.CheckedInAt)
}

// ListByClass returns all attendance records for a given class.
func (r *AttendanceRepository) ListByClass(classID int64) ([]Attendance, error) {
	rows, err := r.db.Query(`
		SELECT id, class_id, user_id, checked_in_at
		FROM attendance WHERE class_id = ?
		ORDER BY checked_in_at`, classID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var records []Attendance
	for rows.Next() {
		var a Attendance
		if err := rows.Scan(&a.ID, &a.ClassID, &a.UserID, &a.CheckedInAt); err != nil {
			return nil, err
		}
		records = append(records, a)
	}
	return records, rows.Err()
}

// ListByUser returns all attendance records for a given user.
func (r *AttendanceRepository) ListByUser(userID int64) ([]Attendance, error) {
	rows, err := r.db.Query(`
		SELECT id, class_id, user_id, checked_in_at
		FROM attendance WHERE user_id = ?
		ORDER BY checked_in_at`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var records []Attendance
	for rows.Next() {
		var a Attendance
		if err := rows.Scan(&a.ID, &a.ClassID, &a.UserID, &a.CheckedInAt); err != nil {
			return nil, err
		}
		records = append(records, a)
	}
	return records, rows.Err()
}
