package order

import (
	"context"
	"fmt"
	"time"

	"boilerplate-go/infrastructure/logger"
	"boilerplate-go/internal/domain/entity"
	"boilerplate-go/internal/domain/provider"
	"boilerplate-go/internal/domain/repository"
	"boilerplate-go/pkg/errors"
)

type OrderUsecase struct {
	userRepo             repository.UserRepository
	paymentProvider      provider.PaymentProvider
	notificationProvider provider.NotificationProvider
	logger               *logger.Logger
}

func NewOrderUsecase(
	userRepo repository.UserRepository,
	paymentProvider provider.PaymentProvider,
	notificationProvider provider.NotificationProvider,
	logger *logger.Logger,
) *OrderUsecase {
	return &OrderUsecase{
		userRepo:             userRepo,
		paymentProvider:      paymentProvider,
		notificationProvider: notificationProvider,
		logger:               logger,
	}
}

func (u *OrderUsecase) ProcessOrder(ctx context.Context, req *entity.CreateOrderRequest) (*entity.OrderResponse, error) {
	u.logger.WithContext(ctx).WithFields(map[string]interface{}{
		"user_id":   req.UserID,
		"amount":    req.Amount,
		"operation": "process_order",
	}).Info("Processing order")

	// 1. Validate user exists
	user, err := u.userRepo.GetByID(ctx, req.UserID)
	if err != nil {
		if errors.IsUserNotFound(err) {
			return nil, fmt.Errorf("user not found: %w", err)
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	// 2. Create payment intent
	paymentIntentReq := &entity.PaymentIntentRequest{
		Amount:      req.Amount,
		Currency:    req.Currency,
		CustomerID:  fmt.Sprintf("%d", user.ID),
		Description: fmt.Sprintf("Order for user %s", user.Username),
	}

	paymentIntent, err := u.paymentProvider.CreatePaymentIntent(ctx, paymentIntentReq)
	if err != nil {
		u.logger.ErrorLogger(ctx, err, "Failed to create payment intent", map[string]interface{}{
			"user_id": req.UserID,
			"amount":  req.Amount,
		})
		return nil, fmt.Errorf("failed to create payment intent: %w", err)
	}

	// 3. Process payment
	paymentReq := &entity.PaymentRequest{
		OrderID:     req.OrderID,
		Amount:      req.Amount,
		Currency:    req.Currency,
		Description: fmt.Sprintf("Order %s for %s", req.OrderID, user.Username),
		CustomerID:  fmt.Sprintf("%d", user.ID),
		Metadata: map[string]interface{}{
			"user_id":  user.ID,
			"username": user.Username,
			"order_id": req.OrderID,
		},
	}

	payment, err := u.paymentProvider.ProcessPayment(ctx, paymentReq)
	if err != nil {
		u.logger.ErrorLogger(ctx, err, "Payment processing failed", map[string]interface{}{
			"user_id":  req.UserID,
			"order_id": req.OrderID,
		})

		// Send failure notification
		go u.sendPaymentFailureNotification(context.Background(), user, req.OrderID, err)

		return nil, fmt.Errorf("payment processing failed: %w", err)
	}

	// 4. Send success notification
	go u.sendOrderConfirmationNotification(context.Background(), user, req.OrderID, payment.ID, req.Amount)

	u.logger.WithContext(ctx).WithFields(map[string]interface{}{
		"user_id":    req.UserID,
		"order_id":   req.OrderID,
		"payment_id": payment.ID,
		"amount":     req.Amount,
	}).Info("Order processed successfully")

	// 5. Return order response
	orderResponse := &entity.OrderResponse{
		OrderID:         req.OrderID,
		PaymentID:       payment.ID,
		PaymentIntentID: paymentIntent.ID,
		Status:          "completed",
		Amount:          req.Amount,
		Currency:        req.Currency,
		ProcessedAt:     time.Now(),
		User:            user,
	}

	return orderResponse, nil
}

func (u *OrderUsecase) GetPaymentStatus(ctx context.Context, paymentID string) (*entity.PaymentStatus, error) {
	u.logger.WithContext(ctx).WithFields(map[string]interface{}{
		"payment_id": paymentID,
		"operation":  "get_payment_status",
	}).Info("Getting payment status")

	status, err := u.paymentProvider.GetPaymentStatus(ctx, paymentID)
	if err != nil {
		u.logger.ErrorLogger(ctx, err, "Failed to get payment status", map[string]interface{}{
			"payment_id": paymentID,
		})
		return nil, fmt.Errorf("failed to get payment status: %w", err)
	}

	return status, nil
}

func (u *OrderUsecase) RefundOrder(ctx context.Context, req *entity.RefundOrderRequest) (*entity.RefundResponse, error) {
	u.logger.WithContext(ctx).WithFields(map[string]interface{}{
		"payment_id": req.PaymentID,
		"user_id":    req.UserID,
		"operation":  "refund_order",
	}).Info("Processing refund")

	// 1. Validate user exists
	user, err := u.userRepo.GetByID(ctx, req.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	// 2. Process refund
	refund, err := u.paymentProvider.RefundPayment(ctx, req.PaymentID)
	if err != nil {
		u.logger.ErrorLogger(ctx, err, "Refund processing failed", map[string]interface{}{
			"payment_id": req.PaymentID,
			"user_id":    req.UserID,
		})
		return nil, fmt.Errorf("refund processing failed: %w", err)
	}

	// 3. Send refund notification
	go u.sendRefundNotification(context.Background(), user, req.PaymentID, refund.ID)

	u.logger.WithContext(ctx).WithFields(map[string]interface{}{
		"payment_id": req.PaymentID,
		"refund_id":  refund.ID,
		"user_id":    req.UserID,
	}).Info("Refund processed successfully")

	return refund, nil
}

// Private helper methods for notifications
func (u *OrderUsecase) sendOrderConfirmationNotification(ctx context.Context, user *entity.User, orderID, paymentID string, amount float64) {
	emailReq := &entity.EmailRequest{
		To:      []string{user.Email},
		Subject: "Order Confirmation",
		Body: fmt.Sprintf(`
Hello %s,

Your order has been confirmed!

Order Details:
- Order ID: %s
- Payment ID: %s
- Amount: $%.2f
- Status: Completed

Thank you for your business!

Best regards,
Boilerplate Team
		`, user.Username, orderID, paymentID, amount),
		Metadata: map[string]interface{}{
			"user_id":    user.ID,
			"order_id":   orderID,
			"payment_id": paymentID,
			"type":       "order_confirmation",
		},
	}

	if _, err := u.notificationProvider.SendEmail(ctx, emailReq); err != nil {
		u.logger.ErrorLogger(ctx, err, "Failed to send order confirmation email", map[string]interface{}{
			"user_id":  user.ID,
			"order_id": orderID,
		})
	}
}

func (u *OrderUsecase) sendPaymentFailureNotification(ctx context.Context, user *entity.User, orderID string, paymentErr error) {
	emailReq := &entity.EmailRequest{
		To:      []string{user.Email},
		Subject: "Payment Failed",
		Body: fmt.Sprintf(`
Hello %s,

We encountered an issue processing your payment for order %s.

Please try again or contact our support team.

Error: %s

Best regards,
Boilerplate Team
		`, user.Username, orderID, paymentErr.Error()),
		Metadata: map[string]interface{}{
			"user_id":  user.ID,
			"order_id": orderID,
			"type":     "payment_failure",
		},
	}

	if _, err := u.notificationProvider.SendEmail(ctx, emailReq); err != nil {
		u.logger.ErrorLogger(ctx, err, "Failed to send payment failure email", map[string]interface{}{
			"user_id":  user.ID,
			"order_id": orderID,
		})
	}
}

func (u *OrderUsecase) sendRefundNotification(ctx context.Context, user *entity.User, paymentID, refundID string) {
	emailReq := &entity.EmailRequest{
		To:      []string{user.Email},
		Subject: "Refund Processed",
		Body: fmt.Sprintf(`
Hello %s,

Your refund has been processed successfully.

Refund Details:
- Original Payment ID: %s
- Refund ID: %s

The refund will appear in your account within 3-5 business days.

Best regards,
Boilerplate Team
		`, user.Username, paymentID, refundID),
		Metadata: map[string]interface{}{
			"user_id":    user.ID,
			"payment_id": paymentID,
			"refund_id":  refundID,
			"type":       "refund_confirmation",
		},
	}

	if _, err := u.notificationProvider.SendEmail(ctx, emailReq); err != nil {
		u.logger.ErrorLogger(ctx, err, "Failed to send refund notification email", map[string]interface{}{
			"user_id":    user.ID,
			"payment_id": paymentID,
		})
	}
}
