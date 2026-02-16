# Backend Stage 2 — Auth Schema, Seeds & Auth APIs

## Overview

Build the database schema for authentication and user management, seed initial data, and fully implement the auth endpoints scaffolded in Stage 1.

---

## 1. Database Migrations (goose SQL)

Create versioned SQL migration files in `backend/migrations/`. Each migration has `up` and `down`.

### 1.1 Users Table

```sql
CREATE TABLE users (
    id            BIGSERIAL PRIMARY KEY,
    name          VARCHAR(255) NOT NULL,
    email         VARCHAR(255) NOT NULL UNIQUE,
    phone         VARCHAR(50),
    address       TEXT,
    password_hash TEXT NOT NULL,
    profile_picture TEXT,           -- URL or file path (NULL if not set)
    status        VARCHAR(20) NOT NULL DEFAULT 'active',  -- active, pending, inactive
    is_super_admin BOOLEAN NOT NULL DEFAULT false,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_status ON users(status);
```

**Constraints:**
- `email` is case-insensitive unique (use `CITEXT` extension or handle in application layer with `LOWER()`)
- `status` must be one of: `active`, `pending`, `inactive`

### 1.2 Roles Table

```sql
CREATE TABLE roles (
    id          BIGSERIAL PRIMARY KEY,
    name        VARCHAR(255) NOT NULL UNIQUE,
    description TEXT,
    is_system   BOOLEAN NOT NULL DEFAULT false,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
```

### 1.3 User-Roles Junction Table

```sql
CREATE TABLE user_roles (
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role_id BIGINT NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
    PRIMARY KEY (user_id, role_id)
);

CREATE INDEX idx_user_roles_role_id ON user_roles(role_id);
```

### 1.4 Permissions Table

```sql
CREATE TABLE permissions (
    id      BIGSERIAL PRIMARY KEY,
    module  VARCHAR(100) NOT NULL,    -- e.g., "Master Data", "Transaction"
    feature VARCHAR(100) NOT NULL,    -- e.g., "Product", "Sales"
    actions TEXT[] NOT NULL            -- PostgreSQL array: {"read","create","update","delete","export"}
);

CREATE UNIQUE INDEX idx_permissions_module_feature ON permissions(module, feature);
```

### 1.5 Role-Permissions Table

```sql
CREATE TABLE role_permissions (
    id            BIGSERIAL PRIMARY KEY,
    role_id       BIGINT NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
    permission_id BIGINT NOT NULL REFERENCES permissions(id) ON DELETE CASCADE,
    actions       TEXT[] NOT NULL,     -- granted actions (subset of permission.actions)
    UNIQUE(role_id, permission_id)
);

CREATE INDEX idx_role_permissions_role_id ON role_permissions(role_id);
```

---

## 2. Seed Data

Create seed files in `backend/seeds/` that run after migrations.

### 2.1 Permissions Seed

| Module | Feature | Available Actions |
|--------|---------|-------------------|
| Master Data | Product | read, create, update, delete, export |
| Master Data | Category | read, create, update, delete |
| Master Data | Supplier | read, create, update, delete, export |
| Transaction | Sales | read, create, update, delete, export |
| Transaction | Purchase | read, create, update, delete, export |
| Report | Sales Report | read, export |
| Report | Purchase Report | read, export |
| Settings | Users | read, create, update, delete |
| Settings | Roles & Permissions | read, create, update, delete |

### 2.2 Roles Seed

| Name | Description | isSystem |
|------|-------------|----------|
| Super Admin | Full system access. Cannot be modified or deleted. | true |
| Manager | Manage products, transactions, and view reports. | false |
| Cashier | Process sales transactions. | false |
| Accountant | View transactions and generate reports. | false |
| Warehouse | Manage product stock and purchase orders. | false |

### 2.3 Role-Permissions Seed

**Super Admin** — full access is implied by `is_super_admin` flag on the user, not stored as individual role_permission entries. The backend always grants all actions to super admin users.

**Manager:**
- Master Data: Product (all), Category (all), Supplier (all)
- Transaction: Sales (read, create, update, export), Purchase (read, create, update, export)
- Report: Sales Report (read, export), Purchase Report (read, export)

**Cashier:**
- Transaction: Sales (read, create)
- Report: Sales Report (read)

