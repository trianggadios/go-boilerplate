package main

import (
	"fmt"

	"boilerplate-go/config"
	"boilerplate-go/infrastructure/logger"
	"boilerplate-go/internal/domain/provider"
	"boilerplate-go/internal/provider/notification"
	"boilerplate-go/internal/provider/payment"
)

// ProviderFactory handles the creation of providers based on configuration
type ProviderFactory struct {
	config *config.Config
	logger *logger.Logger
}

func NewProviderFactory(config *config.Config, logger *logger.Logger) *ProviderFactory {
	return &ProviderFactory{
		config: config,
		logger: logger,
	}
}

// CreatePaymentProvider creates and returns the configured payment provider
func (f *ProviderFactory) CreatePaymentProvider() (provider.PaymentProvider, error) {
	switch f.config.Providers.Payment.Provider {
	case "stripe":
		return f.createStripeProvider(), nil
	case "paypal":
		return f.createPayPalProvider(), nil
	default:
		return nil, fmt.Errorf("unsupported payment provider: %s", f.config.Providers.Payment.Provider)
	}
}

// CreateNotificationProvider creates and returns the unified notification provider
func (f *ProviderFactory) CreateNotificationProvider() (provider.NotificationProvider, error) {
	notificationConfig := notification.UnifiedConfig{
		EmailConfig: notification.EmailConfig{
			BaseURL:   f.config.Providers.Notification.Email.BaseURL,
			APIKey:    f.config.Providers.Notification.Email.APIKey,
			FromEmail: f.config.Providers.Notification.Email.FromEmail,
			Timeout:   f.config.Providers.Notification.Email.Timeout,
		},
		SMSConfig: notification.SMSConfig{
			BaseURL:    f.config.Providers.Notification.SMS.BaseURL,
			APIKey:     f.config.Providers.Notification.SMS.APIKey,
			FromNumber: f.config.Providers.Notification.SMS.FromNumber,
			Timeout:    f.config.Providers.Notification.SMS.Timeout,
		},
	}

	return notification.NewUnifiedNotificationProvider(notificationConfig, f.logger), nil
}

func (f *ProviderFactory) createStripeProvider() provider.PaymentProvider {
	stripeConfig := payment.StripeConfig{
		BaseURL: f.config.Providers.Payment.Stripe.BaseURL,
		APIKey:  f.config.Providers.Payment.Stripe.APIKey,
		Timeout: f.config.Providers.Payment.Stripe.Timeout,
	}

	f.logger.WithFields(map[string]interface{}{
		"provider": "stripe",
		"base_url": stripeConfig.BaseURL,
		"timeout":  stripeConfig.Timeout.String(),
	}).Info("Initializing Stripe payment provider")

	return payment.NewStripeProvider(stripeConfig, f.logger)
}

func (f *ProviderFactory) createPayPalProvider() provider.PaymentProvider {
	paypalConfig := payment.PayPalConfig{
		BaseURL:      f.config.Providers.Payment.PayPal.BaseURL,
		ClientID:     f.config.Providers.Payment.PayPal.ClientID,
		ClientSecret: f.config.Providers.Payment.PayPal.ClientSecret,
		Timeout:      f.config.Providers.Payment.PayPal.Timeout,
	}

	f.logger.WithFields(map[string]interface{}{
		"provider": "paypal",
		"base_url": paypalConfig.BaseURL,
		"timeout":  paypalConfig.Timeout.String(),
	}).Info("Initializing PayPal payment provider")

	return payment.NewPayPalProvider(paypalConfig, f.logger)
}

// ValidateProviderConfiguration validates that all required provider configurations are present
func (f *ProviderFactory) ValidateProviderConfiguration() error {
	// Validate payment provider configuration
	switch f.config.Providers.Payment.Provider {
	case "stripe":
		if f.config.Providers.Payment.Stripe.APIKey == "" {
			return fmt.Errorf("Stripe API key is required")
		}
	case "paypal":
		if f.config.Providers.Payment.PayPal.ClientID == "" || f.config.Providers.Payment.PayPal.ClientSecret == "" {
			return fmt.Errorf("PayPal client ID and secret are required")
		}
	case "":
		f.logger.Warn("No payment provider configured, payment features will be disabled")
	default:
		return fmt.Errorf("unsupported payment provider: %s", f.config.Providers.Payment.Provider)
	}

	// Validate notification provider configuration
	if f.config.Providers.Notification.Email.APIKey == "" {
		f.logger.Warn("Email API key not configured, email notifications will be disabled")
	}

	if f.config.Providers.Notification.SMS.APIKey == "" {
		f.logger.Warn("SMS API key not configured, SMS notifications will be disabled")
	}

	return nil
}
