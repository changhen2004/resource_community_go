package main

import (
	"context"
	"exchangeapp/config"
	"exchangeapp/global"
	"exchangeapp/router"
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
	global.DB = db

	redisClient, err := config.InitRedis(context.Background(), cfg)
	if err != nil {
		log.Fatalf("init redis: %v", err)
	}
	global.RedisDB = redisClient

	if err := config.Migrate(db); err != nil {
		log.Fatalf("migrate database: %v", err)
	}

	r := router.SetUpRouter()

	port := cfg.App.Port
	if port == "" {
		port = "3000"
	}

	r.Run(":" + port)
}
