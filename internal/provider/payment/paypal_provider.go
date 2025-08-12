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

type PayPalProvider struct {
	httpClient   *http.Client
	baseURL      string
	clientID     string
	clientSecret string
	logger       *logger.Logger
	accessToken  string
	tokenExpiry  time.Time
}

type PayPalConfig struct {
	BaseURL      string
	ClientID     string
	ClientSecret string
	Timeout      time.Duration
}

func NewPayPalProvider(config PayPalConfig, logger *logger.Logger) provider.PaymentProvider {
	timeout := config.Timeout
	if timeout == 0 {
		timeout = 30 * time.Second
	}

	return &PayPalProvider{
		httpClient: &http.Client{
			Timeout: timeout,
		},
		baseURL:      config.BaseURL,
		clientID:     config.ClientID,
		clientSecret: config.ClientSecret,
		logger:       logger,
	}
}

func (p *PayPalProvider) ProcessPayment(ctx context.Context, req *entity.PaymentRequest) (*entity.PaymentResponse, error) {
	p.logger.WithContext(ctx).WithFields(map[string]interface{}{
		"provider":  "paypal",
		"amount":    req.Amount,
		"currency":  req.Currency,
		"order_id":  req.OrderID,
		"operation": "process_payment",
	}).Info("Processing payment")

	// Ensure we have a valid access token
	if err := p.ensureValidToken(ctx); err != nil {
		return nil, p.handleError(ctx, err, "token_refresh_failed")
	}

	// Create PayPal order
	orderReq := map[string]interface{}{
		"intent": "CAPTURE",
		"purchase_units": []map[string]interface{}{
			{
				"amount": map[string]interface{}{
					"currency_code": req.Currency,
					"value":         fmt.Sprintf("%.2f", req.Amount),
				},
				"description":  req.Description,
				"reference_id": req.OrderID,
			},
		},
	}

	jsonData, err := json.Marshal(orderReq)
	if err != nil {
		return nil, p.handleError(ctx, err, "json_marshal_failed")
	}

	// Create order
	url := fmt.Sprintf("%s/v2/checkout/orders", p.baseURL)
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, p.handleError(ctx, err, "create_request_failed")
	}

	p.setHeaders(httpReq)

	resp, err := p.httpClient.Do(httpReq)
	if err != nil {
		return nil, p.handleError(ctx, err, "api_call_failed")
	}
	defer resp.Body.Close()

	orderResp, err := p.parseOrderResponse(ctx, resp)
	if err != nil {
		return nil, err
	}

	// Capture the order (for demonstration, in real scenario this would be done after user approval)
	return p.captureOrder(ctx, orderResp["id"].(string), req)
}

func (p *PayPalProvider) RefundPayment(ctx context.Context, paymentID string) (*entity.RefundResponse, error) {
	p.logger.WithContext(ctx).WithFields(map[string]interface{}{
		"provider":   "paypal",
		"payment_id": paymentID,
		"operation":  "refund_payment",
	}).Info("Processing refund")

	if err := p.ensureValidToken(ctx); err != nil {
		return nil, p.handleError(ctx, err, "token_refresh_failed")
	}

	url := fmt.Sprintf("%s/v2/payments/captures/%s/refund", p.baseURL, paymentID)
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer([]byte("{}")))
	if err != nil {
		return nil, p.handleError(ctx, err, "create_request_failed")
	}

	p.setHeaders(httpReq)

	resp, err := p.httpClient.Do(httpReq)
	if err != nil {
		return nil, p.handleError(ctx, err, "api_call_failed")
	}
	defer resp.Body.Close()

	return p.parseRefundResponse(ctx, resp)
}

func (p *PayPalProvider) GetPaymentStatus(ctx context.Context, paymentID string) (*entity.PaymentStatus, error) {
	p.logger.WithContext(ctx).WithFields(map[string]interface{}{
		"provider":   "paypal",
		"payment_id": paymentID,
		"operation":  "get_payment_status",
	}).Info("Getting payment status")

	if err := p.ensureValidToken(ctx); err != nil {
		return nil, p.handleError(ctx, err, "token_refresh_failed")
	}

	url := fmt.Sprintf("%s/v2/payments/captures/%s", p.baseURL, paymentID)
	httpReq, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, p.handleError(ctx, err, "create_request_failed")
	}

	p.setHeaders(httpReq)

	resp, err := p.httpClient.Do(httpReq)
	if err != nil {
		return nil, p.handleError(ctx, err, "api_call_failed")
	}
	defer resp.Body.Close()

	return p.parsePaymentStatusResponse(ctx, resp)
}

