# Go Microservice Boilerplate

A professional Go microservice boilerplate with Clean Architecture, external service providers, JWT authentication, and comprehensive observability features.

## Architecture

This project follows Clean Architecture principles with a robust provider pattern for external services:

- **Domain Layer** (`internal/domain/`): Contains entities, repository interfaces, and provider interfaces
- **Provider Layer** (`internal/provider/`): Contains concrete implementations of external service providers
- **Use Case Layer** (`internal/usecase/`): Contains business logic and application services
- **Infrastructure Layer** (`infrastructure/`): Contains database connections, logging, metrics, and observability
- **Delivery Layer** (`internal/delivery/`): Contains HTTP handlers, middleware, and routes

## Features

### Core Features
- 🏗️ Clean Architecture pattern
- 🚀 Gin HTTP framework with middleware stack
- 🔐 JWT authentication and authorization
- 🗄️ PostgreSQL with optimized connection pooling
- 📝 Structured logging with correlation IDs
- 📊 Prometheus metrics and health checks
- 🛡️ Security best practices (bcrypt, CORS, rate limiting)
- 🐳 Docker and Docker Compose support

### External Service Integration
- 💳 **Payment Providers**: Stripe and PayPal integration
- 📧 **Notification Services**: Email, SMS, and Push notifications
- 📁 **File Storage**: AWS S3 and local storage support
- 🌍 **Geolocation Services**: IP-based location and distance calculation
- 🔄 **Provider Factory Pattern**: Easy switching between providers

### Observability & Operations
- 📈 Prometheus metrics with custom counters
- 🏥 Health checks (liveness, readiness)
- 📋 Request/response logging
- 🔍 Distributed tracing ready
- 🚨 Graceful shutdown
- 📊 Database connection monitoring

## Project Structure

```
boilerplate-go/
├── cmd/api/                     # Application entry point
│   ├── main.go                 # Main application
│   └── provider_factory.go     # Provider factory pattern
├── config/                      # Configuration management
│   └── config.go               # Environment-based configuration
├── internal/
│   ├── domain/                 # Domain layer (Clean Architecture)
│   │   ├── entity/             # Business entities and DTOs
│   │   │   ├── user.go
│   │   │   ├── provider_entities.go
│   │   │   └── order_entities.go
│   │   ├── repository/         # Repository interfaces
│   │   │   ├── user_repository.go
│   │   │   └── user_repository_impl.go
│   │   └── provider/           # External service interfaces
│   │       ├── payment_provider.go
│   │       ├── notification_provider.go
│   │       └── external_service_provider.go
│   ├── provider/               # Provider implementations
│   │   ├── payment/            # Payment service providers
│   │   │   ├── stripe_provider.go
│   │   │   └── paypal_provider.go
│   │   └── notification/       # Notification providers
│   │       ├── email_provider.go
│   │       ├── sms_provider.go
│   │       └── unified_notification_provider.go
│   ├── usecase/                # Business logic layer
│   │   ├── auth/               # Authentication use cases
│   │   ├── user/               # User management use cases
│   │   └── order/              # Order processing use cases
│   └── delivery/http/          # HTTP delivery layer
│       ├── handler/            # HTTP handlers
│       │   ├── auth_handler.go
│       │   ├── user_handler.go
│       │   └── order_handler.go
│       ├── middleware/         # HTTP middleware
│       │   ├── auth_middleware.go
│       │   └── middleware.go
│       └── route/              # Route definitions
│           └── routes.go
├── infrastructure/             # Infrastructure layer
│   ├── database/               # Database connections
│   ├── logger/                 # Structured logging
│   ├── metrics/                # Prometheus metrics
│   └── tracing/                # Distributed tracing
├── pkg/                        # Shared utilities
│   ├── errors/                 # Custom error types
│   ├── hash/                   # Password hashing
│   ├── jwt/                    # JWT utilities
│   └── response/               # HTTP response utilities
├── migrations/                 # Database migrations
├── test/                       # Test suites
│   ├── integration/            # Integration tests
│   └── e2e/                    # End-to-end tests
├── docs/swagger/               # API documentation
├── docker-compose.yml          # Multi-service setup
├── Dockerfile                  # Container configuration
└── Makefile                    # Build automation
```

## Getting Started

