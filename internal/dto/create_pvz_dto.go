package dto

import (
	"time"

	"github.com/google/uuid"
)

type PvzCreateRequest struct {
	Id               uuid.UUID `json:"id" db:"id"`
	RegistrationDate time.Time `json:"registrationDate" db:"registration_date"`
	City             string    `json:"city" db:"city"`
}

type PvzCreateResponse struct {
	Id               uuid.UUID `json:"id" db:"id"`
	RegistrationDate time.Time `json:"registrationDate" db:"registration_date"`
	City             string    `json:"city" db:"city"`
}
