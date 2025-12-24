# PostgreSQL Database API - Auth Service Implementation

## Overview

A production-ready authentication and user management service built with Go, following clean architecture principles and industry best practices.

## Features

✅ **User Management**
- User registration with validation
- Secure login with JWT tokens
- Token refresh mechanism
- User profile updates
- Password management

✅ **Security**
- Bcrypt password hashing (cost: 12)
- JWT-based authentication (HS256)
- Role-based access control (RBAC)
- User activation/deactivation
- Token validation on every request

✅ **Architecture**
- Clean, layered architecture
- Dependency injection
- Interface-based design
- Error handling with context
- Middleware chain

✅ **API**
- RESTful endpoints
- Consistent error responses
- Request/response validation
- Swagger/OpenAPI ready

## Project Structure

```
.
├── cmd/
│   ├── api/
│   │   ├── main.go                    # Main application entry point
│   │   └── main_example.go            # Example setup guide
│   └── migrate/
│       └── migration.go               # Database migrations
├── config/
│   └── config.go                      # Configuration management
├── internal/
│   ├── domain/
│   │   └── user/
│   │       ├── entity.go              # Domain models
│   │       ├── repository.go          # Repository interface
│   │       ├── service.go             # Service interfaces
│   │       └── errors.go              # Domain errors
│   ├── service/
│   │   ├── auth_service.go            # Authentication service
│   │   ├── user_service.go            # User management service
│   │   └── factory.go                 # Service factory
│   ├── handler/
│   │   ├── auth_handler.go            # Auth HTTP handlers
│   │   └── user_handler.go            # User HTTP handlers
│   ├── middleware/
│   │   └── auth_middleware.go         # JWT & auth middleware
│   ├── dto/
│   │   ├── request/
│   │   │   └── user_request.go
│   │   └── response/
│   │       └── user_response.go
│   ├── repository/
│   │   └── user_repository.go         # Repository implementation
│   └── infrastruktur/
│       └── postgres/
│           └── database.go            # Database connection
├── pkg/
│   ├── jwt/
│   │   └── jwt.go                     # JWT manager
│   ├── utils/
│   │   └── hash.go                    # Password hashing utilities
│   ├── validator/
│   │   └── validator.go               # Input validation
│   └── constants/
│       └── constants.go               # App constants
├── migrations/                         # Database migration files
├── AUTH_SERVICE_DOCUMENTATION.md      # API documentation
├── BEST_PRACTICES.md                  # Best practices guide
├── go.mod                             # Go module file
└── README.md                          # This file
```

## Installation & Setup

### Prerequisites
- Go 1.20+
- PostgreSQL 12+
- Echo web framework
- pgx database driver
- golang-jwt library

### Step 1: Install Dependencies

```bash
go get -u github.com/labstack/echo/v4
go get -u github.com/golang-jwt/jwt/v5
go get -u github.com/google/uuid
go get -u github.com/jackc/pgx/v5
go get -u golang.org/x/crypto/bcrypt
go get -u github.com/go-playground/validator/v10
```

### Step 2: Configure Environment

Create a `.env` file in the project root:

```env
# Server Configuration
SERVER_PORT=8080

# JWT Configuration
JWT_SECRET=your_super_secret_key_minimum_32_characters_required
ACCESS_EXPIRATION_HOURS=24
REFRESH_TOKEN_EXPIRATION_HOURS=72
ISSUER=myapp
AUDIENCE=user-myapp

# Database Configuration
DB_HOST=localhost
DB_PORT=5432
DB_NAME=postgres
DB_USER=postgres
DB_PASSWORD=postgres
DB_SSL_MODE=require
```

### Step 3: Database Setup

Create the users table:

```sql
CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    username VARCHAR(50) UNIQUE NOT NULL,
    email VARCHAR(100) UNIQUE NOT NULL,
    password VARCHAR(255) NOT NULL,
    role VARCHAR(50) NOT NULL DEFAULT 'user',
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_email (email),
    INDEX idx_username (username)
);

CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_users_updated_at BEFORE UPDATE ON users
FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
```

### Step 4: Run the Application

```bash
# From cmd/api directory
go run main.go

# Or build and run
go build -o api ./cmd/api/main.go
./api
```

## API Endpoints

### Authentication Routes

#### Register User
```http
POST /auth/register
Content-Type: application/json

{
  "username": "john_doe",
  "email": "john@example.com",
  "password": "SecurePassword123!",
  "confirm_password": "SecurePassword123!",
  "role": "user"
}
```

**Response (201 Created):**
```json
{
  "message": "User registered successfully",
  "user": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "username": "john_doe",
    "email": "john@example.com",
    "role": "user",
    "is_active": true
  }
}
```

#### Login
```http
POST /auth/login
Content-Type: application/json

{
  "login_id": "john@example.com",
  "password": "SecurePassword123!"
}
```

**Response (200 OK):**
```json
{
  "access_token": "eyJhbGc...",
  "refresh_token": "eyJhbGc...",
  "token_type": "Bearer",
  "expires_in": 86400,
  "user": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "username": "john_doe",
    "email": "john@example.com",
    "role": "user",
    "is_active": true
  }
}
```

#### Refresh Token
```http
POST /auth/refresh
Authorization: Bearer <refresh_token>
```

#### Get Current User
```http
GET /auth/me
Authorization: Bearer <access_token>
```

### User Routes

