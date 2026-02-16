package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/go-chi/chi/v5"
	"github.com/lib/pq"
	"github.com/pointofsale/backend/models"
	"github.com/pointofsale/backend/testutil"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

// setupPermissionTestRouter creates a test router with permission endpoints
func setupPermissionTestRouter(t *testing.T) (chi.Router, *gorm.DB, *redis.Client) {
	t.Helper()

	db := testutil.SetupTestDB(t)

	// Setup miniredis
	mr := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})

	permissionHandler := NewPermissionHandler(db, rdb)

	// Setup router
	r := chi.NewRouter()
	r.Route("/api/v1", func(r chi.Router) {
		r.Get("/permissions", permissionHandler.ListPermissions)
		r.Get("/roles/{id}/permissions", permissionHandler.GetRolePermissions)
		r.Put("/roles/{id}/permissions", permissionHandler.UpdateRolePermissions)
	})

	return r, db, rdb
}

// TestListPermissions_Returns200WithAllPermissions verifies listing all permissions
func TestListPermissions_Returns200WithAllPermissions(t *testing.T) {
	router, db, _ := setupPermissionTestRouter(t)
	defer testutil.CleanupTestDB(t, db)

	// Create permissions
	testutil.CreateTestPermission(t, db, func(p *models.Permission) {
		p.Module = "Master Data"
		p.Feature = "Product"
		p.Actions = pq.StringArray{"read", "create", "update", "delete", "export"}
	})
	testutil.CreateTestPermission(t, db, func(p *models.Permission) {
		p.Module = "Master Data"
		p.Feature = "Category"
		p.Actions = pq.StringArray{"read", "create", "update", "delete"}
	})

	req := httptest.NewRequest("GET", "/api/v1/permissions", nil)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var response map[string]interface{}
	err := json.Unmarshal(rr.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Contains(t, response, "data")

	data := response["data"].([]interface{})
	assert.GreaterOrEqual(t, len(data), 2)

	// Verify structure
	firstPerm := data[0].(map[string]interface{})
	assert.Contains(t, firstPerm, "id")
	assert.Contains(t, firstPerm, "module")
	assert.Contains(t, firstPerm, "feature")
	assert.Contains(t, firstPerm, "actions")
}

// TestGetRolePermissions_RegularRole_ReturnsGrantedActions verifies getting role permissions
func TestGetRolePermissions_RegularRole_ReturnsGrantedActions(t *testing.T) {
	router, db, _ := setupPermissionTestRouter(t)
	defer testutil.CleanupTestDB(t, db)

	// Create permission
	perm1 := testutil.CreateTestPermission(t, db, func(p *models.Permission) {
		p.Module = "Master Data"
		p.Feature = "Product"
		p.Actions = pq.StringArray{"read", "create", "update", "delete", "export"}
	})
	perm2 := testutil.CreateTestPermission(t, db, func(p *models.Permission) {
		p.Module = "Master Data"
		p.Feature = "Category"
		p.Actions = pq.StringArray{"read", "create", "update", "delete"}
	})

	// Create role
	role := testutil.CreateTestRole(t, db, func(r *models.Role) {
		r.Name = "Manager"
		r.IsSystem = false
	})

	// Assign permissions to role
	db.Create(&models.RolePermission{
		RoleID:       role.ID,
		PermissionID: perm1.ID,
		Actions:      pq.StringArray{"read", "create", "update"},
	})
	db.Create(&models.RolePermission{
		RoleID:       role.ID,
		PermissionID: perm2.ID,
		Actions:      pq.StringArray{"read"},
	})

	req := httptest.NewRequest("GET", fmt.Sprintf("/api/v1/roles/%d/permissions", role.ID), nil)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var response map[string]interface{}
	json.Unmarshal(rr.Body.Bytes(), &response)

	data := response["data"].(map[string]interface{})
	assert.Equal(t, "Manager", data["roleName"])
	assert.False(t, data["isSystem"].(bool))

	permissions := data["permissions"].([]interface{})
	assert.GreaterOrEqual(t, len(permissions), 2)

	// Verify structure
	firstPerm := permissions[0].(map[string]interface{})
	assert.Contains(t, firstPerm, "permissionId")
	assert.Contains(t, firstPerm, "module")
	assert.Contains(t, firstPerm, "feature")
	assert.Contains(t, firstPerm, "availableActions")
	assert.Contains(t, firstPerm, "grantedActions")
}

// TestGetRolePermissions_SuperAdmin_ReturnsAllGranted verifies super admin gets all permissions
func TestGetRolePermissions_SuperAdmin_ReturnsAllGranted(t *testing.T) {
	router, db, _ := setupPermissionTestRouter(t)
	defer testutil.CleanupTestDB(t, db)

	// Create permissions
	_ = testutil.CreateTestPermission(t, db, func(p *models.Permission) {
		p.Module = "Master Data"
		p.Feature = "Product"
		p.Actions = pq.StringArray{"read", "create", "update", "delete", "export"}
	})

	// Create super admin role
	role := testutil.CreateTestRole(t, db, func(r *models.Role) {
		r.Name = "Super Admin"
		r.IsSystem = true
	})

	// Note: Super Admin should NOT have role_permissions entries, it's computed

	req := httptest.NewRequest("GET", fmt.Sprintf("/api/v1/roles/%d/permissions", role.ID), nil)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var response map[string]interface{}
	json.Unmarshal(rr.Body.Bytes(), &response)

	data := response["data"].(map[string]interface{})
	assert.Equal(t, "Super Admin", data["roleName"])
	assert.True(t, data["isSystem"].(bool))

	permissions := data["permissions"].([]interface{})
	assert.GreaterOrEqual(t, len(permissions), 1)

	// Verify all available actions are granted for super admin
	firstPerm := permissions[0].(map[string]interface{})
	availableActions := firstPerm["availableActions"].([]interface{})
	grantedActions := firstPerm["grantedActions"].([]interface{})
	assert.Equal(t, len(availableActions), len(grantedActions))
}

// TestGetRolePermissions_NotFound_Returns404 verifies error for non-existent role
func TestGetRolePermissions_NotFound_Returns404(t *testing.T) {
	router, db, _ := setupPermissionTestRouter(t)
	defer testutil.CleanupTestDB(t, db)

	req := httptest.NewRequest("GET", "/api/v1/roles/99999/permissions", nil)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusNotFound, rr.Code)
}