### Prerequisites

- **Go** 1.24+ (using latest toolchain)
- **PostgreSQL** 13+
- **Docker & Docker Compose** (optional)
- **Make** (recommended, for build automation)

### Development Tools

This project includes automated installation of development tools:
- **golangci-lint**: Professional Go linting
- **swag**: Swagger documentation generator

**Quick setup:**
```bash
# Install all development tools at once
make install-tools
```

### Quick Start with Docker

1. **Clone and start services:**
```bash
git clone <repository-url>
cd boilerplate-go
docker-compose up -d
```

2. **The API will be available at:**
   - Main API: `http://localhost:8080`
   - Health Check: `http://localhost:8080/health`
   - Metrics: `http://localhost:8080/metrics`

### Manual Installation

1. **Clone the repository:**
```bash
git clone <repository-url>
cd boilerplate-go
```

2. **Install dependencies:**
```bash
go mod tidy
```

3. **Set up environment variables:**
```bash
cp .env.example .env
# Edit .env with your configuration
```

4. **Set up the database:**
```bash
# Create database
createdb boilerplate

# Run migrations
psql -d boilerplate -f migrations/001_create_users_table.sql
```

5. **Install development tools and run:**
```bash
# Install linting and development tools
make install-tools

# Run the application
make run

# Or build and run
make build
./boilerplate-api

# Or using Go directly
go run cmd/api/main.go
```

**Verify everything works:**
```bash
# Check code quality
make fmt lint security

# Run tests
make test

# Build for production
make build
```

## API Endpoints

### Health & Monitoring
- `GET /health` - Application health check
- `GET /ready` - Readiness probe
- `GET /live` - Liveness probe  
- `GET /metrics` - Prometheus metrics

### Authentication
- `POST /api/v1/auth/register` - Register a new user
- `POST /api/v1/auth/login` - Login user

### User Management (Protected)
- `GET /api/v1/user/profile` - Get user profile

### Order Processing (Protected) 
- `POST /api/v1/orders` - Process a new order with payment
- `GET /api/v1/orders/payment/{payment_id}/status` - Get payment status
- `POST /api/v1/orders/refund` - Process order refund
- `POST /api/v1/orders/payment-intent` - Create payment intent

## Environment Variables

### Core Configuration
| Variable | Description | Default |
|----------|-------------|---------|
| `SERVER_PORT` | HTTP server port | `8080` |
| `SERVER_HOST` | HTTP server host | `localhost` |
| `SERVER_READ_TIMEOUT` | HTTP read timeout | `10s` |
| `SERVER_WRITE_TIMEOUT` | HTTP write timeout | `10s` |
| `LOG_LEVEL` | Logging level (debug,info,warn,error) | `info` |

### Database Configuration
| Variable | Description | Default |
|----------|-------------|---------|
| `DB_HOST` | Database host | `localhost` |
| `DB_PORT` | Database port | `5432` |
| `DB_USER` | Database user | `postgres` |
| `DB_PASSWORD` | Database password | `password` |
| `DB_NAME` | Database name | `boilerplate` |
| `DB_SSLMODE` | SSL mode | `disable` |
| `DB_MAX_OPEN_CONNS` | Max open connections | `25` |
| `DB_MAX_IDLE_CONNS` | Max idle connections | `5` |

### Security Configuration
| Variable | Description | Default |
|----------|-------------|---------|
| `JWT_SECRET` | JWT secret key | `your-secret-key` |
| `JWT_EXPIRY_TIME` | JWT expiration time | `24h` |

### Payment Providers
| Variable | Description | Default |
|----------|-------------|---------|
| `PAYMENT_PROVIDER` | Active payment provider (stripe/paypal) | `stripe` |
| `STRIPE_API_KEY` | Stripe API key | `` |
| `STRIPE_BASE_URL` | Stripe API base URL | `https://api.stripe.com/v1` |
| `PAYPAL_CLIENT_ID` | PayPal client ID | `` |
| `PAYPAL_CLIENT_SECRET` | PayPal client secret | `` |
| `PAYPAL_BASE_URL` | PayPal API base URL | `https://api.paypal.com` |

