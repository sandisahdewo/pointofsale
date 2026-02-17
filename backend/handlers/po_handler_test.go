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

func setupPOTestRouter(t *testing.T) (chi.Router, *gorm.DB, *redis.Client, *config.Config) {
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
	poRepo := repositories.NewPORepository(db)
	stockRepo := repositories.NewStockMovementRepository(db)
	seqSvc := services.NewSequenceService(db)
	poSvc := services.NewPOService(db, poRepo, stockRepo, seqSvc)
	poHandler := NewPOHandler(poSvc)

	authMiddleware := middleware.NewAuthMiddleware(cfg.JWTAccessSecret, rdb, userRepo)
	permMiddleware := middleware.NewPermissionMiddleware(db, rdb)

	r := chi.NewRouter()
	r.Route("/api/v1/purchase-orders", func(r chi.Router) {
		r.Use(authMiddleware.Authenticate)
		r.With(permMiddleware.RequirePermission("Transaction", "Purchase Order", "read")).Get("/", poHandler.ListPOs)
		r.With(permMiddleware.RequirePermission("Transaction", "Purchase Order", "read")).Get("/products", poHandler.GetProductsForPO)
		r.With(permMiddleware.RequirePermission("Transaction", "Purchase Order", "read")).Get("/{id}", poHandler.GetPO)
		r.With(permMiddleware.RequirePermission("Transaction", "Purchase Order", "create")).Post("/", poHandler.CreatePO)
		r.With(permMiddleware.RequirePermission("Transaction", "Purchase Order", "update")).Put("/{id}", poHandler.UpdatePO)
		r.With(permMiddleware.RequirePermission("Transaction", "Purchase Order", "delete")).Delete("/{id}", poHandler.DeletePO)
		r.With(permMiddleware.RequirePermission("Transaction", "Purchase Order", "update")).Patch("/{id}/status", poHandler.UpdatePOStatus)
		r.With(permMiddleware.RequirePermission("Transaction", "Purchase Order", "update")).Post("/{id}/receive", poHandler.ReceivePO)
	})

	return r, db, rdb, cfg
}

func setupPOTestUserWithPermission(t *testing.T, db *gorm.DB, actions []string) *models.User {
	t.Helper()

	perm := testutil.CreateTestPermission(t, db, func(p *models.Permission) {
		p.Module = "Transaction"
		p.Feature = "Purchase Order"
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

// createTestPOViaAPI creates a PO through the repo for test setup
func createDraftPO(t *testing.T, db *gorm.DB, supplier *models.Supplier, product *models.Product) *models.PurchaseOrder {
	t.Helper()
	repo := repositories.NewPORepository(db)
	variant := product.Variants[0]
	unit := product.Units[0]
	po := &models.PurchaseOrder{
		PONumber:   fmt.Sprintf("PO-TEST-%d", product.ID),
		SupplierID: supplier.ID,
		Date:       "2026-01-15",
		Status:     "draft",
		Items: []models.PurchaseOrderItem{
			{
				ProductID:    product.ID,
				VariantID:    variant.ID,
				UnitID:       unit.ID,
				UnitName:     unit.Name,
				ProductName:  product.Name,
				VariantLabel: "Default",
				SKU:          variant.SKU,
				CurrentStock: variant.CurrentStock,
				OrderedQty:   10,
				Price:        15000,
			},
		},
	}
	require.NoError(t, repo.Create(po))
	return po
}

func TestListPOs_Returns200WithStatusCounts(t *testing.T) {
	router, db, _, _ := setupPOTestRouter(t)

	user := setupPOTestUserWithPermission(t, db, []string{"read", "create", "update", "delete"})
	token := testutil.GenerateTestAccessToken(t, user.ID, false)

	req := testutil.AuthenticatedRequest(t, "GET", "/api/v1/purchase-orders", nil, token)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	var response map[string]interface{}
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &response))
	assert.Contains(t, response, "data")
	assert.Contains(t, response, "meta")
	assert.Contains(t, response, "statusCounts")
}

func TestListPOs_FilterByStatus_Returns200(t *testing.T) {
	router, db, _, _ := setupPOTestRouter(t)

	user := setupPOTestUserWithPermission(t, db, []string{"read", "create"})
	token := testutil.GenerateTestAccessToken(t, user.ID, false)

	supplier := testutil.CreateTestSupplier(t, db)
	product := testutil.CreateTestProduct(t, db)
	createDraftPO(t, db, supplier, product)

	req := testutil.AuthenticatedRequest(t, "GET", "/api/v1/purchase-orders?status=draft", nil, token)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	var response map[string]interface{}
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &response))
	data := response["data"].([]interface{})
	assert.NotEmpty(t, data)
}

