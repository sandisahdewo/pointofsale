# Backend Stage 3 — User Management & Roles/Permissions APIs

## Overview

Build CRUD APIs for user management and role/permission management. Add authorization middleware to enforce permissions on protected endpoints.

> **Prerequisite**: Stage 2 must be complete (auth tables, seeds, and auth APIs).

---

## 1. User Management API

All endpoints require authentication. Access control is noted per endpoint.

### 1.1 List Users — `GET /api/v1/users`

**Permission**: `Settings > Users > read`

**Query Parameters:**

| Param | Type | Default | Description |
|-------|------|---------|-------------|
| `page` | int | 1 | Page number |
| `pageSize` | int | 10 | Items per page (max: 100) |
| `search` | string | — | Search by name or email (case-insensitive, partial match) |
| `sortBy` | string | `id` | Sort field: `id`, `name`, `email`, `status`, `createdAt` |
| `sortDir` | string | `asc` | Sort direction: `asc` or `desc` |
| `status` | string | — | Filter by status: `active`, `pending`, `inactive` |

**Response (200):**

```json
{
  "data": [
    {
      "id": 1,
      "name": "Super Admin",
      "email": "admin@pointofsale.com",
      "phone": "+62-812-0000-0001",
      "profilePicture": null,
      "status": "active",
      "isSuperAdmin": true,
      "roles": [
        { "id": 1, "name": "Super Admin" }
      ],
      "createdAt": "2026-01-01T00:00:00Z"
    }
  ],
  "meta": {
    "page": 1,
    "pageSize": 10,
    "totalItems": 6,
    "totalPages": 1
  }
}
```

**Notes:**
- Always eager-load roles with user data
- Never include `passwordHash` in any response
- Support combined search + status filter

### 1.2 Get User — `GET /api/v1/users/:id`

**Permission**: `Settings > Users > read`

**Response (200):** Single user object (same shape as list item, plus `address` field).

**Error:** `404` if user not found.

### 1.3 Create User — `POST /api/v1/users`

**Permission**: `Settings > Users > create`

**Description**: Admin creates a new user. Unlike self-registration, admin-created users get `active` status immediately.

**Request Body:**

```json
{
  "name": "Budi Santoso",
  "email": "budi@pointofsale.com",
  "phone": "+62-812-0000-0002",
  "address": "Jakarta",
  "roleIds": [2],
  "profilePicture": null
}
```

**Validation:**
- `name`: required, 2-255 characters
- `email`: required, valid format, unique (case-insensitive)
- `phone`: optional
- `address`: optional
- `roleIds`: optional, array of valid role IDs

**Business Logic:**
1. Validate email uniqueness
2. Generate a random temporary password (16 chars, alphanumeric + special)
3. Hash password with Argon2id
4. Create user with status = `active`
5. Assign roles via `user_roles` junction table
6. Send credentials email via Mailpit (include temporary password, prompt to change on first login)
7. Return created user

**Response (201):** Created user object.

**Error Responses:**
- `400` — Validation errors
- `409` — Email already registered

### 1.4 Update User — `PUT /api/v1/users/:id`

**Permission**: `Settings > Users > update`

**Request Body:**

```json
{
  "name": "Budi Santoso Updated",
  "email": "budi.new@pointofsale.com",
  "phone": "+62-812-0000-9999",
  "address": "Surabaya",
  "roleIds": [2, 5],
  "status": "inactive",
  "profilePicture": "data:image/png;base64,..."
}
```

**Validation:**
- Same as create, plus:
- `status`: must be one of `active`, `inactive` (cannot set to `pending` via update)
- If user `isSuperAdmin`: cannot change `status` or `isSuperAdmin`
- Email uniqueness check must exclude the current user

**Business Logic:**
1. Find user by ID (404 if not found)
2. Apply super admin restrictions
3. Update user fields
4. Sync roles: delete existing `user_roles`, insert new ones
5. Return updated user

**Response (200):** Updated user object.

### 1.5 Delete User — `DELETE /api/v1/users/:id`

**Permission**: `Settings > Users > delete`

