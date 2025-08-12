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
	"boilerplate-go/internal/domain/provider"
)

type EmailProvider struct {
	httpClient *http.Client
	baseURL    string
	apiKey     string
	fromEmail  string
	logger     *logger.Logger
}

type EmailConfig struct {
	BaseURL   string
	APIKey    string
	FromEmail string
	Timeout   time.Duration
}

func NewEmailProvider(config EmailConfig, logger *logger.Logger) provider.EmailProvider {
	timeout := config.Timeout
	if timeout == 0 {
		timeout = 30 * time.Second
	}

	return &EmailProvider{
		httpClient: &http.Client{
			Timeout: timeout,
		},
		baseURL:   config.BaseURL,
		apiKey:    config.APIKey,
		fromEmail: config.FromEmail,
		logger:    logger,
	}
}

func (e *EmailProvider) SendEmail(ctx context.Context, req *entity.EmailRequest) (*entity.EmailResponse, error) {
	e.logger.WithContext(ctx).WithFields(map[string]interface{}{
		"provider":  "email_service",
		"to_count":  len(req.To),
		"subject":   req.Subject,
		"operation": "send_email",
	}).Info("Sending email")

	// Prepare email request
	emailReq := map[string]interface{}{
		"from":    e.fromEmail,
		"to":      req.To,
		"subject": req.Subject,
	}

	if req.CC != nil && len(req.CC) > 0 {
		emailReq["cc"] = req.CC
	}

	if req.BCC != nil && len(req.BCC) > 0 {
		emailReq["bcc"] = req.BCC
	}

	if req.BodyHTML != "" {
		emailReq["html"] = req.BodyHTML
		emailReq["text"] = req.Body
	} else {
		emailReq["text"] = req.Body
	}

	if req.Attachments != nil && len(req.Attachments) > 0 {
		attachments := make([]map[string]interface{}, 0, len(req.Attachments))
		for _, att := range req.Attachments {
			attachments = append(attachments, map[string]interface{}{
				"filename": att.Filename,
				"content":  att.Content,
				"type":     att.MimeType,
			})
		}
		emailReq["attachments"] = attachments
	}

	if req.Metadata != nil {
		emailReq["metadata"] = req.Metadata
	}

	jsonData, err := json.Marshal(emailReq)
	if err != nil {
		return nil, e.handleError(ctx, err, "json_marshal_failed")
	}

	// Create HTTP request
	url := fmt.Sprintf("%s/send", e.baseURL)
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, e.handleError(ctx, err, "create_request_failed")
	}

	e.setHeaders(httpReq)

	// Execute request
	resp, err := e.httpClient.Do(httpReq)
	if err != nil {
		return nil, e.handleError(ctx, err, "api_call_failed")
	}
	defer resp.Body.Close()

	return e.parseEmailResponse(ctx, resp)
}

func (e *EmailProvider) SendBulkEmail(ctx context.Context, req *entity.BulkEmailRequest) (*entity.BulkEmailResponse, error) {
	e.logger.WithContext(ctx).WithFields(map[string]interface{}{
		"provider":    "email_service",
		"email_count": len(req.Emails),
		"operation":   "send_bulk_email",
	}).Info("Sending bulk emails")

	// Prepare bulk email request
	bulkReq := map[string]interface{}{
		"emails": make([]map[string]interface{}, 0, len(req.Emails)),
	}

	for _, email := range req.Emails {
		emailData := map[string]interface{}{
			"from":    e.fromEmail,
			"to":      email.To,
			"subject": email.Subject,
		}

		if email.CC != nil && len(email.CC) > 0 {
			emailData["cc"] = email.CC
		}

		if email.BCC != nil && len(email.BCC) > 0 {
			emailData["bcc"] = email.BCC
		}

		if email.BodyHTML != "" {
			emailData["html"] = email.BodyHTML
			emailData["text"] = email.Body
		} else {
			emailData["text"] = email.Body
		}

		if email.Metadata != nil {
			emailData["metadata"] = email.Metadata
		}

		bulkReq["emails"] = append(bulkReq["emails"].([]map[string]interface{}), emailData)
	}

	jsonData, err := json.Marshal(bulkReq)
	if err != nil {
		return nil, e.handleError(ctx, err, "json_marshal_failed")
	}

	// Create HTTP request
	url := fmt.Sprintf("%s/send-bulk", e.baseURL)
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, e.handleError(ctx, err, "create_request_failed")
	}

	e.setHeaders(httpReq)

	// Execute request
	resp, err := e.httpClient.Do(httpReq)
	if err != nil {
		return nil, e.handleError(ctx, err, "api_call_failed")
	}
	defer resp.Body.Close()

	return e.parseBulkEmailResponse(ctx, resp)
}

