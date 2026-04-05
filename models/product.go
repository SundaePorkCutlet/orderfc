package models

import "time"

type ProductInfo struct {
	Product Product `json:"product"`
}

type Product struct {
	ID          int64   `json:"id"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Price       float64 `json:"price"`
	Stock       int     `json:"stock"`
	CategoryID  int     `json:"category_id"`
}

// ProductStockUpdatedEvent — stock.updated / stock.rollback 발행에 공통 필드 (스키마 v1).
type ProductStockUpdatedEvent struct {
	SchemaVersion int           `json:"schema_version"` // 1 = 현재 필드 집합
	OrderID       int64         `json:"order_id"`
	UserID        int64         `json:"user_id"` // 파티션 키·순서 보장용 (동일 유저 주문 동일 파티션)
	Products      []ProductItem `json:"products"`
	EventTime     time.Time     `json:"event_time"`
}

type ProductItem struct {
	ProductID int64 `json:"product_id"`
	Quantity  int   `json:"quantity"`
}
