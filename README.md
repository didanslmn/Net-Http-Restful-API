# PostgresDB API Service

A robust REST API service built with Go, featuring JWT authentication, user management, product catalog, and order processing. The application follows clean architecture principles with PostgreSQL database and Redis caching.

## Features

- **Authentication & Authorization**
  - JWT-based authentication with access and refresh tokens
  - Role-based access control (User, Admin)
  - Token refresh and revocation
  - Session management

- **User Management**
  - User registration and login
  - Profile management
  - Password change functionality
  - Admin user management

- **Product Management**
  - Product catalog browsing
  - Admin-only product CRUD operations
  - Product search and filtering

- **Order Management**
  - Order creation and tracking
  - Order status updates
  - User order history

- **Infrastructure**
  - PostgreSQL database integration
  - Redis caching and session storage
  - Graceful server shutdown
  - Environment-based configuration

## Tech Stack

- **Language**: Go 1.25.4
- **Framework**: Standard library (net/http)
- **Database**: PostgreSQL
- **Cache**: Redis
- **Authentication**: JWT (RS256)
- **Validation**: go-playground/validator
- **Architecture**: Clean Architecture (Domain-Driven Design)

## Prerequisites

- Go 1.25.4 or higher
- PostgreSQL 12+
- Redis 6+
- RSA key pair for JWT signing (private.pem, public.pem)

## Installation

1. **Clone the repository**
   ```bash
   git clone <repository-url>
   cd postgresDB
   ```

2. **Install dependencies**
   ```bash
   go mod download
   ```

3. **Set up environment variables**

   Create a `.env` file in the root directory:
   ```env
   # Server Configuration
   SERVER_HOST=localhost
   SERVER_PORT=8080
   SERVER_READ_TIMEOUT=15s
   SERVER_WRITE_TIMEOUT=15s

   # Database Configuration
   DB_HOST=localhost
   DB_PORT=5432
   DB_NAME=postgres
   DB_USER=postgres
   DB_PASSWORD=your_password
   DB_SSL_MODE=disable

   # JWT Configuration
   JWT_PRIVATE_KEY_PATH=keys/private.pem
   JWT_PUBLIC_KEY_PATH=keys/public.pem
   JWT_ACCESS_TOKEN_TTL=15m
   JWT_REFRESH_TOKEN_TTL=168h
   ISSUER=myapp
   AUDIENCE=user-myapp

   # Redis Configuration
   REDIS_HOST=localhost
   REDIS_PORT=6379
   REDIS_PASSWORD=
   REDIS_DB=0
   ```

4. **Set up RSA keys**

   Generate RSA key pair and place them in the `keys/` directory:
   ```bash
   mkdir keys
   openssl genrsa -out keys/private.pem 2048
   openssl rsa -in keys/private.pem -pubout -out keys/public.pem
   ```

5. **Set up PostgreSQL database**

   Create the database and run migrations:
   ```bash
   # Create database
   createdb postgresDB

   # Run migrations (if you have migration files)
   go run cmd/migrate/migration.go
   ```

## Running the Application

1. **Development mode**
   ```bash
   go run cmd/api/main.go
   ```

2. **Build and run**
   ```bash
   go build -o bin/api cmd/api/main.go
   ./bin/api
   ```

The server will start on `http://localhost:8080`

## API Endpoints

### Authentication
- `POST /api/v1/auth/register` - User registration
- `POST /api/v1/auth/login` - User login
- `POST /api/v1/auth/refresh` - Refresh access token
- `POST /api/v1/auth/logout` - Logout (requires auth)
- `POST /api/v1/auth/revoke` - Revoke all sessions (requires auth)

### Users
- `GET /api/v1/users` - Get all users (admin only)
- `GET /api/v1/users/{id}` - Get user by ID
- `PUT /api/v1/users/{id}` - Update user
- `POST /api/v1/users/{id}/change-password` - Change password
- `DELETE /api/v1/users/{id}` - Delete user (admin only)

### Products
- `GET /api/v1/products` - List all products
- `GET /api/v1/products/{id}` - Get product by ID
- `POST /api/v1/products` - Create product (admin only)
- `PUT /api/v1/products/{id}` - Update product (admin only)
- `DELETE /api/v1/products/{id}` - Delete product (admin only)

### Orders
- `GET /api/v1/orders` - List user orders
- `GET /api/v1/orders/{id}` - Get order by ID
- `POST /api/v1/orders` - Create order
- `PATCH /api/v1/orders/{id}/status` - Update order status (admin only)

### Health Check
- `GET /api/v1/health` - Health check endpoint

## Project Structure

```
postgresDB/
├── cmd/
│   ├── api/
│   │   └── main.go              # Application entry point
│   └── migrate/
│       └── migration.go         # Database migrations
├── config/
│   └── config.go                # Configuration management
├── internal/
│   ├── delivery/
│   │   ├── handler/             # HTTP handlers
│   │   ├── middleware/          # HTTP middleware
│   │   └── routers/             # Route definitions
│   ├── domain/
│   │   ├── entities/            # Domain entities
│   │   ├── dto/                 # Data transfer objects
│   │   ├── repository/          # Repository interfaces
│   │   └── service/             # Service interfaces
│   ├── infrastruktur/
│   │   ├── cache/               # Redis client
│   │   └── database/            # PostgreSQL connection
│   ├── repository/
│   │   ├── interface.go         # Repository interfaces
│   │   ├── postgres/            # PostgreSQL implementations
│   │   └── redis/               # Redis implementations
│   └── service/                 # Business logic services
├── pkg/
│   ├── jwt/ 
│   └── utils/               # JWT utilities
│   └── validator/               # Validation utilities
├── keys/                        # RSA keys (not committed)
├── migrations/                  # Database migration files
├── .env                         # Environment variables (not committed)
├── .gitignore
├── go.mod
├── go.sum
└── README.md
```

## Database Schema

The application uses PostgreSQL with the following main tables:
- `users` - User accounts
- `products` - Product catalog
- `orders` - Order records
- `order_items` - Order line items

## Authentication

The API uses JWT tokens for authentication:

1. **Access Token**: Short-lived (15 minutes) for API access
2. **Refresh Token**: Long-lived (7 days) for token renewal
3. **Token Storage**: Refresh tokens stored in Redis for validation

Include the access token in the Authorization header:
```
Authorization: Bearer <access_token>
```

## Error Handling

The API returns standardized error responses:

```json
{
  "error": "Error message",
  "code": "ERROR_CODE",
  "details": {}
}
```

Common HTTP status codes:
- `200` - Success
- `201` - Created
- `400` - Bad Request
- `401` - Unauthorized
- `403` - Forbidden
- `404` - Not Found
- `500` - Internal Server Error

## Development

### Running Tests
```bash
go test ./...
```

### Code Formatting
```bash
go fmt ./...
```

### Linting
```bash
go vet ./...
```

## Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
