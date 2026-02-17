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

func setupProductTestRouter(t *testing.T) (chi.Router, *gorm.DB, *redis.Client, *config.Config) {
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

	userRepo := repositories.NewUserRepository(db)
	productRepo := repositories.NewProductRepository(db)
	productService := services.NewProductService(productRepo)
	productHandler := NewProductHandler(productService)
	authMiddleware := middleware.NewAuthMiddleware(cfg.JWTAccessSecret, rdb, userRepo)
	permMiddleware := middleware.NewPermissionMiddleware(db, rdb)

	r := chi.NewRouter()
	r.Route("/api/v1/products", func(r chi.Router) {
		r.Use(authMiddleware.Authenticate)
		r.With(permMiddleware.RequirePermission("Master Data", "Product", "read")).Get("/", productHandler.ListProducts)
		r.With(permMiddleware.RequirePermission("Master Data", "Product", "read")).Get("/{id}", productHandler.GetProduct)
		r.With(permMiddleware.RequirePermission("Master Data", "Product", "create")).Post("/", productHandler.CreateProduct)
		r.With(permMiddleware.RequirePermission("Master Data", "Product", "update")).Put("/{id}", productHandler.UpdateProduct)
		r.With(permMiddleware.RequirePermission("Master Data", "Product", "delete")).Delete("/{id}", productHandler.DeleteProduct)
	})

	return r, db, rdb, cfg
}

