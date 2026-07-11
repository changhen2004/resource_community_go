package main

import (
	"context"
	"exchangeapp/config"
	"exchangeapp/internal/article"
	"exchangeapp/internal/points"
	internalWorker "exchangeapp/internal/worker"
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

	rabbitConn, err := config.InitRabbitMQ(cfg)
	if err != nil {
		log.Fatalf("init rabbitmq: %v", err)
	}
	defer rabbitConn.Close()

	pointsService := points.NewService(points.NewRepo(db, redisClient))
	articleService := article.NewService(article.NewRepo(db, redisClient), nil, pointsService)
	processor := internalWorker.NewProcessor(articleService, pointsService)

	worker, err := internalWorker.NewRunner(rabbitConn, cfg.RabbitMQ.Exchange, cfg.RabbitMQ.Queue, processor)
	if err != nil {
		log.Fatalf("init async worker: %v", err)
	}
	defer worker.Close()

	if err := worker.Run(context.Background()); err != nil {
		log.Fatalf("run async worker: %v", err)
	}
}