func (e *EmailProvider) GetEmailStatus(ctx context.Context, emailID string) (*entity.EmailStatus, error) {
	e.logger.WithContext(ctx).WithFields(map[string]interface{}{
		"provider":  "email_service",
		"email_id":  emailID,
		"operation": "get_email_status",
	}).Info("Getting email status")

	url := fmt.Sprintf("%s/status/%s", e.baseURL, emailID)
	httpReq, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, e.handleError(ctx, err, "create_request_failed")
	}

	e.setHeaders(httpReq)

	resp, err := e.httpClient.Do(httpReq)
	if err != nil {
		return nil, e.handleError(ctx, err, "api_call_failed")
	}
	defer resp.Body.Close()

	return e.parseEmailStatusResponse(ctx, resp)
}

func (e *EmailProvider) setHeaders(req *http.Request) {
	req.Header.Set("Authorization", "Bearer "+e.apiKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "boilerplate-go/1.0")
}

func (e *EmailProvider) handleError(ctx context.Context, err error, operation string) error {
	e.logger.ErrorLogger(ctx, err, "Email service operation failed", map[string]interface{}{
		"provider":  "email_service",
		"operation": operation,
	})
	return fmt.Errorf("email service %s: %w", operation, err)
}

func (e *EmailProvider) parseEmailResponse(ctx context.Context, resp *http.Response) (*entity.EmailResponse, error) {
	var emailResp map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&emailResp); err != nil {
		return nil, e.handleError(ctx, err, "parse_response_failed")
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusAccepted {
		err := fmt.Errorf("email service API error: %d", resp.StatusCode)
		return nil, e.handleError(ctx, err, "api_error")
	}

	response := &entity.EmailResponse{
		ID:        emailResp["id"].(string),
		Status:    emailResp["status"].(string),
		SentAt:    time.Now(),
		MessageID: emailResp["message_id"].(string),
	}

	e.logger.WithContext(ctx).WithFields(map[string]interface{}{
		"email_id":   response.ID,
		"status":     response.Status,
		"message_id": response.MessageID,
	}).Info("Email sent successfully")

	return response, nil
}

func (e *EmailProvider) parseBulkEmailResponse(ctx context.Context, resp *http.Response) (*entity.BulkEmailResponse, error) {
	var bulkResp map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&bulkResp); err != nil {
		return nil, e.handleError(ctx, err, "parse_bulk_response_failed")
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusAccepted {
		err := fmt.Errorf("email service API error: %d", resp.StatusCode)
		return nil, e.handleError(ctx, err, "api_error")
	}

	response := &entity.BulkEmailResponse{
		ID:           bulkResp["id"].(string),
		Status:       bulkResp["status"].(string),
		TotalEmails:  int(bulkResp["total_emails"].(float64)),
		SentEmails:   int(bulkResp["sent_emails"].(float64)),
		FailedEmails: int(bulkResp["failed_emails"].(float64)),
		CreatedAt:    time.Now(),
	}

	return response, nil
}

func (e *EmailProvider) parseEmailStatusResponse(ctx context.Context, resp *http.Response) (*entity.EmailStatus, error) {
	var statusResp map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&statusResp); err != nil {
		return nil, e.handleError(ctx, err, "parse_status_response_failed")
	}

	if resp.StatusCode != http.StatusOK {
		err := fmt.Errorf("email service API error: %d", resp.StatusCode)
		return nil, e.handleError(ctx, err, "api_error")
	}

	status := &entity.EmailStatus{
		ID:     statusResp["id"].(string),
		Status: statusResp["status"].(string),
	}

	// Parse optional timestamp fields
	if deliveredAt, exists := statusResp["delivered_at"]; exists && deliveredAt != nil {
		if timestamp, ok := deliveredAt.(string); ok {
			if t, err := time.Parse(time.RFC3339, timestamp); err == nil {
				status.DeliveredAt = &t
			}
		}
	}

	if openedAt, exists := statusResp["opened_at"]; exists && openedAt != nil {
		if timestamp, ok := openedAt.(string); ok {
			if t, err := time.Parse(time.RFC3339, timestamp); err == nil {
				status.OpenedAt = &t
			}
		}
	}

	if clickedAt, exists := statusResp["clicked_at"]; exists && clickedAt != nil {
		if timestamp, ok := clickedAt.(string); ok {
			if t, err := time.Parse(time.RFC3339, timestamp); err == nil {
				status.ClickedAt = &t
			}
		}
	}

	return status, nil
}
