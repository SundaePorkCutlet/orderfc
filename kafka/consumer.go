package kafka

import (
	"context"
	"encoding/json"
	"orderfc/cmd/order/service"
	"orderfc/infrastructure/constant"
	"orderfc/infrastructure/log"
	"orderfc/models"
	"time"

	"github.com/segmentio/kafka-go"
)

type PaymentSuccessEvent struct {
	Reader        *kafka.Reader
	KafkaProducer *KafkaProducer
	OrderService  *service.OrderService
}

func NewPaymentSuccessEvent(brokers []string, topic string, orderService *service.OrderService, kafkaProducer *KafkaProducer) *PaymentSuccessEvent {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers: brokers,
		Topic:   topic,
		GroupID: "orderfc",
	})

	return &PaymentSuccessEvent{
		Reader:        reader,
		OrderService:  orderService,
		KafkaProducer: kafkaProducer,
	}
}

func (e *PaymentSuccessEvent) Start(ctx context.Context) {
	for {
		msg, err := e.Reader.ReadMessage(ctx)
		if err != nil {
			log.Logger.Error().Err(err).Msg("Failed to read message from Kafka")
			continue
		}
		var event models.PaymentSuccessEvent
		if err := json.Unmarshal(msg.Value, &event); err != nil {
			log.Logger.Error().Err(err).Msg("Failed to unmarshal message from Kafka")
			continue
		}
		err = e.OrderService.UpdateOrderStatus(ctx, event.OrderID, constant.OrderStatusCompleted)
		if err != nil {
			log.Logger.Error().Err(err).Msg("Failed to update order status")
			continue
		}

		orderInfo, err := e.OrderService.GetOrderInfoByOrderID(ctx, event.OrderID)
		if err != nil {
			log.Logger.Error().Err(err).Msg("Failed to get order info")
			continue
		}

		orderDetail, err := e.OrderService.GetOrderDetailByOrderID(ctx, orderInfo.OrderDetailID)
		if err != nil {
			log.Logger.Error().Err(err).Msg("Failed to get order detail")
			continue
		}

		var products []models.CheckoutItem
		err = json.Unmarshal([]byte(orderDetail.Products), &products)
		if err != nil {
			log.Logger.Error().Err(err).Msg("Failed to unmarshal products")
			continue
		}
		publishProductStockUpdatedEvent := models.ProductStockUpdatedEvent{
			OrderID:   event.OrderID,
			Products:  convertCheckoutItemToProductItem(products),
			EventTime: time.Now(),
		}
		err = e.KafkaProducer.PublishProductStockUpdated(ctx, publishProductStockUpdatedEvent)
		if err != nil {
			log.Logger.Error().Err(err).Msg("Failed to publish product stock updated event")
			continue
		}
	}
}

func convertCheckoutItemToProductItem(products []models.CheckoutItem) []models.ProductItem {
	var productItems []models.ProductItem
	for _, product := range products {
		productItems = append(productItems, models.ProductItem{
			ProductID: product.ProductID,
			Quantity:  product.Quantity,
		})
	}
	return productItems
}
