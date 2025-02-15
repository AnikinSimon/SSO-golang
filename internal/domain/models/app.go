package models

import "gorm.io/gorm"

type App struct {
	gorm.Model
	ID     string `gorm:"primaryKey"`
	Name   string `gorm:"unique"`
	Secret string
}
