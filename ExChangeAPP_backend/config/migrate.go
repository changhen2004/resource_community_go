package config

import (
	"exchangeapp/internal/article"
	"exchangeapp/internal/auth"
	"exchangeapp/internal/exchange"

	"gorm.io/gorm"
)

func Migrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&auth.User{},
		&article.Article{},
		&exchange.ExchangeRate{},
	)
}
