package main

import (
	"context"
	"exchangeapp/config"
	"exchangeapp/internal/app"
	"exchangeapp/utils"
	"log"
)

func main() {
	cfg, err := config.LoadConfig("./config")
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	db, err := config.InitDB(cfg)
	if err != nil {
		log.Fatalf("init db: %v", err)
	}

	redisClient, err := config.InitRedis(context.Background(), cfg)
	if err != nil {
		log.Fatalf("init redis: %v", err)
	}

	if err := config.Migrate(db); err != nil {
		log.Fatalf("migrate database: %v", err)
	}

	utils.SetJWTSecret(cfg.Auth.JWTSecret)

	r := app.SetUpRouter(app.Dependencies{
		DB:        db,
		RedisDB:   redisClient,
		UploadDir: cfg.Storage.UploadDir,
	})

	port := cfg.App.Port
	if port == "" {
		port = "3000"
	}

	r.Run(":" + port)
}
