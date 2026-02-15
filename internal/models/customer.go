package models

import "github.com/google/uuid"

type Customer struct {
	ID    uuid.UUID `db:"id" json:"id"`
	Name  string    `db:"name" json:"name"`
	Email string    `db:"email" json:"email"`
	Phone string    `db:"phone" json:"phone"`
}
