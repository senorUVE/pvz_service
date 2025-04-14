package dto

type AuthRequest struct {
	Email    string `json:"email" db:"email"`
	Password string `json:"password" db:"password_salt"`
}

type AuthResponse struct {
	Token string `json:"token"`
}
