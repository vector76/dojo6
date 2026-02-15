package repository

import (
	"database/sql"
	"testing"
	"time"
)

func TestCreateClass(t *testing.T) {
	tests := []struct {
		name    string
		input   Class
		wantErr bool
		check   func(t *testing.T, got Class)
	}{
		{
			name: "basic creation",
			input: Class{
				StartTime:       time.Date(2026, 3, 1, 10, 0, 0, 0, time.UTC),
				DurationMinutes: 60,
				Capacity:        20,
			},
			check: func(t *testing.T, got Class) {
				if got.ID == 0 {
					t.Error("expected non-zero ID")
				}
				if got.DurationMinutes != 60 {
					t.Errorf("duration = %d, want 60", got.DurationMinutes)
				}
				if got.Capacity != 20 {
					t.Errorf("capacity = %d, want 20", got.Capacity)
				}
				if got.CreatedAt.IsZero() {
					t.Error("expected non-zero CreatedAt")
				}
			},
		},
		{
			name: "invalid class_type_id",
			input: Class{
				ClassTypeID:     9999,
				StartTime:       time.Date(2026, 3, 1, 10, 0, 0, 0, time.UTC),
				DurationMinutes: 60,
				Capacity:        20,
			},
			wantErr: true,
		},
		{
			name: "invalid instructor_id",
			input: Class{
				InstructorID:    9999,
				StartTime:       time.Date(2026, 3, 1, 10, 0, 0, 0, time.UTC),
				DurationMinutes: 60,
				Capacity:        20,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := openTestDB(t)
			ctID := seedClassType(t, db)
			instrID := seedInstructor(t, db)

			input := tt.input
			// Use valid FKs unless the test explicitly sets invalid ones.
			if input.ClassTypeID == 0 {
				input.ClassTypeID = ctID
			}
			if input.InstructorID == 0 {
				input.InstructorID = instrID
			}

			got, err := CreateClass(db, input.ClassTypeID, input.InstructorID, input.StartTime, input.DurationMinutes, input.Capacity)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got.ClassTypeID != ctID {
				t.Errorf("ClassTypeID = %d, want %d", got.ClassTypeID, ctID)
			}
			if got.InstructorID != instrID {
				t.Errorf("InstructorID = %d, want %d", got.InstructorID, instrID)
			}
			tt.check(t, got)
		})
	}
}

func TestGetClass(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(t *testing.T, db *sql.DB) int64
		wantErr bool
	}{
		{
			name: "existing class",
			setup: func(t *testing.T, db *sql.DB) int64 {
				ctID := seedClassType(t, db)
				instrID := seedInstructor(t, db)
				c, err := CreateClass(db, ctID, instrID, time.Date(2026, 3, 1, 10, 0, 0, 0, time.UTC), 60, 20)
				if err != nil {
					t.Fatalf("setup: %v", err)
				}
				return c.ID
			},
		},
		{
			name: "non-existent ID",
			setup: func(t *testing.T, db *sql.DB) int64 {
				return 9999
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := openTestDB(t)
			id := tt.setup(t, db)
			got, err := GetClass(db, id)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got.ID != id {
				t.Errorf("ID = %d, want %d", got.ID, id)
			}
			if got.DurationMinutes != 60 {
				t.Errorf("duration = %d, want 60", got.DurationMinutes)
			}
		})
	}
}

func TestListClasses(t *testing.T) {
	tests := []struct {
		name      string
		seedCount int
		wantCount int
	}{
		{name: "empty", seedCount: 0, wantCount: 0},
		{name: "one class", seedCount: 1, wantCount: 1},
		{name: "multiple classes", seedCount: 3, wantCount: 3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := openTestDB(t)
			var ctID, instrID int64
			if tt.seedCount > 0 {
				ctID = seedClassType(t, db)
				instrID = seedInstructor(t, db)
			}
			for i := 0; i < tt.seedCount; i++ {
				startTime := time.Date(2026, 3, 1+i, 10, 0, 0, 0, time.UTC)
				if _, err := CreateClass(db, ctID, instrID, startTime, 60, 20); err != nil {
					t.Fatalf("seed %d: %v", i, err)
				}
			}
			got, err := ListClasses(db)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(got) != tt.wantCount {
				t.Errorf("count = %d, want %d", len(got), tt.wantCount)
			}
		})
	}
}

func TestUpdateClass(t *testing.T) {
	tests := []struct {
		name    string
		modify  func(c *Class)
		wantErr bool
		check   func(t *testing.T, got Class)
	}{
		{
			name: "update time and capacity",
			modify: func(c *Class) {
				c.StartTime = time.Date(2026, 4, 1, 14, 0, 0, 0, time.UTC)
				c.Capacity = 30
			},
			check: func(t *testing.T, got Class) {
				if got.Capacity != 30 {
					t.Errorf("capacity = %d, want 30", got.Capacity)
				}
				want := time.Date(2026, 4, 1, 14, 0, 0, 0, time.UTC)
				if !got.StartTime.Equal(want) {
					t.Errorf("start_time = %v, want %v", got.StartTime, want)
				}
			},
		},
		{
			name: "update duration",
			modify: func(c *Class) {
				c.DurationMinutes = 90
			},
			check: func(t *testing.T, got Class) {
				if got.DurationMinutes != 90 {
					t.Errorf("duration = %d, want 90", got.DurationMinutes)
				}
			},
		},
		{
			name: "invalid instructor FK",
			modify: func(c *Class) {
				c.InstructorID = 9999
			},
			wantErr: true,
		},
		{
			name: "invalid class_type FK",
			modify: func(c *Class) {
				c.ClassTypeID = 9999
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := openTestDB(t)
			ctID := seedClassType(t, db)
			instrID := seedInstructor(t, db)

			orig, err := CreateClass(db, ctID, instrID, time.Date(2026, 3, 1, 10, 0, 0, 0, time.UTC), 60, 20)
			if err != nil {
				t.Fatalf("setup: %v", err)
			}

			updated := orig
			tt.modify(&updated)

			got, err := UpdateClass(db, updated.ID, updated.ClassTypeID, updated.InstructorID, updated.StartTime, updated.DurationMinutes, updated.Capacity)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			tt.check(t, got)
		})
	}
}

func TestUpdateClass_NotFound(t *testing.T) {
	db := openTestDB(t)
	ctID := seedClassType(t, db)
	instrID := seedInstructor(t, db)
	_, err := UpdateClass(db, 9999, ctID, instrID, time.Now(), 60, 20)
	if err == nil {
		t.Fatal("expected error for non-existent ID")
	}
}

func TestDeleteClass(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(t *testing.T, db *sql.DB) int64
		wantErr bool
	}{
		{
			name: "delete existing",
			setup: func(t *testing.T, db *sql.DB) int64 {
				ctID := seedClassType(t, db)
				instrID := seedInstructor(t, db)
				c, err := CreateClass(db, ctID, instrID, time.Date(2026, 3, 1, 10, 0, 0, 0, time.UTC), 60, 20)
				if err != nil {
					t.Fatalf("setup: %v", err)
				}
				return c.ID
			},
		},
		{
			name: "delete non-existent",
			setup: func(t *testing.T, db *sql.DB) int64 {
				return 9999
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := openTestDB(t)
			id := tt.setup(t, db)
			err := DeleteClass(db, id)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			// Verify it's gone.
			_, err = GetClass(db, id)
			if err == nil {
				t.Error("expected error after delete, got nil")
			}
		})
	}
}
