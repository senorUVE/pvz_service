package dto

import (
	"time"

	"github.com/google/uuid"
)

type GetPvzRequest struct {
	StartDate time.Time `query:"startDate"`
	EndDate   time.Time `query:"endDate"`
	Page      int       `query:"page"`
	Limit     int       `query:"limit"`
}

type ReceptionResponse struct {
	Id       uuid.UUID `json:"id" db:"id"`
	DateTime time.Time `json:"dateTime" db:"date_time"`
	PvzId    uuid.UUID `json:"pvzId" db:"pvz_Id"`
	Status   string    `json:"status" db:"status"`
}

type ProductResponse struct {
	Id          uuid.UUID `json:"id" db:"id"`
	DateTime    time.Time `json:"dateTime" db:"date_time"`
	Type        string    `json:"type" db:"type"`
	ReceptionId uuid.UUID `json:"receptionId" db:"reception_Id"`
}

type ReceptionWithProducts struct {
	Reception ReceptionResponse `json:"reception"`
	Products  []ProductResponse `json:"products"`
}

type PVZResponse struct {
	Id               uuid.UUID `json:"id" db:"id"`
	RegistrationDate time.Time `json:"registrationDate" db:"registration_date"`
	City             string    `json:"city" db:"city"`
}

type PVZWithReceptions struct {
	PVZ        PVZResponse             `json:"pvz"`
	Receptions []ReceptionWithProducts `json:"receptions"`
}
