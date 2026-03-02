package models

type PaymentUpdateStatusEvent struct {
	OrderID int64  `json:"order_id"`
	Status  string `json:"status"`
}
