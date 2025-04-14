package dto

import "github.com/google/uuid"

type DeleteProductRequest struct {
	PvzId uuid.UUID `query:"pvzId" db:"pvz_id"`
}
