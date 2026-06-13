package models

type Admin struct {
	ID       int64  `json:"id"`
	Email    string `json:"email"`
	Password string `json:"-"`
}

type AdminLoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type AdminLoginResponse struct {
	Token string `json:"token"`
	Email string `json:"email"`
}
