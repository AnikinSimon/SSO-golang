package models

import "gorm.io/gorm"

type User struct {
	gorm.Model
	ID       string `gorm:"primaryKey"`
	Email    string `gorm:"unique; not null"`
	Passhash []byte `gorm:"not null"`
	IsAdmin  bool   `gorm:"default:false"`
	AppID    string `gorm:"not null"`
	App      App    `gorm:"foreignKey:AppID"`
}
