package payment

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"boilerplate-go/infrastructure/logger"
	"boilerplate-go/internal/domain/entity"
	"boilerplate-go/internal/domain/provider"
)

type StripeProvider struct {
	httpClient *http.Client
	baseURL    string
	apiKey     string
	logger     *logger.Logger
}

type StripeConfig struct {
	BaseURL string
	APIKey  string
	Timeout time.Duration
}

func NewStripeProvider(config StripeConfig, logger *logger.Logger) provider.PaymentProvider {
	timeout := config.Timeout
	if timeout == 0 {
		timeout = 30 * time.Second
	}

	return &StripeProvider{
		httpClient: &http.Client{
			Timeout: timeout,
		},
		baseURL: config.BaseURL,
		apiKey:  config.APIKey,
		logger:  logger,
	}
}

func (s *StripeProvider) ProcessPayment(ctx context.Context, req *entity.PaymentRequest) (*entity.PaymentResponse, error) {
	s.logger.WithContext(ctx).WithFields(map[string]interface{}{
		"provider":  "stripe",
		"amount":    req.Amount,
		"currency":  req.Currency,
		"order_id":  req.OrderID,
		"operation": "process_payment",
	}).Info("Processing payment")

	// Prepare Stripe payment request
	stripeReq := map[string]interface{}{
		"amount":      int(req.Amount * 100), // Convert to cents
		"currency":    req.Currency,
		"description": req.Description,
		"metadata":    req.Metadata,
	}

	if req.CustomerID != "" {
		stripeReq["customer"] = req.CustomerID
	}

	jsonData, err := json.Marshal(stripeReq)
	if err != nil {
		return nil, s.handleError(ctx, err, "json_marshal_failed")
	}

	// Create HTTP request
	url := fmt.Sprintf("%s/charges", s.baseURL)
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, s.handleError(ctx, err, "create_request_failed")
	}

	s.setHeaders(httpReq)

	// Execute request
	resp, err := s.httpClient.Do(httpReq)
	if err != nil {
		return nil, s.handleError(ctx, err, "api_call_failed")
	}
	defer resp.Body.Close()

	// Parse response
	return s.parsePaymentResponse(ctx, resp)
}

func (s *StripeProvider) RefundPayment(ctx context.Context, paymentID string) (*entity.RefundResponse, error) {
	s.logger.WithContext(ctx).WithFields(map[string]interface{}{
		"provider":   "stripe",
		"payment_id": paymentID,
		"operation":  "refund_payment",
	}).Info("Processing refund")

	refundReq := map[string]interface{}{
		"charge": paymentID,
	}

	jsonData, err := json.Marshal(refundReq)
	if err != nil {
		return nil, s.handleError(ctx, err, "json_marshal_failed")
	}

	url := fmt.Sprintf("%s/refunds", s.baseURL)
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, s.handleError(ctx, err, "create_request_failed")
	}

	s.setHeaders(httpReq)

	resp, err := s.httpClient.Do(httpReq)
	if err != nil {
		return nil, s.handleError(ctx, err, "api_call_failed")
	}
	defer resp.Body.Close()

	return s.parseRefundResponse(ctx, resp)
}

func (s *StripeProvider) GetPaymentStatus(ctx context.Context, paymentID string) (*entity.PaymentStatus, error) {
	s.logger.WithContext(ctx).WithFields(map[string]interface{}{
		"provider":   "stripe",
		"payment_id": paymentID,
		"operation":  "get_payment_status",
	}).Info("Getting payment status")

	url := fmt.Sprintf("%s/charges/%s", s.baseURL, paymentID)
	httpReq, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, s.handleError(ctx, err, "create_request_failed")
	}

	s.setHeaders(httpReq)

	resp, err := s.httpClient.Do(httpReq)
	if err != nil {
		return nil, s.handleError(ctx, err, "api_call_failed")
	}
	defer resp.Body.Close()

	return s.parsePaymentStatusResponse(ctx, resp)
}

func (s *StripeProvider) CreatePaymentIntent(ctx context.Context, req *entity.PaymentIntentRequest) (*entity.PaymentIntent, error) {
	s.logger.WithContext(ctx).WithFields(map[string]interface{}{
		"provider":    "stripe",
		"amount":      req.Amount,
		"currency":    req.Currency,
		"customer_id": req.CustomerID,
		"operation":   "create_payment_intent",
	}).Info("Creating payment intent")

	intentReq := map[string]interface{}{
		"amount":      int(req.Amount * 100), // Convert to cents
		"currency":    req.Currency,
		"description": req.Description,
	}

	if req.CustomerID != "" {
		intentReq["customer"] = req.CustomerID
	}

	jsonData, err := json.Marshal(intentReq)
	if err != nil {
		return nil, s.handleError(ctx, err, "json_marshal_failed")
	}

	url := fmt.Sprintf("%s/payment_intents", s.baseURL)
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, s.handleError(ctx, err, "create_request_failed")
	}

	s.setHeaders(httpReq)

	resp, err := s.httpClient.Do(httpReq)
	if err != nil {
		return nil, s.handleError(ctx, err, "api_call_failed")
	}
	defer resp.Body.Close()

	return s.parsePaymentIntentResponse(ctx, resp)
}

