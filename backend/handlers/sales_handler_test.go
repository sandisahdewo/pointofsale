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

func setupSalesTestRouter(t *testing.T) (chi.Router, *gorm.DB, *redis.Client, *config.Config) {
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
	salesRepo := repositories.NewSalesRepository(db)
	seqService := services.NewSequenceService(db)
	salesService := services.NewSalesService(db, salesRepo, seqService)
	salesHandler := NewSalesHandler(salesService)

	authMiddleware := middleware.NewAuthMiddleware(cfg.JWTAccessSecret, rdb, userRepo)
	permMiddleware := middleware.NewPermissionMiddleware(db, rdb)

	r := chi.NewRouter()
	r.Route("/api/v1/sales", func(r chi.Router) {
		r.Use(authMiddleware.Authenticate)
		r.With(permMiddleware.RequirePermission("Transaction", "Sale", "read")).Get("/products/search", salesHandler.ProductSearch)
		r.With(permMiddleware.RequirePermission("Transaction", "Sale", "create")).Post("/checkout", salesHandler.Checkout)
		r.With(permMiddleware.RequirePermission("Transaction", "Sale", "read")).Get("/transactions", salesHandler.ListTransactions)
		r.With(permMiddleware.RequirePermission("Transaction", "Sale", "read")).Get("/transactions/{id}", salesHandler.GetTransaction)
	})

	return r, db, rdb, cfg
}

func setupSalesTestUserWithPermission(t *testing.T, db *gorm.DB, actions []string) *models.User {
	t.Helper()

	perm := testutil.CreateTestPermission(t, db, func(p *models.Permission) {
		p.Module = "Transaction"
		p.Feature = "Sale"
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

func TestProductSearch_MinChars_Returns200(t *testing.T) {
	router, db, _, _ := setupSalesTestRouter(t)
	defer testutil.CleanupTestDB(t, db)

	user := setupSalesTestUserWithPermission(t, db, []string{"read"})
	token := testutil.GenerateTestAccessToken(t, user.ID, false)

	testutil.CreateTestProduct(t, db, func(p *models.Product) {
		p.Name = "SearchableProduct ABC"
		p.Status = "active"
	})

	req := testutil.AuthenticatedRequest(t, "GET", "/api/v1/sales/products/search?q=SearchableProduct", nil, token)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	var response map[string]interface{}
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &response))
	data, ok := response["data"].([]interface{})
	require.True(t, ok)
	assert.NotEmpty(t, data)
}

func TestProductSearch_TooShort_Returns400(t *testing.T) {
	router, db, _, _ := setupSalesTestRouter(t)
	defer testutil.CleanupTestDB(t, db)

	user := setupSalesTestUserWithPermission(t, db, []string{"read"})
	token := testutil.GenerateTestAccessToken(t, user.ID, false)

	req := testutil.AuthenticatedRequest(t, "GET", "/api/v1/sales/products/search?q=ab", nil, token)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestProductSearch_ByName_ReturnsMatching(t *testing.T) {
	router, db, _, _ := setupSalesTestRouter(t)
	defer testutil.CleanupTestDB(t, db)

	user := setupSalesTestUserWithPermission(t, db, []string{"read"})
	token := testutil.GenerateTestAccessToken(t, user.ID, false)

	testutil.CreateTestProduct(t, db, func(p *models.Product) {
		p.Name = "CoolWidget Pro"
		p.Status = "active"
	})

	req := testutil.AuthenticatedRequest(t, "GET", "/api/v1/sales/products/search?q=CoolWidget", nil, token)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	var response map[string]interface{}
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &response))
	data := response["data"].([]interface{})
	require.NotEmpty(t, data)
	first := data[0].(map[string]interface{})
	assert.Equal(t, "CoolWidget Pro", first["name"])
}

func TestProductSearch_BySKU_ReturnsMatching(t *testing.T) {
	router, db, _, _ := setupSalesTestRouter(t)
	defer testutil.CleanupTestDB(t, db)

	user := setupSalesTestUserWithPermission(t, db, []string{"read"})
	token := testutil.GenerateTestAccessToken(t, user.ID, false)

	product := testutil.CreateTestProduct(t, db, func(p *models.Product) {
		p.Name = "SKU Test Product"
		p.Status = "active"
	})
	// Update variant SKU to something specific
	require.NoError(t, db.Model(&models.ProductVariant{}).Where("product_id = ?", product.ID).Update("sku", "UNIQUESKU-XYZ").Error)

	req := testutil.AuthenticatedRequest(t, "GET", "/api/v1/sales/products/search?q=UNIQUESKU", nil, token)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	var response map[string]interface{}
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &response))
	data := response["data"].([]interface{})
	assert.NotEmpty(t, data)
}

