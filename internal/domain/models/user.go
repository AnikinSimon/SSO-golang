package models

import "github.com/google/uuid"

type User struct {
	UUID     uuid.UUID `gorm: "primaryKey"`
	Email    string    `gorm: "unique, index"`
	Passhash []byte		
}
