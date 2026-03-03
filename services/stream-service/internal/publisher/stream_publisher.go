package publisher

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/VladUrsul/livestream-platform/services/stream-service/internal/domain"
	amqp "github.com/rabbitmq/amqp091-go"
)

type StreamPublisher struct {
	channel      *amqp.Channel
	exchangeName string
}

func NewStreamPublisher(conn *amqp.Connection, exchangeName string) (*StreamPublisher, error) {
	ch, err := conn.Channel()
	if err != nil {
		return nil, fmt.Errorf("open channel: %w", err)
	}
	if err := ch.ExchangeDeclare(exchangeName, "fanout", true, false, false, false, nil); err != nil {
		return nil, fmt.Errorf("declare exchange: %w", err)
	}
	return &StreamPublisher{channel: ch, exchangeName: exchangeName}, nil
}

func (p *StreamPublisher) PublishStreamStarted(ctx context.Context, event domain.StreamStartedEvent) error {
	return p.publish("stream.started", event)
}

func (p *StreamPublisher) PublishStreamEnded(ctx context.Context, event domain.StreamEndedEvent) error {
	return p.publish("stream.ended", event)
}

func (p *StreamPublisher) publish(routingKey string, payload any) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal event: %w", err)
	}
	return p.channel.Publish(p.exchangeName, routingKey, false, false,
		amqp.Publishing{ContentType: "application/json", Body: body})
}

func (p *StreamPublisher) Close() { p.channel.Close() }
