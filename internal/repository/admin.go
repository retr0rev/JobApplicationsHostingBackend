package repository

import (
	"database/sql"
	"fmt"
	"jobapps/internal/models"
)

type AdminRepo struct {
	db *sql.DB
}

func NewAdminRepo(db *sql.DB) *AdminRepo {
	return &AdminRepo{db: db}
}

func (r *AdminRepo) FindByEmail(email string) (*models.Admin, error) {
	row := r.db.QueryRow(
		"SELECT id, email, password FROM ADMINS WHERE email = ?",
		email,
	)

	a := &models.Admin{}
	if err := row.Scan(&a.ID, &a.Email, &a.Password); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("query admin: %w", err)
	}
	return a, nil
}

func (r *AdminRepo) FindByID(id int64) (*models.Admin, error) {
	row := r.db.QueryRow(
		"SELECT id, email, password FROM ADMINS WHERE id = ?",
		id,
	)

	a := &models.Admin{}
	if err := row.Scan(&a.ID, &a.Email, &a.Password); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("query admin: %w", err)
	}
	return a, nil
}
