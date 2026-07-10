package model

import "gorm.io/gorm"

type User struct {
	gorm.Model
	Username string `json:"username" binding:"required,min=3,max=50" gorm:"size:50;not null;uniqueIndex"`
	Password string `json:"password" binding:"required,min=6,max=72" gorm:"size:255;not null"`
}
