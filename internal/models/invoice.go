package models

import (
	"time"

	"github.com/google/uuid"
)

type InvoiceStatus string

const (
	InvoiceIssued    InvoiceStatus = "issued"
	InvoicePaid      InvoiceStatus = "paid"
	InvoiceCancelled InvoiceStatus = "cancelled"
)

type Invoice struct {
	ID        uuid.UUID     `db:"id" json:"id"`
	OrderID   uuid.UUID     `db:"order_id" json:"order_id"`
	Number    string        `db:"number" json:"number"`
	Amount    int64         `db:"amount" json:"amount"`
	Status    InvoiceStatus `db:"status" json:"status"`
	IssuedAt  time.Time     `db:"issued_at" json:"issued_at"`
	CreatedAt time.Time     `db:"created_at" json:"created_at"`
	UpdatedAt time.Time     `db:"updated_at" json:"updated_at"`
}
