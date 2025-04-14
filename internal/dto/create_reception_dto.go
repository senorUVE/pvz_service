package dto

import (
	"time"

	"github.com/google/uuid"
)

type CreateReceptionRequest struct {
	PvzId uuid.UUID `json:"pvzId" db:"pvz_id"`
}

type CreateReceptionResponse struct {
	Id       uuid.UUID `json:"id" db:"id"`
	DateTime time.Time `json:"dateTime" db:"date_time"`
	PvzId    uuid.UUID `json:"pvzId" db:"pvz_id"`
	Status   string    `json:"status" db:"status"`
}
