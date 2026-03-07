package consumer

import (
	"context"
	"encoding/json"
	"log"

	"github.com/VladUrsul/livestream-platform/services/notification-service/internal/domain"
	"github.com/VladUrsul/livestream-platform/services/notification-service/internal/hub"
	"github.com/VladUrsul/livestream-platform/services/notification-service/internal/repository"
	amqp "github.com/rabbitmq/amqp091-go"
)

type Consumer struct {
	conn           *amqp.Connection
	repo           repository.NotificationRepository
	hub            *hub.Hub
	userExchange   string
	streamExchange string
	queueName      string
}

func New(
	conn *amqp.Connection,
	repo repository.NotificationRepository,
	h *hub.Hub,
	userExchange, streamExchange, queueName string,
) *Consumer {
	return &Consumer{
		conn:           conn,
		repo:           repo,
		hub:            h,
		userExchange:   userExchange,
		streamExchange: streamExchange,
		queueName:      queueName,
	}
}

func (c *Consumer) Start(ctx context.Context) {
	go c.consume(ctx, c.userExchange, c.handleUserEvent)
	go c.consume(ctx, c.streamExchange, c.handleStreamEvent)
}

func (c *Consumer) consume(ctx context.Context, exchange string, handler func(context.Context, []byte)) {
	ch, err := c.conn.Channel()
	if err != nil {
		log.Printf("[Consumer] channel error for %s: %v", exchange, err)
		return
	}
	defer ch.Close()

	if err := ch.ExchangeDeclare(exchange, "fanout", true, false, false, false, nil); err != nil {
		log.Printf("[Consumer] exchange declare error: %v", err)
		return
	}

	q, err := ch.QueueDeclare(c.queueName+"."+exchange, true, false, false, false, nil)
	if err != nil {
		log.Printf("[Consumer] queue declare error: %v", err)
		return
	}

	if err := ch.QueueBind(q.Name, "", exchange, false, nil); err != nil {
		log.Printf("[Consumer] queue bind error: %v", err)
		return
	}

	msgs, err := ch.Consume(q.Name, "", false, false, false, false, nil)
	if err != nil {
		log.Printf("[Consumer] consume error: %v", err)
		return
	}

	log.Printf("[Consumer] listening on exchange=%s", exchange)

	for {
		select {
		case <-ctx.Done():
			return
		case msg, ok := <-msgs:
			if !ok {
				return
			}
			handler(ctx, msg.Body)
			msg.Ack(false)
		}
	}
}

func (c *Consumer) handleUserEvent(ctx context.Context, body []byte) {
	var evt domain.UserFollowedEvent
	if err := json.Unmarshal(body, &evt); err != nil {
		return
	}
	// Create notification for the followee
	n := &domain.Notification{
		UserID:    evt.FolloweeID,
		Type:      domain.NotifyFollowed,
		Title:     "New follower",
		Body:      "@" + evt.FollowerUsername + " started following you",
		ActorID:   evt.FollowerID,
		ActorName: evt.FollowerUsername,
	}
	if err := c.repo.Create(ctx, n); err != nil {
		log.Printf("[Consumer] create notification error: %v", err)
		return
	}
	// Push to WebSocket if user is online
	unread, _ := c.repo.GetUnreadCount(ctx, evt.FolloweeID)
	c.hub.Push(evt.FolloweeID, &domain.WSPush{
		Type:         "notification",
		Notification: n,
		UnreadCount:  unread,
	})
	log.Printf("[Consumer] notified %s of new follower @%s", evt.FolloweeID, evt.FollowerUsername)
}

func (c *Consumer) handleStreamEvent(ctx context.Context, body []byte) {
	var evt domain.StreamStartedEvent
	if err := json.Unmarshal(body, &evt); err != nil {
		return
	}
	// TODO: fetch followers of evt.UserID from user-service
	// For now, notify the streamer themselves as a test
	n := &domain.Notification{
		UserID:    evt.UserID,
		Type:      domain.NotifyStreamLive,
		Title:     "Stream started",
		Body:      "@" + evt.Username + " is now live: " + evt.Title,
		ActorID:   evt.UserID,
		ActorName: evt.Username,
	}
	if err := c.repo.Create(ctx, n); err != nil {
		return
	}
	unread, _ := c.repo.GetUnreadCount(ctx, evt.UserID)
	c.hub.Push(evt.UserID, &domain.WSPush{
		Type:         "notification",
		Notification: n,
		UnreadCount:  unread,
	})
}
