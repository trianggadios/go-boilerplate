package auth

import (
	"boilerplate-go/config"
	"boilerplate-go/internal/domain/entity"
	"boilerplate-go/pkg/errors"
	"boilerplate-go/pkg/hash"
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockUserRepository is a mock implementation of UserRepository
type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) Create(ctx context.Context, user *entity.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepository) GetByID(ctx context.Context, id int) (*entity.User, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*entity.User), args.Error(1)
}

func (m *MockUserRepository) GetByUsername(ctx context.Context, username string) (*entity.User, error) {
	args := m.Called(ctx, username)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.User), args.Error(1)
}

func (m *MockUserRepository) GetByEmail(ctx context.Context, email string) (*entity.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.User), args.Error(1)
}

func (m *MockUserRepository) Update(ctx context.Context, user *entity.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepository) Delete(ctx context.Context, id int) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func TestAuthUsecase_Register(t *testing.T) {
	tests := []struct {
		name          string
		request       *entity.RegisterRequest
		setupMock     func(*MockUserRepository)
		expectedError string
	}{
		{
			name: "successful registration",
			request: &entity.RegisterRequest{
				Username: "testuser",
				Email:    "test@example.com",
				Password: "password123",
			},
			setupMock: func(repo *MockUserRepository) {
				repo.On("GetByUsername", mock.Anything, "testuser").Return(nil, errors.ErrUserNotFound)
				repo.On("GetByEmail", mock.Anything, "test@example.com").Return(nil, errors.ErrUserNotFound)
				repo.On("Create", mock.Anything, mock.AnythingOfType("*entity.User")).Return(nil)
			},
			expectedError: "",
		},
		{
			name: "username already exists",
			request: &entity.RegisterRequest{
				Username: "existinguser",
				Email:    "test@example.com",
				Password: "password123",
			},
			setupMock: func(repo *MockUserRepository) {
				existingUser := &entity.User{
					ID:       1,
					Username: "existinguser",
					Email:    "existing@example.com",
				}
				repo.On("GetByUsername", mock.Anything, "existinguser").Return(existingUser, nil)
			},
			expectedError: "user already exists",
		},
		{
			name: "email already exists",
			request: &entity.RegisterRequest{
				Username: "testuser",
				Email:    "existing@example.com",
				Password: "password123",
			},
			setupMock: func(repo *MockUserRepository) {
				repo.On("GetByUsername", mock.Anything, "testuser").Return(nil, errors.ErrUserNotFound)
				existingUser := &entity.User{
					ID:       1,
					Username: "existinguser",
					Email:    "existing@example.com",
				}
				repo.On("GetByEmail", mock.Anything, "existing@example.com").Return(existingUser, nil)
			},
			expectedError: "user already exists",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			mockRepo := new(MockUserRepository)
			tt.setupMock(mockRepo)

			jwtConfig := config.JWTConfig{
				SecretKey:  "test-secret",
				ExpiryTime: 24 * time.Hour,
			}

			authUsecase := NewAuthUsecase(mockRepo, jwtConfig)
			ctx := context.Background()

			// Execute
			user, err := authUsecase.Register(ctx, tt.request)

			// Assert
			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				assert.Nil(t, user)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, user)
				assert.Equal(t, tt.request.Username, user.Username)
				assert.Equal(t, tt.request.Email, user.Email)
				assert.NotEmpty(t, user.Password)                      // Should be hashed
				assert.NotEqual(t, tt.request.Password, user.Password) // Should not be plain text
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestAuthUsecase_Login(t *testing.T) {
	tests := []struct {
		name          string
		request       *entity.LoginRequest
		setupMock     func(*MockUserRepository)
		expectedError string
	}{
		{
			name: "successful login",
			request: &entity.LoginRequest{
				Username: "testuser",
				Password: "password123",
			},
			setupMock: func(repo *MockUserRepository) {
				// Use a properly generated bcrypt hash
				hashedPassword, _ := hash.HashPassword("password123")
				user := &entity.User{
					ID:       1,
					Username: "testuser",
					Email:    "test@example.com",
					Password: hashedPassword,
				}
				repo.On("GetByUsername", mock.Anything, "testuser").Return(user, nil)
			},
			expectedError: "",
		},
		{
			name: "user not found",
			request: &entity.LoginRequest{
				Username: "nonexistent",
				Password: "password123",
			},
			setupMock: func(repo *MockUserRepository) {
				repo.On("GetByUsername", mock.Anything, "nonexistent").Return(nil, errors.ErrUserNotFound)
			},
			expectedError: "invalid credentials",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			mockRepo := new(MockUserRepository)
			tt.setupMock(mockRepo)

			jwtConfig := config.JWTConfig{
				SecretKey:  "test-secret",
				ExpiryTime: 24 * time.Hour,
			}

			authUsecase := NewAuthUsecase(mockRepo, jwtConfig)
			ctx := context.Background()

			// Execute
			loginResponse, err := authUsecase.Login(ctx, tt.request)

			// Assert
			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				assert.Nil(t, loginResponse)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, loginResponse)
				assert.NotEmpty(t, loginResponse.Token)
				assert.NotNil(t, loginResponse.User)
				assert.Equal(t, tt.request.Username, loginResponse.User.Username)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}