#### Get User by ID
```http
GET /users/{id}
Authorization: Bearer <access_token>
```

#### Update User
```http
PUT /users/{id}
Authorization: Bearer <access_token>
Content-Type: application/json

{
  "username": "new_username",
  "email": "new@example.com",
  "role": "manager",
  "is_active": true
}
```

#### Change Password
```http
POST /users/{id}/change-password
Authorization: Bearer <access_token>
Content-Type: application/json

{
  "old_password": "OldPassword123!",
  "new_password": "NewPassword123!"
}
```

#### Delete User
```http
DELETE /users/{id}
Authorization: Bearer <access_token>
```

## Error Responses

All errors follow a consistent format:

```json
{
  "error": "ERROR_CODE",
  "message": "Human-readable message",
  "code": "ERROR_CATEGORY"
}
```

### Common Error Codes

| Code | Message | Status |
|------|---------|--------|
| `INVALID_CREDENTIALS` | Invalid email/username or password | 401 |
| `USER_NOT_FOUND` | User not found | 404 |
| `USER_ALREADY_EXISTS` | Email or username already exists | 409 |
| `MISSING_TOKEN` | Authorization header is required | 401 |
| `INVALID_TOKEN` | Invalid or expired token | 401 |
| `USER_INACTIVE` | User account is inactive | 403 |
| `FORBIDDEN` | Insufficient permissions | 403 |
| `INTERNAL_ERROR` | An unexpected error occurred | 500 |

## Security Considerations

### For Production

1. **Environment Variables**: Never commit secrets to version control
2. **HTTPS**: Always use HTTPS in production
3. **Rate Limiting**: Implement rate limiting on auth endpoints
4. **CORS**: Configure CORS properly for your domain
5. **Token Blacklist**: Implement token revocation for logout
6. **Email Verification**: Add email verification for registrations
7. **Password Reset**: Implement secure password reset flow
8. **Audit Logging**: Log all authentication events
9. **2FA**: Consider adding two-factor authentication
10. **API Versioning**: Use API versioning for backward compatibility

### Password Requirements

- Minimum 8 characters
- Mix of uppercase, lowercase, numbers, and special characters
- Not matching username or email

### Token Details

- **Access Token**: 24 hours expiration (configurable)
- **Refresh Token**: 72 hours expiration (configurable)
- **Algorithm**: HS256
- **Audience**: Verified on validation
- **Issuer**: Verified on validation

## Testing

### Running Tests

```bash
# Run all tests
go test ./...

# Run with coverage
go test -cover ./...

# Run specific test
go test -run TestRegisterSuccess ./internal/service
```

### Example Test

```go
func TestRegisterSuccess(t *testing.T) {
	mockRepo := new(MockUserRepository)
	mockJWT := new(MockJWTManager)

	mockRepo.On("ExistsByEmail", mock.Anything, "test@example.com").Return(false, nil)
	mockRepo.On("ExistsByUsername", mock.Anything, "testuser").Return(false, nil)
	mockRepo.On("CreateUser", mock.Anything, mock.Anything).Return(nil)

	authService := service.NewAuthService(mockRepo, mockJWT, nil)
	resp, err := authService.Register(context.Background(), request.RegisterRequest{
		Username: "testuser",
		Email:    "test@example.com",
		Password: "SecurePassword123!",
		Role:     "user",
	})

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, "test@example.com", resp.User.Email)
}
```

## Documentation

- **[AUTH_SERVICE_DOCUMENTATION.md](./AUTH_SERVICE_DOCUMENTATION.md)** - Comprehensive API documentation
- **[BEST_PRACTICES.md](./BEST_PRACTICES.md)** - Detailed best practices guide
- **[cmd/api/main_example.go](./cmd/api/main_example.go)** - Example setup and routing

## Troubleshooting

### Common Issues

#### 1. JWT Secret Too Short
```
error: JWT_SECRET must be at least 32 characters
```
**Solution**: Set JWT_SECRET environment variable to at least 32 characters

#### 2. Database Connection Failed
```
error: failed to connect to database
```
**Solution**: Verify database credentials and connection string in .env file

#### 3. Invalid Token Format
```
error: Invalid authorization header format
```
**Solution**: Use `Authorization: Bearer <token>` header format

#### 4. User Already Exists
```
error: Email or username already exists
```
**Solution**: Register with a unique email and username

## Contributing

1. Follow the architecture guidelines in `BEST_PRACTICES.md`
2. Write tests for new features
3. Use meaningful commit messages
4. Keep code clean and well-documented

## Performance Optimization

- Database queries use connection pooling
- JWT validation is performed on every protected request
- User status is verified during token validation
- Consider implementing caching for frequently accessed users

## Future Enhancements

1. **OAuth2/OpenID Connect**: External authentication providers
2. **Email Verification**: Verify emails before account activation
3. **Password Reset**: Forgot password functionality
4. **2FA**: Two-factor authentication
5. **Session Management**: Track and manage user sessions
6. **Audit Logging**: Comprehensive audit trail
7. **API Rate Limiting**: Prevent abuse and brute force
8. **Social Login**: Google, GitHub, etc.

## License

This project is part of the PostgreSQL Database basis_data course.

## Support

For issues or questions, please refer to the documentation or create an issue.

---

**Built with Clean Architecture Principles** ✨

For detailed implementation patterns and best practices, see [BEST_PRACTICES.md](./BEST_PRACTICES.md)
