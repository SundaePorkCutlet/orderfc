package repository

import (
	"context"
	"encoding/json"
	"orderfc/infrastructure/constant"
	"orderfc/infrastructure/log"
	"orderfc/models"
	"time"

	"gorm.io/gorm"
)

func (r *OrderRepository) WithTransaction(ctx context.Context, fn func(tx *gorm.DB) error) error {
	tx := r.Database.WithContext(ctx).Begin()

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		}
	}()

	if err := fn(tx); err != nil {
		tx.Rollback()
		return err
	}

	if err := tx.Commit().Error; err != nil {
		return err
	}

	return nil
}

func (r *OrderRepository) InsertOrderTx(ctx context.Context, tx *gorm.DB, order *models.Order) error {
	err := tx.WithContext(ctx).Create(order).Error
	return err
}

func (r *OrderRepository) InsertOrderDetailTx(ctx context.Context, tx *gorm.DB, orderDetail *models.OrderDetail) error {
	err := tx.WithContext(ctx).Create(orderDetail).Table("order_details").Error
	return err
}

func (r *OrderRepository) CheckIdempotencyToken(ctx context.Context, idempotencyToken string) (bool, error) {
	var log models.OrderRequestLog
	err := r.Database.WithContext(ctx).Table("order_request_logs").First(&log, "idempotency_token = ?", idempotencyToken).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (r *OrderRepository) SaveIdempotencyToken(ctx context.Context, idempotencyToken string) error {
	log := models.OrderRequestLog{
		IdempotencyToken: idempotencyToken,
		CreateTime:       time.Now(),
	}

	err := r.Database.WithContext(ctx).Table("order_request_logs").Create(&log).Error
	if err != nil {
		return err
	}
	return nil

}

func (r *OrderRepository) GetOrderHistoryByUserId(ctx context.Context, params models.OrderHistoryParam) ([]models.OrderHistoryResponse, error) {
	var queryResults []models.OrderHistoryResult
	query := r.Database.WithContext(ctx).Table("orders").
		Select("orders.*, order_details.products, order_details.order_history").
		Joins("JOIN order_details ON orders.order_detail_id = order_details.id").
		Where("user_id = ?", params.UserID)

	if params.Status > 0 {
		query = query.Where("status = ?", params.Status)
	}
	err := query.Order("orders.id DESC").Scan(&queryResults).Error
	if err != nil {
		return nil, err
	}
	var results []models.OrderHistoryResponse
	for _, result := range queryResults {
		var products []models.CheckoutItem
		var orderHistory []models.StatusHistory
		err := json.Unmarshal([]byte(result.Products), &products)
		if err != nil {
			log.Logger.Info().Err(err).Msg("Error unmarshalling products")
			return nil, err
		}
		err = json.Unmarshal([]byte(result.OrderHistory), &orderHistory)
		if err != nil {
			log.Logger.Info().Err(err).Msg("Error unmarshalling order history")
			return nil, err
		}
		results = append(results, models.OrderHistoryResponse{
			OrderID:         result.Id,
			TotalAmount:     result.Amount,
			TotalQty:        result.TotalQty,
			PaymentMethod:   result.PaymentMethod,
			ShippingAddress: result.ShippingAddress,
			Products:        products,
			History:         orderHistory,
			Status:          constant.OrderStatusMap[result.Status],
		})
	}

	return results, nil
}

func (r *OrderRepository) UpdateOrderStatus(ctx context.Context, orderID int64, status int) error {
	err := r.Database.WithContext(ctx).Table("orders").Model(&models.Order{}).Where("id = ?", orderID).Updates(map[string]interface{}{
		"status":      status,
		"update_time": time.Now(),
	}).Error
	if err != nil {
		log.Logger.Error().Err(err).Int64("order_id", orderID).Msg("Failed to update order status")
		return err
	}
	return nil
}

func (r *OrderRepository) GetOrderInfoByOrderID(ctx context.Context, orderID int64) (*models.Order, error) {
	var order models.Order
	err := r.Database.WithContext(ctx).Table("orders").Where("id = ?", orderID).First(&order).Error
	if err != nil {
		return nil, err
	}
	return &order, nil
}

func (r *OrderRepository) GetOrderDetailByID(ctx context.Context, orderDetailID int64) (*models.OrderDetail, error) {
	var orderDetail models.OrderDetail
	err := r.Database.WithContext(ctx).Table("order_details").Where("id = ?", orderDetailID).First(&orderDetail).Error
	if err != nil {
		return nil, err
	}
	return &orderDetail, nil
}

func (r *OrderRepository) GetDailySalesReport(ctx context.Context, days int) ([]models.DailySalesReport, error) {
	var results []models.DailySalesReport
	query := `
		WITH daily_sales AS (
			SELECT
				DATE(create_time) as sale_date,
				COUNT(*) as order_count,
				COALESCE(SUM(amount), 0) as total_revenue,
				COALESCE(AVG(amount), 0) as avg_order_value,
				COALESCE(SUM(total_qty), 0) as total_items
			FROM orders
			WHERE create_time >= NOW() - INTERVAL '1 day' * ?
			GROUP BY DATE(create_time)
		)
		SELECT
			TO_CHAR(sale_date, 'YYYY-MM-DD') as sale_date,
			order_count,
			ROUND(total_revenue::numeric, 2) as total_revenue,
			ROUND(avg_order_value::numeric, 2) as avg_order_value,
			total_items,
			ROUND(SUM(total_revenue) OVER (ORDER BY sale_date)::numeric, 2) as cumulative_revenue,
			ROW_NUMBER() OVER (ORDER BY total_revenue DESC) as revenue_rank
		FROM daily_sales
		ORDER BY sale_date DESC
	`
	err := r.Database.WithContext(ctx).Raw(query, days).Scan(&results).Error
	return results, err
}
