package kafka

import (
	"context"
	"orderfc/cmd/order/repository"
	"orderfc/infrastructure/log"
	"time"
)

type OrderOutboxPublisher struct {
	OrderRepo *repository.OrderRepository
	Producer  *KafkaProducer
	Interval  time.Duration
	BatchSize int
}

func NewOrderOutboxPublisher(orderRepo *repository.OrderRepository, producer *KafkaProducer) *OrderOutboxPublisher {
	return &OrderOutboxPublisher{
		OrderRepo: orderRepo,
		Producer:  producer,
		Interval:  2 * time.Second,
		BatchSize: 20,
	}
}

func (p *OrderOutboxPublisher) Start(ctx context.Context) {
	ticker := time.NewTicker(p.Interval)
	defer ticker.Stop()

	for {
		p.publishPending(ctx)

		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
	}
}

func (p *OrderOutboxPublisher) publishPending(ctx context.Context) {
	events, err := p.OrderRepo.GetPendingOutboxEvents(ctx, p.BatchSize)
	if err != nil {
		log.Logger.Error().Err(err).Msg("Failed to fetch order outbox events")
		return
	}

	for _, event := range events {
		publishCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
		err := p.Producer.PublishRaw(publishCtx, event.Topic, event.EventKey, []byte(event.Payload))
		cancel()

		if err != nil {
			log.Logger.Error().Err(err).Int64("event_id", event.ID).Str("topic", event.Topic).Msg("Failed to publish order outbox event")
			if markErr := p.OrderRepo.MarkOutboxEventFailed(ctx, event.ID, err); markErr != nil {
				log.Logger.Error().Err(markErr).Int64("event_id", event.ID).Msg("Failed to mark order outbox event failed")
			}
			continue
		}

		if err := p.OrderRepo.MarkOutboxEventPublished(ctx, event.ID); err != nil {
			log.Logger.Error().Err(err).Int64("event_id", event.ID).Msg("Failed to mark order outbox event published")
		}
	}
}
