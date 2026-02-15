package models

import "github.com/google/uuid"

type Category struct {
	ID   uuid.UUID `db:"id" json:"id"`
	Name string    `db:"name" json:"name"`
	Slug string    `db:"slug" json:"slug"`
}
