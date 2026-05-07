package service

import (
	"context"
	"orderfc/cmd/order/repository"
	"orderfc/models"

	"gorm.io/gorm"
)

type OrderService struct {
	OrderRepo repository.OrderRepository
}

func NewOrderService(orderRepo repository.OrderRepository) *OrderService {
	return &OrderService{OrderRepo: orderRepo}
}

func (s *OrderService) CheckIdempotencyToken(ctx context.Context, idempotencyToken string) (bool, error) {
	return s.OrderRepo.CheckIdempotencyToken(ctx, idempotencyToken)
}

func (s *OrderService) SaveIdempotencyToken(ctx context.Context, idempotencyToken string) error {
	return s.OrderRepo.SaveIdempotencyToken(ctx, idempotencyToken)
}

func (s *OrderService) ReserveIdempotencyToken(ctx context.Context, idempotencyToken, requestHash string) (*models.OrderRequestLog, bool, error) {
	return s.OrderRepo.ReserveIdempotencyToken(ctx, idempotencyToken, requestHash)
}

func (s *OrderService) MarkIdempotencyTokenFailed(ctx context.Context, idempotencyToken string, processErr error) error {
	return s.OrderRepo.MarkIdempotencyTokenFailed(ctx, idempotencyToken, processErr)
}

func (s *OrderService) SaveOrderAndOrderDetail(ctx context.Context, order *models.Order, orderDetail *models.OrderDetail) (int64, error) {
	return s.SaveOrderAndOrderDetailWithOutbox(ctx, order, orderDetail, nil)
}

func (s *OrderService) SaveOrderAndOrderDetailWithOutbox(
	ctx context.Context,
	order *models.Order,
	orderDetail *models.OrderDetail,
	buildEvents func(orderID int64) ([]models.OrderOutboxEvent, error),
) (int64, error) {
	var orderId int64
	err := s.OrderRepo.WithTransaction(ctx, func(tx *gorm.DB) error {
		err := s.OrderRepo.InsertOrderDetailTx(ctx, tx, orderDetail)
		if err != nil {
			return err
		}

		order.OrderDetailID = orderDetail.ID
		err = s.OrderRepo.InsertOrderTx(ctx, tx, order)
		if err != nil {
			return err
		}
		orderId = order.ID
		if buildEvents != nil {
			events, err := buildEvents(orderId)
			if err != nil {
				return err
			}
			if err := s.OrderRepo.InsertOrderOutboxEventsTx(ctx, tx, events); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return 0, err
	}
	return orderId, nil
}

func (s *OrderService) SaveOrderAndOrderDetailWithOutboxAndIdempotency(
	ctx context.Context,
	order *models.Order,
	orderDetail *models.OrderDetail,
	idempotencyToken string,
	buildEvents func(orderID int64) ([]models.OrderOutboxEvent, error),
) (int64, error) {
	var orderId int64
	err := s.OrderRepo.WithTransaction(ctx, func(tx *gorm.DB) error {
		if err := s.OrderRepo.InsertOrderDetailTx(ctx, tx, orderDetail); err != nil {
			return err
		}

		order.OrderDetailID = orderDetail.ID
		if err := s.OrderRepo.InsertOrderTx(ctx, tx, order); err != nil {
			return err
		}
		orderId = order.ID

		if buildEvents != nil {
			events, err := buildEvents(orderId)
			if err != nil {
				return err
			}
			if err := s.OrderRepo.InsertOrderOutboxEventsTx(ctx, tx, events); err != nil {
				return err
			}
		}

		if idempotencyToken != "" {
			if err := s.OrderRepo.MarkIdempotencyTokenSucceededTx(ctx, tx, idempotencyToken, orderId); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return 0, err
	}
	return orderId, nil
}

func (s *OrderService) GetOrderHistoryByUserId(ctx context.Context, params models.OrderHistoryParam) ([]models.OrderHistoryResponse, error) {
	results, err := s.OrderRepo.GetOrderHistoryByUserId(ctx, params)
	if err != nil {
		return nil, err
	}
	return results, nil
}

func (s *OrderService) GetProductInfo(ctx context.Context, productID int64) (models.Product, error) {
	product, err := s.OrderRepo.GetProductInfo(ctx, productID)
	if err != nil {
		return models.Product{}, err
	}
	return product, nil
}

func (s *OrderService) UpdateOrderStatus(ctx context.Context, orderID int64, status int) error {
	err := s.OrderRepo.UpdateOrderStatus(ctx, orderID, status)
	if err != nil {
		return err
	}
	return nil
}

func (s *OrderService) GetOrderInfoByOrderID(ctx context.Context, orderID int64) (*models.Order, error) {
	order, err := s.OrderRepo.GetOrderInfoByOrderID(ctx, orderID)
	if err != nil {
		return nil, err
	}
	return order, nil
}

func (s *OrderService) GetOrderDetailByID(ctx context.Context, orderDetailID int64) (*models.OrderDetail, error) {
	orderDetail, err := s.OrderRepo.GetOrderDetailByID(ctx, orderDetailID)
	if err != nil {
		return nil, err
	}
	return orderDetail, nil
}

func (s *OrderService) GetDailySalesReport(ctx context.Context, days int) ([]models.DailySalesReport, error) {
	return s.OrderRepo.GetDailySalesReport(ctx, days)
}
