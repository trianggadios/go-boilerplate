package repository

import (
	"boilerplate-go/infrastructure/database"
	"boilerplate-go/infrastructure/logger"
	"boilerplate-go/infrastructure/metrics"
	"boilerplate-go/internal/domain/entity"
	"boilerplate-go/pkg/errors"
	"context"
	"database/sql"
	"fmt"
	"time"
)

// userRepositoryImpl implements the UserRepository interface
type userRepositoryImpl struct {
	db      *database.PostgresDB
	logger  *logger.Logger
	metrics *metrics.Metrics
}

// NewUserRepository creates a new user repository implementation
func NewUserRepository(db *database.PostgresDB, log *logger.Logger, m *metrics.Metrics) UserRepository {
	return &userRepositoryImpl{
		db:      db,
		logger:  log,
		metrics: m,
	}
}

func (r *userRepositoryImpl) Create(ctx context.Context, user *entity.User) error {
	start := time.Now()
	operation := "INSERT"
	table := "users"

	query := `
		INSERT INTO users (username, email, password, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id`

	now := time.Now()
	err := r.db.DB.QueryRowContext(ctx, query,
		user.Username, user.Email, user.Password, now, now).Scan(&user.ID)

	// Record metrics and logs
	duration := time.Since(start)
	r.metrics.RecordDatabaseQuery(operation, table, duration, err)
	r.logger.DatabaseLogger(ctx, operation, table, duration.String(), err)

	if err != nil {
		r.logger.ErrorLogger(ctx, err, "Failed to create user", map[string]interface{}{
			"username": user.Username,
			"email":    user.Email,
		})
		return fmt.Errorf("failed to create user: %w", err)
	}

	user.CreatedAt = now
	user.UpdatedAt = now
	return nil
}

func (r *userRepositoryImpl) GetByID(ctx context.Context, id int) (*entity.User, error) {
	start := time.Now()
	operation := "SELECT"
	table := "users"

	query := `
		SELECT id, username, email, password, created_at, updated_at
		FROM users
		WHERE id = $1`

	user := &entity.User{}
	err := r.db.DB.QueryRowContext(ctx, query, id).Scan(
		&user.ID, &user.Username, &user.Email, &user.Password,
		&user.CreatedAt, &user.UpdatedAt)

	// Record metrics and logs
	duration := time.Since(start)
	r.metrics.RecordDatabaseQuery(operation, table, duration, err)
	r.logger.DatabaseLogger(ctx, operation, table, duration.String(), err)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.ErrUserNotFound
		}
		r.logger.ErrorLogger(ctx, err, "Failed to get user by ID", map[string]interface{}{
			"user_id": id,
		})
		return nil, fmt.Errorf("failed to get user by id: %w", err)
	}

	return user, nil
}

func (r *userRepositoryImpl) GetByUsername(ctx context.Context, username string) (*entity.User, error) {
	start := time.Now()
	operation := "SELECT"
	table := "users"

	query := `
		SELECT id, username, email, password, created_at, updated_at
		FROM users
		WHERE username = $1`

	user := &entity.User{}
	err := r.db.DB.QueryRowContext(ctx, query, username).Scan(
		&user.ID, &user.Username, &user.Email, &user.Password,
		&user.CreatedAt, &user.UpdatedAt)

	// Record metrics and logs
	duration := time.Since(start)
	r.metrics.RecordDatabaseQuery(operation, table, duration, err)
	r.logger.DatabaseLogger(ctx, operation, table, duration.String(), err)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.ErrUserNotFound
		}
		r.logger.ErrorLogger(ctx, err, "Failed to get user by username", map[string]interface{}{
			"username": username,
		})
		return nil, fmt.Errorf("failed to get user by username: %w", err)
	}

	return user, nil
}

func (r *userRepositoryImpl) GetByEmail(ctx context.Context, email string) (*entity.User, error) {
	start := time.Now()
	operation := "SELECT"
	table := "users"

	query := `
		SELECT id, username, email, password, created_at, updated_at
		FROM users
		WHERE email = $1`

	user := &entity.User{}
	err := r.db.DB.QueryRowContext(ctx, query, email).Scan(
		&user.ID, &user.Username, &user.Email, &user.Password,
		&user.CreatedAt, &user.UpdatedAt)

	// Record metrics and logs
	duration := time.Since(start)
	r.metrics.RecordDatabaseQuery(operation, table, duration, err)
	r.logger.DatabaseLogger(ctx, operation, table, duration.String(), err)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.ErrUserNotFound
		}
		r.logger.ErrorLogger(ctx, err, "Failed to get user by email", map[string]interface{}{
			"email": email,
		})
		return nil, fmt.Errorf("failed to get user by email: %w", err)
	}

	return user, nil
}

func (r *userRepositoryImpl) Update(ctx context.Context, user *entity.User) error {
	start := time.Now()
	operation := "UPDATE"
	table := "users"

	query := `
		UPDATE users
		SET username = $1, email = $2, password = $3, updated_at = $4
		WHERE id = $5`

	user.UpdatedAt = time.Now()
	_, err := r.db.DB.ExecContext(ctx, query,
		user.Username, user.Email, user.Password, user.UpdatedAt, user.ID)

	// Record metrics and logs
	duration := time.Since(start)
	r.metrics.RecordDatabaseQuery(operation, table, duration, err)
	r.logger.DatabaseLogger(ctx, operation, table, duration.String(), err)

	if err != nil {
		r.logger.ErrorLogger(ctx, err, "Failed to update user", map[string]interface{}{
			"user_id":  user.ID,
			"username": user.Username,
			"email":    user.Email,
		})
		return fmt.Errorf("failed to update user: %w", err)
	}

	return nil
}

func (r *userRepositoryImpl) Delete(ctx context.Context, id int) error {
	start := time.Now()
	operation := "DELETE"
	table := "users"

	query := `DELETE FROM users WHERE id = $1`

	_, err := r.db.DB.ExecContext(ctx, query, id)

	// Record metrics and logs
	duration := time.Since(start)
	r.metrics.RecordDatabaseQuery(operation, table, duration, err)
	r.logger.DatabaseLogger(ctx, operation, table, duration.String(), err)

	if err != nil {
		r.logger.ErrorLogger(ctx, err, "Failed to delete user", map[string]interface{}{
			"user_id": id,
		})
		return fmt.Errorf("failed to delete user: %w", err)
	}

	return nil
}
