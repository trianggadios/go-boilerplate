package entity

import "time"

// Order related entities for use case integration
type CreateOrderRequest struct {
	OrderID   string  `json:"order_id" binding:"required"`
	UserID    int     `json:"user_id" binding:"required"`
	Amount    float64 `json:"amount" binding:"required,gt=0"`
	Currency  string  `json:"currency" binding:"required"`
	UserEmail string  `json:"user_email" binding:"required,email"`
}

type OrderResponse struct {
	OrderID         string    `json:"order_id"`
	PaymentID       string    `json:"payment_id"`
	PaymentIntentID string    `json:"payment_intent_id"`
	Status          string    `json:"status"`
	Amount          float64   `json:"amount"`
	Currency        string    `json:"currency"`
	ProcessedAt     time.Time `json:"processed_at"`
	User            *User     `json:"user"`
}

type RefundOrderRequest struct {
	PaymentID string `json:"payment_id" binding:"required"`
	UserID    int    `json:"user_id" binding:"required"`
	Reason    string `json:"reason,omitempty"`
}
