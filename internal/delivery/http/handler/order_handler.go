package handler

import (
	"net/http"
	"strconv"

	"boilerplate-go/infrastructure/logger"
	"boilerplate-go/infrastructure/metrics"
	"boilerplate-go/internal/domain/entity"
	"boilerplate-go/internal/usecase/order"
	"boilerplate-go/pkg/response"

	"github.com/gin-gonic/gin"
)

type OrderHandler struct {
	orderUsecase *order.OrderUsecase
	logger       *logger.Logger
	metrics      *metrics.Metrics
}

func NewOrderHandler(orderUsecase *order.OrderUsecase, logger *logger.Logger, metrics *metrics.Metrics) *OrderHandler {
	return &OrderHandler{
		orderUsecase: orderUsecase,
		logger:       logger,
		metrics:      metrics,
	}
}

// ProcessOrder godoc
// @Summary Process a new order
// @Description Process a new order with payment
// @Tags orders
// @Accept json
// @Produce json
// @Param request body entity.CreateOrderRequest true "Order request"
// @Success 200 {object} response.Response{data=entity.OrderResponse}
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Security BearerAuth
// @Router /orders [post]
func (h *OrderHandler) ProcessOrder(c *gin.Context) {
	var req entity.CreateOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.ErrorLogger(c.Request.Context(), err, "Invalid order request", map[string]interface{}{
			"endpoint": "/orders",
			"method":   "POST",
		})
		response.BadRequest(c, "Invalid request format", err.Error())
		return
	}

	// Get user ID from JWT context
	userID, exists := c.Get("user_id")
	if !exists {
		h.logger.WithContext(c.Request.Context()).Error("User ID not found in context")
		response.Unauthorized(c, "Authentication required", "user_id not found in token")
		return
	}

	req.UserID = userID.(int)

	// Process the order
	orderResponse, err := h.orderUsecase.ProcessOrder(c.Request.Context(), &req)
	if err != nil {
		h.metrics.IncrementCounter("order_processing_failures")
		h.logger.ErrorLogger(c.Request.Context(), err, "Failed to process order", map[string]interface{}{
			"user_id":  req.UserID,
			"order_id": req.OrderID,
			"amount":   req.Amount,
		})
		response.InternalServerError(c, "Failed to process order", err.Error())
		return
	}

	h.metrics.IncrementCounter("order_processing_success")
	h.logger.WithContext(c.Request.Context()).WithFields(map[string]interface{}{
		"user_id":    req.UserID,
		"order_id":   req.OrderID,
		"payment_id": orderResponse.PaymentID,
	}).Info("Order processed successfully")

	response.Success(c, http.StatusOK, "Order processed successfully", orderResponse)
}

// GetPaymentStatus godoc
// @Summary Get payment status
// @Description Get the status of a payment by payment ID
// @Tags orders
// @Accept json
// @Produce json
// @Param payment_id path string true "Payment ID"
// @Success 200 {object} response.Response{data=entity.PaymentStatus}
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Security BearerAuth
// @Router /orders/payment/{payment_id}/status [get]
func (h *OrderHandler) GetPaymentStatus(c *gin.Context) {
	paymentID := c.Param("payment_id")
	if paymentID == "" {
		response.BadRequest(c, "Payment ID is required", "payment_id parameter is missing")
		return
	}

	status, err := h.orderUsecase.GetPaymentStatus(c.Request.Context(), paymentID)
	if err != nil {
		h.logger.ErrorLogger(c.Request.Context(), err, "Failed to get payment status", map[string]interface{}{
			"payment_id": paymentID,
		})
		response.InternalServerError(c, "Failed to get payment status", err.Error())
		return
	}

	h.logger.WithContext(c.Request.Context()).WithFields(map[string]interface{}{
		"payment_id": paymentID,
		"status":     status.Status,
	}).Info("Payment status retrieved successfully")

	response.Success(c, http.StatusOK, "Payment status retrieved", status)
}

// RefundOrder godoc
// @Summary Refund an order
// @Description Process a refund for an order
// @Tags orders
// @Accept json
// @Produce json
// @Param request body entity.RefundOrderRequest true "Refund request"
// @Success 200 {object} response.Response{data=entity.RefundResponse}
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Security BearerAuth
// @Router /orders/refund [post]
func (h *OrderHandler) RefundOrder(c *gin.Context) {
	var req entity.RefundOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.ErrorLogger(c.Request.Context(), err, "Invalid refund request", map[string]interface{}{
			"endpoint": "/orders/refund",
			"method":   "POST",
		})
		response.BadRequest(c, "Invalid request format", err.Error())
		return
	}

	// Get user ID from JWT context
	userID, exists := c.Get("user_id")
	if !exists {
		response.Unauthorized(c, "Authentication required", "user_id not found in token")
		return
	}

	req.UserID = userID.(int)

	// Process the refund
	refundResponse, err := h.orderUsecase.RefundOrder(c.Request.Context(), &req)
	if err != nil {
		h.metrics.IncrementCounter("order_refund_failures")
		h.logger.ErrorLogger(c.Request.Context(), err, "Failed to process refund", map[string]interface{}{
			"user_id":    req.UserID,
			"payment_id": req.PaymentID,
		})
		response.InternalServerError(c, "Failed to process refund", err.Error())
		return
	}

	h.metrics.IncrementCounter("order_refund_success")
	h.logger.WithContext(c.Request.Context()).WithFields(map[string]interface{}{
		"user_id":    req.UserID,
		"payment_id": req.PaymentID,
		"refund_id":  refundResponse.ID,
	}).Info("Refund processed successfully")

	response.Success(c, http.StatusOK, "Refund processed successfully", refundResponse)
}

// CreatePaymentIntent godoc
// @Summary Create payment intent
// @Description Create a payment intent for client-side payment processing
// @Tags orders
// @Accept json
// @Produce json
// @Param request body entity.PaymentIntentRequest true "Payment intent request"
// @Success 200 {object} response.Response{data=entity.PaymentIntent}
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Security BearerAuth
// @Router /orders/payment-intent [post]
func (h *OrderHandler) CreatePaymentIntent(c *gin.Context) {
	var req entity.PaymentIntentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.ErrorLogger(c.Request.Context(), err, "Invalid payment intent request", map[string]interface{}{
			"endpoint": "/orders/payment-intent",
			"method":   "POST",
		})
		response.BadRequest(c, "Invalid request format", err.Error())
		return
	}

	// Get user ID from JWT context
	userID, exists := c.Get("user_id")
	if !exists {
		response.Unauthorized(c, "Authentication required", "user_id not found in token")
		return
	}

	req.CustomerID = strconv.Itoa(userID.(int))

	// Create payment intent (this would typically go through a use case)
	// For demonstration, we'll call the provider directly
	// In real implementation, this should go through a use case
	response.Success(c, http.StatusOK, "Payment intent creation not fully implemented", map[string]string{
		"message":     "This endpoint needs to be connected to the payment use case",
		"customer_id": req.CustomerID,
		"amount":      strconv.FormatFloat(req.Amount, 'f', 2, 64),
	})
}
