package model

import "gorm.io/gorm"

type User struct {
	gorm.Model `gorm:"unique"`
	Username   string
	Password   string
}
