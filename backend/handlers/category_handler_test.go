package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/go-chi/chi/v5"
	"github.com/pointofsale/backend/config"
	"github.com/pointofsale/backend/middleware"
	"github.com/pointofsale/backend/models"
	"github.com/pointofsale/backend/repositories"
	"github.com/pointofsale/backend/services"
	"github.com/pointofsale/backend/testutil"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

// setupCategoryTestRouter creates a test router with category endpoints
func setupCategoryTestRouter(t *testing.T) (chi.Router, *gorm.DB, *redis.Client, *config.Config) {
	t.Helper()

	db := testutil.SetupTestDB(t)

	mr := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})

	cfg := &config.Config{
		FrontendURL:      "http://localhost:3000",
		JWTAccessSecret:  testutil.TestJWTAccessSecret,
		JWTRefreshSecret: testutil.TestJWTRefreshSecret,
		JWTAccessExpiry:  15 * time.Minute,
		JWTRefreshExpiry: 7 * 24 * time.Hour,
	}

	// Initialize layers
	userRepo := repositories.NewUserRepository(db)
	categoryRepo := repositories.NewCategoryRepository(db)
	categoryService := services.NewCategoryService(categoryRepo)
	categoryHandler := NewCategoryHandler(categoryService)
	authMiddleware := middleware.NewAuthMiddleware(cfg.JWTAccessSecret, rdb, userRepo)
	permMiddleware := middleware.NewPermissionMiddleware(db, rdb)

	r := chi.NewRouter()
	r.Route("/api/v1/categories", func(r chi.Router) {
		r.Use(authMiddleware.Authenticate)
		r.With(permMiddleware.RequirePermission("Master Data", "Category", "read")).Get("/", categoryHandler.ListCategories)
		r.With(permMiddleware.RequirePermission("Master Data", "Category", "read")).Get("/{id}", categoryHandler.GetCategory)
		r.With(permMiddleware.RequirePermission("Master Data", "Category", "create")).Post("/", categoryHandler.CreateCategory)
		r.With(permMiddleware.RequirePermission("Master Data", "Category", "update")).Put("/{id}", categoryHandler.UpdateCategory)
		r.With(permMiddleware.RequirePermission("Master Data", "Category", "delete")).Delete("/{id}", categoryHandler.DeleteCategory)
	})

	return r, db, rdb, cfg
}

// setupCategoryTestUserWithPermission creates a user with category CRUD permissions
func setupCategoryTestUserWithPermission(t *testing.T, db *gorm.DB, actions []string) *models.User {
	t.Helper()

	perm := testutil.CreateTestPermission(t, db, func(p *models.Permission) {
		p.Module = "Master Data"
		p.Feature = "Category"
		p.Actions = actions
	})

	role := testutil.CreateTestRole(t, db)

	rolePerm := &models.RolePermission{
		RoleID:       role.ID,
		PermissionID: perm.ID,
		Actions:      actions,
	}
	require.NoError(t, db.Create(rolePerm).Error)

	user := testutil.CreateTestUser(t, db, func(u *models.User) {
		u.Roles = []models.Role{*role}
	})

	return user
}

// createTestCategoryInDB inserts a category directly into the database
func createTestCategoryInDB(t *testing.T, db *gorm.DB, name, description string) *models.Category {
	t.Helper()
	cat := &models.Category{
		Name:        name,
		Description: description,
	}
	require.NoError(t, db.Create(cat).Error)
	return cat
}

func TestListCategories_Returns200WithPagination(t *testing.T) {
	router, db, _, _ := setupCategoryTestRouter(t)
	defer testutil.CleanupTestDB(t, db)

	// Create user with read permission
	user := setupCategoryTestUserWithPermission(t, db, []string{"read", "create", "update", "delete"})
	token := testutil.GenerateTestAccessToken(t, user.ID, false)

	// Create test categories
	for i := 1; i <= 15; i++ {
		createTestCategoryInDB(t, db, fmt.Sprintf("Category %d", i), fmt.Sprintf("Desc %d", i))
	}

	// Request page 1
	req := testutil.AuthenticatedRequest(t, "GET", "/api/v1/categories?page=1&pageSize=10", nil, token)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var response map[string]interface{}
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &response))

	data, ok := response["data"].([]interface{})
	require.True(t, ok)
	assert.Len(t, data, 10)

	meta, ok := response["meta"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, float64(15), meta["totalItems"])
	assert.Equal(t, float64(2), meta["totalPages"])
}

