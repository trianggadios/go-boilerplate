package provider

import (
	"boilerplate-go/internal/domain/entity"
	"context"
)

// NotificationProvider defines the contract for notification operations
type NotificationProvider interface {
	SendEmail(ctx context.Context, req *entity.EmailRequest) (*entity.EmailResponse, error)
	SendSMS(ctx context.Context, req *entity.SMSRequest) (*entity.SMSResponse, error)
	SendPushNotification(ctx context.Context, req *entity.PushNotificationRequest) (*entity.PushNotificationResponse, error)
}

// EmailProvider defines specific email operations
type EmailProvider interface {
	SendEmail(ctx context.Context, req *entity.EmailRequest) (*entity.EmailResponse, error)
	SendBulkEmail(ctx context.Context, req *entity.BulkEmailRequest) (*entity.BulkEmailResponse, error)
	GetEmailStatus(ctx context.Context, emailID string) (*entity.EmailStatus, error)
}
