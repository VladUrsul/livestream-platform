package consumer

import (
	"context"
	"encoding/json"
	"log"

	"github.com/VladUrsul/livestream-platform/services/user-service/internal/domain"
	"github.com/VladUrsul/livestream-platform/services/user-service/internal/service"
	"github.com/google/uuid"
	amqp "github.com/rabbitmq/amqp091-go"
)

type Consumer struct {
	conn           *amqp.Connection
	svc            service.UserService
	authExchange   string
	streamExchange string
	queueName      string
}

func New(conn *amqp.Connection, svc service.UserService, authExchange, streamExchange, queueName string) *Consumer {
	return &Consumer{conn, svc, authExchange, streamExchange, queueName}
}

func (c *Consumer) Start(ctx context.Context) {
	go c.consume(ctx, c.authExchange)
	go c.consume(ctx, c.streamExchange)
}

func (c *Consumer) consume(ctx context.Context, exchange string) {
	ch, err := c.conn.Channel()
	if err != nil {
		log.Printf("[Consumer] channel error (%s): %v", exchange, err)
		return
	}
	defer ch.Close()

	ch.ExchangeDeclare(exchange, "fanout", true, false, false, false, nil)

	q, err := ch.QueueDeclare(c.queueName+"."+exchange, true, false, false, false, nil)
	if err != nil {
		log.Printf("[Consumer] queue error: %v", err)
		return
	}
	ch.QueueBind(q.Name, "", exchange, false, nil)

	msgs, _ := ch.Consume(q.Name, "", false, false, false, false, nil)
	log.Printf("[Consumer] listening on exchange=%s queue=%s", exchange, q.Name)

	for {
		select {
		case <-ctx.Done():
			return
		case msg, ok := <-msgs:
			if !ok {
				return
			}
			c.handle(ctx, msg)
		}
	}
}

func (c *Consumer) handle(ctx context.Context, msg amqp.Delivery) {
	switch msg.RoutingKey {

	case "user.registered":
		var e domain.UserRegisteredEvent
		if err := json.Unmarshal(msg.Body, &e); err != nil {
			msg.Nack(false, false)
			return
		}
		if err := c.svc.CreateFromEvent(ctx, e); err != nil {
			log.Printf("[Consumer] create profile @%s: %v", e.Username, err)
			msg.Nack(false, true)
			return
		}
		log.Printf("[Consumer] ✓ profile created @%s", e.Username)
		msg.Ack(false)

	case "stream.started":
		var e struct {
			UserID string `json:"user_id"`
		}
		json.Unmarshal(msg.Body, &e)
		if id, err := uuid.Parse(e.UserID); err == nil {
			c.svc.SetLiveStatus(ctx, id, true)
		}
		msg.Ack(false)

	case "stream.ended":
		var e struct {
			UserID string `json:"user_id"`
		}
		json.Unmarshal(msg.Body, &e)
		if id, err := uuid.Parse(e.UserID); err == nil {
			c.svc.SetLiveStatus(ctx, id, false)
		}
		msg.Ack(false)

	default:
		msg.Ack(false)
	}
}
