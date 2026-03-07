package scheduler

import (
	"context"
	"log"
	"time"

	"github.com/VladUrsul/livestream-platform/services/notification-service/internal/repository"
)

type Scheduler struct {
	repo        repository.NotificationRepository
	interval    time.Duration
	expireAfter time.Duration
}

func New(repo repository.NotificationRepository) *Scheduler {
	return &Scheduler{
		repo:        repo,
		interval:    1 * time.Hour,
		expireAfter: 30 * 24 * time.Hour, // 30 days
	}
}

func (s *Scheduler) Run(ctx context.Context) {
	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()
	log.Println("[Scheduler] started — expiring read notifications after 30 days")
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			before := time.Now().Add(-s.expireAfter)
			n, err := s.repo.DeleteExpired(ctx, before)
			if err != nil {
				log.Printf("[Scheduler] delete expired error: %v", err)
			} else if n > 0 {
				log.Printf("[Scheduler] deleted %d expired notifications", n)
			}
		}
	}
}
