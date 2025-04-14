package dto

import (
	"time"

	"github.com/google/uuid"
)

type CloseLastReceptionRequest struct {
	PvzId uuid.UUID `query:"pvzId" db:"pvz_id"`
}

type CloseLastReceptionResponse struct {
	Id       uuid.UUID `json:"id" db:"id"`
	DateTime time.Time `json:"dateTime" db:"date_time"`
	PvzId    uuid.UUID `json:"pvzId" db:"pvz_id"`
	Status   string    `json:"status" db:"status"`
}