func TestProductSearch_ByBarcode_ReturnsMatching(t *testing.T) {
	router, db, _, _ := setupSalesTestRouter(t)
	defer testutil.CleanupTestDB(t, db)

	user := setupSalesTestUserWithPermission(t, db, []string{"read"})
	token := testutil.GenerateTestAccessToken(t, user.ID, false)

	product := testutil.CreateTestProduct(t, db, func(p *models.Product) {
		p.Name = "Barcode Test Product"
		p.Status = "active"
	})
	require.NoError(t, db.Model(&models.ProductVariant{}).Where("product_id = ?", product.ID).Update("barcode", "8901234XYZABC").Error)

	req := testutil.AuthenticatedRequest(t, "GET", "/api/v1/sales/products/search?q=8901234XYZ", nil, token)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	var response map[string]interface{}
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &response))
	data := response["data"].([]interface{})
	assert.NotEmpty(t, data)
}

func TestProductSearch_InactiveProducts_Excluded(t *testing.T) {
	router, db, _, _ := setupSalesTestRouter(t)
	defer testutil.CleanupTestDB(t, db)

	user := setupSalesTestUserWithPermission(t, db, []string{"read"})
	token := testutil.GenerateTestAccessToken(t, user.ID, false)

	testutil.CreateTestProduct(t, db, func(p *models.Product) {
		p.Name = "InvisibleInactive Product"
		p.Status = "inactive"
	})

	req := testutil.AuthenticatedRequest(t, "GET", "/api/v1/sales/products/search?q=InvisibleInactive", nil, token)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	var response map[string]interface{}
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &response))
	data := response["data"].([]interface{})
	assert.Empty(t, data)
}

func TestProductSearch_Max10Results(t *testing.T) {
	router, db, _, _ := setupSalesTestRouter(t)
	defer testutil.CleanupTestDB(t, db)

	user := setupSalesTestUserWithPermission(t, db, []string{"read"})
	token := testutil.GenerateTestAccessToken(t, user.ID, false)

	for i := 0; i < 12; i++ {
		testutil.CreateTestProduct(t, db, func(p *models.Product) {
			p.Name = "MaxResults TestAlpha Product"
			p.Status = "active"
		})
	}

	req := testutil.AuthenticatedRequest(t, "GET", "/api/v1/sales/products/search?q=MaxResults", nil, token)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	var response map[string]interface{}
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &response))
	data := response["data"].([]interface{})
	assert.LessOrEqual(t, len(data), 10)
}

func TestProductSearch_NoAuth_Returns401(t *testing.T) {
	router, db, _, _ := setupSalesTestRouter(t)
	defer testutil.CleanupTestDB(t, db)

	req := httptest.NewRequest("GET", "/api/v1/sales/products/search?q=test", nil)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusUnauthorized, rr.Code)
}

func TestCheckout_ValidBody_Returns201WithReceipt(t *testing.T) {
	router, db, _, _ := setupSalesTestRouter(t)
	defer testutil.CleanupTestDB(t, db)

	user := setupSalesTestUserWithPermission(t, db, []string{"read", "create"})
	token := testutil.GenerateTestAccessToken(t, user.ID, false)

	product := testutil.CreateTestProduct(t, db)
	variant := product.Variants[0]
	unit := product.Units[0]

	body := fmt.Sprintf(`{
		"paymentMethod": "cash",
		"items": [
			{"productId": %d, "variantId": "%s", "unitId": %d, "quantity": 2}
		]
	}`, product.ID, variant.ID, unit.ID)

	req := testutil.AuthenticatedRequest(t, "POST", "/api/v1/sales/checkout", strings.NewReader(body), token)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusCreated, rr.Code)
	data := testutil.AssertSuccessResponse(t, rr, http.StatusCreated)
	assert.NotNil(t, data["transactionNumber"])
	assert.NotNil(t, data["grandTotal"])
}

func TestCheckout_InsufficientStock_Returns400(t *testing.T) {
	router, db, _, _ := setupSalesTestRouter(t)
	defer testutil.CleanupTestDB(t, db)

	user := setupSalesTestUserWithPermission(t, db, []string{"read", "create"})
	token := testutil.GenerateTestAccessToken(t, user.ID, false)

	product := testutil.CreateTestProduct(t, db)
	variant := product.Variants[0]
	unit := product.Units[0]

	// Set stock to 0
	require.NoError(t, db.Model(&models.ProductVariant{}).Where("id = ?", variant.ID).Update("current_stock", 0).Error)

	body := fmt.Sprintf(`{
		"paymentMethod": "cash",
		"items": [
			{"productId": %d, "variantId": "%s", "unitId": %d, "quantity": 1}
		]
	}`, product.ID, variant.ID, unit.ID)

	req := testutil.AuthenticatedRequest(t, "POST", "/api/v1/sales/checkout", strings.NewReader(body), token)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestCheckout_NoAuth_Returns401(t *testing.T) {
	router, db, _, _ := setupSalesTestRouter(t)
	defer testutil.CleanupTestDB(t, db)

	req := httptest.NewRequest("POST", "/api/v1/sales/checkout", strings.NewReader(`{}`))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusUnauthorized, rr.Code)
}