### Notification Services
| Variable | Description | Default |
|----------|-------------|---------|
| `EMAIL_API_KEY` | Email service API key | `` |
| `EMAIL_SERVICE_URL` | Email service URL | `https://api.mailgun.net/v3` |
| `EMAIL_FROM` | Default sender email | `noreply@boilerplate.com` |
| `SMS_API_KEY` | SMS service API key | `` |
| `SMS_SERVICE_URL` | SMS service URL | `https://api.twilio.com/2010-04-01` |
| `SMS_FROM` | Default sender number | `+1234567890` |

### File Storage
| Variable | Description | Default |
|----------|-------------|---------|
| `FILE_STORAGE_PROVIDER` | Storage provider (s3/local) | `local` |
| `AWS_REGION` | AWS region | `us-east-1` |
| `AWS_S3_BUCKET` | S3 bucket name | `` |
| `AWS_ACCESS_KEY_ID` | AWS access key | `` |
| `AWS_SECRET_ACCESS_KEY` | AWS secret key | `` |
| `LOCAL_STORAGE_PATH` | Local storage path | `./uploads` |

## API Usage Examples

### Authentication Flow

**Register a new user:**
```bash
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "username": "johndoe",
    "email": "john@example.com",
    "password": "password123"
  }'
```

**Login:**
```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "username": "johndoe",
    "password": "password123"
  }'
```

**Access protected endpoint:**
```bash
curl -X GET http://localhost:8080/api/v1/user/profile \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

### Order Processing

**Process an order:**
```bash
curl -X POST http://localhost:8080/api/v1/orders \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "order_id": "order-123",
    "amount": 99.99,
    "currency": "USD",
    "user_email": "john@example.com"
  }'
```

**Check payment status:**
```bash
curl -X GET http://localhost:8080/api/v1/orders/payment/payment-123/status \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

**Process refund:**
```bash
curl -X POST http://localhost:8080/api/v1/orders/refund \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "payment_id": "payment-123",
    "reason": "Customer request"
  }'
```

## Provider Configuration

### Payment Provider Switching

Switch between payment providers via environment variable:

```bash
# Use Stripe
PAYMENT_PROVIDER=stripe
STRIPE_API_KEY=sk_test_...

# Use PayPal  
PAYMENT_PROVIDER=paypal
PAYPAL_CLIENT_ID=your_client_id
PAYPAL_CLIENT_SECRET=your_client_secret
```

### Adding New Providers

1. **Create interface in domain layer:**
```go
// internal/domain/provider/new_provider.go
type NewServiceProvider interface {
    DoSomething(ctx context.Context, req *entity.Request) (*entity.Response, error)
}
```

2. **Implement concrete provider:**
```go
// internal/provider/new_service/implementation.go
type NewServiceClient struct {
    // implementation
}

func (c *NewServiceClient) DoSomething(ctx context.Context, req *entity.Request) (*entity.Response, error) {
    // implementation
}
```

3. **Update provider factory:**
```go
// cmd/api/provider_factory.go
func (f *ProviderFactory) CreateNewServiceProvider() provider.NewServiceProvider {
    // factory logic
}
```

## Development

### Build Commands

```bash
# Install development tools (golangci-lint, swag)
make install-tools

# Build the application
make build

# Run the application
make run

# Run tests
make test

# Run tests with coverage
make test-coverage

# Format code
make fmt

# Lint code  
make lint

# Run security checks
make security

# Generate API docs
make docs
```

### Testing

```bash
# Run unit tests
go test ./...

# Run with coverage
go test -cover ./...

# Run tests with verbose output
make test

# Run tests with coverage report
make test-coverage

# Run integration tests
go test ./test/integration/...

# Run specific test
go test -run TestAuthUsecase_Login ./internal/usecase/auth/
```

### Docker Development

```bash
# Build Docker image
docker build -t boilerplate-go .

# Run with Docker Compose
docker-compose up -d

# View logs
docker-compose logs -f api

# Stop services
docker-compose down
```

## Production Deployment

### Docker Production

```bash
# Build production image
docker build -t boilerplate-go:latest .

# Run in production mode
docker run -p 8080:8080 \
  -e DB_HOST=your-db-host \
  -e JWT_SECRET=your-secret \
  -e STRIPE_API_KEY=your-stripe-key \
  boilerplate-go:latest
```

### Kubernetes Deployment

