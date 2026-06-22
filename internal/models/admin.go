package models

type Admin struct {
	ID        int64  `json:"id"`
	Email     string `json:"email"`
	Password  string `json:"-"`
	AdminRole string `json:"admin_role"`
}

const (
	AdminRoleSuperAdmin = "super_admin"
	AdminRoleModerator  = "moderator"
)

type AdminLoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type AdminLoginResponse struct {
	Token     string `json:"token"`
	Email     string `json:"email"`
	AdminRole string `json:"admin_role"`
}

type CreateAdminRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Role     string `json:"role"` // "super_admin" | "moderator"
}
