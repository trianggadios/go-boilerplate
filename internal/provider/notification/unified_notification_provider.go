package notification

import (
	"context"
	"time"

	"boilerplate-go/infrastructure/logger"
	"boilerplate-go/internal/domain/entity"
	"boilerplate-go/internal/domain/provider"
)

// UnifiedNotificationProvider implements the NotificationProvider interface
// and coordinates between different notification channels
type UnifiedNotificationProvider struct {
	emailProvider provider.EmailProvider
	smsProvider   *SMSProvider
	logger        *logger.Logger
}

type UnifiedConfig struct {
	EmailConfig EmailConfig
	SMSConfig   SMSConfig
}

func NewUnifiedNotificationProvider(config UnifiedConfig, logger *logger.Logger) provider.NotificationProvider {
	emailProvider := NewEmailProvider(config.EmailConfig, logger)
	smsProvider := NewSMSProvider(config.SMSConfig, logger)

	return &UnifiedNotificationProvider{
		emailProvider: emailProvider,
		smsProvider:   smsProvider,
		logger:        logger,
	}
}

func (u *UnifiedNotificationProvider) SendEmail(ctx context.Context, req *entity.EmailRequest) (*entity.EmailResponse, error) {
	u.logger.WithContext(ctx).WithFields(map[string]interface{}{
		"provider":  "unified_notification",
		"channel":   "email",
		"operation": "send_email",
	}).Info("Routing email through unified provider")

	return u.emailProvider.SendEmail(ctx, req)
}

func (u *UnifiedNotificationProvider) SendSMS(ctx context.Context, req *entity.SMSRequest) (*entity.SMSResponse, error) {
	u.logger.WithContext(ctx).WithFields(map[string]interface{}{
		"provider":  "unified_notification",
		"channel":   "sms",
		"operation": "send_sms",
	}).Info("Routing SMS through unified provider")

	return u.smsProvider.SendSMS(ctx, req)
}

func (u *UnifiedNotificationProvider) SendPushNotification(ctx context.Context, req *entity.PushNotificationRequest) (*entity.PushNotificationResponse, error) {
	u.logger.WithContext(ctx).WithFields(map[string]interface{}{
		"provider":  "unified_notification",
		"channel":   "push",
		"operation": "send_push_notification",
	}).Info("Push notification not implemented yet")

	// TODO: Implement push notification provider
	// For now, return a mock response
	response := &entity.PushNotificationResponse{
		ID:           "mock-push-id",
		Status:       "not_implemented",
		SentAt:       time.Now(),
		SuccessCount: 0,
		FailureCount: len(req.DeviceTokens),
	}

	return response, nil
}