func TestListCategories_WithSearch_FiltersResults(t *testing.T) {
	router, db, _, _ := setupCategoryTestRouter(t)
	defer testutil.CleanupTestDB(t, db)

	user := setupCategoryTestUserWithPermission(t, db, []string{"read"})
	token := testutil.GenerateTestAccessToken(t, user.ID, false)

	createTestCategoryInDB(t, db, "Electronics", "Gadgets")
	createTestCategoryInDB(t, db, "Clothing", "Apparel")
	createTestCategoryInDB(t, db, "Food", "Edible items")

	req := testutil.AuthenticatedRequest(t, "GET", "/api/v1/categories?search=electronics", nil, token)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var response map[string]interface{}
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &response))

	data, ok := response["data"].([]interface{})
	require.True(t, ok)
	assert.Len(t, data, 1)
}

func TestListCategories_WithSort_OrdersResults(t *testing.T) {
	router, db, _, _ := setupCategoryTestRouter(t)
	defer testutil.CleanupTestDB(t, db)

	user := setupCategoryTestUserWithPermission(t, db, []string{"read"})
	token := testutil.GenerateTestAccessToken(t, user.ID, false)

	createTestCategoryInDB(t, db, "Zebra", "")
	createTestCategoryInDB(t, db, "Apple", "")
	createTestCategoryInDB(t, db, "Mango", "")

	req := testutil.AuthenticatedRequest(t, "GET", "/api/v1/categories?sortBy=name&sortDir=asc", nil, token)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var response map[string]interface{}
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &response))

	data, ok := response["data"].([]interface{})
	require.True(t, ok)
	assert.Len(t, data, 3)

	first := data[0].(map[string]interface{})
	last := data[2].(map[string]interface{})
	assert.Equal(t, "Apple", first["name"])
	assert.Equal(t, "Zebra", last["name"])
}

func TestGetCategory_Exists_Returns200(t *testing.T) {
	router, db, _, _ := setupCategoryTestRouter(t)
	defer testutil.CleanupTestDB(t, db)

	user := setupCategoryTestUserWithPermission(t, db, []string{"read"})
	token := testutil.GenerateTestAccessToken(t, user.ID, false)

	cat := createTestCategoryInDB(t, db, "Electronics", "Gadgets and devices")

	req := testutil.AuthenticatedRequest(t, "GET", fmt.Sprintf("/api/v1/categories/%d", cat.ID), nil, token)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	data := testutil.AssertSuccessResponse(t, rr, http.StatusOK)
	assert.Equal(t, "Electronics", data["name"])
	assert.Equal(t, "Gadgets and devices", data["description"])
}

func TestGetCategory_NotFound_Returns404(t *testing.T) {
	router, db, _, _ := setupCategoryTestRouter(t)
	defer testutil.CleanupTestDB(t, db)

	user := setupCategoryTestUserWithPermission(t, db, []string{"read"})
	token := testutil.GenerateTestAccessToken(t, user.ID, false)

	req := testutil.AuthenticatedRequest(t, "GET", "/api/v1/categories/99999", nil, token)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusNotFound, rr.Code)
}

func TestCreateCategory_ValidBody_Returns201(t *testing.T) {
	router, db, _, _ := setupCategoryTestRouter(t)
	defer testutil.CleanupTestDB(t, db)

	user := setupCategoryTestUserWithPermission(t, db, []string{"read", "create"})
	token := testutil.GenerateTestAccessToken(t, user.ID, false)

	body := `{"name":"Electronics","description":"Electronic devices"}`
	req := testutil.AuthenticatedRequest(t, "POST", "/api/v1/categories", strings.NewReader(body), token)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusCreated, rr.Code)

	data := testutil.AssertSuccessResponse(t, rr, http.StatusCreated)
	assert.Equal(t, "Electronics", data["name"])
	assert.Equal(t, "Electronic devices", data["description"])
	assert.NotZero(t, data["id"])
}

func TestCreateCategory_MissingName_Returns400(t *testing.T) {
	router, db, _, _ := setupCategoryTestRouter(t)
	defer testutil.CleanupTestDB(t, db)

	user := setupCategoryTestUserWithPermission(t, db, []string{"read", "create"})
	token := testutil.GenerateTestAccessToken(t, user.ID, false)

	body := `{"name":"","description":"Some desc"}`
	req := testutil.AuthenticatedRequest(t, "POST", "/api/v1/categories", strings.NewReader(body), token)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestCreateCategory_NoAuth_Returns401(t *testing.T) {
	router, db, _, _ := setupCategoryTestRouter(t)
	defer testutil.CleanupTestDB(t, db)

	body := `{"name":"Electronics","description":"Gadgets"}`
	req := httptest.NewRequest("POST", "/api/v1/categories", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusUnauthorized, rr.Code)
}

func TestCreateCategory_NoPermission_Returns403(t *testing.T) {
	router, db, _, _ := setupCategoryTestRouter(t)
	defer testutil.CleanupTestDB(t, db)

	// Create user with only read permission (not create)
	user := setupCategoryTestUserWithPermission(t, db, []string{"read"})
	token := testutil.GenerateTestAccessToken(t, user.ID, false)

	body := `{"name":"Electronics","description":"Gadgets"}`
	req := testutil.AuthenticatedRequest(t, "POST", "/api/v1/categories", strings.NewReader(body), token)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusForbidden, rr.Code)
}

