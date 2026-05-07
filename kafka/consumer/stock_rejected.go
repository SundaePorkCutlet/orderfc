package consumer

import (
	"context"
	"encoding/json"
	"orderfc/cmd/order/service"
	"orderfc/infrastructure/constant"
	"orderfc/infrastructure/log"
	"orderfc/models"

	"github.com/segmentio/kafka-go"
)

type StockRejectedConsumer struct {
	Reader       *kafka.Reader
	OrderService *service.OrderService
}

func NewStockRejectedConsumer(brokers []string, topic string, orderService *service.OrderService) *StockRejectedConsumer {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers: brokers,
		Topic:   topic,
		GroupID: "orderfc",
	})
	return &StockRejectedConsumer{
		Reader:       reader,
		OrderService: orderService,
	}
}

func (c *StockRejectedConsumer) Start(ctx context.Context) {
	for {
		msg, err := c.Reader.ReadMessage(ctx)
		if err != nil {
			log.Logger.Error().Err(err).Msg("Failed to read stock.rejected message")
			continue
		}

		var event models.StockReservationEvent
		if err := json.Unmarshal(msg.Value, &event); err != nil {
			log.Logger.Error().Err(err).Msg("Failed to unmarshal stock.rejected message")
			continue
		}

		if err := c.OrderService.UpdateOrderStatus(ctx, event.OrderID, constant.OrderStatusCancelled); err != nil {
			log.Logger.Error().Err(err).Int64("order_id", event.OrderID).Msg("Failed to cancel order after stock rejection")
			continue
		}

		log.Logger.Info().Int64("order_id", event.OrderID).Str("reason", event.Reason).Msg("Order cancelled after stock rejection")
	}
}
