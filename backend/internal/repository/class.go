package repository

import (
	"database/sql"
	"fmt"
	"time"
)

// CreateClass inserts a new class and returns it with populated fields.
func CreateClass(db *sql.DB, classTypeID, instructorID int64, startTime time.Time, durationMinutes, capacity int) (Class, error) {
	res, err := db.Exec(
		`INSERT INTO classes (class_type_id, instructor_id, start_time, duration_minutes, capacity)
		 VALUES (?, ?, ?, ?, ?)`,
		classTypeID, instructorID, startTime.UTC().Format(time.RFC3339), durationMinutes, capacity,
	)
	if err != nil {
		return Class{}, fmt.Errorf("insert class: %w", err)
	}
	id, _ := res.LastInsertId()
	return GetClass(db, id)
}

// GetClass retrieves a class by ID.
func GetClass(db *sql.DB, id int64) (Class, error) {
	var c Class
	var startTime, createdAt, updatedAt string
	err := db.QueryRow(
		`SELECT id, class_type_id, instructor_id, start_time, duration_minutes, capacity, created_at, updated_at
		 FROM classes WHERE id = ?`, id,
	).Scan(&c.ID, &c.ClassTypeID, &c.InstructorID, &startTime, &c.DurationMinutes, &c.Capacity, &createdAt, &updatedAt)
	if err != nil {
		return Class{}, fmt.Errorf("get class %d: %w", id, err)
	}
	c.StartTime = parseTime(startTime)
	c.CreatedAt = parseTime(createdAt)
	c.UpdatedAt = parseTime(updatedAt)
	return c, nil
}

// ListClasses returns all classes ordered by start_time.
func ListClasses(db *sql.DB) ([]Class, error) {
	rows, err := db.Query(
		`SELECT id, class_type_id, instructor_id, start_time, duration_minutes, capacity, created_at, updated_at
		 FROM classes ORDER BY start_time`)
	if err != nil {
		return nil, fmt.Errorf("list classes: %w", err)
	}
	defer rows.Close()

	var result []Class
	for rows.Next() {
		var c Class
		var startTime, createdAt, updatedAt string
		if err := rows.Scan(&c.ID, &c.ClassTypeID, &c.InstructorID, &startTime, &c.DurationMinutes, &c.Capacity, &createdAt, &updatedAt); err != nil {
			return nil, fmt.Errorf("scan class: %w", err)
		}
		c.StartTime = parseTime(startTime)
		c.CreatedAt = parseTime(createdAt)
		c.UpdatedAt = parseTime(updatedAt)
		result = append(result, c)
	}
	return result, rows.Err()
}

// UpdateClass updates all mutable fields of a class.
func UpdateClass(db *sql.DB, id, classTypeID, instructorID int64, startTime time.Time, durationMinutes, capacity int) (Class, error) {
	res, err := db.Exec(
		`UPDATE classes SET class_type_id = ?, instructor_id = ?, start_time = ?, duration_minutes = ?, capacity = ?, updated_at = CURRENT_TIMESTAMP
		 WHERE id = ?`,
		classTypeID, instructorID, startTime.UTC().Format(time.RFC3339), durationMinutes, capacity, id,
	)
	if err != nil {
		return Class{}, fmt.Errorf("update class %d: %w", id, err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return Class{}, fmt.Errorf("class %d not found", id)
	}
	return GetClass(db, id)
}

// DeleteClass removes a class by ID.
func DeleteClass(db *sql.DB, id int64) error {
	res, err := db.Exec(`DELETE FROM classes WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("delete class %d: %w", id, err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return fmt.Errorf("class %d not found", id)
	}
	return nil
}
