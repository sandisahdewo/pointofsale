package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/lib/pq"
	"github.com/pointofsale/backend/models"
	"github.com/pointofsale/backend/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRequirePermission_SuperAdmin_AlwaysAllows(t *testing.T) {
	// Setup test DB and Redis
	db := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, db)
	rdb := testutil.SetupTestRedis(t)
	defer testutil.CleanupTestRedis(t, rdb)

	// Create super admin user
	superAdmin := testutil.CreateTestSuperAdmin(t, db)

	// Create permission middleware
	permMiddleware := NewPermissionMiddleware(db, rdb)

	// Create a test handler that sets status 200
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("success"))
	})

	// Wrap with permission middleware requiring a permission
	handler := permMiddleware.RequirePermission("Settings", "Users", "delete")(testHandler)

	// Create request with super admin context
	req := httptest.NewRequest("DELETE", "/test", nil)
	ctx := context.WithValue(req.Context(), UserIDKey, superAdmin.ID)
	ctx = context.WithValue(ctx, IsSuperAdminKey, true)
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	// Assert: Super admin should always be allowed
	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, "success", rr.Body.String())
}

func TestRequirePermission_UserWithPermission_Allows(t *testing.T) {
	// Setup
	db := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, db)
	rdb := testutil.SetupTestRedis(t)
	defer testutil.CleanupTestRedis(t, rdb)

	// Create permission
	permission := testutil.CreateTestPermission(t, db, func(p *models.Permission) {
		p.Module = "Settings"
		p.Feature = "Users"
		p.Actions = pq.StringArray{"read", "create", "update", "delete"}
	})

	// Create role with permission
	role := testutil.CreateTestRole(t, db, func(r *models.Role) {
		r.Name = "Manager"
	})

	// Grant permission to role
	rolePermission := &models.RolePermission{
		RoleID:       role.ID,
		PermissionID: permission.ID,
		Actions:      pq.StringArray{"read", "create", "update"},
	}
	require.NoError(t, db.Create(rolePermission).Error)

	// Create user with role
	user := testutil.CreateTestUser(t, db, func(u *models.User) {
		u.Name = "Test Manager"
		u.Roles = []models.Role{*role}
	})

	// Create middleware and handler
	permMiddleware := NewPermissionMiddleware(db, rdb)
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("success"))
	})

	// Test with "update" action (granted)
	handler := permMiddleware.RequirePermission("Settings", "Users", "update")(testHandler)
	req := httptest.NewRequest("PUT", "/test", nil)
	ctx := context.WithValue(req.Context(), UserIDKey, user.ID)
	ctx = context.WithValue(ctx, IsSuperAdminKey, false)
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	// Assert: User should be allowed
	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestRequirePermission_UserWithoutPermission_Returns403(t *testing.T) {
	// Setup
	db := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, db)
	rdb := testutil.SetupTestRedis(t)
	defer testutil.CleanupTestRedis(t, rdb)

	// Create permission
	permission := testutil.CreateTestPermission(t, db, func(p *models.Permission) {
		p.Module = "Settings"
		p.Feature = "Users"
		p.Actions = pq.StringArray{"read", "create", "update", "delete"}
	})

	// Create role with limited permission (no "delete" action)
	role := testutil.CreateTestRole(t, db, func(r *models.Role) {
		r.Name = "Viewer"
	})

	rolePermission := &models.RolePermission{
		RoleID:       role.ID,
		PermissionID: permission.ID,
		Actions:      pq.StringArray{"read"}, // Only read, no delete
	}
	require.NoError(t, db.Create(rolePermission).Error)

	// Create user with role
	user := testutil.CreateTestUser(t, db, func(u *models.User) {
		u.Roles = []models.Role{*role}
	})

	// Create middleware and handler
	permMiddleware := NewPermissionMiddleware(db, rdb)
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("success"))
	})

	// Test with "delete" action (NOT granted)
	handler := permMiddleware.RequirePermission("Settings", "Users", "delete")(testHandler)
	req := httptest.NewRequest("DELETE", "/test", nil)
	ctx := context.WithValue(req.Context(), UserIDKey, user.ID)
	ctx = context.WithValue(ctx, IsSuperAdminKey, false)
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	// Assert: Should return 403
	assert.Equal(t, http.StatusForbidden, rr.Code)

	var response map[string]string
	err := json.Unmarshal(rr.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "You don't have permission to perform this action", response["error"])
	assert.Equal(t, "FORBIDDEN", response["code"])
}

