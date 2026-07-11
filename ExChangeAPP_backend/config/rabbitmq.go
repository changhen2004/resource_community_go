package config

import (
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
)

func InitRabbitMQ(cfg *Config) (*amqp.Connection, error) {
	conn, err := amqp.Dial(cfg.RabbitMQ.URL)
	if err != nil {
		return nil, err
	}

	log.Println("rabbitmq initialized")
	return conn, nil
}