**Business Logic:**
1. Find user by ID (404 if not found)
2. Block deletion of super admin users → return `403` with message "Super admin cannot be deleted"
3. Block self-deletion → return `403` with message "Cannot delete your own account"
4. Delete user (CASCADE removes user_roles entries)

**Response (200):**

```json
{
  "message": "User deleted successfully"
}
```

### 1.6 Approve User — `PATCH /api/v1/users/:id/approve`

**Permission**: `Settings > Users > update`

**Business Logic:**
1. Find user by ID (404 if not found)
2. Check status is `pending` → return `400` if not
3. Update status to `active`
4. Send account approved email via Mailpit

**Response (200):**

```json
{
  "data": { /* updated user */ },
  "message": "User approved successfully"
}
```

### 1.7 Reject User — `DELETE /api/v1/users/:id/reject`

**Permission**: `Settings > Users > delete`

**Business Logic:**
1. Find user by ID (404 if not found)
2. Check status is `pending` → return `400` if not
3. Delete user from database
4. Send rejection email via Mailpit

**Response (200):**

```json
{
  "message": "User registration rejected"
}
```

### 1.8 Profile Picture Upload — `POST /api/v1/users/:id/profile-picture`

**Permission**: `Settings > Users > update` (or user updating own profile)

**Request**: `multipart/form-data` with `image` field.

**Validation:**
- File type: JPEG, PNG, WebP only
- Max file size: 2 MB
- Recommended dimensions: 200x200 px (resize/crop on upload)

**Business Logic:**
1. Validate file
2. Save to disk (e.g., `backend/uploads/profiles/{userId}_{timestamp}.{ext}`)
3. Update user's `profile_picture` field with the URL/path
4. Delete old profile picture file if exists

**Response (200):**

```json
{
  "data": {
    "profilePicture": "/uploads/profiles/1_1708099200.jpg"
  }
}
```

**Serving Uploads:**
- Configure a static file server in Chi for the `/uploads/` path
- In production, use cloud storage (S3, etc.) — for now, local disk is fine

---

## 2. Role Management API

### 2.1 List Roles — `GET /api/v1/roles`

**Permission**: `Settings > Roles & Permissions > read`

**Query Parameters:**

| Param | Type | Default | Description |
|-------|------|---------|-------------|
| `page` | int | 1 | Page number |
| `pageSize` | int | 10 | Items per page (max: 100) |
| `search` | string | — | Search by name or description |
| `sortBy` | string | `id` | Sort field: `id`, `name`, `description` |
| `sortDir` | string | `asc` | Sort direction |

**Response (200):**

```json
{
  "data": [
    {
      "id": 1,
      "name": "Super Admin",
      "description": "Full system access. Cannot be modified or deleted.",
      "isSystem": true,
      "userCount": 1,
      "createdAt": "2026-01-01T00:00:00Z"
    }
  ],
  "meta": {
    "page": 1,
    "pageSize": 10,
    "totalItems": 5,
    "totalPages": 1
  }
}
```

**Notes:**
- `userCount` is computed from `user_roles` junction table (COUNT of users per role)
- Include `userCount` in list response

### 2.2 Get Role — `GET /api/v1/roles/:id`

**Permission**: `Settings > Roles & Permissions > read`

**Response (200):** Single role object with `userCount`.

### 2.3 Create Role — `POST /api/v1/roles`

**Permission**: `Settings > Roles & Permissions > create`

**Request Body:**

```json
{
  "name": "Supervisor",
  "description": "Supervise daily operations"
}
```

**Validation:**
- `name`: required, unique (case-insensitive), 2-255 characters
- `description`: optional

**Response (201):** Created role object.

### 2.4 Update Role — `PUT /api/v1/roles/:id`

**Permission**: `Settings > Roles & Permissions > update`

**Business Logic:**
- Block editing system roles (`isSystem: true`) → return `403`
- Name uniqueness must exclude current role

**Response (200):** Updated role object.

### 2.5 Delete Role — `DELETE /api/v1/roles/:id`

**Permission**: `Settings > Roles & Permissions > delete`

