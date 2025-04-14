package dto

import (
	"github.com/google/uuid"
)

type RegisterRequest struct {
	Email    string `json:"email" db:"email"`
	Password string `json:"password" db:"password_salt"`
	Role     string `json:"role" db:"role"`
}

type RegisterResponse struct {
	Id    uuid.UUID `json:"id" db:"id"`
	Email string    `json:"email" db:"email"`
	Role  string    `json:"role" db:"role"`
}
