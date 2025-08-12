package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

// Config holds all configuration for our application.
type Config struct {
	Server    ServerConfig
	Database  DatabaseConfig
	JWT       JWTConfig
	Providers ProvidersConfig
}

// ServerConfig holds server configuration.
type ServerConfig struct {
	Port           string
	Host           string
	ReadTimeout    time.Duration
	WriteTimeout   time.Duration
	MaxHeaderBytes int
}

// DatabaseConfig holds database configuration.
type DatabaseConfig struct {
	Host            string
	Port            string
	User            string
	Password        string
	DBName          string
	SSLMode         string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
}

// JWTConfig holds JWT configuration.
type JWTConfig struct {
	SecretKey  string
	ExpiryTime time.Duration
}

// ProvidersConfig holds external providers configuration.
type ProvidersConfig struct {
	Payment      PaymentConfig
	Notification NotificationConfig
	FileStorage  FileStorageConfig
}

// PaymentConfig holds payment provider configuration.
type PaymentConfig struct {
	Provider string
	Stripe   StripeConfig
	PayPal   PayPalConfig
}

// StripeConfig holds Stripe-specific configuration.
type StripeConfig struct {
	BaseURL string
	APIKey  string
	Timeout time.Duration
}

// PayPalConfig holds PayPal-specific configuration.
type PayPalConfig struct {
	BaseURL      string
	ClientID     string
	ClientSecret string
	Timeout      time.Duration
}

// NotificationConfig holds notification provider configuration.
type NotificationConfig struct {
	Email EmailConfig
	SMS   SMSConfig
}

// EmailConfig holds email service configuration.
type EmailConfig struct {
	BaseURL   string
	APIKey    string
	FromEmail string
	Timeout   time.Duration
}

// SMSConfig holds SMS service configuration.
type SMSConfig struct {
	BaseURL    string
	APIKey     string
	FromNumber string
	Timeout    time.Duration
}

// FileStorageConfig holds file storage configuration.
type FileStorageConfig struct {
	Provider string
	S3       S3Config
	Local    LocalStorageConfig
}

// S3Config holds AWS S3 configuration.
type S3Config struct {
	Region          string
	Bucket          string
	AccessKeyID     string
	SecretAccessKey string
	Endpoint        string
}

// LocalStorageConfig holds local file storage configuration.
type LocalStorageConfig struct {
	BasePath string
}

func LoadConfig() *Config {
	return &Config{
		Server: ServerConfig{
			Port:           getEnv("SERVER_PORT", "8080"),
			Host:           getEnv("SERVER_HOST", "localhost"),
			ReadTimeout:    getDurationEnv("SERVER_READ_TIMEOUT", 10*time.Second),
			WriteTimeout:   getDurationEnv("SERVER_WRITE_TIMEOUT", 10*time.Second),
			MaxHeaderBytes: getIntEnv("SERVER_MAX_HEADER_BYTES", 1<<20),
		},
		Database: DatabaseConfig{
			Host:            getEnv("DB_HOST", "localhost"),
			Port:            getEnv("DB_PORT", "5432"),
			User:            getEnv("DB_USER", "postgres"),
			Password:        getEnv("DB_PASSWORD", "password"),
			DBName:          getEnv("DB_NAME", "boilerplate"),
			SSLMode:         getEnv("DB_SSLMODE", "disable"),
			MaxOpenConns:    getIntEnv("DB_MAX_OPEN_CONNS", 25),
			MaxIdleConns:    getIntEnv("DB_MAX_IDLE_CONNS", 5),
			ConnMaxLifetime: getDurationEnv("DB_CONN_MAX_LIFETIME", 5*time.Minute),
		},
		JWT: JWTConfig{
			SecretKey:  getEnv("JWT_SECRET", "your-secret-key"),
			ExpiryTime: getDurationEnv("JWT_EXPIRY_TIME", 24*time.Hour),
		},
		Providers: ProvidersConfig{
			Payment: PaymentConfig{
				Provider: getEnv("PAYMENT_PROVIDER", "stripe"),
				Stripe: StripeConfig{
					BaseURL: getEnv("STRIPE_BASE_URL", "https://api.stripe.com/v1"),
					APIKey:  getEnv("STRIPE_API_KEY", ""),
					Timeout: getDurationEnv("STRIPE_TIMEOUT", 30*time.Second),
				},
				PayPal: PayPalConfig{
					BaseURL:      getEnv("PAYPAL_BASE_URL", "https://api.paypal.com"),
					ClientID:     getEnv("PAYPAL_CLIENT_ID", ""),
					ClientSecret: getEnv("PAYPAL_CLIENT_SECRET", ""),
					Timeout:      getDurationEnv("PAYPAL_TIMEOUT", 30*time.Second),
				},
			},
			Notification: NotificationConfig{
				Email: EmailConfig{
					BaseURL:   getEnv("EMAIL_SERVICE_URL", "https://api.mailgun.net/v3"),
					APIKey:    getEnv("EMAIL_API_KEY", ""),
					FromEmail: getEnv("EMAIL_FROM", "noreply@boilerplate.com"),
					Timeout:   getDurationEnv("EMAIL_TIMEOUT", 30*time.Second),
				},
				SMS: SMSConfig{
					BaseURL:    getEnv("SMS_SERVICE_URL", "https://api.twilio.com/2010-04-01"),
					APIKey:     getEnv("SMS_API_KEY", ""),
					FromNumber: getEnv("SMS_FROM", "+1234567890"),
					Timeout:    getDurationEnv("SMS_TIMEOUT", 30*time.Second),
				},
			},
			FileStorage: FileStorageConfig{
				Provider: getEnv("FILE_STORAGE_PROVIDER", "local"),
				S3: S3Config{
					Region:          getEnv("AWS_REGION", "us-east-1"),
					Bucket:          getEnv("AWS_S3_BUCKET", ""),
					AccessKeyID:     getEnv("AWS_ACCESS_KEY_ID", ""),
					SecretAccessKey: getEnv("AWS_SECRET_ACCESS_KEY", ""),
					Endpoint:        getEnv("AWS_S3_ENDPOINT", ""),
				},
				Local: LocalStorageConfig{
					BasePath: getEnv("LOCAL_STORAGE_PATH", "./uploads"),
				},
			},
		},
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getIntEnv(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
		fmt.Printf("Warning: invalid value for %s, using default\n", key)
	}
	return defaultValue
}

func getDurationEnv(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
		fmt.Printf("Warning: invalid duration value for %s, using default\n", key)
	}
	return defaultValue
}