func (s *StripeProvider) setHeaders(req *http.Request) {
	req.Header.Set("Authorization", "Bearer "+s.apiKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "boilerplate-go/1.0")
}

func (s *StripeProvider) handleError(ctx context.Context, err error, operation string) error {
	s.logger.ErrorLogger(ctx, err, "Stripe operation failed", map[string]interface{}{
		"provider":  "stripe",
		"operation": operation,
	})
	return fmt.Errorf("stripe %s: %w", operation, err)
}

func (s *StripeProvider) parsePaymentResponse(ctx context.Context, resp *http.Response) (*entity.PaymentResponse, error) {
	var stripeResp map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&stripeResp); err != nil {
		return nil, s.handleError(ctx, err, "parse_response_failed")
	}

	if resp.StatusCode != http.StatusOK {
		err := fmt.Errorf("stripe API error: %d", resp.StatusCode)
		return nil, s.handleError(ctx, err, "api_error")
	}

	paymentResp := &entity.PaymentResponse{
		ID:            stripeResp["id"].(string),
		Status:        stripeResp["status"].(string),
		Amount:        float64(stripeResp["amount"].(float64)) / 100, // Convert from cents
		Currency:      stripeResp["currency"].(string),
		TransactionID: stripeResp["balance_transaction"].(string),
		CreatedAt:     time.Unix(int64(stripeResp["created"].(float64)), 0),
	}

	if metadata, ok := stripeResp["metadata"].(map[string]interface{}); ok {
		paymentResp.Metadata = metadata
	}

	s.logger.WithContext(ctx).WithFields(map[string]interface{}{
		"payment_id": paymentResp.ID,
		"status":     paymentResp.Status,
		"amount":     paymentResp.Amount,
	}).Info("Payment processed successfully")

	return paymentResp, nil
}

func (s *StripeProvider) parseRefundResponse(ctx context.Context, resp *http.Response) (*entity.RefundResponse, error) {
	var stripeResp map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&stripeResp); err != nil {
		return nil, s.handleError(ctx, err, "parse_response_failed")
	}

	if resp.StatusCode != http.StatusOK {
		err := fmt.Errorf("stripe API error: %d", resp.StatusCode)
		return nil, s.handleError(ctx, err, "api_error")
	}

	refundResp := &entity.RefundResponse{
		ID:        stripeResp["id"].(string),
		PaymentID: stripeResp["charge"].(string),
		Amount:    float64(stripeResp["amount"].(float64)) / 100,
		Status:    stripeResp["status"].(string),
		CreatedAt: time.Unix(int64(stripeResp["created"].(float64)), 0),
	}

	return refundResp, nil
}

func (s *StripeProvider) parsePaymentStatusResponse(ctx context.Context, resp *http.Response) (*entity.PaymentStatus, error) {
	var stripeResp map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&stripeResp); err != nil {
		return nil, s.handleError(ctx, err, "parse_response_failed")
	}

	if resp.StatusCode != http.StatusOK {
		err := fmt.Errorf("stripe API error: %d", resp.StatusCode)
		return nil, s.handleError(ctx, err, "api_error")
	}

	statusResp := &entity.PaymentStatus{
		ID:        stripeResp["id"].(string),
		Status:    stripeResp["status"].(string),
		Amount:    float64(stripeResp["amount"].(float64)) / 100,
		UpdatedAt: time.Now(),
	}

	return statusResp, nil
}

func (s *StripeProvider) parsePaymentIntentResponse(ctx context.Context, resp *http.Response) (*entity.PaymentIntent, error) {
	var stripeResp map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&stripeResp); err != nil {
		return nil, s.handleError(ctx, err, "parse_response_failed")
	}

	if resp.StatusCode != http.StatusOK {
		err := fmt.Errorf("stripe API error: %d", resp.StatusCode)
		return nil, s.handleError(ctx, err, "api_error")
	}

	intentResp := &entity.PaymentIntent{
		ID:           stripeResp["id"].(string),
		ClientSecret: stripeResp["client_secret"].(string),
		Status:       stripeResp["status"].(string),
	}

	return intentResp, nil
}