```yaml
# k8s-deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: boilerplate-api
spec:
  replicas: 3
  selector:
    matchLabels:
      app: boilerplate-api
  template:
    metadata:
      labels:
        app: boilerplate-api
    spec:
      containers:
      - name: api
        image: boilerplate-go:latest
        ports:
        - containerPort: 8080
        env:
        - name: DB_HOST
          value: "postgres-service"
        livenessProbe:
          httpGet:
            path: /live
            port: 8080
        readinessProbe:
          httpGet:
            path: /ready
            port: 8080
```

## Monitoring & Observability

### Prometheus Metrics

Available at `/metrics` endpoint:

- `http_requests_total` - Total HTTP requests
- `http_request_duration_seconds` - Request duration
- `database_connections_active` - Active DB connections
- `database_queries_total` - Database query count
- `auth_attempts_total` - Authentication attempts

### Health Checks

- `/health` - Overall health status
- `/ready` - Readiness probe (for K8s)
- `/live` - Liveness probe (for K8s)

### Logging

Structured JSON logging with correlation IDs:

```json
{
  "level": "info",
  "msg": "Order processed successfully",
  "correlation_id": "req-123",
  "user_id": 1,
  "order_id": "order-456",
  "payment_id": "pay-789",
  "timestamp": "2024-01-01T12:00:00Z"
}
```

## Security Features

- 🔐 **JWT Authentication** with configurable expiry
- 🛡️ **bcrypt Password Hashing** with proper cost
- 🚫 **CORS Protection** with configurable origins  
- ⏱️ **Rate Limiting** to prevent abuse
- 🆔 **Request ID Middleware** for tracing
- 🔒 **Secure Headers** middleware
- 📝 **Input Validation** with struct tags
- 🚨 **Error Handling** without information leakage

## Contributing

1. **Fork the repository**
2. **Create your feature branch** (`git checkout -b feature/amazing-feature`)
3. **Follow the coding standards** (run `make lint` and `make fmt`)
4. **Add tests** for new functionality
5. **Commit your changes** (`git commit -m 'Add some amazing feature'`)
6. **Push to the branch** (`git push origin feature/amazing-feature`)
7. **Open a Pull Request**

### Coding Standards

- Follow [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
- Write unit tests with >80% coverage
- Use meaningful commit messages
- Document all public APIs
- Run quality checks before submitting:
  ```bash
  make fmt         # Format code
  make lint        # Run linter
  make security    # Security checks
  make test        # Run tests
  ```

### Code Quality Tools

This project uses professional Go linting and code quality tools:

- **golangci-lint**: Comprehensive linting with 20+ linters
- **gosec**: Security vulnerability scanner
- **goimports**: Import formatting and organization
- **gocritic**: Advanced code analysis
- **stylecheck**: Go style guide compliance

The linting configuration is in `.golangci.yml` and includes:
- Essential linters (errcheck, govet, staticcheck)
- Code quality checks (gocritic, gocyclo, misspell)
- Security analysis (gosec)
- Style enforcement (goimports, stylecheck)

**First time setup:**
```bash
# Install all development tools
make install-tools

# Verify installation
make lint
```

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Troubleshooting

### Common Issues

**Linting fails with "golangci-lint not found":**
```bash
# Install development tools
make install-tools

# Or install manually
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
```

**Import errors after adding new dependencies:**
```bash
# Clean and update dependencies
go mod tidy
make fmt  # Fix import organization
```

**Build fails on Windows:**
```bash
# Ensure Go is properly installed
go version

# Check GOPATH and GOROOT
go env GOPATH
go env GOROOT

# Clean module cache if needed
go clean -modcache
```

**Database connection issues:**
```bash
# Check PostgreSQL is running
pg_isready

# Verify connection string in .env
# Default: postgres://postgres:password@localhost:5432/boilerplate?sslmode=disable
```

## Support

- 📫 **Issues**: [GitHub Issues](https://github.com/your-repo/boilerplate-go/issues)  
- 📖 **Documentation**: [API Docs](http://localhost:8080/docs)
- 💬 **Discussions**: [GitHub Discussions](https://github.com/your-repo/boilerplate-go/discussions)
- 🔧 **Development Tools**: All tools auto-install via `make install-tools`

---

**Built with ❤️ using Go and Clean Architecture principles**