func setupProductTestUserWithPermission(t *testing.T, db *gorm.DB, actions []string) *models.User {
	t.Helper()

	perm := testutil.CreateTestPermission(t, db, func(p *models.Permission) {
		p.Module = "Master Data"
		p.Feature = "Product"
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

func minimalProductPayload(categoryID, supplierID, rackID uint) string {
	return fmt.Sprintf(`{
		"name":"Rice",
		"description":"Premium rice",
		"categoryId":%d,
		"priceSetting":"fixed",
		"hasVariants":false,
		"status":"active",
		"supplierIds":[%d],
		"units":[
			{"name":"Kg","isBase":true}
		],
		"variants":[
			{
				"sku":"RC-001",
				"barcode":"8901234567000",
				"attributes":[],
				"pricingTiers":[{"minQty":1,"value":15000}],
				"rackIds":[%d]
			}
		]
	}`, categoryID, supplierID, rackID)
}

func TestCreateProduct_MinimalValid_Returns201(t *testing.T) {
	router, db, _, _ := setupProductTestRouter(t)
	defer testutil.CleanupTestDB(t, db)

	user := setupProductTestUserWithPermission(t, db, []string{"read", "create", "update", "delete"})
	token := testutil.GenerateTestAccessToken(t, user.ID, false)

	category := testutil.CreateTestCategory(t, db)
	supplier := testutil.CreateTestSupplier(t, db)
	rack := testutil.CreateTestRack(t, db)

	req := testutil.AuthenticatedRequest(
		t,
		"POST",
		"/api/v1/products",
		strings.NewReader(minimalProductPayload(category.ID, supplier.ID, rack.ID)),
		token,
	)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusCreated, rr.Code)
	data := testutil.AssertSuccessResponse(t, rr, http.StatusCreated)
	assert.Equal(t, "Rice", data["name"])
	assert.Equal(t, "fixed", data["priceSetting"])
	assert.Equal(t, false, data["hasVariants"])
}

func TestListProducts_Returns200WithVariantCount(t *testing.T) {
	router, db, _, _ := setupProductTestRouter(t)
	defer testutil.CleanupTestDB(t, db)

	user := setupProductTestUserWithPermission(t, db, []string{"read", "create"})
	token := testutil.GenerateTestAccessToken(t, user.ID, false)

	category := testutil.CreateTestCategory(t, db)
	supplier := testutil.CreateTestSupplier(t, db)
	rack := testutil.CreateTestRack(t, db)

	createReq := testutil.AuthenticatedRequest(
		t,
		"POST",
		"/api/v1/products",
		strings.NewReader(minimalProductPayload(category.ID, supplier.ID, rack.ID)),
		token,
	)
	createRR := httptest.NewRecorder()
	router.ServeHTTP(createRR, createReq)
	require.Equal(t, http.StatusCreated, createRR.Code)

	listReq := testutil.AuthenticatedRequest(t, "GET", "/api/v1/products?page=1&pageSize=10", nil, token)
	listRR := httptest.NewRecorder()
	router.ServeHTTP(listRR, listReq)

	assert.Equal(t, http.StatusOK, listRR.Code)

	var response map[string]interface{}
	require.NoError(t, json.Unmarshal(listRR.Body.Bytes(), &response))
	data, ok := response["data"].([]interface{})
	require.True(t, ok)
	require.NotEmpty(t, data)

	first := data[0].(map[string]interface{})
	assert.Contains(t, first, "variantCount")
}

func TestGetProduct_ReturnsFullNestedData(t *testing.T) {
	router, db, _, _ := setupProductTestRouter(t)
	defer testutil.CleanupTestDB(t, db)

	user := setupProductTestUserWithPermission(t, db, []string{"read", "create"})
	token := testutil.GenerateTestAccessToken(t, user.ID, false)

	category := testutil.CreateTestCategory(t, db)
	supplier := testutil.CreateTestSupplier(t, db)
	rack := testutil.CreateTestRack(t, db)

	createReq := testutil.AuthenticatedRequest(
		t,
		"POST",
		"/api/v1/products",
		strings.NewReader(minimalProductPayload(category.ID, supplier.ID, rack.ID)),
		token,
	)
	createRR := httptest.NewRecorder()
	router.ServeHTTP(createRR, createReq)
	require.Equal(t, http.StatusCreated, createRR.Code)
	created := testutil.AssertSuccessResponse(t, createRR, http.StatusCreated)
	productID := uint(created["id"].(float64))

	getReq := testutil.AuthenticatedRequest(t, "GET", fmt.Sprintf("/api/v1/products/%d", productID), nil, token)
	getRR := httptest.NewRecorder()
	router.ServeHTTP(getRR, getReq)

	assert.Equal(t, http.StatusOK, getRR.Code)
	data := testutil.AssertSuccessResponse(t, getRR, http.StatusOK)
	assert.Equal(t, "Rice", data["name"])
	assert.Contains(t, data, "units")
	assert.Contains(t, data, "variants")
	assert.Contains(t, data, "suppliers")
}

func TestUpdateProduct_UnitsWithStock_Returns409(t *testing.T) {
	router, db, _, _ := setupProductTestRouter(t)
	defer testutil.CleanupTestDB(t, db)

	user := setupProductTestUserWithPermission(t, db, []string{"read", "create", "update"})
	token := testutil.GenerateTestAccessToken(t, user.ID, false)

	category := testutil.CreateTestCategory(t, db)
	supplier := testutil.CreateTestSupplier(t, db)
	rack := testutil.CreateTestRack(t, db)

	createReq := testutil.AuthenticatedRequest(
		t,
		"POST",
		"/api/v1/products",
		strings.NewReader(minimalProductPayload(category.ID, supplier.ID, rack.ID)),
		token,
	)
	createRR := httptest.NewRecorder()
	router.ServeHTTP(createRR, createReq)
	require.Equal(t, http.StatusCreated, createRR.Code)
	created := testutil.AssertSuccessResponse(t, createRR, http.StatusCreated)
	productID := uint(created["id"].(float64))

	require.NoError(t, db.Model(&models.ProductVariant{}).Where("product_id = ?", productID).Update("current_stock", 10).Error)

	updateBody := fmt.Sprintf(`{
		"name":"Rice",
		"description":"Premium rice",
		"categoryId":%d,
		"priceSetting":"fixed",
		"hasVariants":false,
		"status":"active",
		"supplierIds":[%d],
		"units":[
			{"name":"Gram","isBase":true}
		],
		"variants":[
			{
				"sku":"RC-001",
				"barcode":"8901234567000",
				"attributes":[],
				"pricingTiers":[{"minQty":1,"value":15000}],
				"rackIds":[%d]
			}
		]
	}`, category.ID, supplier.ID, rack.ID)

	updateReq := testutil.AuthenticatedRequest(
		t,
		"PUT",
		fmt.Sprintf("/api/v1/products/%d", productID),
		strings.NewReader(updateBody),
		token,
	)
	updateRR := httptest.NewRecorder()
	router.ServeHTTP(updateRR, updateReq)

	assert.Equal(t, http.StatusConflict, updateRR.Code)
}

func TestDeleteProduct_NoStock_Returns200(t *testing.T) {
	router, db, _, _ := setupProductTestRouter(t)
	defer testutil.CleanupTestDB(t, db)

	user := setupProductTestUserWithPermission(t, db, []string{"read", "create", "delete"})
	token := testutil.GenerateTestAccessToken(t, user.ID, false)

	category := testutil.CreateTestCategory(t, db)
	supplier := testutil.CreateTestSupplier(t, db)
	rack := testutil.CreateTestRack(t, db)

	createReq := testutil.AuthenticatedRequest(
		t,
		"POST",
		"/api/v1/products",
		strings.NewReader(minimalProductPayload(category.ID, supplier.ID, rack.ID)),
		token,
	)
	createRR := httptest.NewRecorder()
	router.ServeHTTP(createRR, createReq)
	require.Equal(t, http.StatusCreated, createRR.Code)
	created := testutil.AssertSuccessResponse(t, createRR, http.StatusCreated)
	productID := uint(created["id"].(float64))

	deleteReq := testutil.AuthenticatedRequest(t, "DELETE", fmt.Sprintf("/api/v1/products/%d", productID), nil, token)
	deleteRR := httptest.NewRecorder()
	router.ServeHTTP(deleteRR, deleteReq)

	assert.Equal(t, http.StatusOK, deleteRR.Code)

	var body map[string]string
	require.NoError(t, json.Unmarshal(deleteRR.Body.Bytes(), &body))
	assert.Equal(t, "Product deleted successfully", body["message"])
}
