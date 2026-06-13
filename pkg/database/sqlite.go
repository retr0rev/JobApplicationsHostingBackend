package database

import (
	"database/sql"
	"fmt"
	"path/filepath"

	_ "modernc.org/sqlite"
)

func NewDB(dbPath string) (*sql.DB, error) {
	absPath, err := filepath.Abs(dbPath)
	if err != nil {
		return nil, fmt.Errorf("resolve db path: %w", err)
	}

	db, err := sql.Open("sqlite", absPath)
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("ping database: %w", err)
	}

	db.SetMaxOpenConns(1)

	if err := migrate(db); err != nil {
		return nil, fmt.Errorf("migrate: %w", err)
	}

	return db, nil
}

func migrate(db *sql.DB) error {
	schema := []string{
		`CREATE TABLE IF NOT EXISTS ADMINS (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			email TEXT NOT NULL UNIQUE,
			password TEXT NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS CLIENTS (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			c_email TEXT NOT NULL UNIQUE,
			c_password TEXT NOT NULL,
			phone_number TEXT,
			verified INTEGER NOT NULL DEFAULT 0,
			verify_token_hash TEXT,
			verify_token_expiry DATETIME,
			reset_token_hash TEXT,
			reset_token_expiry DATETIME
		)`,
		`CREATE TABLE IF NOT EXISTS JOBSAPPS (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			client_id INTEGER NOT NULL,
			jobtitle TEXT NOT NULL,
			description TEXT NOT NULL,
			status TEXT NOT NULL DEFAULT 'pending',
			category TEXT NOT NULL DEFAULT '',
			location TEXT NOT NULL DEFAULT '',
			FOREIGN KEY (client_id) REFERENCES CLIENTS(id)
		)`,
	}

	for _, s := range schema {
		if _, err := db.Exec(s); err != nil {
			return err
		}
	}
	return nil
}
