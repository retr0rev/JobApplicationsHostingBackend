package models

type Client struct {
	ID                 int64   `json:"id"`
	Email              string  `json:"email"`
	Password           string  `json:"-"`
	PhoneNumber        *string `json:"phone_number,omitempty"`
	Verified           int64   `json:"verified"`
	VerifyTokenHash    *string `json:"-"`
	VerifyTokenExpiry  *string `json:"-"`
	ResetTokenHash     *string `json:"-"`
	ResetTokenExpiry   *string `json:"-"`
}

type SignupRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Phone    string `json:"phone,omitempty"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type AuthResponse struct {
	Token string `json:"token"`
	Email string `json:"email"`
}

type ForgotPasswordRequest struct {
	Email string `json:"email"`
}

type ResetPasswordRequest struct {
	Token       string `json:"token"`
	NewPassword string `json:"new_password"`
}

type MessageResponse struct {
	Message string `json:"message"`
}