// TestUpdateRolePermissions_ValidData_Returns200 verifies updating role permissions
func TestUpdateRolePermissions_ValidData_Returns200(t *testing.T) {
	router, db, _ := setupPermissionTestRouter(t)
	defer testutil.CleanupTestDB(t, db)

	// Create permissions
	perm1 := testutil.CreateTestPermission(t, db, func(p *models.Permission) {
		p.Module = "Master Data"
		p.Feature = "Product"
		p.Actions = pq.StringArray{"read", "create", "update", "delete"}
	})
	perm2 := testutil.CreateTestPermission(t, db, func(p *models.Permission) {
		p.Module = "Master Data"
		p.Feature = "Category"
		p.Actions = pq.StringArray{"read", "create", "update", "delete"}
	})

	// Create role
	role := testutil.CreateTestRole(t, db, func(r *models.Role) {
		r.Name = "Manager"
		r.IsSystem = false
	})

	// Create user with this role (for cache invalidation test)
	user := testutil.CreateTestUser(t, db)
	db.Exec("INSERT INTO user_roles (user_id, role_id) VALUES (?, ?)", user.ID, role.ID)

	body := fmt.Sprintf(`{
		"permissions": [
			{
				"permissionId": %d,
				"actions": ["read", "create"]
			},
			{
				"permissionId": %d,
				"actions": ["read"]
			}
		]
	}`, perm1.ID, perm2.ID)

	req := httptest.NewRequest("PUT", fmt.Sprintf("/api/v1/roles/%d/permissions", role.ID), strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	// Verify permissions were updated
	var rolePerms []models.RolePermission
	db.Where("role_id = ?", role.ID).Find(&rolePerms)
	assert.Len(t, rolePerms, 2)
}

// TestUpdateRolePermissions_SystemRole_Returns403 verifies system role protection
func TestUpdateRolePermissions_SystemRole_Returns403(t *testing.T) {
	router, db, _ := setupPermissionTestRouter(t)
	defer testutil.CleanupTestDB(t, db)

	role := testutil.CreateTestRole(t, db, func(r *models.Role) {
		r.Name = "Super Admin"
		r.IsSystem = true
	})

	body := `{"permissions": []}`

	req := httptest.NewRequest("PUT", fmt.Sprintf("/api/v1/roles/%d/permissions", role.ID), strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusForbidden, rr.Code)

	var response map[string]interface{}
	json.Unmarshal(rr.Body.Bytes(), &response)

	assert.Contains(t, response["error"], "System role")
}

// TestUpdateRolePermissions_InvalidPermissionId_Returns400 verifies validation
func TestUpdateRolePermissions_InvalidPermissionId_Returns400(t *testing.T) {
	router, db, _ := setupPermissionTestRouter(t)
	defer testutil.CleanupTestDB(t, db)

	role := testutil.CreateTestRole(t, db, func(r *models.Role) {
		r.Name = "Manager"
		r.IsSystem = false
	})

	body := `{
		"permissions": [
			{
				"permissionId": 99999,
				"actions": ["read"]
			}
		]
	}`

	req := httptest.NewRequest("PUT", fmt.Sprintf("/api/v1/roles/%d/permissions", role.ID), strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

// TestUpdateRolePermissions_InvalidAction_FiltersOut verifies invalid actions are filtered
func TestUpdateRolePermissions_InvalidAction_FiltersOut(t *testing.T) {
	router, db, _ := setupPermissionTestRouter(t)
	defer testutil.CleanupTestDB(t, db)

	perm := testutil.CreateTestPermission(t, db, func(p *models.Permission) {
		p.Module = "Master Data"
		p.Feature = "Product"
		p.Actions = pq.StringArray{"read", "create", "update"}
	})

	role := testutil.CreateTestRole(t, db, func(r *models.Role) {
		r.Name = "Manager"
		r.IsSystem = false
	})

	// Request includes "delete" which is not in available actions
	body := fmt.Sprintf(`{
		"permissions": [
			{
				"permissionId": %d,
				"actions": ["read", "create", "delete"]
			}
		]
	}`, perm.ID)

	req := httptest.NewRequest("PUT", fmt.Sprintf("/api/v1/roles/%d/permissions", role.ID), strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	// Verify only valid actions were saved
	var rolePerm models.RolePermission
	err := db.Where("role_id = ? AND permission_id = ?", role.ID, perm.ID).First(&rolePerm).Error
	require.NoError(t, err)

	// Should only have "read" and "create", "delete" should be filtered out
	assert.Contains(t, []string(rolePerm.Actions), "read")
	assert.Contains(t, []string(rolePerm.Actions), "create")
	assert.NotContains(t, []string(rolePerm.Actions), "delete")
}

// TestUpdateRolePermissions_InvalidJSON_Returns400 verifies JSON parsing
func TestUpdateRolePermissions_InvalidJSON_Returns400(t *testing.T) {
	router, db, _ := setupPermissionTestRouter(t)
	defer testutil.CleanupTestDB(t, db)

	role := testutil.CreateTestRole(t, db, func(r *models.Role) {
		r.Name = "Manager"
	})

	body := `{invalid json`

	req := httptest.NewRequest("PUT", fmt.Sprintf("/api/v1/roles/%d/permissions", role.ID), strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}
