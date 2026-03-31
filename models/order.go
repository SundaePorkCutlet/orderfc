package models

import "time"

type OrderDetail struct {
	ID           int64  `gorm:"primaryKey;autoIncrement" json:"id"`
	Products     string `gorm:"type:text;not null" json:"products"`
	OrderHistory string `gorm:"type:text;not null" json:"order_history"`
}

type Order struct {
	ID              int64       `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID          int64       `gorm:"type:bigint;not null;index:idx_orders_user_status" json:"user_id"`
	Amount          float64     `gorm:"type:numeric;not null" json:"amount"`
	TotalQty        int         `gorm:"type:integer;not null" json:"total_qty"`
	PaymentMethod   string      `gorm:"type:varchar(50)" json:"payment_method"`
	ShippingAddress string      `gorm:"type:text" json:"shipping_address"`
	Status          int         `gorm:"type:integer;not null;index:idx_orders_user_status;index:idx_orders_status_time" json:"status"`
	OrderDetailID   int64       `gorm:"type:bigint" json:"order_detail_id"`
	OrderDetail     OrderDetail `gorm:"foreignKey:OrderDetailID;constraint:OnDelete:CASCADE" json:"order_detail"`
	CreateTime      time.Time   `gorm:"type:timestamp;default:CURRENT_TIMESTAMP;index:idx_orders_status_time" json:"create_time"`
	UpdateTime      time.Time   `gorm:"type:timestamp;default:CURRENT_TIMESTAMP" json:"update_time"`
}

type DailySalesReport struct {
	SaleDate          string  `json:"sale_date" gorm:"column:sale_date"`
	OrderCount        int     `json:"order_count" gorm:"column:order_count"`
	TotalRevenue      float64 `json:"total_revenue" gorm:"column:total_revenue"`
	AvgOrderValue     float64 `json:"avg_order_value" gorm:"column:avg_order_value"`
	TotalItems        int     `json:"total_items" gorm:"column:total_items"`
	CumulativeRevenue float64 `json:"cumulative_revenue" gorm:"column:cumulative_revenue"`
	RevenueRank       int     `json:"revenue_rank" gorm:"column:revenue_rank"`
}

type OrderRequestLog struct {
	ID               int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	IdempotencyToken string    `gorm:"type:text;unique;not null" json:"idempotency_token"`
	CreateTime       time.Time `gorm:"type:timestamp;default:CURRENT_TIMESTAMP" json:"create_time"`
}

type CheckoutItem struct {
	ProductID int64   `json:"product_id"`
	Quantity  int     `json:"quantity"`
	Price     float64 `json:"price"`
}

type CheckoutRequest struct {
	UserID           int64          `json:"user_id"`
	Items            []CheckoutItem `json:"items"`
	PaymentMethod    string         `json:"payment_method"`
	ShippingAddress  string         `json:"shipping_address"`
	IdempotencyToken string         `json:"idempotency_token"`
}

type OrderHistoryParam struct {
	UserID int64 `json:"user_id"`
	Status int   `json:"status"`
}

type OrderHistoryResponse struct {
	OrderID         int64           `json:"order_id"`
	TotalAmount     float64         `json:"total_amount"`
	TotalQty        int             `json:"total_qty"`
	PaymentMethod   string          `json:"payment_method"`
	ShippingAddress string          `json:"shipping_address"`
	Products        []CheckoutItem  `json:"products"`
	History         []StatusHistory `json:"history"`
	Status          string          `json:"status"`
}

type StatusHistory struct {
	Status    int    `json:"status"`
	Timestamp string `json:"timestamp"`
}

type OrderHistoryResult struct {
	Id              int64 `json:"id" gorm:"column:id"`
	Amount          float64
	TotalQty        int
	Status          int
	PaymentMethod   string
	ShippingAddress string
	Products        string `gorm:"column:products"`
	OrderHistory    string `gorm:"column:order_history"`
}

type OrderCreatedEvent struct {
	OrderID         int64   `json:"order_id"`
	UserID          int64   `json:"user_id"`
	TotalAmount     float64 `json:"total_amount"`
	PaymentMethod   string  `json:"payment_method"`
	ShippingAddress string  `json:"shipping_address"`
}
