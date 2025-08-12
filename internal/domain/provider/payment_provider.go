package provider

import (
	"boilerplate-go/internal/domain/entity"
	"context"
)

// PaymentProvider defines the contract for payment operations
type PaymentProvider interface {
	ProcessPayment(ctx context.Context, req *entity.PaymentRequest) (*entity.PaymentResponse, error)
	RefundPayment(ctx context.Context, paymentID string) (*entity.RefundResponse, error)
	GetPaymentStatus(ctx context.Context, paymentID string) (*entity.PaymentStatus, error)
	CreatePaymentIntent(ctx context.Context, req *entity.PaymentIntentRequest) (*entity.PaymentIntent, error)
}
