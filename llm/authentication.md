# Authentication Implementation

## Overview

This implementation provides secure password-based authentication using bcrypt for password verification. The system queries the `users` table by email and verifies the provided plaintext password against the stored bcrypt hash.

## Database Schema

The `users` table has the following structure:

```sql
CREATE TABLE `users` (
  `id` int NOT NULL AUTO_INCREMENT,
  `uid` varchar(255) DEFAULT NULL,
  `email` varchar(255) NOT NULL,
  `display_name` varchar(255) DEFAULT NULL,
  `created_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  `password` varchar(255) DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `email` (`email`),
  UNIQUE KEY `uid` (`uid`)
) ENGINE=InnoDB AUTO_INCREMENT=3 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
```

## Implementation Details

### 1. Dependencies

Added `golang.org/x/crypto/bcrypt` to `go.mod`:

```
golang.org/x/crypto/bcrypt v0.17.0
```

### 2. Data Structures

**User struct** (`database.go`):

```go
type User struct {
    ID          int     `json:"id"`
    UID         *string `json:"uid"`
    Email       string  `json:"email"`
    DisplayName *string `json:"display_name"`
    CreatedAt   string  `json:"created_at"`
    Password    *string `json:"-"` // Excluded from JSON responses
}
```

**LoginRequest struct**:

```go
type LoginRequest struct {
    Email    string `json:"email"`
    Password string `json:"password"`
}
```

**LoginResponse struct**:

```go
type LoginResponse struct {
    User  User   `json:"user"`
    Token string `json:"token"`
}
```

### 3. Core Functions

**getUserByEmail** (`database.go`):

- Queries the database for a user by email
- Returns `nil` if user not found
- Handles nullable fields (uid, display_name, password)
- Returns user struct with password hash

**verifyPassword** (`database.go`):

- Uses `bcrypt.CompareHashAndPassword()` for secure comparison
- Returns error if passwords don't match
- Handles bcrypt-specific errors

### 4. Login Handler

**loginHandler** (`handlers.go`):

- Accepts POST requests to `/api/v1/login`
- Validates request body and required fields
- Queries user by email
- Verifies password using bcrypt
- Returns user data and token on success
- Returns appropriate error messages for security

## Route Protection

All API routes except `/api/v1/login` and `/api/v1/health` now require authentication.

### Authentication Header

Protected routes require an `Authorization` header:

```
Authorization: Bearer dummy-token
```

### Protected Routes

- `/api/v1/categories` (GET, POST)
- `/api/v1/categories/{id}` (GET, PUT, DELETE)
- `/api/v1/subcategories` (POST)
- `/api/v1/subcategories/{id}` (GET, PUT, DELETE)
- `/api/v1/expenses` (GET, POST)
- `/api/v1/expenses/{id}` (DELETE)
- `/api/v1/subcategories-by-expense-count` (GET)

### Public Routes

- `/api/v1/login` (POST)
- `/api/v1/health` (GET)

## API Endpoint

### POST /api/v1/login

**Request Body:**

```json
{
  "email": "user@example.com",
  "password": "plaintext_password"
}
```

**Success Response (200):**

```json
{
  "user": {
    "id": 1,
    "uid": "user123",
    "email": "user@example.com",
    "display_name": "John Doe",
    "created_at": "2024-01-01T00:00:00Z"
  },
  "token": "dummy-token"
}
```

**Error Responses:**

- `400 Bad Request`: Invalid request body or missing fields
- `401 Unauthorized`: Invalid email or password
- `405 Method Not Allowed`: Wrong HTTP method
- `500 Internal Server Error`: Database or server errors

## Security Features

1. **Secure Password Comparison**: Uses bcrypt's `CompareHashAndPassword()` for timing-attack-resistant comparison
2. **Generic Error Messages**: Returns same error for invalid email/password to prevent user enumeration
3. **Input Validation**: Validates required fields before processing
4. **SQL Injection Protection**: Uses parameterized queries
5. **Password Exclusion**: Password hash is excluded from JSON responses

## Usage Example

```bash
curl -X POST http://localhost:8082/api/v1/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "your_password"
  }'
```

## Testing

### Test Authentication

Use the provided `test_protected_routes.js` script to test the authentication and route protection:

```bash
cd api
node test_protected_routes.js
```

### Manual Testing with curl

**Login:**

```bash
curl -X POST http://localhost:8082/api/v1/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "your_password"
  }'
```

**Access protected route without auth (should fail):**

```bash
curl http://localhost:8082/api/v1/categories
```

**Access protected route with auth (should succeed):**

```bash
curl http://localhost:8082/api/v1/categories \
  -H "Authorization: Bearer dummy-token"
```

## Notes

- The current implementation returns a "dummy-token". In production, you should implement JWT token generation
- Password hashing for registration is not implemented (as per requirements)
- The system assumes bcrypt hashes are already stored in the database
- Error handling follows security best practices to prevent information leakage
