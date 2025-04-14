package dto

import (
	"time"

	"github.com/google/uuid"
)

type AddProductRequest struct {
	Type  string    `json:"type" db:"type"`
	PvzId uuid.UUID `query:"pvzId" db:"pvz_id"`
}

type AddProductResponse struct {
	Id          uuid.UUID `json:"id" db:"id"`
	DateTime    time.Time `json:"dateTime" db:"date_time"`
	Type        string    `json:"type" db:"type"`
	ReceptionId uuid.UUID `json:"receptionId" db:"reception_id"`
}
