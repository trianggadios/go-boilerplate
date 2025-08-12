package provider

import (
	"boilerplate-go/internal/domain/entity"
	"context"
)

// UserServiceProvider defines the contract for external user service operations
type UserServiceProvider interface {
	GetUserProfile(ctx context.Context, userID string) (*entity.ExternalUserProfile, error)
	ValidateUser(ctx context.Context, userID string) (*entity.UserValidation, error)
	UpdateUserProfile(ctx context.Context, userID string, req *entity.UpdateUserProfileRequest) error
}

// GeolocationProvider defines the contract for geolocation services
type GeolocationProvider interface {
	GetLocationByIP(ctx context.Context, ipAddress string) (*entity.LocationInfo, error)
	GetDistanceBetween(ctx context.Context, from, to *entity.Coordinates) (*entity.DistanceInfo, error)
}

// FileStorageProvider defines the contract for file storage operations
type FileStorageProvider interface {
	UploadFile(ctx context.Context, req *entity.FileUploadRequest) (*entity.FileUploadResponse, error)
	DownloadFile(ctx context.Context, fileID string) (*entity.FileDownloadResponse, error)
	DeleteFile(ctx context.Context, fileID string) error
	GetFileInfo(ctx context.Context, fileID string) (*entity.FileInfo, error)
}