func TestRequirePermission_UserWithMultipleRoles_ChecksAll(t *testing.T) {
	// Setup
	db := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, db)
	rdb := testutil.SetupTestRedis(t)
	defer testutil.CleanupTestRedis(t, rdb)

	// Create permission
	permission := testutil.CreateTestPermission(t, db, func(p *models.Permission) {
		p.Module = "Inventory"
		p.Feature = "Stock"
		p.Actions = pq.StringArray{"read", "update"}
	})

	// Create two roles: role1 has no permission, role2 has permission
	role1 := testutil.CreateTestRole(t, db, func(r *models.Role) {
		r.Name = "Cashier"
	})

	role2 := testutil.CreateTestRole(t, db, func(r *models.Role) {
		r.Name = "Warehouse"
	})

	// Grant permission only to role2
	rolePermission := &models.RolePermission{
		RoleID:       role2.ID,
		PermissionID: permission.ID,
		Actions:      pq.StringArray{"read", "update"},
	}
	require.NoError(t, db.Create(rolePermission).Error)

	// Create user with BOTH roles
	user := testutil.CreateTestUser(t, db, func(u *models.User) {
		u.Roles = []models.Role{*role1, *role2}
	})

	// Create middleware and handler
	permMiddleware := NewPermissionMiddleware(db, rdb)
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("success"))
	})

	handler := permMiddleware.RequirePermission("Inventory", "Stock", "update")(testHandler)
	req := httptest.NewRequest("PUT", "/test", nil)
	ctx := context.WithValue(req.Context(), UserIDKey, user.ID)
	ctx = context.WithValue(ctx, IsSuperAdminKey, false)
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	// Assert: User should be allowed because one of their roles grants permission
	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestRequirePermission_CachedPermissions_UsesCache(t *testing.T) {
	// Setup
	db := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, db)
	rdb := testutil.SetupTestRedis(t)
	defer testutil.CleanupTestRedis(t, rdb)

	// Create permission
	permission := testutil.CreateTestPermission(t, db, func(p *models.Permission) {
		p.Module = "Settings"
		p.Feature = "Roles"
		p.Actions = pq.StringArray{"read", "update"}
	})

	role := testutil.CreateTestRole(t, db)
	rolePermission := &models.RolePermission{
		RoleID:       role.ID,
		PermissionID: permission.ID,
		Actions:      pq.StringArray{"read", "update"},
	}
	require.NoError(t, db.Create(rolePermission).Error)

	user := testutil.CreateTestUser(t, db, func(u *models.User) {
		u.Roles = []models.Role{*role}
	})

	permMiddleware := NewPermissionMiddleware(db, rdb)
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	handler := permMiddleware.RequirePermission("Settings", "Roles", "read")(testHandler)

	// First request: should cache permissions
	req1 := httptest.NewRequest("GET", "/test", nil)
	ctx1 := context.WithValue(req1.Context(), UserIDKey, user.ID)
	ctx1 = context.WithValue(ctx1, IsSuperAdminKey, false)
	req1 = req1.WithContext(ctx1)

	rr1 := httptest.NewRecorder()
	handler.ServeHTTP(rr1, req1)
	assert.Equal(t, http.StatusOK, rr1.Code)

	// Verify cache was set
	cacheKey := buildPermissionCacheKey(user.ID)
	cachedData := rdb.Get(context.Background(), cacheKey).Val()
	assert.NotEmpty(t, cachedData, "permissions should be cached in Redis")

	// Second request: should use cache (we can't directly verify cache hit,
	// but we can verify the result is still correct)
	req2 := httptest.NewRequest("GET", "/test", nil)
	ctx2 := context.WithValue(req2.Context(), UserIDKey, user.ID)
	ctx2 = context.WithValue(ctx2, IsSuperAdminKey, false)
	req2 = req2.WithContext(ctx2)

	rr2 := httptest.NewRecorder()
	handler.ServeHTTP(rr2, req2)
	assert.Equal(t, http.StatusOK, rr2.Code)
}

func TestRequirePermission_CacheInvalidation_RefreshesAfterRoleChange(t *testing.T) {
	// Setup
	db := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, db)
	rdb := testutil.SetupTestRedis(t)
	defer testutil.CleanupTestRedis(t, rdb)

	// Create permission
	permission := testutil.CreateTestPermission(t, db, func(p *models.Permission) {
		p.Module = "Reports"
		p.Feature = "Sales"
		p.Actions = pq.StringArray{"read", "export"}
	})

	role := testutil.CreateTestRole(t, db)
	rolePermission := &models.RolePermission{
		RoleID:       role.ID,
		PermissionID: permission.ID,
		Actions:      pq.StringArray{"read"}, // Initially only read
	}
	require.NoError(t, db.Create(rolePermission).Error)

	user := testutil.CreateTestUser(t, db, func(u *models.User) {
		u.Roles = []models.Role{*role}
	})

	permMiddleware := NewPermissionMiddleware(db, rdb)
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// First request: user should be denied "export" action
	handler := permMiddleware.RequirePermission("Reports", "Sales", "export")(testHandler)
	req1 := httptest.NewRequest("POST", "/test", nil)
	ctx1 := context.WithValue(req1.Context(), UserIDKey, user.ID)
	ctx1 = context.WithValue(ctx1, IsSuperAdminKey, false)
	req1 = req1.WithContext(ctx1)

	rr1 := httptest.NewRecorder()
	handler.ServeHTTP(rr1, req1)
	assert.Equal(t, http.StatusForbidden, rr1.Code)

	// Update role permissions to grant "export" action
	rolePermission.Actions = pq.StringArray{"read", "export"}
	require.NoError(t, db.Save(rolePermission).Error)

	// Invalidate cache
	err := InvalidatePermissionCache(context.Background(), rdb, user.ID)
	require.NoError(t, err)

	// Second request: user should now be allowed "export" action
	req2 := httptest.NewRequest("POST", "/test", nil)
	ctx2 := context.WithValue(req2.Context(), UserIDKey, user.ID)
	ctx2 = context.WithValue(ctx2, IsSuperAdminKey, false)
	req2 = req2.WithContext(ctx2)

	rr2 := httptest.NewRecorder()
	handler.ServeHTTP(rr2, req2)
	assert.Equal(t, http.StatusOK, rr2.Code)
}

func TestRequirePermission_NoUserInContext_Returns401(t *testing.T) {
	// Setup
	db := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, db)
	rdb := testutil.SetupTestRedis(t)
	defer testutil.CleanupTestRedis(t, rdb)

	permMiddleware := NewPermissionMiddleware(db, rdb)
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	handler := permMiddleware.RequirePermission("Settings", "Users", "read")(testHandler)

	// Request without user context (should be set by auth middleware)
	req := httptest.NewRequest("GET", "/test", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	// Assert: Should return 401 unauthorized
	assert.Equal(t, http.StatusUnauthorized, rr.Code)
}