**Accountant:**
- Transaction: Sales (read, export), Purchase (read, export)
- Report: Sales Report (read, export), Purchase Report (read, export)

**Warehouse:**
- Master Data: Product (read, update), Supplier (read)
- Transaction: Purchase (read, create, update)

### 2.4 Super Admin User Seed

Create one super admin user:
- Name: `Super Admin`
- Email: `admin@pointofsale.com`
- Password: `Admin@12345` (hashed with Argon2id)
- Phone: `+62-812-0000-0001`
- Status: `active`
- isSuperAdmin: `true`
- Roles: [Super Admin]

---

## 3. GORM Models

Define GORM model structs in `backend/models/` that map to the tables above. Use proper struct tags for JSON serialization (omit `password_hash` from JSON output).

```go
type User struct {
    ID             uint      `json:"id" gorm:"primaryKey"`
    Name           string    `json:"name"`
    Email          string    `json:"email" gorm:"uniqueIndex"`
    Phone          string    `json:"phone,omitempty"`
    Address        string    `json:"address,omitempty"`
    PasswordHash   string    `json:"-" gorm:"column:password_hash"`  // never expose
    ProfilePicture string    `json:"profilePicture,omitempty" gorm:"column:profile_picture"`
    Status         string    `json:"status" gorm:"default:active"`
    IsSuperAdmin   bool      `json:"isSuperAdmin" gorm:"column:is_super_admin;default:false"`
    CreatedAt      time.Time `json:"createdAt"`
    UpdatedAt      time.Time `json:"updatedAt"`
    Roles          []Role    `json:"roles" gorm:"many2many:user_roles;"`
}

type Role struct {
    ID          uint      `json:"id" gorm:"primaryKey"`
    Name        string    `json:"name" gorm:"uniqueIndex"`
    Description string    `json:"description,omitempty"`
    IsSystem    bool      `json:"isSystem" gorm:"column:is_system;default:false"`
    CreatedAt   time.Time `json:"createdAt"`
    UpdatedAt   time.Time `json:"updatedAt"`
}

type Permission struct {
    ID      uint           `json:"id" gorm:"primaryKey"`
    Module  string         `json:"module"`
    Feature string         `json:"feature"`
    Actions pq.StringArray `json:"actions" gorm:"type:text[]"`
}

type RolePermission struct {
    ID           uint           `json:"id" gorm:"primaryKey"`
    RoleID       uint           `json:"roleId" gorm:"column:role_id"`
    PermissionID uint           `json:"permissionId" gorm:"column:permission_id"`
    Actions      pq.StringArray `json:"actions" gorm:"type:text[]"`
    Role         Role           `json:"-" gorm:"foreignKey:RoleID"`
    Permission   Permission     `json:"-" gorm:"foreignKey:PermissionID"`
}
```

---

## 4. Auth API Endpoints

All endpoints return JSON. Use consistent error response format:

```json
{
  "error": "Human-readable error message",
  "code": "ERROR_CODE"
}
```

Success responses:

```json
{
  "data": { ... },
  "message": "Optional success message"
}
```

### 4.1 Register — `POST /api/v1/auth/register`

**Description**: Self-registration for new users. New users get `pending` status and must be approved by a super admin.

**Request Body:**

```json
{
  "name": "John Doe",
  "email": "john@example.com",
  "password": "SecurePass123!",
  "confirmPassword": "SecurePass123!"
}
```

**Validation:**
- `name`: required, 2-255 characters
- `email`: required, valid email format, unique (case-insensitive)
- `password`: required, min 8 characters, must contain: 1 uppercase, 1 lowercase, 1 digit, 1 special character
- `confirmPassword`: must match `password`

