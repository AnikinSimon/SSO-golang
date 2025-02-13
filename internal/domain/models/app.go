package models

import (
	"github.com/google/uuid"
)

type App struct {
	ID     uuid.UUID `gorm: "primaryKey"`
	Name   string
	Secret string
}