func TestGetPO_WithItems_Returns200(t *testing.T) {
	router, db, _, _ := setupPOTestRouter(t)

	user := setupPOTestUserWithPermission(t, db, []string{"read"})
	token := testutil.GenerateTestAccessToken(t, user.ID, false)

	supplier := testutil.CreateTestSupplier(t, db)
	product := testutil.CreateTestProduct(t, db)
	po := createDraftPO(t, db, supplier, product)

	req := testutil.AuthenticatedRequest(t, "GET", fmt.Sprintf("/api/v1/purchase-orders/%d", po.ID), nil, token)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	data := testutil.AssertSuccessResponse(t, rr, http.StatusOK)
	assert.Equal(t, float64(po.ID), data["id"])
	items, ok := data["items"].([]interface{})
	require.True(t, ok)
	assert.NotEmpty(t, items)
}

func TestGetPO_NotFound_Returns404(t *testing.T) {
	router, db, _, _ := setupPOTestRouter(t)

	user := setupPOTestUserWithPermission(t, db, []string{"read"})
	token := testutil.GenerateTestAccessToken(t, user.ID, false)

	req := testutil.AuthenticatedRequest(t, "GET", "/api/v1/purchase-orders/99999", nil, token)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusNotFound, rr.Code)
}

func TestCreatePO_ValidBody_Returns201(t *testing.T) {
	router, db, _, _ := setupPOTestRouter(t)

	user := setupPOTestUserWithPermission(t, db, []string{"read", "create"})
	token := testutil.GenerateTestAccessToken(t, user.ID, false)

	supplier := testutil.CreateTestSupplier(t, db)
	product := testutil.CreateTestProduct(t, db)
	variant := product.Variants[0]
	unit := product.Units[0]

	body := fmt.Sprintf(`{
		"supplierId": %d,
		"date": "2026-01-15",
		"notes": "Test purchase",
		"items": [
			{
				"productId": %d,
				"variantId": "%s",
				"unitId": %d,
				"orderedQty": 5,
				"price": 10000
			}
		]
	}`, supplier.ID, product.ID, variant.ID, unit.ID)

	req := testutil.AuthenticatedRequest(t, "POST", "/api/v1/purchase-orders", strings.NewReader(body), token)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusCreated, rr.Code)
	data := testutil.AssertSuccessResponse(t, rr, http.StatusCreated)
	assert.NotEmpty(t, data["poNumber"])
	assert.Equal(t, "draft", data["status"])
}

