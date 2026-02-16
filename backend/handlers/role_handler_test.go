package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/pointofsale/backend/models"
	"github.com/pointofsale/backend/repositories"
	"github.com/pointofsale/backend/services"
	"github.com/pointofsale/backend/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

// setupRoleTestRouter creates a test router with role endpoints
func setupRoleTestRouter(t *testing.T) (chi.Router, *gorm.DB) {
	t.Helper()

	db := testutil.SetupTestDB(t)

	// Initialize layers
	roleRepo := repositories.NewRoleRepository(db)
	roleService := services.NewRoleService(roleRepo)
	roleHandler := NewRoleHandler(roleService)

	// Setup router
	r := chi.NewRouter()
	r.Route("/api/v1/roles", func(r chi.Router) {
		r.Get("/", roleHandler.ListRoles)
		r.Get("/{id}", roleHandler.GetRole)
		r.Post("/", roleHandler.CreateRole)
		r.Put("/{id}", roleHandler.UpdateRole)
		r.Delete("/{id}", roleHandler.DeleteRole)
	})

	return r, db
}

// TestListRoles_Returns200WithUserCounts verifies list endpoint with user counts
func TestListRoles_Returns200WithUserCounts(t *testing.T) {
	router, db := setupRoleTestRouter(t)
	defer testutil.CleanupTestDB(t, db)

	// Create roles
	role1 := testutil.CreateTestRole(t, db, func(r *models.Role) {
		r.Name = "Manager"
	})
	_ = testutil.CreateTestRole(t, db, func(r *models.Role) {
		r.Name = "Cashier"
	})

	// Create users and assign roles
	user1 := testutil.CreateTestUser(t, db)
	db.Exec("INSERT INTO user_roles (user_id, role_id) VALUES (?, ?)", user1.ID, role1.ID)

	user2 := testutil.CreateTestUser(t, db)
	db.Exec("INSERT INTO user_roles (user_id, role_id) VALUES (?, ?)", user2.ID, role1.ID)

	// Request
	req := httptest.NewRequest("GET", "/api/v1/roles", nil)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var response map[string]interface{}
	err := json.Unmarshal(rr.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Contains(t, response, "data")
	assert.Contains(t, response, "meta")

	data := response["data"].([]interface{})
	assert.GreaterOrEqual(t, len(data), 2)

	// Verify userCount field exists
	firstRole := data[0].(map[string]interface{})
	assert.Contains(t, firstRole, "userCount")
}

// TestListRoles_WithSearch_FiltersResults verifies search functionality
func TestListRoles_WithSearch_FiltersResults(t *testing.T) {
	router, db := setupRoleTestRouter(t)
	defer testutil.CleanupTestDB(t, db)

	testutil.CreateTestRole(t, db, func(r *models.Role) {
		r.Name = "Manager"
		r.Description = "Manages operations"
	})
	testutil.CreateTestRole(t, db, func(r *models.Role) {
		r.Name = "Cashier"
		r.Description = "Handles cash"
	})
	testutil.CreateTestRole(t, db, func(r *models.Role) {
		r.Name = "Warehouse"
		r.Description = "Manages inventory"
	})

	// Search for "cash"
	req := httptest.NewRequest("GET", "/api/v1/roles?search=cash", nil)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var response map[string]interface{}
	json.Unmarshal(rr.Body.Bytes(), &response)

	data := response["data"].([]interface{})
	assert.Equal(t, 1, len(data))

	role := data[0].(map[string]interface{})
	assert.Equal(t, "Cashier", role["name"])
}

// TestListRoles_WithPagination_ReturnsCorrectPage verifies pagination
func TestListRoles_WithPagination_ReturnsCorrectPage(t *testing.T) {
	router, db := setupRoleTestRouter(t)
	defer testutil.CleanupTestDB(t, db)

	// Create 5 roles
	for i := 1; i <= 5; i++ {
		testutil.CreateTestRole(t, db, func(r *models.Role) {
			r.Name = fmt.Sprintf("Role %d", i)
		})
	}

	// Request page 1 with pageSize 2
	req := httptest.NewRequest("GET", "/api/v1/roles?page=1&pageSize=2", nil)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var response map[string]interface{}
	json.Unmarshal(rr.Body.Bytes(), &response)

	data := response["data"].([]interface{})
	assert.Equal(t, 2, len(data))

	meta := response["meta"].(map[string]interface{})
	assert.Equal(t, float64(1), meta["page"])
	assert.Equal(t, float64(2), meta["pageSize"])
	assert.Equal(t, float64(5), meta["totalItems"])
	assert.Equal(t, float64(3), meta["totalPages"])
}

// TestListRoles_WithSort_ReturnsOrdered verifies sorting
func TestListRoles_WithSort_ReturnsOrdered(t *testing.T) {
	router, db := setupRoleTestRouter(t)
	defer testutil.CleanupTestDB(t, db)

	testutil.CreateTestRole(t, db, func(r *models.Role) {
		r.Name = "Zebra"
	})
	testutil.CreateTestRole(t, db, func(r *models.Role) {
		r.Name = "Alpha"
	})
	testutil.CreateTestRole(t, db, func(r *models.Role) {
		r.Name = "Beta"
	})

	// Sort by name desc
	req := httptest.NewRequest("GET", "/api/v1/roles?sortBy=name&sortDir=desc", nil)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var response map[string]interface{}
	json.Unmarshal(rr.Body.Bytes(), &response)

	data := response["data"].([]interface{})
	assert.GreaterOrEqual(t, len(data), 3)

	// First should be Zebra
	firstRole := data[0].(map[string]interface{})
	assert.Equal(t, "Zebra", firstRole["name"])
}

// TestGetRole_Exists_Returns200 verifies getting a role by ID
func TestGetRole_Exists_Returns200(t *testing.T) {
	router, db := setupRoleTestRouter(t)
	defer testutil.CleanupTestDB(t, db)

	role := testutil.CreateTestRole(t, db, func(r *models.Role) {
		r.Name = "Manager"
		r.Description = "Manages operations"
	})

	req := httptest.NewRequest("GET", fmt.Sprintf("/api/v1/roles/%d", role.ID), nil)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var response map[string]interface{}
	json.Unmarshal(rr.Body.Bytes(), &response)

	data := response["data"].(map[string]interface{})
	assert.Equal(t, "Manager", data["name"])
	assert.Equal(t, "Manages operations", data["description"])
}

// TestGetRole_NotFound_Returns404 verifies 404 for non-existent ID
func TestGetRole_NotFound_Returns404(t *testing.T) {
	router, db := setupRoleTestRouter(t)
	defer testutil.CleanupTestDB(t, db)

	req := httptest.NewRequest("GET", "/api/v1/roles/99999", nil)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusNotFound, rr.Code)

	var response map[string]interface{}
	json.Unmarshal(rr.Body.Bytes(), &response)

	assert.Contains(t, response, "error")
	assert.Contains(t, response, "code")
}