**Business Logic:**
1. Check email uniqueness (case-insensitive)
2. Hash password with Argon2id
3. Create user with status = `pending`, isSuperAdmin = `false`
4. Send welcome email via Mailpit (include user's name, inform them registration is pending approval)
5. Return created user (without password)

**Response (201):**

```json
{
  "data": {
    "id": 7,
    "name": "John Doe",
    "email": "john@example.com",
    "status": "pending",
    "createdAt": "2026-02-16T10:00:00Z"
  },
  "message": "Registration successful. Your account is pending approval."
}
```

**Error Responses:**
- `400` — Validation errors
- `409` — Email already registered

### 4.2 Login — `POST /api/v1/auth/login`

**Request Body:**

```json
{
  "email": "admin@pointofsale.com",
  "password": "Admin@12345"
}
```

**Validation:**
- `email`: required, valid email format
- `password`: required

**Business Logic:**
1. Find user by email (case-insensitive)
2. Verify password with Argon2id
3. Check user status: only `active` users can login
   - `pending`: return error "Account is pending approval"
   - `inactive`: return error "Account has been deactivated"
4. Generate access token (15 min) and refresh token (7 days)
5. Store refresh token in Redis with key `refresh:{jti}` and TTL = 7 days
6. Return tokens and user data (with roles)

**Response (200):**

```json
{
  "data": {
    "user": {
      "id": 1,
      "name": "Super Admin",
      "email": "admin@pointofsale.com",
      "status": "active",
      "isSuperAdmin": true,
      "roles": [
        { "id": 1, "name": "Super Admin" }
      ]
    },
    "accessToken": "eyJhbGci...",
    "refreshToken": "eyJhbGci...",
    "expiresAt": "2026-02-16T10:15:00Z"
  }
}
```

**Error Responses:**
- `400` — Validation errors
- `401` — Invalid email or password
- `403` — Account pending/inactive

### 4.3 Refresh Token — `POST /api/v1/auth/refresh`

**Request Body:**

```json
{
  "refreshToken": "eyJhbGci..."
}
```

**Business Logic:**
1. Validate the refresh token (signature, expiry)
2. Check if the token's `jti` exists in Redis (not blacklisted)
3. Blacklist the old refresh token in Redis (set with short TTL for cleanup)
4. Generate new access + refresh token pair
5. Store new refresh token in Redis

**Response (200):**

```json
{
  "data": {
    "accessToken": "eyJhbGci...",
    "refreshToken": "eyJhbGci...",
    "expiresAt": "2026-02-16T10:30:00Z"
  }
}
```

**Error Responses:**
- `401` — Invalid or expired refresh token

### 4.4 Logout — `POST /api/v1/auth/logout`

**Headers:** `Authorization: Bearer {accessToken}`

**Request Body:**

```json
{
  "refreshToken": "eyJhbGci..."
}
```

**Business Logic:**
1. Extract access token from Authorization header
2. Blacklist access token in Redis (key `blacklist:{jti}`, TTL = remaining expiry time)
3. Delete refresh token from Redis
4. Return success

**Response (200):**

```json
{
  "message": "Logged out successfully"
}
```

### 4.5 Forgot Password — `POST /api/v1/auth/forgot-password`

**Request Body:**

```json
{
  "email": "john@example.com"
}
```

**Business Logic:**
1. Find user by email
2. If user exists and is active, generate a password reset token (random 32-byte hex string)
3. Store reset token in Redis with key `reset:{token}` and value `{userId}`, TTL = 1 hour
4. Send password reset email via Mailpit (include reset link: `{FRONTEND_URL}/reset-password?token={token}`)
5. Always return success (don't reveal if email exists)

**Response (200):**

```json
{
  "message": "If the email exists, a reset link has been sent."
}
```

### 4.6 Reset Password — `POST /api/v1/auth/reset-password`

**Request Body:**

```json
{
  "token": "abc123...",
  "password": "NewSecurePass123!",
  "confirmPassword": "NewSecurePass123!"
}
```

**Business Logic:**
1. Validate token exists in Redis (`reset:{token}`)
2. Validate new password (same rules as register)
3. Hash new password with Argon2id
4. Update user's password in database
5. Delete the reset token from Redis
6. Blacklist all existing refresh tokens for this user (force re-login)

**Response (200):**

```json
{
  "message": "Password reset successfully. Please login with your new password."
}
```

**Error Responses:**
- `400` — Invalid or expired token, validation errors

### 4.7 Get Current User — `GET /api/v1/auth/me`

**Headers:** `Authorization: Bearer {accessToken}`

**Business Logic:**
1. Extract user ID from JWT claims
2. Fetch user with roles and permissions
3. For super admin, return all permissions as granted

**Response (200):**

```json
{
  "data": {
    "id": 1,
    "name": "Super Admin",
    "email": "admin@pointofsale.com",
    "phone": "+62-812-0000-0001",
    "address": "",
    "profilePicture": null,
    "status": "active",
    "isSuperAdmin": true,
    "roles": [
      { "id": 1, "name": "Super Admin" }
    ],
    "permissions": [
      {
        "module": "Master Data",
        "feature": "Product",
        "actions": ["read", "create", "update", "delete", "export"]
      }
    ]
  }
}
```

---

## 5. Auth Middleware Enhancement

Update the JWT auth middleware from Stage 1:

1. Extract Bearer token from `Authorization` header
2. Validate token signature and expiry
3. Check if token `jti` is blacklisted in Redis
4. Load user from database (cache in Redis for performance, TTL = 5 min)
5. Inject user context (`user_id`, `is_super_admin`, `roles`) into request context
6. Return 401 if any validation fails

---

## 6. Email Templates

Create simple HTML email templates for:

1. **Welcome Email** (on registration):
   - Subject: "Welcome to Point of Sale — Registration Pending"
   - Body: Greeting with user name, inform them registration is pending admin approval

2. **Password Reset Email**:
   - Subject: "Point of Sale — Password Reset"
   - Body: Reset link (valid for 1 hour), security notice

3. **Account Approved Email** (sent when admin approves a pending user):
   - Subject: "Point of Sale — Account Approved"
   - Body: Inform user they can now login

Use Go's `html/template` for email rendering. Send via SMTP (Mailpit in development).

---

## 7. Utility Functions

### Password Hashing (`utils/password.go`)
- `HashPassword(password string) (string, error)` — Argon2id hash
- `VerifyPassword(password, hash string) (bool, error)` — verify against hash
- Parameters: memory=64MB, iterations=3, parallelism=4, saltLength=16, keyLength=32

### JWT (`utils/jwt.go`)
- `GenerateAccessToken(user User) (string, error)`
- `GenerateRefreshToken(user User) (string, error)`
- `ValidateToken(tokenString string, secret string) (*Claims, error)`
- `GenerateResetToken() (string, error)` — crypto/rand hex string

### Validation (`utils/validation.go`)
- `ValidateEmail(email string) bool`
- `ValidatePassword(password string) []string` — returns list of unmet requirements
- `ValidateRequired(field, name string) string`

---

## 8. TDD Workflow

**Follow strict TDD for this stage. Write tests BEFORE implementation.**

### 8.1 Test Infrastructure (do first)

Set up the shared test infrastructure before writing any feature tests:

1. Create `backend/testutil/db.go` — test database connection, migration runner, cleanup
2. Create `backend/testutil/fixtures.go` — factory functions for users, roles
3. Create `backend/testutil/auth.go` — helper to generate JWT tokens for test requests
4. Create `backend/testutil/assert.go` — custom JSON response assertion helpers
5. Configure a `pointofsale_test` database in docker-compose (or use the same postgres with a separate DB)

### 8.2 TDD Order — Utilities First

Write and test utility functions before anything else:

**`utils/password_test.go`** (write tests first, then implement):
- `TestHashPassword_ValidPassword_ReturnsHash` — hash a password, verify it's not empty and not equal to the input
- `TestHashPassword_DifferentSalts_DifferentHashes` — same password hashed twice produces different hashes
- `TestVerifyPassword_CorrectPassword_ReturnsTrue`
- `TestVerifyPassword_WrongPassword_ReturnsFalse`
- `TestVerifyPassword_InvalidHash_ReturnsError`

**`utils/jwt_test.go`**:
- `TestGenerateAccessToken_ValidUser_ReturnsToken`
- `TestGenerateRefreshToken_ValidUser_ReturnsToken`
- `TestValidateToken_ValidToken_ReturnsClaims`
- `TestValidateToken_ExpiredToken_ReturnsError`
- `TestValidateToken_InvalidSignature_ReturnsError`
- `TestValidateToken_MalformedToken_ReturnsError`

**`utils/validation_test.go`**:
- `TestValidateEmail_ValidFormats_ReturnsTrue` — test multiple valid emails
- `TestValidateEmail_InvalidFormats_ReturnsFalse` — test multiple invalid emails
- `TestValidatePassword_StrongPassword_ReturnsNoErrors`
- `TestValidatePassword_TooShort_ReturnsError`
- `TestValidatePassword_NoUppercase_ReturnsError`
- `TestValidatePassword_NoDigit_ReturnsError`
- `TestValidatePassword_NoSpecialChar_ReturnsError`

### 8.3 TDD Order — Repository Layer

**`repositories/user_repository_test.go`**:
- `TestCreateUser_ValidUser_Succeeds`
- `TestCreateUser_DuplicateEmail_ReturnsError`
- `TestFindUserByEmail_Exists_ReturnsUser`
- `TestFindUserByEmail_NotExists_ReturnsNil`
- `TestFindUserByID_WithRoles_EagerLoadsRoles`

### 8.4 TDD Order — Service Layer

**`services/auth_service_test.go`** (use mocked repository):
- `TestRegister_ValidInput_CreatesUserWithPendingStatus`
- `TestRegister_DuplicateEmail_ReturnsConflictError`
- `TestRegister_PasswordMismatch_ReturnsValidationError`
- `TestRegister_WeakPassword_ReturnsValidationError`
- `TestLogin_ActiveUser_ReturnsTokens`
- `TestLogin_PendingUser_ReturnsForbiddenError`
- `TestLogin_InactiveUser_ReturnsForbiddenError`
- `TestLogin_WrongPassword_ReturnsUnauthorizedError`
- `TestLogin_NonExistentEmail_ReturnsUnauthorizedError`
- `TestRefreshToken_ValidToken_ReturnsNewPair`
- `TestRefreshToken_BlacklistedToken_ReturnsError`
- `TestRefreshToken_ExpiredToken_ReturnsError`
- `TestLogout_ValidTokens_BlacklistsBoth`
- `TestForgotPassword_ExistingEmail_StoresTokenInRedis`
- `TestForgotPassword_NonExistingEmail_StillReturnsSuccess`
- `TestResetPassword_ValidToken_UpdatesPassword`
- `TestResetPassword_ExpiredToken_ReturnsError`
- `TestGetCurrentUser_ValidId_ReturnsUserWithPermissions`

### 8.5 TDD Order — Handler Layer (Integration Tests)

**`handlers/auth_handler_test.go`** (full HTTP integration tests):
- `TestRegisterHandler_ValidBody_Returns201`
- `TestRegisterHandler_MissingFields_Returns400`
- `TestRegisterHandler_DuplicateEmail_Returns409`
- `TestLoginHandler_ValidCredentials_Returns200WithTokens`
- `TestLoginHandler_InvalidCredentials_Returns401`
- `TestLoginHandler_PendingUser_Returns403`
- `TestRefreshHandler_ValidToken_Returns200`
- `TestRefreshHandler_InvalidToken_Returns401`
- `TestLogoutHandler_Authenticated_Returns200`
- `TestLogoutHandler_NoAuth_Returns401`
- `TestForgotPasswordHandler_ValidEmail_Returns200`
- `TestResetPasswordHandler_ValidToken_Returns200`
- `TestMeHandler_Authenticated_ReturnsUserData`
- `TestMeHandler_NoAuth_Returns401`

### 8.6 TDD Order — Middleware

**`middleware/auth_middleware_test.go`**:
- `TestAuthMiddleware_ValidToken_SetsUserContext`
- `TestAuthMiddleware_NoAuthHeader_Returns401`
- `TestAuthMiddleware_InvalidToken_Returns401`
- `TestAuthMiddleware_BlacklistedToken_Returns401`
- `TestAuthMiddleware_ExpiredToken_Returns401`

---

## 9. Deliverables

After completing this stage:

1. All auth-related tables exist in PostgreSQL (created by goose migrations)
2. Seed data is populated (super admin user, roles, permissions, role-permissions)
3. `POST /api/v1/auth/register` creates a pending user and sends welcome email
4. `POST /api/v1/auth/login` returns JWT tokens for active users
5. `POST /api/v1/auth/refresh` rotates token pairs
6. `POST /api/v1/auth/logout` blacklists tokens
7. `POST /api/v1/auth/forgot-password` sends reset email via Mailpit
8. `POST /api/v1/auth/reset-password` resets password with valid token
9. `GET /api/v1/auth/me` returns authenticated user with roles/permissions
10. Emails are visible in Mailpit Web UI at `http://localhost:8025`
11. **All tests pass** (`go test ./...` from `backend/`)
12. **Test coverage** for utils, services, and handlers ≥ 80%