**Business Logic:**
1. Block deletion of system roles → return `403`
2. Delete role (CASCADE removes `user_roles` and `role_permissions` entries)
3. Return success

**Response (200):**

```json
{
  "message": "Role deleted successfully"
}
```

---

## 3. Permission Management API

### 3.1 List All Permissions — `GET /api/v1/permissions`

**Permission**: `Settings > Roles & Permissions > read`

Returns the full permission tree (seed data). Not paginated — always returns all permissions.

**Response (200):**

```json
{
  "data": [
    {
      "id": 1,
      "module": "Master Data",
      "feature": "Product",
      "actions": ["read", "create", "update", "delete", "export"]
    },
    {
      "id": 2,
      "module": "Master Data",
      "feature": "Category",
      "actions": ["read", "create", "update", "delete"]
    }
  ]
}
```

### 3.2 Get Role Permissions — `GET /api/v1/roles/:id/permissions`

**Permission**: `Settings > Roles & Permissions > read`

Returns the permission assignments for a specific role.

**Response (200):**

```json
{
  "data": {
    "roleId": 2,
    "roleName": "Manager",
    "isSystem": false,
    "permissions": [
      {
        "permissionId": 1,
        "module": "Master Data",
        "feature": "Product",
        "availableActions": ["read", "create", "update", "delete", "export"],
        "grantedActions": ["read", "create", "update", "delete", "export"]
      },
      {
        "permissionId": 2,
        "module": "Master Data",
        "feature": "Category",
        "availableActions": ["read", "create", "update", "delete"],
        "grantedActions": ["read", "create", "update", "delete"]
      }
    ]
  }
}
```

**For Super Admin role:** Return all permissions with all available actions granted (computed, not stored).

### 3.3 Update Role Permissions — `PUT /api/v1/roles/:id/permissions`

**Permission**: `Settings > Roles & Permissions > update`

**Request Body:**

```json
{
  "permissions": [
    {
      "permissionId": 1,
      "actions": ["read", "create", "update"]
    },
    {
      "permissionId": 2,
      "actions": ["read"]
    }
  ]
}
```

**Validation:**
- Block updating system role permissions → return `403`
- Each `permissionId` must exist
- Each action must be a valid action for the given permission
- Omitted permissions = no access (delete existing `role_permissions` for those)

