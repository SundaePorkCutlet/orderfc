package usecase

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"orderfc/cmd/order/service"
	"orderfc/infrastructure/constant"
	"orderfc/kafka"
	"orderfc/models"
	"time"
)

var (
	ErrIdempotencyInProgress   = errors.New("idempotency request is still processing")
	ErrIdempotencyKeyReused    = errors.New("idempotency token reused with different request")
	ErrIdempotencyPreviousFail = errors.New("idempotency request previously failed")
)

type OrderUsecase struct {
	OrderService  service.OrderService
	KafkaProducer *kafka.KafkaProducer
}

func NewOrderUsecase(orderService service.OrderService, kafkaProducer *kafka.KafkaProducer) *OrderUsecase {
	return &OrderUsecase{OrderService: orderService, KafkaProducer: kafkaProducer}
}

func (u *OrderUsecase) CheckOutOrder(ctx context.Context, checkoutRequest *models.CheckoutRequest) (int64, error) {
	err := u.validateProducts(ctx, checkoutRequest.Items)
	if err != nil {
		return 0, err
	}

	if checkoutRequest.IdempotencyToken != "" {
		requestHash, err := hashCheckoutRequest(checkoutRequest)
		if err != nil {
			return 0, err
		}
		idem, reserved, err := u.OrderService.ReserveIdempotencyToken(ctx, checkoutRequest.IdempotencyToken, requestHash)
		if err != nil {
			return 0, err
		}
		if !reserved {
			if idem.RequestHash != requestHash {
				return 0, ErrIdempotencyKeyReused
			}
			switch idem.Status {
			case models.IdempotencyStatusSucceeded:
				if idem.OrderID > 0 {
					return idem.OrderID, nil
				}
				return 0, ErrIdempotencyInProgress
			case models.IdempotencyStatusProcessing:
				return 0, ErrIdempotencyInProgress
			case models.IdempotencyStatusFailed:
				return 0, ErrIdempotencyPreviousFail
			default:
				return 0, ErrIdempotencyInProgress
			}
		}
	}

	totalQty, totalAmount := u.calculateItemSummary(ctx, checkoutRequest.Items)

	products, history := u.constructOrderDetail(ctx, checkoutRequest.Items)

	orderDetail := &models.OrderDetail{
		Products:     products,
		OrderHistory: history,
	}

	order := &models.Order{
		UserID:          checkoutRequest.UserID,
		Amount:          totalAmount,
		TotalQty:        totalQty,
		PaymentMethod:   checkoutRequest.PaymentMethod,
		ShippingAddress: checkoutRequest.ShippingAddress,
		Status:          constant.OrderStatusCreated,
	}

	orderId, err := u.OrderService.SaveOrderAndOrderDetailWithOutboxAndIdempotency(ctx, order, orderDetail, checkoutRequest.IdempotencyToken, func(orderID int64) ([]models.OrderOutboxEvent, error) {
		orderCreatedEvent := models.OrderCreatedEvent{
			OrderID:         orderID,
			UserID:          checkoutRequest.UserID,
			TotalAmount:     totalAmount,
			PaymentMethod:   checkoutRequest.PaymentMethod,
			ShippingAddress: checkoutRequest.ShippingAddress,
			Products:        convertCheckoutItemToProductItem(checkoutRequest.Items),
		}
		orderCreatedPayload, err := json.Marshal(orderCreatedEvent)
		if err != nil {
			return nil, err
		}

		eventKey := fmt.Sprintf("order-%d", orderID)

		return []models.OrderOutboxEvent{
			{
				Topic:    "order.created",
				EventKey: eventKey,
				Payload:  string(orderCreatedPayload),
				Status:   models.OrderOutboxStatusPending,
			},
		}, nil
	})
	if err != nil {
		if checkoutRequest.IdempotencyToken != "" {
			_ = u.OrderService.MarkIdempotencyTokenFailed(ctx, checkoutRequest.IdempotencyToken, err)
		}
		return 0, err
	}
	return orderId, nil

}

func (u *OrderUsecase) validateProducts(ctx context.Context, items []models.CheckoutItem) error {

	for _, item := range items {
		productInfo, err := u.OrderService.GetProductInfo(ctx, item.ProductID)
		if err != nil {
			return err
		}

		if productInfo.Stock < item.Quantity {
			return fmt.Errorf("product stock is not enough for product %d", item.ProductID)
		}

		if item.Quantity <= 0 || item.Quantity > 1000 {
			return fmt.Errorf("quantity must be between 1 and 1000 for product %d", item.ProductID)
		}

		if item.Price != productInfo.Price {
			return fmt.Errorf("price mismatch for product %d", item.ProductID)
		}
	}
	return nil
}

func (u *OrderUsecase) calculateItemSummary(ctx context.Context, items []models.CheckoutItem) (int, float64) {
	var totalQty int
	var totalAmount float64
	for _, item := range items {
		totalAmount += item.Price * float64(item.Quantity)
		totalQty += item.Quantity
	}
	return totalQty, totalAmount
}

func (u *OrderUsecase) constructOrderDetail(ctx context.Context, items []models.CheckoutItem) (string, string) {
	productJson, err := json.Marshal(items)
	if err != nil {
		return "", ""
	}
	history := []map[string]interface{}{
		{
			"status": constant.OrderStatusCreated,
			"time":   time.Now(),
		},
	}
	historyJson, err := json.Marshal(history)
	if err != nil {
		return "", ""
	}
	return string(productJson), string(historyJson)
}

func (u *OrderUsecase) GetOrderHistoryByUserId(ctx context.Context, params models.OrderHistoryParam) ([]models.OrderHistoryResponse, error) {
	results, err := u.OrderService.GetOrderHistoryByUserId(ctx, params)
	if err != nil {
		return nil, err
	}
	return results, nil
}

func (u *OrderUsecase) GetProductInfo(ctx context.Context, productID int64) (models.Product, error) {
	product, err := u.OrderService.GetProductInfo(ctx, productID)
	if err != nil {
		return models.Product{}, err
	}
	return product, nil
}

func (u *OrderUsecase) GetDailySalesReport(ctx context.Context, days int) ([]models.DailySalesReport, error) {
	if days <= 0 {
		days = 30
	}
	return u.OrderService.GetDailySalesReport(ctx, days)
}

func convertCheckoutItemToProductItem(items []models.CheckoutItem) []models.ProductItem {
	var productItems []models.ProductItem
	for _, item := range items {
		productItems = append(productItems, models.ProductItem{
			ProductID: item.ProductID,
			Quantity:  item.Quantity,
		})
	}
	return productItems
}

func hashCheckoutRequest(req *models.CheckoutRequest) (string, error) {
	payload := struct {
		UserID          int64                 `json:"user_id"`
		Items           []models.CheckoutItem `json:"items"`
		PaymentMethod   string                `json:"payment_method"`
		ShippingAddress string                `json:"shipping_address"`
	}{
		UserID:          req.UserID,
		Items:           req.Items,
		PaymentMethod:   req.PaymentMethod,
		ShippingAddress: req.ShippingAddress,
	}
	b, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:]), nil
}
