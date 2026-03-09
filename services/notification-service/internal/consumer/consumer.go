package consumer

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/VladUrsul/livestream-platform/services/notification-service/internal/domain"
	"github.com/VladUrsul/livestream-platform/services/notification-service/internal/hub"
	"github.com/VladUrsul/livestream-platform/services/notification-service/internal/repository"
	"github.com/google/uuid"
	amqp "github.com/rabbitmq/amqp091-go"
)

type Consumer struct {
	conn           *amqp.Connection
	repo           repository.NotificationRepository
	hub            *hub.Hub
	userServiceURL string
	userExchange   string
	streamExchange string
	queueName      string
}

func New(
	conn *amqp.Connection,
	repo repository.NotificationRepository,
	h *hub.Hub,
	userServiceURL, userExchange, streamExchange, queueName string,
) *Consumer {
	return &Consumer{
		conn:           conn,
		repo:           repo,
		hub:            h,
		userServiceURL: userServiceURL,
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
	var peek struct {
		EventType string `json:"event_type"`
	}
	if err := json.Unmarshal(body, &peek); err != nil {
		log.Printf("[Consumer] failed to peek stream event: %v", err)
		return
	}
	if peek.EventType != "stream.started" {
		log.Printf("[Consumer] ignoring stream event type=%q", peek.EventType)
		return
	}

	var evt domain.StreamStartedEvent
	if err := json.Unmarshal(body, &evt); err != nil {
		log.Printf("[Consumer] failed to unmarshal stream event: %v", err)
		return
	}

	log.Printf("[Consumer] stream event received: userID=%s username=%s title=%s", evt.UserID, evt.Username, evt.Title)

	// Fetch followers
	followers, err := c.getFollowers(ctx, evt.UserID)
	if err != nil {
		log.Printf("[Consumer] getFollowers error for %s: %v", evt.UserID, err)
		followers = []uuid.UUID{}
	}

	log.Printf("[Consumer] found %d followers for @%s", len(followers), evt.Username)

	for _, followerID := range followers {
		n := &domain.Notification{
			UserID:    followerID,
			Type:      domain.NotifyStreamLive,
			Title:     "@" + evt.Username + " is live",
			Body:      evt.Username + " started streaming: " + evt.Title,
			ActorID:   evt.UserID,
			ActorName: evt.Username,
		}
		if err := c.repo.Create(ctx, n); err != nil {
			log.Printf("[Consumer] create notification error: %v", err)
			continue
		}
		unread, _ := c.repo.GetUnreadCount(ctx, followerID)
		c.hub.Push(followerID, &domain.WSPush{
			Type:         "notification",
			Notification: n,
			UnreadCount:  unread,
		})
		log.Printf("[Consumer] notified follower %s of stream by @%s", followerID, evt.Username)
	}

	// Also notify the streamer themselves
	n := &domain.Notification{
		UserID:    evt.UserID,
		Type:      domain.NotifyStreamLive,
		Title:     "Your stream is live",
		Body:      "Your stream \"" + evt.Title + "\" is now live",
		ActorID:   evt.UserID,
		ActorName: evt.Username,
	}
	if err := c.repo.Create(ctx, n); err == nil {
		unread, _ := c.repo.GetUnreadCount(ctx, evt.UserID)
		c.hub.Push(evt.UserID, &domain.WSPush{
			Type:         "notification",
			Notification: n,
			UnreadCount:  unread,
		})
	}
}

func (c *Consumer) getFollowers(ctx context.Context, userID uuid.UUID) ([]uuid.UUID, error) {
	url := fmt.Sprintf("%s/api/v1/users/internal/%s/follower-ids", c.userServiceURL, userID)
	log.Printf("[Consumer] fetching followers from %s", url)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	log.Printf("[Consumer] follower-ids response status: %d", resp.StatusCode)

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	log.Printf("[Consumer] follower-ids response body: %s", string(data))

	var result struct {
		FollowerIDs []uuid.UUID `json:"follower_ids"`
	}
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, err
	}
	return result.FollowerIDs, nil
}
