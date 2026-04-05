package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"orderfc/models"

	"github.com/segmentio/kafka-go"
)

type KafkaProducer struct {
	writer *kafka.Writer
}

func NewKafkaProducer(brokers []string) *KafkaProducer {
	writer := &kafka.Writer{
		Addr:     kafka.TCP(brokers...),
		Balancer: &kafka.Hash{}, // Message.Key 기준 파티션 (user_id 기반 순서 보장)
	}
	return &KafkaProducer{writer: writer}
}

func stockEventPartitionKey(userID, orderID int64) []byte {
	if userID > 0 {
		return []byte(fmt.Sprintf("user-%d", userID))
	}
	return []byte(fmt.Sprintf("order-%d", orderID))
}

func (p *KafkaProducer) Close() error {
	return p.writer.Close()
}

func (p *KafkaProducer) PublishOrderCreated(ctx context.Context, event interface{}) error {
	json, err := json.Marshal(event)
	if err != nil {
		return err
	}
	msg := kafka.Message{
		Key:   []byte(fmt.Sprintf("order-%d", event.(models.OrderCreatedEvent).OrderID)),
		Value: json,
		Topic: "order.created",
	}
	return p.writer.WriteMessages(ctx, msg)
}

func (p *KafkaProducer) PublishProductStockUpdated(ctx context.Context, event models.ProductStockUpdatedEvent) error {
	json, err := json.Marshal(event)
	if err != nil {
		return err
	}
	msg := kafka.Message{
		Key:   stockEventPartitionKey(event.UserID, event.OrderID),
		Value: json,
		Topic: "stock.updated",
	}
	return p.writer.WriteMessages(ctx, msg)
}

func (p *KafkaProducer) PublishStockRollback(ctx context.Context, event models.ProductStockUpdatedEvent) error {
	json, err := json.Marshal(event)
	if err != nil {
		return err
	}
	msg := kafka.Message{
		Key:   stockEventPartitionKey(event.UserID, event.OrderID),
		Value: json,
		Topic: "stock.rollback",
	}
	return p.writer.WriteMessages(ctx, msg)
}
