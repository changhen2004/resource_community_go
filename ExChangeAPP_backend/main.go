package main

import (
	"context"
	"exchangeapp/config"
	"exchangeapp/internal/app"
	"exchangeapp/internal/asyncjob"
	"exchangeapp/utils"
	"log"
	"time"
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
	rabbitConn, err := config.InitRabbitMQ(cfg)
	if err != nil {
		log.Fatalf("init rabbitmq: %v", err)
	}
	defer rabbitConn.Close()

	publisher, err := asyncjob.NewRabbitPublisher(
		rabbitConn,
		cfg.RabbitMQ.Exchange,
		cfg.RabbitMQ.Queue,
	)
	if err != nil {
		log.Fatalf("init async publisher: %v", err)
	}
	defer publisher.Close()

	if err := config.Migrate(db); err != nil {
		log.Fatalf("migrate database: %v", err)
	}

	utils.SetJWTSecret(cfg.Auth.JWTSecret)

	r := app.SetUpRouter(app.Dependencies{
		DB:                   db,
		RedisDB:              redisClient,
		UploadDir:            cfg.Storage.UploadDir,
		EnablePprof:          cfg.Observability.EnablePprof,
		AsyncPublisher:       publisher,
		SlowRequestThreshold: time.Duration(cfg.Observability.SlowRequestThresholdM) * time.Millisecond,
	})

	port := cfg.App.Port
	if port == "" {
		port = "3000"
	}

	r.Run(":" + port)
}
