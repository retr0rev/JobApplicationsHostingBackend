package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"jobapps/internal/email"
	"jobapps/internal/middleware"
	"jobapps/internal/models"
	"jobapps/internal/repository"

	"golang.org/x/crypto/bcrypt"
)

type VerifyHandler struct {
	clientRepo  *repository.ClientRepo
	emailSender email.Sender
}

func NewVerifyHandler(clientRepo *repository.ClientRepo, es email.Sender) *VerifyHandler {
	return &VerifyHandler{clientRepo: clientRepo, emailSender: es}
}

func (h *VerifyHandler) VerifyEmail(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")
	if token == "" {
		http.Error(w, `{"error":"missing token"}`, http.StatusBadRequest)
		return
	}

	client, err := h.clientRepo.FindByVerifyToken(token)
	if err != nil {
		http.Error(w, `{"error":"server error"}`, http.StatusInternalServerError)
		return
	}
	if client == nil {
		http.Error(w, `{"error":"invalid or expired token"}`, http.StatusNotFound)
		return
	}

	if client.VerifyTokenExpiry != nil {
		expiry, err := time.Parse(time.RFC3339, *client.VerifyTokenExpiry)
		if err == nil && time.Now().UTC().After(expiry) {
			http.Error(w, `{"error":"verification token has expired"}`, http.StatusGone)
			return
		}
	}

	if err := h.clientRepo.SetVerified(client.ID); err != nil {
		http.Error(w, `{"error":"failed to verify email"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(models.MessageResponse{
		Message: "Email verified successfully. You can now log in.",
	})
}

func (h *VerifyHandler) ForgotPassword(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 4096)

	var req models.ForgotPasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request body"}`, http.StatusBadRequest)
		return
	}

	req.Email = middleware.NormalizeEmail(req.Email)

	if req.Email == "" {
		http.Error(w, `{"error":"email is required"}`, http.StatusBadRequest)
		return
	}

	client, err := h.clientRepo.FindByEmail(req.Email)
	if err != nil {
		http.Error(w, `{"error":"server error"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	genericResponse := models.MessageResponse{
		Message: "If that email is registered, a reset link has been sent.",
	}

	if client == nil {
		equalizeTiming()
		json.NewEncoder(w).Encode(genericResponse)
		return
	}

	resetToken, err := generateTokenHex()
	if err != nil {
		http.Error(w, `{"error":"server error"}`, http.StatusInternalServerError)
		return
	}

	expiry := time.Now().Add(1 * time.Hour).UTC().Format(time.RFC3339)

	if err := h.clientRepo.SetResetToken(req.Email, resetToken, expiry); err != nil {
		http.Error(w, `{"error":"server error"}`, http.StatusInternalServerError)
		return
	}

	resetURL := fmt.Sprintf("%s/api/auth/reset-password?token=%s", baseURL(), resetToken)
	subj, body := email.BuildResetEmail(resetURL)
	h.emailSender.Send(client.Email, subj, body)

	json.NewEncoder(w).Encode(genericResponse)
}

func equalizeTiming() {
	dummy := []byte("timing-equalization-placeholder-hash-input")
	_, _ = bcrypt.GenerateFromPassword(dummy, bcrypt.DefaultCost)
}

func (h *VerifyHandler) ResetPassword(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 4096)

	var req models.ResetPasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request body"}`, http.StatusBadRequest)
		return
	}

	if req.Token == "" || req.NewPassword == "" {
		http.Error(w, `{"error":"token and new_password are required"}`, http.StatusBadRequest)
		return
	}

	if errMsg := middleware.ValidatePassword(req.NewPassword); errMsg != "" {
		http.Error(w, `{"error":"`+errMsg+`"}`, http.StatusBadRequest)
		return
	}

	client, err := h.clientRepo.FindByResetToken(req.Token)
	if err != nil {
		http.Error(w, `{"error":"server error"}`, http.StatusInternalServerError)
		return
	}
	if client == nil {
		http.Error(w, `{"error":"invalid or expired token"}`, http.StatusNotFound)
		return
	}

	if client.ResetTokenExpiry != nil {
		expiry, err := time.Parse(time.RFC3339, *client.ResetTokenExpiry)
		if err == nil && time.Now().UTC().After(expiry) {
			http.Error(w, `{"error":"reset token has expired"}`, http.StatusGone)
			return
		}
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, `{"error":"server error"}`, http.StatusInternalServerError)
		return
	}

	if err := h.clientRepo.UpdatePassword(client.ID, string(hash)); err != nil {
		http.Error(w, `{"error":"failed to reset password"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(models.MessageResponse{
		Message: "Password reset successfully. You can now log in.",
	})
}