func TestCreatePO_InvalidSupplier_Returns400(t *testing.T) {
	router, db, _, _ := setupPOTestRouter(t)

	user := setupPOTestUserWithPermission(t, db, []string{"read", "create"})
	token := testutil.GenerateTestAccessToken(t, user.ID, false)

	body := `{
		"supplierId": 99999,
		"date": "2026-01-15",
		"items": [{"productId": 1, "variantId": "uuid", "unitId": 1, "orderedQty": 1}]
	}`

	req := testutil.AuthenticatedRequest(t, "POST", "/api/v1/purchase-orders", strings.NewReader(body), token)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestCreatePO_NoItems_Returns400(t *testing.T) {
	router, db, _, _ := setupPOTestRouter(t)

	user := setupPOTestUserWithPermission(t, db, []string{"read", "create"})
	token := testutil.GenerateTestAccessToken(t, user.ID, false)

	supplier := testutil.CreateTestSupplier(t, db)
	body := fmt.Sprintf(`{
		"supplierId": %d,
		"date": "2026-01-15",
		"items": []
	}`, supplier.ID)

	req := testutil.AuthenticatedRequest(t, "POST", "/api/v1/purchase-orders", strings.NewReader(body), token)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestUpdatePO_DraftPO_Returns200(t *testing.T) {
	router, db, _, _ := setupPOTestRouter(t)

	user := setupPOTestUserWithPermission(t, db, []string{"read", "create", "update"})
	token := testutil.GenerateTestAccessToken(t, user.ID, false)

	supplier := testutil.CreateTestSupplier(t, db)
	product := testutil.CreateTestProduct(t, db)
	po := createDraftPO(t, db, supplier, product)

	variant := product.Variants[0]
	unit := product.Units[0]

	body := fmt.Sprintf(`{
		"supplierId": %d,
		"date": "2026-02-01",
		"notes": "Updated notes",
		"items": [
			{
				"productId": %d,
				"variantId": "%s",
				"unitId": %d,
				"orderedQty": 20,
				"price": 12000
			}
		]
	}`, supplier.ID, product.ID, variant.ID, unit.ID)

	req := testutil.AuthenticatedRequest(t, "PUT", fmt.Sprintf("/api/v1/purchase-orders/%d", po.ID), strings.NewReader(body), token)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestUpdatePO_SentPO_Returns403(t *testing.T) {
	router, db, _, _ := setupPOTestRouter(t)

	user := setupPOTestUserWithPermission(t, db, []string{"read", "create", "update"})
	token := testutil.GenerateTestAccessToken(t, user.ID, false)

	supplier := testutil.CreateTestSupplier(t, db)
	product := testutil.CreateTestProduct(t, db)
	po := createDraftPO(t, db, supplier, product)

	// Move PO to sent status
	require.NoError(t, db.Model(po).Update("status", "sent").Error)

	variant := product.Variants[0]
	unit := product.Units[0]

	body := fmt.Sprintf(`{
		"supplierId": %d,
		"date": "2026-02-01",
		"items": [{"productId": %d, "variantId": "%s", "unitId": %d, "orderedQty": 5, "price": 1000}]
	}`, supplier.ID, product.ID, variant.ID, unit.ID)

	req := testutil.AuthenticatedRequest(t, "PUT", fmt.Sprintf("/api/v1/purchase-orders/%d", po.ID), strings.NewReader(body), token)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusForbidden, rr.Code)
}

func TestDeletePO_DraftPO_Returns200(t *testing.T) {
	router, db, _, _ := setupPOTestRouter(t)

	user := setupPOTestUserWithPermission(t, db, []string{"read", "create", "delete"})
	token := testutil.GenerateTestAccessToken(t, user.ID, false)

	supplier := testutil.CreateTestSupplier(t, db)
	product := testutil.CreateTestProduct(t, db)
	po := createDraftPO(t, db, supplier, product)

	req := testutil.AuthenticatedRequest(t, "DELETE", fmt.Sprintf("/api/v1/purchase-orders/%d", po.ID), nil, token)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestDeletePO_SentPO_Returns403(t *testing.T) {
	router, db, _, _ := setupPOTestRouter(t)

	user := setupPOTestUserWithPermission(t, db, []string{"read", "create", "delete"})
	token := testutil.GenerateTestAccessToken(t, user.ID, false)

	supplier := testutil.CreateTestSupplier(t, db)
	product := testutil.CreateTestProduct(t, db)
	po := createDraftPO(t, db, supplier, product)

	require.NoError(t, db.Model(po).Update("status", "sent").Error)

	req := testutil.AuthenticatedRequest(t, "DELETE", fmt.Sprintf("/api/v1/purchase-orders/%d", po.ID), nil, token)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusForbidden, rr.Code)
}

func TestUpdatePOStatus_DraftToSent_Returns200(t *testing.T) {
	router, db, _, _ := setupPOTestRouter(t)

	user := setupPOTestUserWithPermission(t, db, []string{"read", "create", "update"})
	token := testutil.GenerateTestAccessToken(t, user.ID, false)

	supplier := testutil.CreateTestSupplier(t, db)
	product := testutil.CreateTestProduct(t, db)
	po := createDraftPO(t, db, supplier, product)

	body := `{"status": "sent"}`
	req := testutil.AuthenticatedRequest(t, "PATCH", fmt.Sprintf("/api/v1/purchase-orders/%d/status", po.ID), strings.NewReader(body), token)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	data := testutil.AssertSuccessResponse(t, rr, http.StatusOK)
	assert.Equal(t, "sent", data["status"])
}

func TestUpdatePOStatus_InvalidTransition_Returns400(t *testing.T) {
	router, db, _, _ := setupPOTestRouter(t)

	user := setupPOTestUserWithPermission(t, db, []string{"read", "create", "update"})
	token := testutil.GenerateTestAccessToken(t, user.ID, false)

	supplier := testutil.CreateTestSupplier(t, db)
	product := testutil.CreateTestProduct(t, db)
	po := createDraftPO(t, db, supplier, product)

	body := `{"status": "completed"}` // draft -> completed is invalid
	req := testutil.AuthenticatedRequest(t, "PATCH", fmt.Sprintf("/api/v1/purchase-orders/%d/status", po.ID), strings.NewReader(body), token)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestReceivePO_ValidBody_Returns200_UpdatesStock(t *testing.T) {
	router, db, _, _ := setupPOTestRouter(t)

	user := setupPOTestUserWithPermission(t, db, []string{"read", "create", "update"})
	token := testutil.GenerateTestAccessToken(t, user.ID, false)

	supplier := testutil.CreateTestSupplier(t, db)
	product := testutil.CreateTestProduct(t, db)
	variant := product.Variants[0]
	initialStock := variant.CurrentStock

	po := createDraftPO(t, db, supplier, product)

	// Move to sent status
	require.NoError(t, db.Model(po).Update("status", "sent").Error)

	// Get item ID
	loadedPO := &models.PurchaseOrder{}
	require.NoError(t, db.Preload("Items").First(loadedPO, po.ID).Error)
	require.NotEmpty(t, loadedPO.Items)
	itemID := loadedPO.Items[0].ID

	body := fmt.Sprintf(`{
		"receivedDate": "2026-01-20",
		"paymentMethod": "cash",
		"items": [
			{
				"itemId": "%s",
				"receivedQty": 8,
				"receivedPrice": 14000,
				"isVerified": true
			}
		]
	}`, itemID)

	req := testutil.AuthenticatedRequest(t, "POST", fmt.Sprintf("/api/v1/purchase-orders/%d/receive", po.ID), strings.NewReader(body), token)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	// Verify stock updated
	var updatedVariant models.ProductVariant
	require.NoError(t, db.First(&updatedVariant, "id = ?", variant.ID).Error)
	// unit.ToBaseUnit = 1, so stockDelta = 8 * 1 = 8
	assert.Equal(t, initialStock+8, updatedVariant.CurrentStock)
}

func TestReceivePO_NonSentPO_Returns400(t *testing.T) {
	router, db, _, _ := setupPOTestRouter(t)

	user := setupPOTestUserWithPermission(t, db, []string{"read", "create", "update"})
	token := testutil.GenerateTestAccessToken(t, user.ID, false)

	supplier := testutil.CreateTestSupplier(t, db)
	product := testutil.CreateTestProduct(t, db)
	po := createDraftPO(t, db, supplier, product)

	// PO is still in draft (non-sent)
	// Move to received status to simulate invalid state
	require.NoError(t, db.Model(po).Update("status", "received").Error)

	body := `{
		"receivedDate": "2026-01-20",
		"paymentMethod": "cash",
		"items": []
	}`

	req := testutil.AuthenticatedRequest(t, "POST", fmt.Sprintf("/api/v1/purchase-orders/%d/receive", po.ID), strings.NewReader(body), token)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestReceivePO_BankTransferNoBankAccount_Returns400(t *testing.T) {
	router, db, _, _ := setupPOTestRouter(t)

	user := setupPOTestUserWithPermission(t, db, []string{"read", "create", "update"})
	token := testutil.GenerateTestAccessToken(t, user.ID, false)

	supplier := testutil.CreateTestSupplier(t, db)
	product := testutil.CreateTestProduct(t, db)
	po := createDraftPO(t, db, supplier, product)

	require.NoError(t, db.Model(po).Update("status", "sent").Error)

	body := `{
		"receivedDate": "2026-01-20",
		"paymentMethod": "bank_transfer",
		"items": []
	}`

	req := testutil.AuthenticatedRequest(t, "POST", fmt.Sprintf("/api/v1/purchase-orders/%d/receive", po.ID), strings.NewReader(body), token)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestGetProductsForPO_ReturnsFilteredProducts(t *testing.T) {
	router, db, _, _ := setupPOTestRouter(t)

	user := setupPOTestUserWithPermission(t, db, []string{"read"})
	token := testutil.GenerateTestAccessToken(t, user.ID, false)

	supplier := testutil.CreateTestSupplier(t, db)
	// Create a product with the supplier
	product := testutil.CreateTestProduct(t, db)
	require.NoError(t, db.Exec("INSERT INTO product_suppliers (product_id, supplier_id) VALUES (?, ?)", product.ID, supplier.ID).Error)

	req := testutil.AuthenticatedRequest(t, "GET", fmt.Sprintf("/api/v1/purchase-orders/products?supplierId=%d", supplier.ID), nil, token)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	var response map[string]interface{}
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &response))
	assert.Contains(t, response, "data")
}