func TestCheckout_NoPermission_Returns403(t *testing.T) {
	router, db, _, _ := setupSalesTestRouter(t)
	defer testutil.CleanupTestDB(t, db)

	// User with read-only permission (no "create")
	user := setupSalesTestUserWithPermission(t, db, []string{"read"})
	token := testutil.GenerateTestAccessToken(t, user.ID, false)

	req := testutil.AuthenticatedRequest(t, "POST", "/api/v1/sales/checkout", strings.NewReader(`{}`), token)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusForbidden, rr.Code)
}

func TestCheckout_VerifyStockDeducted(t *testing.T) {
	router, db, _, _ := setupSalesTestRouter(t)
	defer testutil.CleanupTestDB(t, db)

	user := setupSalesTestUserWithPermission(t, db, []string{"read", "create"})
	token := testutil.GenerateTestAccessToken(t, user.ID, false)

	product := testutil.CreateTestProduct(t, db)
	variant := product.Variants[0]
	unit := product.Units[0]
	initialStock := variant.CurrentStock

	body := fmt.Sprintf(`{
		"paymentMethod": "card",
		"items": [
			{"productId": %d, "variantId": "%s", "unitId": %d, "quantity": 3}
		]
	}`, product.ID, variant.ID, unit.ID)

	req := testutil.AuthenticatedRequest(t, "POST", "/api/v1/sales/checkout", strings.NewReader(body), token)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	require.Equal(t, http.StatusCreated, rr.Code)

	var updatedVariant models.ProductVariant
	require.NoError(t, db.First(&updatedVariant, "id = ?", variant.ID).Error)
	assert.Equal(t, initialStock-3, updatedVariant.CurrentStock)
}

func TestCheckout_VerifyStockMovementCreated(t *testing.T) {
	router, db, _, _ := setupSalesTestRouter(t)
	defer testutil.CleanupTestDB(t, db)

	user := setupSalesTestUserWithPermission(t, db, []string{"read", "create"})
	token := testutil.GenerateTestAccessToken(t, user.ID, false)

	product := testutil.CreateTestProduct(t, db)
	variant := product.Variants[0]
	unit := product.Units[0]

	body := fmt.Sprintf(`{
		"paymentMethod": "qris",
		"items": [
			{"productId": %d, "variantId": "%s", "unitId": %d, "quantity": 5}
		]
	}`, product.ID, variant.ID, unit.ID)

	req := testutil.AuthenticatedRequest(t, "POST", "/api/v1/sales/checkout", strings.NewReader(body), token)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	require.Equal(t, http.StatusCreated, rr.Code)

	var movements []models.StockMovement
	require.NoError(t, db.Where("variant_id = ? AND reference_type = ?", variant.ID, "sales_transaction").Find(&movements).Error)
	require.Len(t, movements, 1)
	assert.Equal(t, -5, movements[0].Quantity)
	assert.Equal(t, "sales", movements[0].MovementType)
}

func TestListTransactions_Returns200WithPagination(t *testing.T) {
	router, db, _, _ := setupSalesTestRouter(t)
	defer testutil.CleanupTestDB(t, db)

	user := setupSalesTestUserWithPermission(t, db, []string{"read", "create"})
	token := testutil.GenerateTestAccessToken(t, user.ID, false)

	product := testutil.CreateTestProduct(t, db)
	variant := product.Variants[0]
	unit := product.Units[0]

	// Create 2 transactions
	for i := 0; i < 2; i++ {
		body := fmt.Sprintf(`{
			"paymentMethod": "cash",
			"items": [
				{"productId": %d, "variantId": "%s", "unitId": %d, "quantity": 1}
			]
		}`, product.ID, variant.ID, unit.ID)
		checkReq := testutil.AuthenticatedRequest(t, "POST", "/api/v1/sales/checkout", strings.NewReader(body), token)
		checkRR := httptest.NewRecorder()
		router.ServeHTTP(checkRR, checkReq)
		require.Equal(t, http.StatusCreated, checkRR.Code)
	}

	req := testutil.AuthenticatedRequest(t, "GET", "/api/v1/sales/transactions?page=1&pageSize=10", nil, token)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	var response map[string]interface{}
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &response))
	data, ok := response["data"].([]interface{})
	require.True(t, ok)
	assert.GreaterOrEqual(t, len(data), 2)
	assert.Contains(t, response, "meta")
}