**Business Logic:**
1. Delete all existing `role_permissions` for this role
2. Insert new entries from request body
3. Filter out invalid actions (actions not in the permission's `actions` array)
4. Return updated role permissions

**Response (200):** Same format as `GET /api/v1/roles/:id/permissions`.

---

## 4. Authorization Middleware

Create a reusable authorization middleware for protecting endpoints.

### 4.1 Permission Check Middleware

```go
// RequirePermission returns middleware that checks if the user has the specified action
// on the specified module/feature.
func RequirePermission(module, feature, action string) func(http.Handler) http.Handler
```

**Logic:**
1. Extract user from request context (set by auth middleware)
2. If `user.isSuperAdmin` → always allow (bypass all checks)
3. Load user's role-permission mappings:
   - Get user's role IDs from `user_roles`
   - Get all `role_permissions` for those role IDs
   - Check if any role grants the required `action` on the specified `module` + `feature`
4. If allowed → proceed to handler
5. If denied → return `403 Forbidden`

**Caching:**
- Cache the user's computed permissions in Redis (key: `perms:{userId}`, TTL: 5 min)
- Invalidate cache when roles or role_permissions are updated

### 4.2 Usage Example

```go
r.Route("/api/v1/users", func(r chi.Router) {
    r.Use(middleware.RequireAuth)
    r.With(middleware.RequirePermission("Settings", "Users", "read")).Get("/", handlers.ListUsers)
    r.With(middleware.RequirePermission("Settings", "Users", "create")).Post("/", handlers.CreateUser)
    r.With(middleware.RequirePermission("Settings", "Users", "update")).Put("/{id}", handlers.UpdateUser)
    r.With(middleware.RequirePermission("Settings", "Users", "delete")).Delete("/{id}", handlers.DeleteUser)
})
```

### 4.3 Error Response

```json
{
  "error": "You don't have permission to perform this action",
  "code": "FORBIDDEN"
}
```

---

## 5. Pagination Helper

Create a reusable pagination utility for list endpoints.

```go
type PaginationParams struct {
    Page     int    `json:"page"`
    PageSize int    `json:"pageSize"`
    Search   string `json:"search"`
    SortBy   string `json:"sortBy"`
    SortDir  string `json:"sortDir"`
}

type PaginationMeta struct {
    Page       int `json:"page"`
    PageSize   int `json:"pageSize"`
    TotalItems int `json:"totalItems"`
    TotalPages int `json:"totalPages"`
}

type PaginatedResponse struct {
    Data interface{}    `json:"data"`
    Meta PaginationMeta `json:"meta"`
}
```

- Parse pagination params from query string with defaults
- Validate `pageSize` (min 1, max 100)
- Validate `sortDir` (must be `asc` or `desc`)
- Validate `sortBy` against allowed fields per endpoint (prevent SQL injection)
- Calculate `offset` = (page - 1) * pageSize
- Apply GORM `Offset()`, `Limit()`, `Order()`, `Count()` for total

---

## 6. Additional Seeds

Seed additional test users (beyond the super admin from Stage 2):

| Name | Email | Phone | Roles | Status |
|------|-------|-------|-------|--------|
| Budi Santoso | budi@pointofsale.com | +62-812-0000-0002 | Manager | active |
| Siti Rahayu | siti@pointofsale.com | +62-812-0000-0003 | Cashier | active |
| Ahmad Wijaya | ahmad@pointofsale.com | +62-812-0000-0004 | Warehouse, Accountant | active |
| Dewi Lestari | dewi@pointofsale.com | +62-812-0000-0005 | Cashier | inactive |
| Rizky Pratama | rizky@pointofsale.com | +62-812-0000-0006 | — | pending |

All users have password: `Password@123` (hashed).

---

## 7. TDD Workflow

**Follow strict TDD. Write all tests BEFORE their implementations.**

### 7.1 TDD Order — Pagination Helper (utility, test first)

**`utils/pagination_test.go`**:
- `TestParsePaginationParams_Defaults_ReturnsDefaults`
- `TestParsePaginationParams_ValidValues_ParsesCorrectly`
- `TestParsePaginationParams_PageSizeExceedsMax_CapsAt100`
- `TestParsePaginationParams_InvalidSortDir_DefaultsToAsc`
- `TestParsePaginationParams_InvalidSortBy_ReturnsError`

### 7.2 TDD Order — User Repository

**`repositories/user_repository_test.go`** (extend from Stage 2):
- `TestListUsers_Pagination_ReturnsCorrectPage`
- `TestListUsers_SearchByName_FiltersCorrectly`
- `TestListUsers_SearchByEmail_FiltersCorrectly`
- `TestListUsers_FilterByStatus_ReturnsMatchingOnly`
- `TestListUsers_SortByName_ReturnsOrdered`
- `TestListUsers_CombinedSearchAndFilter_Works`
- `TestUpdateUser_ValidData_UpdatesFields`
- `TestUpdateUser_SyncRoles_ReplacesRoles`
- `TestDeleteUser_Exists_RemovesUser`
- `TestDeleteUser_CascadesUserRoles`

### 7.3 TDD Order — User Service

**`services/user_service_test.go`** (mocked repository):
- `TestCreateUser_ValidInput_GeneratesPasswordAndSendsEmail`
- `TestCreateUser_DuplicateEmail_ReturnsConflict`
- `TestUpdateUser_SuperAdmin_BlocksStatusChange`
- `TestUpdateUser_NonExistent_ReturnsNotFound`
- `TestDeleteUser_SuperAdmin_ReturnsForbidden`
- `TestDeleteUser_SelfDeletion_ReturnsForbidden`
- `TestApproveUser_PendingUser_SetsActive`
- `TestApproveUser_ActiveUser_ReturnsBadRequest`
- `TestRejectUser_PendingUser_DeletesUser`
- `TestRejectUser_ActiveUser_ReturnsBadRequest`

### 7.4 TDD Order — User Handler (Integration Tests)

**`handlers/user_handler_test.go`**:
- `TestListUsers_Authenticated_Returns200WithPagination`
- `TestListUsers_WithSearch_FiltersResults`
- `TestListUsers_NoAuth_Returns401`
- `TestListUsers_NoPermission_Returns403`
- `TestGetUser_Exists_Returns200`
- `TestGetUser_NotFound_Returns404`
- `TestCreateUser_ValidBody_Returns201`
- `TestCreateUser_DuplicateEmail_Returns409`
- `TestCreateUser_MissingName_Returns400`
- `TestUpdateUser_ValidBody_Returns200`
- `TestUpdateUser_SuperAdminStatusChange_Returns403`
- `TestDeleteUser_Regular_Returns200`
- `TestDeleteUser_SuperAdmin_Returns403`
- `TestDeleteUser_Self_Returns403`
- `TestApproveUser_Pending_Returns200`
- `TestApproveUser_Active_Returns400`
- `TestRejectUser_Pending_Returns200`
- `TestUploadProfilePicture_ValidImage_Returns200`
- `TestUploadProfilePicture_InvalidFileType_Returns400`
- `TestUploadProfilePicture_TooLarge_Returns400`

### 7.5 TDD Order — Role Repository & Service

**`repositories/role_repository_test.go`**:
- `TestListRoles_WithUserCount_ReturnsCorrectCounts`
- `TestCreateRole_ValidData_Succeeds`
- `TestCreateRole_DuplicateName_ReturnsError`
- `TestDeleteRole_CascadesPermissionsAndUserRoles`

**`services/role_service_test.go`**:
- `TestCreateRole_Valid_Succeeds`
- `TestCreateRole_DuplicateName_ReturnsConflict`
- `TestUpdateRole_SystemRole_ReturnsForbidden`
- `TestDeleteRole_SystemRole_ReturnsForbidden`
- `TestDeleteRole_Regular_CleansUpRelations`

### 7.6 TDD Order — Role Handler

**`handlers/role_handler_test.go`**:
- `TestListRoles_Returns200WithUserCounts`
- `TestCreateRole_ValidBody_Returns201`
- `TestCreateRole_DuplicateName_Returns409`
- `TestUpdateRole_SystemRole_Returns403`
- `TestDeleteRole_SystemRole_Returns403`
- `TestDeleteRole_Regular_Returns200`

### 7.7 TDD Order — Permissions

**`handlers/permission_handler_test.go`**:
- `TestListPermissions_Returns200WithAllPermissions`
- `TestGetRolePermissions_RegularRole_ReturnsGrantedActions`
- `TestGetRolePermissions_SuperAdmin_ReturnsAllGranted`
- `TestUpdateRolePermissions_ValidData_Returns200`
- `TestUpdateRolePermissions_SystemRole_Returns403`
- `TestUpdateRolePermissions_InvalidPermissionId_Returns400`
- `TestUpdateRolePermissions_InvalidAction_FiltersOut`

### 7.8 TDD Order — Authorization Middleware

**`middleware/permission_middleware_test.go`**:
- `TestRequirePermission_SuperAdmin_AlwaysAllows`
- `TestRequirePermission_UserWithPermission_Allows`
- `TestRequirePermission_UserWithoutPermission_Returns403`
- `TestRequirePermission_UserWithMultipleRoles_ChecksAll`
- `TestRequirePermission_CachedPermissions_UsesCache`
- `TestRequirePermission_CacheInvalidation_RefreshesAfterRoleChange`

---

## 8. Deliverables

After completing this stage:

1. Full user CRUD works (list, get, create, update, delete)
2. User approval and rejection flows work with email notifications
3. Profile picture upload and serving works
4. Full role CRUD works (with system role protection)
5. Permission listing and role-permission assignment works
6. Authorization middleware protects all endpoints by module/feature/action
7. Super admin bypasses all permission checks
8. Pagination, search, sort work consistently across list endpoints
9. All user/role changes invalidate relevant permission caches
10. **All tests pass** (`go test ./...`)
11. **Test coverage** ≥ 80% across handlers and services
