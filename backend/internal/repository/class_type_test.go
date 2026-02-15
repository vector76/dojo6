package repository

import (
	"database/sql"
	"testing"
)

func TestCreateClassType(t *testing.T) {
	tests := []struct {
		name        string
		input       ClassType
		wantErr     bool
		checkResult func(t *testing.T, got ClassType)
	}{
		{
			name:  "basic creation",
			input: ClassType{Name: "Karate", Description: "Traditional martial art"},
			checkResult: func(t *testing.T, got ClassType) {
				if got.ID == 0 {
					t.Error("expected non-zero ID")
				}
				if got.Name != "Karate" {
					t.Errorf("name = %q, want %q", got.Name, "Karate")
				}
				if got.Description != "Traditional martial art" {
					t.Errorf("description = %q, want %q", got.Description, "Traditional martial art")
				}
				if got.CreatedAt.IsZero() {
					t.Error("expected non-zero CreatedAt")
				}
				if got.UpdatedAt.IsZero() {
					t.Error("expected non-zero UpdatedAt")
				}
			},
		},
		{
			name:  "empty description",
			input: ClassType{Name: "Yoga"},
			checkResult: func(t *testing.T, got ClassType) {
				if got.Name != "Yoga" {
					t.Errorf("name = %q, want %q", got.Name, "Yoga")
				}
			},
		},
		{
			name:    "empty name fails",
			input:   ClassType{Name: ""},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := openTestDB(t)
			got, err := CreateClassType(db, tt.input.Name, tt.input.Description)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			tt.checkResult(t, got)
		})
	}
}

func TestGetClassType(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(t *testing.T, db *sql.DB) int64
		wantErr bool
	}{
		{
			name: "existing class type",
			setup: func(t *testing.T, db *sql.DB) int64 {
				ct, err := CreateClassType(db, "Karate", "desc")
				if err != nil {
					t.Fatalf("setup: %v", err)
				}
				return ct.ID
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
			got, err := GetClassType(db, id)
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
			if got.Name != "Karate" {
				t.Errorf("name = %q, want %q", got.Name, "Karate")
			}
		})
	}
}

func TestListClassTypes(t *testing.T) {
	tests := []struct {
		name      string
		seedCount int
		wantCount int
	}{
		{name: "empty list", seedCount: 0, wantCount: 0},
		{name: "one item", seedCount: 1, wantCount: 1},
		{name: "multiple items", seedCount: 3, wantCount: 3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := openTestDB(t)
			for i := 0; i < tt.seedCount; i++ {
				if _, err := CreateClassType(db, "Type"+string(rune('A'+i)), ""); err != nil {
					t.Fatalf("seed %d: %v", i, err)
				}
			}
			got, err := ListClassTypes(db)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(got) != tt.wantCount {
				t.Errorf("count = %d, want %d", len(got), tt.wantCount)
			}
		})
	}
}

func TestUpdateClassType(t *testing.T) {
	tests := []struct {
		name    string
		newName string
		newDesc string
		wantErr bool
	}{
		{name: "update name and description", newName: "Judo", newDesc: "Updated"},
		{name: "update name only", newName: "Aikido", newDesc: ""},
		{name: "empty name fails", newName: "", newDesc: "desc", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := openTestDB(t)
			ct, err := CreateClassType(db, "Original", "Original desc")
			if err != nil {
				t.Fatalf("setup: %v", err)
			}

			got, err := UpdateClassType(db, ct.ID, tt.newName, tt.newDesc)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got.Name != tt.newName {
				t.Errorf("name = %q, want %q", got.Name, tt.newName)
			}
			if got.Description != tt.newDesc {
				t.Errorf("description = %q, want %q", got.Description, tt.newDesc)
			}
		})
	}
}

func TestUpdateClassType_NotFound(t *testing.T) {
	db := openTestDB(t)
	_, err := UpdateClassType(db, 9999, "Name", "Desc")
	if err == nil {
		t.Fatal("expected error for non-existent ID")
	}
}

func TestDeleteClassType(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(t *testing.T, db *sql.DB) int64
		wantErr bool
	}{
		{
			name: "delete existing",
			setup: func(t *testing.T, db *sql.DB) int64 {
				ct, err := CreateClassType(db, "ToDelete", "")
				if err != nil {
					t.Fatalf("setup: %v", err)
				}
				return ct.ID
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
			err := DeleteClassType(db, id)
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
			_, err = GetClassType(db, id)
			if err == nil {
				t.Error("expected error after delete, got nil")
			}
		})
	}
}