func (p *PayPalProvider) CreatePaymentIntent(ctx context.Context, req *entity.PaymentIntentRequest) (*entity.PaymentIntent, error) {
	p.logger.WithContext(ctx).WithFields(map[string]interface{}{
		"provider":    "paypal",
		"amount":      req.Amount,
		"currency":    req.Currency,
		"customer_id": req.CustomerID,
		"operation":   "create_payment_intent",
	}).Info("Creating payment intent")

	// For PayPal, payment intent is similar to creating an order
	if err := p.ensureValidToken(ctx); err != nil {
		return nil, p.handleError(ctx, err, "token_refresh_failed")
	}

	orderReq := map[string]interface{}{
		"intent": "CAPTURE",
		"purchase_units": []map[string]interface{}{
			{
				"amount": map[string]interface{}{
					"currency_code": req.Currency,
					"value":         fmt.Sprintf("%.2f", req.Amount),
				},
				"description": req.Description,
			},
		},
	}

	jsonData, err := json.Marshal(orderReq)
	if err != nil {
		return nil, p.handleError(ctx, err, "json_marshal_failed")
	}

	url := fmt.Sprintf("%s/v2/checkout/orders", p.baseURL)
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, p.handleError(ctx, err, "create_request_failed")
	}

	p.setHeaders(httpReq)

	resp, err := p.httpClient.Do(httpReq)
	if err != nil {
		return nil, p.handleError(ctx, err, "api_call_failed")
	}
	defer resp.Body.Close()

	return p.parsePaymentIntentResponse(ctx, resp)
}

func (p *PayPalProvider) ensureValidToken(ctx context.Context) error {
	if p.accessToken != "" && time.Now().Before(p.tokenExpiry) {
		return nil
	}

	return p.refreshAccessToken(ctx)
}

func (p *PayPalProvider) refreshAccessToken(ctx context.Context) error {
	tokenReq := "grant_type=client_credentials"

	url := fmt.Sprintf("%s/v1/oauth2/token", p.baseURL)
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBufferString(tokenReq))
	if err != nil {
		return err
	}

	httpReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	httpReq.SetBasicAuth(p.clientID, p.clientSecret)

	resp, err := p.httpClient.Do(httpReq)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var tokenResp map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return err
	}

	p.accessToken = tokenResp["access_token"].(string)
	expiresIn := int64(tokenResp["expires_in"].(float64))
	p.tokenExpiry = time.Now().Add(time.Duration(expiresIn-60) * time.Second) // Refresh 60s before expiry

	return nil
}

func (p *PayPalProvider) captureOrder(ctx context.Context, orderID string, req *entity.PaymentRequest) (*entity.PaymentResponse, error) {
	url := fmt.Sprintf("%s/v2/checkout/orders/%s/capture", p.baseURL, orderID)
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer([]byte("{}")))
	if err != nil {
		return nil, p.handleError(ctx, err, "create_capture_request_failed")
	}

	p.setHeaders(httpReq)

	resp, err := p.httpClient.Do(httpReq)
	if err != nil {
		return nil, p.handleError(ctx, err, "capture_api_call_failed")
	}
	defer resp.Body.Close()

	return p.parseCaptureResponse(ctx, resp)
}

func (p *PayPalProvider) setHeaders(req *http.Request) {
	req.Header.Set("Authorization", "Bearer "+p.accessToken)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "boilerplate-go/1.0")
}

func (p *PayPalProvider) handleError(ctx context.Context, err error, operation string) error {
	p.logger.ErrorLogger(ctx, err, "PayPal operation failed", map[string]interface{}{
		"provider":  "paypal",
		"operation": operation,
	})
	return fmt.Errorf("paypal %s: %w", operation, err)
}

