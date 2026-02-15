package models

import (
	"database/sql"
	"testing"
)

// seedPaymentDeps creates the prerequisite data for payment tests:
// an admin user and a student user. Returns (adminID, studentID).
func seedPaymentDeps(t *testing.T, db *sql.DB) (int64, int64) {
	t.Helper()

	res, err := db.Exec(`
		INSERT INTO users (name, email, phone, role, password_hash)
		VALUES ('Admin', 'admin@example.com', '555-0001', 'admin', 'hashed')`)
	if err != nil {
		t.Fatalf("seed admin: %v", err)
	}
	adminID, _ := res.LastInsertId()

	res, err = db.Exec(`
		INSERT INTO users (name, email, phone, role, password_hash)
		VALUES ('Student', 'student@example.com', '555-0002', 'user', 'hashed')`)
	if err != nil {
		t.Fatalf("seed student: %v", err)
	}
	studentID, _ := res.LastInsertId()

	return adminID, studentID
}

func TestPaymentRepository_Create(t *testing.T) {
	db := setupTestDB(t)
	repo := NewPaymentRepository(db)
	adminID, studentID := seedPaymentDeps(t, db)

	tests := []struct {
		name    string
		payment Payment
		wantErr bool
	}{
		{
			name: "valid payment",
			payment: Payment{
				UserID:     studentID,
				Amount:     50.00,
				Date:       "2026-02-01",
				Note:       "Monthly dues",
				RecordedBy: adminID,
			},
		},
		{
			name: "payment with no note",
			payment: Payment{
				UserID:     studentID,
				Amount:     25.00,
				Date:       "2026-02-15",
				RecordedBy: adminID,
			},
		},
		{
			name: "non-existent user",
			payment: Payment{
				UserID:     9999,
				Amount:     50.00,
				Date:       "2026-02-01",
				RecordedBy: adminID,
			},
			wantErr: true,
		},
		{
			name: "non-existent recorder",
			payment: Payment{
				UserID:     studentID,
				Amount:     50.00,
				Date:       "2026-02-01",
				RecordedBy: 9999,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := tt.payment
			err := repo.Create(&p)
			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if p.ID == 0 {
				t.Error("expected non-zero ID")
			}
			if p.CreatedAt.IsZero() {
				t.Error("expected CreatedAt to be set")
			}
			if p.UpdatedAt.IsZero() {
				t.Error("expected UpdatedAt to be set")
			}
		})
	}
}

func TestPaymentRepository_GetByID(t *testing.T) {
	db := setupTestDB(t)
	repo := NewPaymentRepository(db)
	adminID, studentID := seedPaymentDeps(t, db)

	created := &Payment{
		UserID:     studentID,
		Amount:     75.00,
		Date:       "2026-02-10",
		Note:       "Drop-in",
		RecordedBy: adminID,
	}
	if err := repo.Create(created); err != nil {
		t.Fatalf("create: %v", err)
	}

	tests := []struct {
		name    string
		id      int64
		wantErr bool
	}{
		{name: "existing payment", id: created.ID},
		{name: "non-existent payment", id: 9999, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var p Payment
			err := repo.GetByID(tt.id, &p)
			if tt.wantErr {
				if err != sql.ErrNoRows {
					t.Errorf("expected ErrNoRows, got %v", err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if p.Amount != 75.00 {
				t.Errorf("expected amount 75.00, got %f", p.Amount)
			}
			if p.Note != "Drop-in" {
				t.Errorf("expected note 'Drop-in', got '%s'", p.Note)
			}
			if p.RecordedBy != adminID {
				t.Errorf("expected recorded_by %d, got %d", adminID, p.RecordedBy)
			}
		})
	}
}

func TestPaymentRepository_ListByUser(t *testing.T) {
	db := setupTestDB(t)
	repo := NewPaymentRepository(db)
	adminID, studentID := seedPaymentDeps(t, db)

	p1 := &Payment{UserID: studentID, Amount: 50.00, Date: "2026-01-01", RecordedBy: adminID}
	p2 := &Payment{UserID: studentID, Amount: 75.00, Date: "2026-02-01", RecordedBy: adminID}
	if err := repo.Create(p1); err != nil {
		t.Fatalf("create p1: %v", err)
	}
	if err := repo.Create(p2); err != nil {
		t.Fatalf("create p2: %v", err)
	}

	payments, err := repo.ListByUser(studentID)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(payments) != 2 {
		t.Fatalf("expected 2 payments, got %d", len(payments))
	}
	// Should be ordered by date DESC.
	if payments[0].Date != "2026-02-01" {
		t.Errorf("expected first payment date '2026-02-01', got '%s'", payments[0].Date)
	}
	if payments[1].Date != "2026-01-01" {
		t.Errorf("expected second payment date '2026-01-01', got '%s'", payments[1].Date)
	}

	// Non-existent user should return empty.
	empty, err := repo.ListByUser(9999)
	if err != nil {
		t.Fatalf("list empty: %v", err)
	}
	if len(empty) != 0 {
		t.Errorf("expected 0 payments for non-existent user, got %d", len(empty))
	}
}

func TestPaymentRepository_ListByUser_OrdersDescending(t *testing.T) {
	db := setupTestDB(t)
	repo := NewPaymentRepository(db)
	adminID, studentID := seedPaymentDeps(t, db)

	// Create payments on the same date — should order by ID DESC.
	p1 := &Payment{UserID: studentID, Amount: 10.00, Date: "2026-03-01", Note: "first", RecordedBy: adminID}
	p2 := &Payment{UserID: studentID, Amount: 20.00, Date: "2026-03-01", Note: "second", RecordedBy: adminID}
	if err := repo.Create(p1); err != nil {
		t.Fatalf("create p1: %v", err)
	}
	if err := repo.Create(p2); err != nil {
		t.Fatalf("create p2: %v", err)
	}

	payments, err := repo.ListByUser(studentID)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(payments) != 2 {
		t.Fatalf("expected 2 payments, got %d", len(payments))
	}
	// Same date, so ordered by ID DESC — second created first.
	if payments[0].Note != "second" {
		t.Errorf("expected first result 'second', got '%s'", payments[0].Note)
	}
	if payments[1].Note != "first" {
		t.Errorf("expected second result 'first', got '%s'", payments[1].Note)
	}
}
