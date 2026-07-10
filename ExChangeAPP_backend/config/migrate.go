package config

import (
	"exchangeapp/model"

	"gorm.io/gorm"
)

func Migrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&model.User{},
		&model.Article{},
		&model.ExchangeRate{},
	)
}
