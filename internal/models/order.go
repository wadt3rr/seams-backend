package models

import (
	"time"

	"github.com/google/uuid"
)

type OrderStatus string

const (
	OrderNew       OrderStatus = "new"
	OrderConfirmed OrderStatus = "confirmed"
	OrderPreparing OrderStatus = "prepairing"
	OrderReady     OrderStatus = "ready"
	OrderDelivered OrderStatus = "delivered"
	OrderCancelled OrderStatus = "cancelled"
)

type PaymentMethod string

const (
	PaymentInvoice    PaymentMethod = "invoice"
	PaymentSBP        PaymentMethod = "sbp"
	PaymentOnDelivery PaymentMethod = "cash_on_delivery"
)

type FullfillmentMethod string

const (
	FullfillmentPickup FullfillmentMethod = "pickup"
)

type Order struct {
	ID         uuid.UUID `json:"id"`
	CustomerID uuid.UUID `json:"customer_id"`

	Status OrderStatus `json:"status"`
	Total  int64       `json:"total"`

	Items []OrderItem `json:"items"`

	PaymentMethod      PaymentMethod      `json:"payment_method"`
	FullfillmentMethod FullfillmentMethod `json:"fullfillment_method"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type OrderItem struct {
	ID        uuid.UUID `json:"id"`
	ProductID uuid.UUID `json:"product_id"`
	Quantity  int       `json:"quantity"`
	Price     int64     `json:"price"` // цена на момент покупки
	Total     int64     `json:"total"` // Price * Quantity
}

type OrderDetailsDTO struct {
	ID                 uuid.UUID          `json:"id"`
	CustomerName       string             `json:"customer_name"`
	Status             OrderStatus        `json:"status"`
	PaymentMethod      PaymentMethod      `json:"payment_method"`
	FullfillmentMethod FullfillmentMethod `json:"fullfillment_method"`
	Items              []OrderItemDTO     `json:"items"`
	Total              int64              `json:"total"`
	CreatedAt          time.Time          `json:"created_at"`
	Invoice            *Invoice           `json:"invoice"`
}

type OrderItemDTO struct {
	ProductID   uuid.UUID `json:"product_id"`
	ProductName string    `json:"product_name"`
	Quantity    int       `json:"quantity"`
	Price       int64     `json:"price"`
	Total       int64     `json:"total"`
}
