package models

import (
	"time"

	"github.com/google/uuid"
)

type Request struct {
	ID         uuid.UUID `json:"id"`
	CustomerID uuid.UUID `json:"customer_id"`

	Description string `json:"desc"`
	FilePath    string `json:"file_path"`

	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}
