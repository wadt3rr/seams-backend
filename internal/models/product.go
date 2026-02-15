package models

import (
	"github.com/google/uuid"
)

type Product struct {
	ID          uuid.UUID `db:"id" json:"id"`
	Name        string    `db:"name" json:"name"`
	Sku         string    `db:"sku" json:"sku"`
	Price       int64     `db:"price" json:"price"`
	CategoryID  uuid.UUID `db:"category_id" json:"category_id"`
	Description *string   `db:"description" json:"description,omitempty"`
}
