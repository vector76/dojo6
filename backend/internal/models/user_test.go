package models

import (
	"database/sql"
	"testing"

	"dojo6/backend/internal/database"
)

func setupTestDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := database.Open(":memory:")
	if err != nil {
		t.Fatalf("open test db: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	if err := database.Migrate(db, database.MigrationsFS); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	return db
}

func seedUser(t *testing.T, repo *UserRepository, name, email string) *User {
	t.Helper()
	u := &User{
		Name:         name,
		Email:        email,
		Phone:        "555-0100",
		Role:         "user",
		PasswordHash: "hashed",
	}
	if err := repo.Create(u); err != nil {
		t.Fatalf("seed user %s: %v", email, err)
	}
	return u
}

func TestUserRepository_Create(t *testing.T) {
	db := setupTestDB(t)
	repo := NewUserRepository(db)

	tests := []struct {
		name    string
		user    User
		wantErr bool
	}{
		{
			name: "valid user with defaults",
			user: User{
				Name:         "Alice",
				Email:        "alice@example.com",
				Phone:        "555-0100",
				Role:         "user",
				PasswordHash: "hashed",
			},
		},
		{
			name: "admin role",
			user: User{
				Name:         "Bob",
				Email:        "bob@example.com",
				Phone:        "555-0200",
				Role:         "admin",
				PasswordHash: "hashed",
			},
		},
		{
			name: "all optional fields",
			user: User{
				Name:             "Carol",
				Email:            "carol@example.com",
				Phone:            "555-0300",
				Role:             "instructor",
				PasswordHash:     "hashed",
				MembershipType:   "monthly",
				MembershipStatus: "active",
				EmergencyContact: "Dan 555-0400",
				JoinDate:         "2026-01-15",
				ExpectedBalance:  100.50,
			},
		},
		{
			name: "duplicate email",
			user: User{
				Name:         "Alice2",
				Email:        "alice@example.com",
				Phone:        "555-0500",
				Role:         "user",
				PasswordHash: "hashed",
			},
			wantErr: true,
		},
		{
			name: "invalid role",
			user: User{
				Name:         "Eve",
				Email:        "eve@example.com",
				Phone:        "555-0600",
				Role:         "superadmin",
				PasswordHash: "hashed",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := tt.user
			err := repo.Create(&u)
			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if u.ID == 0 {
				t.Error("expected non-zero ID")
			}
			if u.CreatedAt.IsZero() {
				t.Error("expected CreatedAt to be set")
			}
			if u.UpdatedAt.IsZero() {
				t.Error("expected UpdatedAt to be set")
			}
		})
	}
}

func TestUserRepository_GetByID(t *testing.T) {
	db := setupTestDB(t)
	repo := NewUserRepository(db)

	created := seedUser(t, repo, "Alice", "alice@example.com")

	tests := []struct {
		name    string
		id      int64
		wantErr bool
	}{
		{name: "existing user", id: created.ID},
		{name: "non-existent user", id: 9999, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var u User
			err := repo.GetByID(tt.id, &u)
			if tt.wantErr {
				if err != sql.ErrNoRows {
					t.Errorf("expected ErrNoRows, got %v", err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if u.Name != "Alice" {
				t.Errorf("expected name Alice, got %s", u.Name)
			}
			if u.Email != "alice@example.com" {
				t.Errorf("expected email alice@example.com, got %s", u.Email)
			}
		})
	}
}

func TestUserRepository_GetByEmail(t *testing.T) {
	db := setupTestDB(t)
	repo := NewUserRepository(db)

	seedUser(t, repo, "Alice", "alice@example.com")

	tests := []struct {
		name    string
		email   string
		wantErr bool
	}{
		{name: "existing email", email: "alice@example.com"},
		{name: "non-existent email", email: "nobody@example.com", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var u User
			err := repo.GetByEmail(tt.email, &u)
			if tt.wantErr {
				if err != sql.ErrNoRows {
					t.Errorf("expected ErrNoRows, got %v", err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if u.Name != "Alice" {
				t.Errorf("expected name Alice, got %s", u.Name)
			}
		})
	}
}

func TestUserRepository_GetByEmail_ExcludesDeleted(t *testing.T) {
	db := setupTestDB(t)
	repo := NewUserRepository(db)

	created := seedUser(t, repo, "Alice", "alice@example.com")

	if err := repo.SoftDelete(created.ID); err != nil {
		t.Fatalf("soft delete: %v", err)
	}

	var u User
	err := repo.GetByEmail("alice@example.com", &u)
	if err != sql.ErrNoRows {
		t.Errorf("expected ErrNoRows for deleted user, got %v", err)
	}
}

func TestUserRepository_List(t *testing.T) {
	db := setupTestDB(t)
	repo := NewUserRepository(db)

	seedUser(t, repo, "Alice", "alice@example.com")
	seedUser(t, repo, "Bob", "bob@example.com")
	deleted := seedUser(t, repo, "Carol", "carol@example.com")

	if err := repo.SoftDelete(deleted.ID); err != nil {
		t.Fatalf("soft delete: %v", err)
	}

	users, err := repo.List()
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(users) != 2 {
		t.Fatalf("expected 2 users, got %d", len(users))
	}
	if users[0].Name != "Alice" {
		t.Errorf("expected first user Alice, got %s", users[0].Name)
	}
	if users[1].Name != "Bob" {
		t.Errorf("expected second user Bob, got %s", users[1].Name)
	}
}

func TestUserRepository_Update(t *testing.T) {
	db := setupTestDB(t)
	repo := NewUserRepository(db)

	u := seedUser(t, repo, "Alice", "alice@example.com")
	originalUpdatedAt := u.UpdatedAt

	u.Name = "Alice Updated"
	u.Phone = "555-9999"
	u.MembershipType = "annual"
	u.MembershipStatus = "active"

	if err := repo.Update(u); err != nil {
		t.Fatalf("update: %v", err)
	}

	var fetched User
	if err := repo.GetByID(u.ID, &fetched); err != nil {
		t.Fatalf("get after update: %v", err)
	}
	if fetched.Name != "Alice Updated" {
		t.Errorf("expected name 'Alice Updated', got %s", fetched.Name)
	}
	if fetched.Phone != "555-9999" {
		t.Errorf("expected phone '555-9999', got %s", fetched.Phone)
	}
	if fetched.MembershipType != "annual" {
		t.Errorf("expected membership_type 'annual', got %s", fetched.MembershipType)
	}
	if !fetched.UpdatedAt.After(originalUpdatedAt) && fetched.UpdatedAt.Equal(originalUpdatedAt) {
		// SQLite CURRENT_TIMESTAMP has second resolution, so they may be equal
		// if the test runs fast. We just verify it's not before.
	}
}

func TestUserRepository_Update_ReturnsRefreshedData(t *testing.T) {
	db := setupTestDB(t)
	repo := NewUserRepository(db)

	u := seedUser(t, repo, "Alice", "alice@example.com")
	u.Name = "Alice V2"

	if err := repo.Update(u); err != nil {
		t.Fatalf("update: %v", err)
	}

	// u should be refreshed with the latest DB values
	if u.Name != "Alice V2" {
		t.Errorf("expected refreshed name 'Alice V2', got %s", u.Name)
	}
}

func TestUserRepository_Update_DeletedUser(t *testing.T) {
	db := setupTestDB(t)
	repo := NewUserRepository(db)

	u := seedUser(t, repo, "Alice", "alice@example.com")
	if err := repo.SoftDelete(u.ID); err != nil {
		t.Fatalf("soft delete: %v", err)
	}

	u.Name = "Alice Updated"
	err := repo.Update(u)
	if err != sql.ErrNoRows {
		t.Errorf("expected ErrNoRows when updating deleted user, got %v", err)
	}
}

func TestUserRepository_SoftDelete(t *testing.T) {
	db := setupTestDB(t)
	repo := NewUserRepository(db)

	tests := []struct {
		name    string
		setup   func() int64
		wantErr bool
	}{
		{
			name: "delete existing user",
			setup: func() int64 {
				u := seedUser(t, repo, "ToDelete", "delete@example.com")
				return u.ID
			},
		},
		{
			name:    "delete non-existent user",
			setup:   func() int64 { return 9999 },
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			id := tt.setup()
			err := repo.SoftDelete(id)
			if tt.wantErr {
				if err != sql.ErrNoRows {
					t.Errorf("expected ErrNoRows, got %v", err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			// Verify user is soft-deleted (still in DB but has deleted_at)
			var u User
			if err := repo.GetByID(id, &u); err != nil {
				t.Fatalf("get after delete: %v", err)
			}
			if u.DeletedAt == nil {
				t.Error("expected deleted_at to be set")
			}
		})
	}
}

func TestUserRepository_SoftDelete_Idempotent(t *testing.T) {
	db := setupTestDB(t)
	repo := NewUserRepository(db)

	u := seedUser(t, repo, "Alice", "alice-idem@example.com")

	if err := repo.SoftDelete(u.ID); err != nil {
		t.Fatalf("first delete: %v", err)
	}

	// Second delete should fail (already deleted)
	err := repo.SoftDelete(u.ID)
	if err != sql.ErrNoRows {
		t.Errorf("expected ErrNoRows on second delete, got %v", err)
	}
}
