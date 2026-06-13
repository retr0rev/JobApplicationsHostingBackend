package handlers

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"jobapps/internal/email"
	"jobapps/internal/middleware"
	"jobapps/internal/models"
	"jobapps/internal/repository"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type CompanyHandler struct {
	clientRepo  *repository.ClientRepo
	emailSender email.Sender
}

func NewCompanyHandler(clientRepo *repository.ClientRepo, es email.Sender) *CompanyHandler {
	return &CompanyHandler{clientRepo: clientRepo, emailSender: es}
}

func generateTokenHex() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func baseURL() string {
	u := os.Getenv("APP_URL")
	if u == "" {
		u = "http://localhost:8080"
	}
	return u
}

func (h *CompanyHandler) Signup(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 4096)

	var req models.SignupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request body"}`, http.StatusBadRequest)
		return
	}

	req.Email = middleware.NormalizeEmail(req.Email)
	req.Phone = strings.TrimSpace(req.Phone)

	if !middleware.ValidateEmail(req.Email) {
		http.Error(w, `{"error":"invalid email format"}`, http.StatusBadRequest)
		return
	}

	if errMsg := middleware.ValidatePassword(req.Password); errMsg != "" {
		http.Error(w, `{"error":"`+errMsg+`"}`, http.StatusBadRequest)
		return
	}

	existing, _ := h.clientRepo.FindByEmail(req.Email)
	if existing != nil {
		http.Error(w, `{"error":"email already registered"}`, http.StatusConflict)
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, `{"error":"failed to create account"}`, http.StatusInternalServerError)
		return
	}

	verifyToken, err := generateTokenHex()
	if err != nil {
		http.Error(w, `{"error":"failed to create account"}`, http.StatusInternalServerError)
		return
	}

	expiry := time.Now().Add(24 * time.Hour).UTC().Format(time.RFC3339)

	client, err := h.clientRepo.Create(req.Email, string(hash), req.Phone, verifyToken, expiry)
	if err != nil {
		http.Error(w, `{"error":"failed to create account"}`, http.StatusInternalServerError)
		return
	}

	verifyURL := fmt.Sprintf("%s/api/auth/verify?token=%s", baseURL(), verifyToken)
	subj, body := email.BuildVerifyEmail(verifyURL)
	h.emailSender.Send(client.Email, subj, body)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(models.MessageResponse{
		Message: "Account created. Check your email for the verification link.",
	})
}

func (h *CompanyHandler) Login(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 4096)

	var req models.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request body"}`, http.StatusBadRequest)
		return
	}

	req.Email = middleware.NormalizeEmail(req.Email)

	if req.Email == "" || req.Password == "" {
		http.Error(w, `{"error":"email and password required"}`, http.StatusBadRequest)
		return
	}

	client, err := h.clientRepo.FindByEmail(req.Email)
	if err != nil {
		http.Error(w, `{"error":"server error"}`, http.StatusInternalServerError)
		return
	}
	if client == nil {
		http.Error(w, `{"error":"invalid email or password"}`, http.StatusUnauthorized)
		return
	}

	if client.Verified == 0 {
		http.Error(w, `{"error":"please verify your email before logging in"}`, http.StatusForbidden)
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(client.Password), []byte(req.Password)); err != nil {
		http.Error(w, `{"error":"invalid email or password"}`, http.StatusUnauthorized)
		return
	}

	token, err := h.generateToken(client.ID)
	if err != nil {
		http.Error(w, `{"error":"server error"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(models.AuthResponse{Token: token, Email: client.Email})
}

func (h *CompanyHandler) generateToken(clientID int64) (string, error) {
	claims := jwt.MapClaims{
		"client_id": clientID,
		"role":      "client",
	}
	return middleware.GenerateToken(claims)
}