// TestGetRole_InvalidID_Returns400 verifies 400 for invalid ID format
func TestGetRole_InvalidID_Returns400(t *testing.T) {
	router, db := setupRoleTestRouter(t)
	defer testutil.CleanupTestDB(t, db)

	req := httptest.NewRequest("GET", "/api/v1/roles/invalid", nil)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

// TestCreateRole_ValidBody_Returns201 verifies role creation
func TestCreateRole_ValidBody_Returns201(t *testing.T) {
	router, db := setupRoleTestRouter(t)
	defer testutil.CleanupTestDB(t, db)

	body := `{
		"name": "Supervisor",
		"description": "Supervises daily operations"
	}`

	req := httptest.NewRequest("POST", "/api/v1/roles", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusCreated, rr.Code)

	var response map[string]interface{}
	json.Unmarshal(rr.Body.Bytes(), &response)

	data := response["data"].(map[string]interface{})
	assert.Equal(t, "Supervisor", data["name"])
	assert.Equal(t, "Supervises daily operations", data["description"])
	assert.False(t, data["isSystem"].(bool))
}

// TestCreateRole_DuplicateName_Returns409 verifies conflict on duplicate name
func TestCreateRole_DuplicateName_Returns409(t *testing.T) {
	router, db := setupRoleTestRouter(t)
	defer testutil.CleanupTestDB(t, db)

	testutil.CreateTestRole(t, db, func(r *models.Role) {
		r.Name = "Manager"
	})

	body := `{
		"name": "Manager",
		"description": "Another manager"
	}`

	req := httptest.NewRequest("POST", "/api/v1/roles", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusConflict, rr.Code)

	var response map[string]interface{}
	json.Unmarshal(rr.Body.Bytes(), &response)

	assert.Contains(t, response["error"], "already exists")
}

// TestCreateRole_EmptyName_Returns400 verifies validation
func TestCreateRole_EmptyName_Returns400(t *testing.T) {
	router, db := setupRoleTestRouter(t)
	defer testutil.CleanupTestDB(t, db)

	body := `{
		"name": "",
		"description": "Test"
	}`

	req := httptest.NewRequest("POST", "/api/v1/roles", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

// TestCreateRole_InvalidJSON_Returns400 verifies JSON parsing
func TestCreateRole_InvalidJSON_Returns400(t *testing.T) {
	router, db := setupRoleTestRouter(t)
	defer testutil.CleanupTestDB(t, db)

	body := `{invalid json`

	req := httptest.NewRequest("POST", "/api/v1/roles", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

// TestUpdateRole_ValidBody_Returns200 verifies role update
func TestUpdateRole_ValidBody_Returns200(t *testing.T) {
	router, db := setupRoleTestRouter(t)
	defer testutil.CleanupTestDB(t, db)

	role := testutil.CreateTestRole(t, db, func(r *models.Role) {
		r.Name = "OldName"
	})

	body := `{
		"name": "NewName",
		"description": "Updated description"
	}`

	req := httptest.NewRequest("PUT", fmt.Sprintf("/api/v1/roles/%d", role.ID), strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var response map[string]interface{}
	json.Unmarshal(rr.Body.Bytes(), &response)

	data := response["data"].(map[string]interface{})
	assert.Equal(t, "NewName", data["name"])
	assert.Equal(t, "Updated description", data["description"])
}

// TestUpdateRole_SystemRole_Returns403 verifies system role protection
func TestUpdateRole_SystemRole_Returns403(t *testing.T) {
	router, db := setupRoleTestRouter(t)
	defer testutil.CleanupTestDB(t, db)

	role := testutil.CreateTestRole(t, db, func(r *models.Role) {
		r.Name = "Super Admin"
		r.IsSystem = true
	})

	body := `{
		"name": "NewName",
		"description": "Test"
	}`

	req := httptest.NewRequest("PUT", fmt.Sprintf("/api/v1/roles/%d", role.ID), strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusForbidden, rr.Code)

	var response map[string]interface{}
	json.Unmarshal(rr.Body.Bytes(), &response)

	assert.Contains(t, response["error"], "System roles cannot be modified")
}

// TestUpdateRole_NotFound_Returns404 verifies 404 for non-existent role
func TestUpdateRole_NotFound_Returns404(t *testing.T) {
	router, db := setupRoleTestRouter(t)
	defer testutil.CleanupTestDB(t, db)

	body := `{
		"name": "NewName"
	}`

	req := httptest.NewRequest("PUT", "/api/v1/roles/99999", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusNotFound, rr.Code)
}

// TestDeleteRole_Regular_Returns200 verifies role deletion
func TestDeleteRole_Regular_Returns200(t *testing.T) {
	router, db := setupRoleTestRouter(t)
	defer testutil.CleanupTestDB(t, db)

	role := testutil.CreateTestRole(t, db, func(r *models.Role) {
		r.Name = "ToDelete"
	})

	req := httptest.NewRequest("DELETE", fmt.Sprintf("/api/v1/roles/%d", role.ID), nil)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var response map[string]interface{}
	json.Unmarshal(rr.Body.Bytes(), &response)

	assert.Contains(t, response["message"], "deleted successfully")
}

// TestDeleteRole_SystemRole_Returns403 verifies system role protection
func TestDeleteRole_SystemRole_Returns403(t *testing.T) {
	router, db := setupRoleTestRouter(t)
	defer testutil.CleanupTestDB(t, db)

	role := testutil.CreateTestRole(t, db, func(r *models.Role) {
		r.Name = "Super Admin"
		r.IsSystem = true
	})

	req := httptest.NewRequest("DELETE", fmt.Sprintf("/api/v1/roles/%d", role.ID), nil)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusForbidden, rr.Code)

	var response map[string]interface{}
	json.Unmarshal(rr.Body.Bytes(), &response)

	assert.Contains(t, response["error"], "System roles cannot be deleted")
}

// TestDeleteRole_NotFound_Returns404 verifies 404 for non-existent role
func TestDeleteRole_NotFound_Returns404(t *testing.T) {
	router, db := setupRoleTestRouter(t)
	defer testutil.CleanupTestDB(t, db)

	req := httptest.NewRequest("DELETE", "/api/v1/roles/99999", nil)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusNotFound, rr.Code)
}
