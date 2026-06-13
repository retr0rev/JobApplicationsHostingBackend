package repository

import (
	"database/sql"
	"fmt"
	"jobapps/internal/models"
)

type JobRepo struct {
	db *sql.DB
}

func NewJobRepo(db *sql.DB) *JobRepo {
	return &JobRepo{db: db}
}

func (r *JobRepo) Create(clientID int64, title, description, category, location string) (*models.JobApp, error) {
	res, err := r.db.Exec(
		"INSERT INTO JOBSAPPS (client_id, jobtitle, description, status, category, location) VALUES (?, ?, ?, ?, ?, ?)",
		clientID, title, description, models.StatusPending, category, location,
	)
	if err != nil {
		return nil, fmt.Errorf("insert job: %w", err)
	}

	id, _ := res.LastInsertId()
	return &models.JobApp{
		ID:          id,
		ClientID:    clientID,
		JobTitle:    title,
		Description: description,
		Status:      models.StatusPending,
		Category:    category,
		Location:    location,
	}, nil
}

func (r *JobRepo) ListByClient(clientID int64) ([]models.JobApp, error) {
	rows, err := r.db.Query(
		"SELECT id, client_id, jobtitle, description, status, category, location FROM JOBSAPPS WHERE client_id = ? ORDER BY id DESC",
		clientID,
	)
	if err != nil {
		return nil, fmt.Errorf("query jobs: %w", err)
	}
	defer rows.Close()

	var jobs []models.JobApp
	for rows.Next() {
		var j models.JobApp
		if err := rows.Scan(&j.ID, &j.ClientID, &j.JobTitle, &j.Description, &j.Status, &j.Category, &j.Location); err != nil {
			return nil, fmt.Errorf("scan job: %w", err)
		}
		jobs = append(jobs, j)
	}
	return jobs, rows.Err()
}

func (r *JobRepo) ListAll() ([]models.JobApp, error) {
	rows, err := r.db.Query(
		"SELECT id, client_id, jobtitle, description, status, category, location FROM JOBSAPPS ORDER BY id DESC",
	)
	if err != nil {
		return nil, fmt.Errorf("query all jobs: %w", err)
	}
	defer rows.Close()

	var jobs []models.JobApp
	for rows.Next() {
		var j models.JobApp
		if err := rows.Scan(&j.ID, &j.ClientID, &j.JobTitle, &j.Description, &j.Status, &j.Category, &j.Location); err != nil {
			return nil, fmt.Errorf("scan job: %w", err)
		}
		jobs = append(jobs, j)
	}
	return jobs, rows.Err()
}

func (r *JobRepo) FindByID(id, clientID int64) (*models.JobApp, error) {
	row := r.db.QueryRow(
		"SELECT id, client_id, jobtitle, description, status, category, location FROM JOBSAPPS WHERE id = ? AND client_id = ?",
		id, clientID,
	)

	var j models.JobApp
	if err := row.Scan(&j.ID, &j.ClientID, &j.JobTitle, &j.Description, &j.Status, &j.Category, &j.Location); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("query job: %w", err)
	}
	return &j, nil
}

func (r *JobRepo) FindByIDAdmin(id int64) (*models.JobApp, error) {
	row := r.db.QueryRow(
		"SELECT id, client_id, jobtitle, description, status, category, location FROM JOBSAPPS WHERE id = ?",
		id,
	)

	var j models.JobApp
	if err := row.Scan(&j.ID, &j.ClientID, &j.JobTitle, &j.Description, &j.Status, &j.Category, &j.Location); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("query job: %w", err)
	}
	return &j, nil
}

func (r *JobRepo) UpdateStatus(id int64, status string) error {
	res, err := r.db.Exec(
		"UPDATE JOBSAPPS SET status = ? WHERE id = ?",
		status, id,
	)
	if err != nil {
		return fmt.Errorf("update job status: %w", err)
	}

	affected, _ := res.RowsAffected()
	if affected == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (r *JobRepo) Update(id, clientID int64, title, description, category, location string) (*models.JobApp, error) {
	_, err := r.db.Exec(
		"UPDATE JOBSAPPS SET jobtitle = ?, description = ?, category = ?, location = ? WHERE id = ? AND client_id = ?",
		title, description, category, location, id, clientID,
	)
	if err != nil {
		return nil, fmt.Errorf("update job: %w", err)
	}
	return r.FindByID(id, clientID)
}

func (r *JobRepo) Delete(id, clientID int64) error {
	res, err := r.db.Exec(
		"DELETE FROM JOBSAPPS WHERE id = ? AND client_id = ?",
		id, clientID,
	)
	if err != nil {
		return fmt.Errorf("delete job: %w", err)
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("rows affected: %w", err)
	}
	if affected == 0 {
		return sql.ErrNoRows
	}
	return nil
}
