package notification

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"boilerplate-go/infrastructure/logger"
	"boilerplate-go/internal/domain/entity"
)

type SMSProvider struct {
	httpClient *http.Client
	baseURL    string
	apiKey     string
	fromNumber string
	logger     *logger.Logger
}

type SMSConfig struct {
	BaseURL    string
	APIKey     string
	FromNumber string
	Timeout    time.Duration
}

func NewSMSProvider(config SMSConfig, logger *logger.Logger) *SMSProvider {
	timeout := config.Timeout
	if timeout == 0 {
		timeout = 30 * time.Second
	}

	return &SMSProvider{
		httpClient: &http.Client{
			Timeout: timeout,
		},
		baseURL:    config.BaseURL,
		apiKey:     config.APIKey,
		fromNumber: config.FromNumber,
		logger:     logger,
	}
}

func (s *SMSProvider) SendSMS(ctx context.Context, req *entity.SMSRequest) (*entity.SMSResponse, error) {
	s.logger.WithContext(ctx).WithFields(map[string]interface{}{
		"provider":  "sms_service",
		"to":        req.To,
		"operation": "send_sms",
	}).Info("Sending SMS")

	// Prepare SMS request
	smsReq := map[string]interface{}{
		"to":      req.To,
		"message": req.Message,
	}

	// Use from number from request or default
	if req.From != "" {
		smsReq["from"] = req.From
	} else {
		smsReq["from"] = s.fromNumber
	}

	jsonData, err := json.Marshal(smsReq)
	if err != nil {
		return nil, s.handleError(ctx, err, "json_marshal_failed")
	}

	// Create HTTP request
	url := fmt.Sprintf("%s/send", s.baseURL)
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

	return s.parseSMSResponse(ctx, resp)
}

func (s *SMSProvider) setHeaders(req *http.Request) {
	req.Header.Set("Authorization", "Bearer "+s.apiKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "boilerplate-go/1.0")
}

func (s *SMSProvider) handleError(ctx context.Context, err error, operation string) error {
	s.logger.ErrorLogger(ctx, err, "SMS service operation failed", map[string]interface{}{
		"provider":  "sms_service",
		"operation": operation,
	})
	return fmt.Errorf("sms service %s: %w", operation, err)
}

func (s *SMSProvider) parseSMSResponse(ctx context.Context, resp *http.Response) (*entity.SMSResponse, error) {
	var smsResp map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&smsResp); err != nil {
		return nil, s.handleError(ctx, err, "parse_response_failed")
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusAccepted {
		err := fmt.Errorf("SMS service API error: %d", resp.StatusCode)
		return nil, s.handleError(ctx, err, "api_error")
	}

	response := &entity.SMSResponse{
		ID:        smsResp["id"].(string),
		Status:    smsResp["status"].(string),
		SentAt:    time.Now(),
		MessageID: smsResp["message_id"].(string),
	}

	s.logger.WithContext(ctx).WithFields(map[string]interface{}{
		"sms_id":     response.ID,
		"status":     response.Status,
		"message_id": response.MessageID,
	}).Info("SMS sent successfully")

	return response, nil
}
