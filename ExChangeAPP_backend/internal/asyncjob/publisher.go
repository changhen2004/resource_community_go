package asyncjob

import (
	"context"
	"encoding/json"

	amqp "github.com/rabbitmq/amqp091-go"
)

type Publisher interface {
	Publish(ctx context.Context, job Job) error
}

type NoopPublisher struct{}

func (NoopPublisher) Publish(context.Context, Job) error {
	return nil
}

type RabbitPublisher struct {
	channel  *amqp.Channel
	exchange string
	queue    string
}

func NewRabbitPublisher(conn *amqp.Connection, exchange, queue string) (*RabbitPublisher, error) {
	ch, err := conn.Channel()
	if err != nil {
		return nil, err
	}

	if err := ch.ExchangeDeclare(exchange, "direct", true, false, false, false, nil); err != nil {
		_ = ch.Close()
		return nil, err
	}
	if _, err := ch.QueueDeclare(queue, true, false, false, false, nil); err != nil {
		_ = ch.Close()
		return nil, err
	}
	if err := ch.QueueBind(queue, queue, exchange, false, nil); err != nil {
		_ = ch.Close()
		return nil, err
	}

	return &RabbitPublisher{
		channel:  ch,
		exchange: exchange,
		queue:    queue,
	}, nil
}

func (p *RabbitPublisher) Publish(ctx context.Context, job Job) error {
	body, err := json.Marshal(job)
	if err != nil {
		return err
	}

	return p.channel.PublishWithContext(ctx, p.exchange, p.queue, false, false, amqp.Publishing{
		ContentType: "application/json",
		Body:        body,
	})
}

func (p *RabbitPublisher) Close() error {
	if p == nil || p.channel == nil {
		return nil
	}
	return p.channel.Close()
}