func (p *PayPalProvider) parseOrderResponse(ctx context.Context, resp *http.Response) (map[string]interface{}, error) {
	var paypalResp map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&paypalResp); err != nil {
		return nil, p.handleError(ctx, err, "parse_order_response_failed")
	}

	if resp.StatusCode != http.StatusCreated {
		err := fmt.Errorf("paypal API error: %d", resp.StatusCode)
		return nil, p.handleError(ctx, err, "api_error")
	}

	return paypalResp, nil
}

func (p *PayPalProvider) parseCaptureResponse(ctx context.Context, resp *http.Response) (*entity.PaymentResponse, error) {
	var paypalResp map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&paypalResp); err != nil {
		return nil, p.handleError(ctx, err, "parse_capture_response_failed")
	}

	if resp.StatusCode != http.StatusCreated {
		err := fmt.Errorf("paypal API error: %d", resp.StatusCode)
		return nil, p.handleError(ctx, err, "api_error")
	}

	// Extract capture details from the response
	purchaseUnits := paypalResp["purchase_units"].([]interface{})
	firstUnit := purchaseUnits[0].(map[string]interface{})
	payments := firstUnit["payments"].(map[string]interface{})
	captures := payments["captures"].([]interface{})
	capture := captures[0].(map[string]interface{})

	amount := capture["amount"].(map[string]interface{})

	paymentResp := &entity.PaymentResponse{
		ID:            capture["id"].(string),
		Status:        capture["status"].(string),
		Amount:        parseFloat(amount["value"].(string)),
		Currency:      amount["currency_code"].(string),
		TransactionID: paypalResp["id"].(string),
		CreatedAt:     time.Now(),
	}

	return paymentResp, nil
}

func (p *PayPalProvider) parseRefundResponse(ctx context.Context, resp *http.Response) (*entity.RefundResponse, error) {
	var paypalResp map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&paypalResp); err != nil {
		return nil, p.handleError(ctx, err, "parse_refund_response_failed")
	}

	if resp.StatusCode != http.StatusCreated {
		err := fmt.Errorf("paypal API error: %d", resp.StatusCode)
		return nil, p.handleError(ctx, err, "api_error")
	}

	amount := paypalResp["amount"].(map[string]interface{})

	refundResp := &entity.RefundResponse{
		ID:        paypalResp["id"].(string),
		PaymentID: paypalResp["id"].(string),
		Amount:    parseFloat(amount["value"].(string)),
		Status:    paypalResp["status"].(string),
		CreatedAt: time.Now(),
	}

	return refundResp, nil
}

func (p *PayPalProvider) parsePaymentStatusResponse(ctx context.Context, resp *http.Response) (*entity.PaymentStatus, error) {
	var paypalResp map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&paypalResp); err != nil {
		return nil, p.handleError(ctx, err, "parse_status_response_failed")
	}

	if resp.StatusCode != http.StatusOK {
		err := fmt.Errorf("paypal API error: %d", resp.StatusCode)
		return nil, p.handleError(ctx, err, "api_error")
	}

	amount := paypalResp["amount"].(map[string]interface{})

	statusResp := &entity.PaymentStatus{
		ID:        paypalResp["id"].(string),
		Status:    paypalResp["status"].(string),
		Amount:    parseFloat(amount["value"].(string)),
		UpdatedAt: time.Now(),
	}

	return statusResp, nil
}

func (p *PayPalProvider) parsePaymentIntentResponse(ctx context.Context, resp *http.Response) (*entity.PaymentIntent, error) {
	var paypalResp map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&paypalResp); err != nil {
		return nil, p.handleError(ctx, err, "parse_intent_response_failed")
	}

	if resp.StatusCode != http.StatusCreated {
		err := fmt.Errorf("paypal API error: %d", resp.StatusCode)
		return nil, p.handleError(ctx, err, "api_error")
	}

	// Extract approval URL for client
	links := paypalResp["links"].([]interface{})
	var approvalURL string
	for _, link := range links {
		linkMap := link.(map[string]interface{})
		if linkMap["rel"].(string) == "approve" {
			approvalURL = linkMap["href"].(string)
			break
		}
	}

	intentResp := &entity.PaymentIntent{
		ID:           paypalResp["id"].(string),
		ClientSecret: approvalURL, // Using approval URL as client secret equivalent
		Status:       paypalResp["status"].(string),
	}

	return intentResp, nil
}

func parseFloat(s string) float64 {
	var f float64
	fmt.Sscanf(s, "%f", &f)
	return f
}
