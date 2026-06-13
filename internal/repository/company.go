package repository

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"jobapps/internal/models"
)

type ClientRepo struct {
	db *sql.DB
}

func NewClientRepo(db *sql.DB) *ClientRepo {
	return &ClientRepo{db: db}
}

func HashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

func scanClient(row scanner) (*models.Client, error) {
	c := &models.Client{}
	err := row.Scan(
		&c.ID, &c.Email, &c.Password, &c.PhoneNumber,
		&c.Verified, &c.VerifyTokenHash, &c.VerifyTokenExpiry,
		&c.ResetTokenHash, &c.ResetTokenExpiry,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("scan client: %w", err)
	}
	return c, nil
}

type scanner interface {
	Scan(dest ...interface{}) error
}

func (r *ClientRepo) Create(email, password, phone, verifyToken, verifyExpiry string) (*models.Client, error) {
	var phoneArg interface{}
	var phonePtr *string
	if phone != "" {
		phoneArg = phone
		phonePtr = &phone
	} else {
		phoneArg = nil
	}

	tokenHash := HashToken(verifyToken)

	res, err := r.db.Exec(
		`INSERT INTO CLIENTS (c_email, c_password, phone_number, verified, verify_token_hash, verify_token_expiry)
		 VALUES (?, ?, ?, 0, ?, ?)`,
		email, password, phoneArg, tokenHash, verifyExpiry,
	)
	if err != nil {
		return nil, fmt.Errorf("insert client: %w", err)
	}

	id, _ := res.LastInsertId()

	return &models.Client{
		ID:                id,
		Email:             email,
		Password:          password,
		PhoneNumber:       phonePtr,
		Verified:          0,
		VerifyTokenHash:   &tokenHash,
		VerifyTokenExpiry: &verifyExpiry,
	}, nil
}

func (r *ClientRepo) FindByEmail(email string) (*models.Client, error) {
	row := r.db.QueryRow(
		`SELECT id, c_email, c_password, phone_number,
		        verified, verify_token_hash, verify_token_expiry,
		        reset_token_hash, reset_token_expiry
		 FROM CLIENTS WHERE c_email = ?`,
		email,
	)
	return scanClient(row)
}

func (r *ClientRepo) SetVerified(id int64) error {
	_, err := r.db.Exec(
		"UPDATE CLIENTS SET verified = 1, verify_token_hash = NULL, verify_token_expiry = NULL WHERE id = ?",
		id,
	)
	return err
}

func (r *ClientRepo) FindByVerifyToken(token string) (*models.Client, error) {
	tokenHash := HashToken(token)
	row := r.db.QueryRow(
		`SELECT id, c_email, c_password, phone_number,
		        verified, verify_token_hash, verify_token_expiry,
		        reset_token_hash, reset_token_expiry
		 FROM CLIENTS WHERE verify_token_hash = ?`,
		tokenHash,
	)
	return scanClient(row)
}

func (r *ClientRepo) SetResetToken(email, token, expiry string) error {
	tokenHash := HashToken(token)
	_, err := r.db.Exec(
		"UPDATE CLIENTS SET reset_token_hash = ?, reset_token_expiry = ? WHERE c_email = ?",
		tokenHash, expiry, email,
	)
	return err
}

func (r *ClientRepo) FindByResetToken(token string) (*models.Client, error) {
	tokenHash := HashToken(token)
	row := r.db.QueryRow(
		`SELECT id, c_email, c_password, phone_number,
		        verified, verify_token_hash, verify_token_expiry,
		        reset_token_hash, reset_token_expiry
		 FROM CLIENTS WHERE reset_token_hash = ?`,
		tokenHash,
	)
	return scanClient(row)
}

func (r *ClientRepo) UpdatePassword(id int64, hash string) error {
	_, err := r.db.Exec(
		"UPDATE CLIENTS SET c_password = ?, reset_token_hash = NULL, reset_token_expiry = NULL WHERE id = ?",
		hash, id,
	)
	return err
}