func TestUpdateCategory_ValidBody_Returns200(t *testing.T) {
	router, db, _, _ := setupCategoryTestRouter(t)
	defer testutil.CleanupTestDB(t, db)

	user := setupCategoryTestUserWithPermission(t, db, []string{"read", "update"})
	token := testutil.GenerateTestAccessToken(t, user.ID, false)

	cat := createTestCategoryInDB(t, db, "Old Name", "Old desc")

	body := `{"name":"New Name","description":"New desc"}`
	req := testutil.AuthenticatedRequest(t, "PUT", fmt.Sprintf("/api/v1/categories/%d", cat.ID), strings.NewReader(body), token)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	data := testutil.AssertSuccessResponse(t, rr, http.StatusOK)
	assert.Equal(t, "New Name", data["name"])
	assert.Equal(t, "New desc", data["description"])
}

func TestUpdateCategory_NotFound_Returns404(t *testing.T) {
	router, db, _, _ := setupCategoryTestRouter(t)
	defer testutil.CleanupTestDB(t, db)

	user := setupCategoryTestUserWithPermission(t, db, []string{"read", "update"})
	token := testutil.GenerateTestAccessToken(t, user.ID, false)

	body := `{"name":"New Name"}`
	req := testutil.AuthenticatedRequest(t, "PUT", "/api/v1/categories/99999", strings.NewReader(body), token)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusNotFound, rr.Code)
}

func TestDeleteCategory_Unreferenced_Returns200(t *testing.T) {
	router, db, _, _ := setupCategoryTestRouter(t)
	defer testutil.CleanupTestDB(t, db)

	user := setupCategoryTestUserWithPermission(t, db, []string{"read", "delete"})
	token := testutil.GenerateTestAccessToken(t, user.ID, false)

	cat := createTestCategoryInDB(t, db, "To Delete", "Will be deleted")

	req := testutil.AuthenticatedRequest(t, "DELETE", fmt.Sprintf("/api/v1/categories/%d", cat.ID), nil, token)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	// Verify it's deleted
	var count int64
	db.Model(&models.Category{}).Where("id = ?", cat.ID).Count(&count)
	assert.Equal(t, int64(0), count)
}

func TestDeleteCategory_ReferencedByProduct_Returns409(t *testing.T) {
	router, db, _, _ := setupCategoryTestRouter(t)
	defer testutil.CleanupTestDB(t, db)

	user := setupCategoryTestUserWithPermission(t, db, []string{"read", "delete"})
	token := testutil.GenerateTestAccessToken(t, user.ID, false)

	cat := createTestCategoryInDB(t, db, "Used Category", "Referenced")

	// Create a products table and insert a product referencing this category
	// Since products table may not exist, we need to create it for this test
	if db.Migrator().HasTable("products") {
		db.Exec("INSERT INTO products (name, category_id, created_at, updated_at) VALUES (?, ?, NOW(), NOW())", "Test Product", cat.ID)
	} else {
		// Create a minimal products table for testing
		db.Exec("CREATE TABLE IF NOT EXISTS products (id BIGSERIAL PRIMARY KEY, name VARCHAR(255), category_id BIGINT, created_at TIMESTAMPTZ DEFAULT NOW(), updated_at TIMESTAMPTZ DEFAULT NOW())")
		db.Exec("INSERT INTO products (name, category_id) VALUES (?, ?)", "Test Product", cat.ID)
	}

	req := testutil.AuthenticatedRequest(t, "DELETE", fmt.Sprintf("/api/v1/categories/%d", cat.ID), nil, token)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusConflict, rr.Code)

	var response map[string]interface{}
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &response))
	errorMsg, ok := response["error"].(string)
	require.True(t, ok)
	assert.Contains(t, errorMsg, "product(s)")
}

func TestCreateCategory_SuperAdmin_Returns201(t *testing.T) {
	router, db, _, _ := setupCategoryTestRouter(t)
	defer testutil.CleanupTestDB(t, db)

	// Super admin bypasses permission checks
	superAdmin := testutil.CreateTestSuperAdmin(t, db)
	token := testutil.GenerateTestAccessToken(t, superAdmin.ID, true)

	body := `{"name":"Super Category","description":"Created by super admin"}`
	req := testutil.AuthenticatedRequest(t, "POST", "/api/v1/categories", strings.NewReader(body), token)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusCreated, rr.Code)
}
