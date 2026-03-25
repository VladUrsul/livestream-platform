package publisher

import (
	"encoding/json"
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
)

type Publisher struct {
	ch       *amqp.Channel
	exchange string
}

func New(conn *amqp.Connection, exchange string) (*Publisher, error) {
	ch, err := conn.Channel()
	if err != nil {
		return nil, err
	}
	if err := ch.ExchangeDeclare(exchange, "fanout", true, false, false, false, nil); err != nil {
		return nil, err
	}
	return &Publisher{ch, exchange}, nil
}

func (p *Publisher) Publish(routingKey string, payload any) error {
	if p.ch == nil {
		return fmt.Errorf("RabbitMQ channel not initialized")
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	return p.ch.Publish(p.exchange, routingKey, false, false,
		amqp.Publishing{ContentType: "application/json", Body: body})
}

func (p *Publisher) Close() { p.ch.Close() }
