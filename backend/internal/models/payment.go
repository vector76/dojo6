package models

import (
	"database/sql"
	"time"
)

type Payment struct {
	ID         int64
	UserID     int64
	Amount     float64
	Date       string
	Note       string
	RecordedBy int64
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

type PaymentRepository struct {
	db *sql.DB
}

func NewPaymentRepository(db *sql.DB) *PaymentRepository {
	return &PaymentRepository{db: db}
}

// Create records a payment.
func (r *PaymentRepository) Create(p *Payment) error {
	result, err := r.db.Exec(`
		INSERT INTO payments (user_id, amount, date, note, recorded_by)
		VALUES (?, ?, ?, ?, ?)`,
		p.UserID, p.Amount, p.Date, p.Note, p.RecordedBy,
	)
	if err != nil {
		return err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	p.ID = id
	return r.GetByID(p.ID, p)
}

// GetByID fetches a single payment record.
func (r *PaymentRepository) GetByID(id int64, p *Payment) error {
	return r.db.QueryRow(`
		SELECT id, user_id, amount, date, note, recorded_by, created_at, updated_at
		FROM payments WHERE id = ?`, id,
	).Scan(&p.ID, &p.UserID, &p.Amount, &p.Date, &p.Note, &p.RecordedBy, &p.CreatedAt, &p.UpdatedAt)
}

// ListByUser returns all payments for a given user, ordered by date descending.
func (r *PaymentRepository) ListByUser(userID int64) ([]Payment, error) {
	rows, err := r.db.Query(`
		SELECT id, user_id, amount, date, note, recorded_by, created_at, updated_at
		FROM payments WHERE user_id = ?
		ORDER BY date DESC, id DESC`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var payments []Payment
	for rows.Next() {
		var p Payment
		if err := rows.Scan(&p.ID, &p.UserID, &p.Amount, &p.Date, &p.Note, &p.RecordedBy, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, err
		}
		payments = append(payments, p)
	}
	return payments, rows.Err()
}