func TestListTransactions_FilterByDate_Works(t *testing.T) {
	router, db, _, _ := setupSalesTestRouter(t)
	defer testutil.CleanupTestDB(t, db)

	user := setupSalesTestUserWithPermission(t, db, []string{"read", "create"})
	token := testutil.GenerateTestAccessToken(t, user.ID, false)

	product := testutil.CreateTestProduct(t, db)
	variant := product.Variants[0]
	unit := product.Units[0]

	// Create a transaction
	body := fmt.Sprintf(`{
		"paymentMethod": "cash",
		"items": [
			{"productId": %d, "variantId": "%s", "unitId": %d, "quantity": 1}
		]
	}`, product.ID, variant.ID, unit.ID)
	checkReq := testutil.AuthenticatedRequest(t, "POST", "/api/v1/sales/checkout", strings.NewReader(body), token)
	checkRR := httptest.NewRecorder()
	router.ServeHTTP(checkRR, checkReq)
	require.Equal(t, http.StatusCreated, checkRR.Code)

	today := time.Now().Format("2006-01-02")
	req := testutil.AuthenticatedRequest(t, "GET", "/api/v1/sales/transactions?dateFrom="+today, nil, token)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	var response map[string]interface{}
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &response))
	data := response["data"].([]interface{})
	assert.NotEmpty(t, data)
}

func TestListTransactions_FilterByPaymentMethod_Works(t *testing.T) {
	router, db, _, _ := setupSalesTestRouter(t)
	defer testutil.CleanupTestDB(t, db)

	user := setupSalesTestUserWithPermission(t, db, []string{"read", "create"})
	token := testutil.GenerateTestAccessToken(t, user.ID, false)

	product := testutil.CreateTestProduct(t, db)
	variant := product.Variants[0]
	unit := product.Units[0]

	// Create cash transaction
	body := fmt.Sprintf(`{
		"paymentMethod": "cash",
		"items": [
			{"productId": %d, "variantId": "%s", "unitId": %d, "quantity": 1}
		]
	}`, product.ID, variant.ID, unit.ID)
	checkReq := testutil.AuthenticatedRequest(t, "POST", "/api/v1/sales/checkout", strings.NewReader(body), token)
	checkRR := httptest.NewRecorder()
	router.ServeHTTP(checkRR, checkReq)
	require.Equal(t, http.StatusCreated, checkRR.Code)

	req := testutil.AuthenticatedRequest(t, "GET", "/api/v1/sales/transactions?paymentMethod=cash", nil, token)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	var response map[string]interface{}
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &response))
	data := response["data"].([]interface{})
	assert.NotEmpty(t, data)
	for _, item := range data {
		m := item.(map[string]interface{})
		assert.Equal(t, "cash", m["paymentMethod"])
	}
}

func TestGetTransaction_ReturnsReceiptData(t *testing.T) {
	router, db, _, _ := setupSalesTestRouter(t)
	defer testutil.CleanupTestDB(t, db)

	user := setupSalesTestUserWithPermission(t, db, []string{"read", "create"})
	token := testutil.GenerateTestAccessToken(t, user.ID, false)

	product := testutil.CreateTestProduct(t, db)
	variant := product.Variants[0]
	unit := product.Units[0]

	body := fmt.Sprintf(`{
		"paymentMethod": "card",
		"items": [
			{"productId": %d, "variantId": "%s", "unitId": %d, "quantity": 2}
		]
	}`, product.ID, variant.ID, unit.ID)
	checkReq := testutil.AuthenticatedRequest(t, "POST", "/api/v1/sales/checkout", strings.NewReader(body), token)
	checkRR := httptest.NewRecorder()
	router.ServeHTTP(checkRR, checkReq)
	require.Equal(t, http.StatusCreated, checkRR.Code)

	created := testutil.AssertSuccessResponse(t, checkRR, http.StatusCreated)
	txID := uint(created["id"].(float64))

	req := testutil.AuthenticatedRequest(t, "GET", fmt.Sprintf("/api/v1/sales/transactions/%d", txID), nil, token)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	data := testutil.AssertSuccessResponse(t, rr, http.StatusOK)
	assert.NotNil(t, data["transactionNumber"])
	assert.NotNil(t, data["items"])
	assert.Equal(t, "card", data["paymentMethod"])
}
