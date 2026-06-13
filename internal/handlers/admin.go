package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"jobapps/internal/middleware"
	"jobapps/internal/models"
	"jobapps/internal/repository"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type AdminHandler struct {
	adminRepo *repository.AdminRepo
	jobRepo   *repository.JobRepo
}

func NewAdminHandler(adminRepo *repository.AdminRepo, jobRepo *repository.JobRepo) *AdminHandler {
	return &AdminHandler{adminRepo: adminRepo, jobRepo: jobRepo}
}

func (h *AdminHandler) Login(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 4096)

	var req models.AdminLoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request body"}`, http.StatusBadRequest)
		return
	}

	req.Email = middleware.NormalizeEmail(req.Email)

	if req.Email == "" || req.Password == "" {
		http.Error(w, `{"error":"email and password required"}`, http.StatusBadRequest)
		return
	}

	admin, err := h.adminRepo.FindByEmail(req.Email)
	if err != nil {
		http.Error(w, `{"error":"server error"}`, http.StatusInternalServerError)
		return
	}
	if admin == nil {
		http.Error(w, `{"error":"invalid email or password"}`, http.StatusUnauthorized)
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(admin.Password), []byte(req.Password)); err != nil {
		http.Error(w, `{"error":"invalid email or password"}`, http.StatusUnauthorized)
		return
	}

	claims := jwt.MapClaims{
		"admin_id": admin.ID,
		"role":     "admin",
	}
	token, err := middleware.GenerateToken(claims)
	if err != nil {
		http.Error(w, `{"error":"server error"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(models.AdminLoginResponse{Token: token, Email: admin.Email})
}

func (h *AdminHandler) ListJobs(w http.ResponseWriter, r *http.Request) {
	jobs, err := h.jobRepo.ListAll()
	if err != nil {
		http.Error(w, `{"error":"failed to fetch jobs"}`, http.StatusInternalServerError)
		return
	}

	if jobs == nil {
		jobs = []models.JobApp{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(jobs)
}

func (h *AdminHandler) UpdateStatus(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 1024)

	jobID, err := parseIDParam(r)
	if err != nil {
		http.Error(w, `{"error":"invalid job id"}`, http.StatusBadRequest)
		return
	}

	var req models.UpdateJobStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request body"}`, http.StatusBadRequest)
		return
	}

	if req.Status != models.StatusApproved && req.Status != models.StatusRejected {
		http.Error(w, `{"error":"status must be 'approved' or 'rejected'"}`, http.StatusBadRequest)
		return
	}

	if err := h.jobRepo.UpdateStatus(jobID, req.Status); err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, `{"error":"job not found"}`, http.StatusNotFound)
			return
		}
		http.Error(w, `{"error":"failed to update job status"}`, http.StatusInternalServerError)
		return
	}

	job, _ := h.jobRepo.FindByIDAdmin(jobID)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(job)
}
