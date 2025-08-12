package entity

import "time"

// Payment related entities
type PaymentRequest struct {
	OrderID     string                 `json:"order_id"`
	Amount      float64                `json:"amount"`
	Currency    string                 `json:"currency"`
	Description string                 `json:"description"`
	CustomerID  string                 `json:"customer_id"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

type PaymentResponse struct {
	ID            string                 `json:"id"`
	Status        string                 `json:"status"`
	Amount        float64                `json:"amount"`
	Currency      string                 `json:"currency"`
	TransactionID string                 `json:"transaction_id"`
	CreatedAt     time.Time              `json:"created_at"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
}

type RefundResponse struct {
	ID        string    `json:"id"`
	PaymentID string    `json:"payment_id"`
	Amount    float64   `json:"amount"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

type PaymentStatus struct {
	ID        string    `json:"id"`
	Status    string    `json:"status"`
	Amount    float64   `json:"amount"`
	UpdatedAt time.Time `json:"updated_at"`
}

type PaymentIntentRequest struct {
	Amount      float64 `json:"amount"`
	Currency    string  `json:"currency"`
	CustomerID  string  `json:"customer_id"`
	Description string  `json:"description"`
}

type PaymentIntent struct {
	ID           string `json:"id"`
	ClientSecret string `json:"client_secret"`
	Status       string `json:"status"`
}

// Notification related entities
type EmailRequest struct {
	To          []string               `json:"to"`
	CC          []string               `json:"cc,omitempty"`
	BCC         []string               `json:"bcc,omitempty"`
	Subject     string                 `json:"subject"`
	Body        string                 `json:"body"`
	BodyHTML    string                 `json:"body_html,omitempty"`
	Attachments []EmailAttachment      `json:"attachments,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

type EmailResponse struct {
	ID        string    `json:"id"`
	Status    string    `json:"status"`
	SentAt    time.Time `json:"sent_at"`
	MessageID string    `json:"message_id"`
}

type EmailAttachment struct {
	Filename string `json:"filename"`
	Content  []byte `json:"content"`
	MimeType string `json:"mime_type"`
}

type SMSRequest struct {
	To      string `json:"to"`
	Message string `json:"message"`
	From    string `json:"from,omitempty"`
}

type SMSResponse struct {
	ID        string    `json:"id"`
	Status    string    `json:"status"`
	SentAt    time.Time `json:"sent_at"`
	MessageID string    `json:"message_id"`
}

type PushNotificationRequest struct {
	DeviceTokens []string               `json:"device_tokens"`
	Title        string                 `json:"title"`
	Body         string                 `json:"body"`
	Data         map[string]interface{} `json:"data,omitempty"`
}

type PushNotificationResponse struct {
	ID           string    `json:"id"`
	Status       string    `json:"status"`
	SentAt       time.Time `json:"sent_at"`
	SuccessCount int       `json:"success_count"`
	FailureCount int       `json:"failure_count"`
}

type BulkEmailRequest struct {
	Emails []EmailRequest `json:"emails"`
}

type BulkEmailResponse struct {
	ID           string    `json:"id"`
	Status       string    `json:"status"`
	TotalEmails  int       `json:"total_emails"`
	SentEmails   int       `json:"sent_emails"`
	FailedEmails int       `json:"failed_emails"`
	CreatedAt    time.Time `json:"created_at"`
}

type EmailStatus struct {
	ID          string     `json:"id"`
	Status      string     `json:"status"`
	DeliveredAt *time.Time `json:"delivered_at,omitempty"`
	OpenedAt    *time.Time `json:"opened_at,omitempty"`
	ClickedAt   *time.Time `json:"clicked_at,omitempty"`
}

// External service related entities
type ExternalUserProfile struct {
	ID        string                 `json:"id"`
	Username  string                 `json:"username"`
	Email     string                 `json:"email"`
	FullName  string                 `json:"full_name"`
	Avatar    string                 `json:"avatar"`
	Verified  bool                   `json:"verified"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt time.Time              `json:"created_at"`
	UpdatedAt time.Time              `json:"updated_at"`
}

type UserValidation struct {
	UserID    string     `json:"user_id"`
	IsValid   bool       `json:"is_valid"`
	Reason    string     `json:"reason,omitempty"`
	ExpiresAt *time.Time `json:"expires_at,omitempty"`
}

type UpdateUserProfileRequest struct {
	FullName string                 `json:"full_name,omitempty"`
	Avatar   string                 `json:"avatar,omitempty"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// Geolocation related entities
type Coordinates struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

type LocationInfo struct {
	IP          string      `json:"ip"`
	Country     string      `json:"country"`
	CountryCode string      `json:"country_code"`
	City        string      `json:"city"`
	Region      string      `json:"region"`
	Coordinates Coordinates `json:"coordinates"`
	Timezone    string      `json:"timezone"`
}

type DistanceInfo struct {
	Distance float64 `json:"distance"` // in kilometers
	Duration int     `json:"duration"` // in seconds
	Unit     string  `json:"unit"`
}

// File storage related entities
type FileUploadRequest struct {
	FileName    string            `json:"file_name"`
	Content     []byte            `json:"content"`
	ContentType string            `json:"content_type"`
	Path        string            `json:"path,omitempty"`
	Metadata    map[string]string `json:"metadata,omitempty"`
}

type FileUploadResponse struct {
	ID         string    `json:"id"`
	URL        string    `json:"url"`
	Path       string    `json:"path"`
	Size       int64     `json:"size"`
	MimeType   string    `json:"mime_type"`
	UploadedAt time.Time `json:"uploaded_at"`
}

type FileDownloadResponse struct {
	ID          string            `json:"id"`
	FileName    string            `json:"file_name"`
	Content     []byte            `json:"content"`
	ContentType string            `json:"content_type"`
	Size        int64             `json:"size"`
	Metadata    map[string]string `json:"metadata,omitempty"`
}

type FileInfo struct {
	ID          string            `json:"id"`
	FileName    string            `json:"file_name"`
	Path        string            `json:"path"`
	Size        int64             `json:"size"`
	ContentType string            `json:"content_type"`
	URL         string            `json:"url"`
	Metadata    map[string]string `json:"metadata,omitempty"`
	UploadedAt  time.Time         `json:"uploaded_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
}
