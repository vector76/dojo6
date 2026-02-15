package models

import (
	"database/sql"
	"time"
)

type User struct {
	ID               int64
	Name             string
	Email            string
	Phone            string
	Role             string
	PasswordHash     string
	MembershipType   string
	MembershipStatus string
	EmergencyContact string
	JoinDate         string
	ExpectedBalance  float64
	DeletedAt        *time.Time
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(u *User) error {
	result, err := r.db.Exec(`
		INSERT INTO users (name, email, phone, role, password_hash, membership_type, membership_status, emergency_contact, join_date, expected_balance)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		u.Name, u.Email, u.Phone, u.Role, u.PasswordHash,
		u.MembershipType, u.MembershipStatus, u.EmergencyContact, u.JoinDate, u.ExpectedBalance,
	)
	if err != nil {
		return err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	u.ID = id
	return r.GetByID(u.ID, u)
}

func (r *UserRepository) GetByID(id int64, u *User) error {
	return r.db.QueryRow(`
		SELECT id, name, email, phone, role, password_hash, membership_type, membership_status,
		       emergency_contact, join_date, expected_balance, deleted_at, created_at, updated_at
		FROM users WHERE id = ?`, id,
	).Scan(
		&u.ID, &u.Name, &u.Email, &u.Phone, &u.Role, &u.PasswordHash,
		&u.MembershipType, &u.MembershipStatus, &u.EmergencyContact,
		&u.JoinDate, &u.ExpectedBalance, &u.DeletedAt, &u.CreatedAt, &u.UpdatedAt,
	)
}

func (r *UserRepository) GetByEmail(email string, u *User) error {
	return r.db.QueryRow(`
		SELECT id, name, email, phone, role, password_hash, membership_type, membership_status,
		       emergency_contact, join_date, expected_balance, deleted_at, created_at, updated_at
		FROM users WHERE email = ? AND deleted_at IS NULL`, email,
	).Scan(
		&u.ID, &u.Name, &u.Email, &u.Phone, &u.Role, &u.PasswordHash,
		&u.MembershipType, &u.MembershipStatus, &u.EmergencyContact,
		&u.JoinDate, &u.ExpectedBalance, &u.DeletedAt, &u.CreatedAt, &u.UpdatedAt,
	)
}

// List returns all non-deleted users.
func (r *UserRepository) List() ([]User, error) {
	rows, err := r.db.Query(`
		SELECT id, name, email, phone, role, password_hash, membership_type, membership_status,
		       emergency_contact, join_date, expected_balance, deleted_at, created_at, updated_at
		FROM users WHERE deleted_at IS NULL
		ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []User
	for rows.Next() {
		var u User
		if err := rows.Scan(
			&u.ID, &u.Name, &u.Email, &u.Phone, &u.Role, &u.PasswordHash,
			&u.MembershipType, &u.MembershipStatus, &u.EmergencyContact,
			&u.JoinDate, &u.ExpectedBalance, &u.DeletedAt, &u.CreatedAt, &u.UpdatedAt,
		); err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	return users, rows.Err()
}

func (r *UserRepository) Update(u *User) error {
	result, err := r.db.Exec(`
		UPDATE users SET name = ?, email = ?, phone = ?, role = ?, password_hash = ?,
		       membership_type = ?, membership_status = ?, emergency_contact = ?,
		       join_date = ?, expected_balance = ?, updated_at = CURRENT_TIMESTAMP
		WHERE id = ? AND deleted_at IS NULL`,
		u.Name, u.Email, u.Phone, u.Role, u.PasswordHash,
		u.MembershipType, u.MembershipStatus, u.EmergencyContact,
		u.JoinDate, u.ExpectedBalance, u.ID,
	)
	if err != nil {
		return err
	}
	n, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return sql.ErrNoRows
	}
	return r.GetByID(u.ID, u)
}

func (r *UserRepository) SoftDelete(id int64) error {
	result, err := r.db.Exec(`
		UPDATE users SET deleted_at = CURRENT_TIMESTAMP, updated_at = CURRENT_TIMESTAMP
		WHERE id = ? AND deleted_at IS NULL`, id)
	if err != nil {
		return err
	}
	n, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return sql.ErrNoRows
	}
	return nil
}
