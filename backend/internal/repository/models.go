package repository

import "time"

// ClassType represents a category of class offered by the dojo (e.g. "Karate", "Yoga").
type ClassType struct {
	ID          int64
	Name        string
	Description string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// Class represents a scheduled class instance.
type Class struct {
	ID              int64
	ClassTypeID     int64
	InstructorID    int64
	StartTime       time.Time
	DurationMinutes int
	Capacity        int
	CreatedAt       time.Time
	UpdatedAt       time.Time
}
