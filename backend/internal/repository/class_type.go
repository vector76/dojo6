package repository

import (
	"database/sql"
	"errors"
	"fmt"
)

// CreateClassType inserts a new class type and returns it with populated fields.
func CreateClassType(db *sql.DB, name, description string) (ClassType, error) {
	if name == "" {
		return ClassType{}, errors.New("name is required")
	}

	res, err := db.Exec(
		`INSERT INTO class_types (name, description) VALUES (?, ?)`,
		name, description,
	)
	if err != nil {
		return ClassType{}, fmt.Errorf("insert class_type: %w", err)
	}
	id, _ := res.LastInsertId()
	return GetClassType(db, id)
}

// GetClassType retrieves a class type by ID.
func GetClassType(db *sql.DB, id int64) (ClassType, error) {
	var ct ClassType
	var createdAt, updatedAt string
	err := db.QueryRow(
		`SELECT id, name, description, created_at, updated_at FROM class_types WHERE id = ?`, id,
	).Scan(&ct.ID, &ct.Name, &ct.Description, &createdAt, &updatedAt)
	if err != nil {
		return ClassType{}, fmt.Errorf("get class_type %d: %w", id, err)
	}
	ct.CreatedAt = parseTime(createdAt)
	ct.UpdatedAt = parseTime(updatedAt)
	return ct, nil
}

// ListClassTypes returns all class types ordered by name.
func ListClassTypes(db *sql.DB) ([]ClassType, error) {
	rows, err := db.Query(`SELECT id, name, description, created_at, updated_at FROM class_types ORDER BY name`)
	if err != nil {
		return nil, fmt.Errorf("list class_types: %w", err)
	}
	defer rows.Close()

	var result []ClassType
	for rows.Next() {
		var ct ClassType
		var createdAt, updatedAt string
		if err := rows.Scan(&ct.ID, &ct.Name, &ct.Description, &createdAt, &updatedAt); err != nil {
			return nil, fmt.Errorf("scan class_type: %w", err)
		}
		ct.CreatedAt = parseTime(createdAt)
		ct.UpdatedAt = parseTime(updatedAt)
		result = append(result, ct)
	}
	return result, rows.Err()
}

// UpdateClassType updates the name and description of an existing class type.
func UpdateClassType(db *sql.DB, id int64, name, description string) (ClassType, error) {
	if name == "" {
		return ClassType{}, errors.New("name is required")
	}

	res, err := db.Exec(
		`UPDATE class_types SET name = ?, description = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`,
		name, description, id,
	)
	if err != nil {
		return ClassType{}, fmt.Errorf("update class_type %d: %w", id, err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return ClassType{}, fmt.Errorf("class_type %d not found", id)
	}
	return GetClassType(db, id)
}

// DeleteClassType removes a class type by ID.
func DeleteClassType(db *sql.DB, id int64) error {
	res, err := db.Exec(`DELETE FROM class_types WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("delete class_type %d: %w", id, err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return fmt.Errorf("class_type %d not found", id)
	}
	return nil
}
