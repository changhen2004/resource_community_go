package worker

import (
	"context"
	"encoding/json"
	"errors"
	"log"

	"exchangeapp/internal/asyncjob"

	amqp "github.com/rabbitmq/amqp091-go"
)

type Runner struct {
	channel   *amqp.Channel
	queue     string
	processor *Processor
}

func NewRunner(conn *amqp.Connection, exchange, queue string, processor *Processor) (*Runner, error) {
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

	return &Runner{
		channel:   ch,
		queue:     queue,
		processor: processor,
	}, nil
}

func (w *Runner) Run(ctx context.Context) error {
	msgs, err := w.channel.Consume(w.queue, "", false, false, false, false, nil)
	if err != nil {
		return err
	}

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case msg, ok := <-msgs:
			if !ok {
				return errors.New("worker delivery channel closed")
			}

			var job asyncjob.Job
			if err := json.Unmarshal(msg.Body, &job); err != nil {
				log.Printf("invalid async job payload: %v", err)
				_ = msg.Nack(false, false)
				continue
			}

			if err := w.processor.Handle(ctx, job); err != nil {
				log.Printf("process async job failed: type=%s err=%v", job.Type, err)
				_ = msg.Nack(false, true)
				continue
			}

			_ = msg.Ack(false)
		}
	}
}

func (w *Runner) Close() error {
	if w == nil || w.channel == nil {
		return nil
	}
	return w.channel.Close()
}
