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

// setupSupplierTestRouter creates a test router with supplier endpoints
func setupSupplierTestRouter(t *testing.T) (chi.Router, *gorm.DB) {
	t.Helper()

	db := testutil.SetupTestDB(t)

	// Initialize layers
	supplierRepo := repositories.NewSupplierRepository(db)
	supplierService := services.NewSupplierService(supplierRepo)
	supplierHandler := NewSupplierHandler(supplierService)

	// Setup router
	r := chi.NewRouter()
	r.Route("/api/v1/suppliers", func(r chi.Router) {
		r.Get("/", supplierHandler.ListSuppliers)
		r.Get("/{id}", supplierHandler.GetSupplier)
		r.Post("/", supplierHandler.CreateSupplier)
		r.Put("/{id}", supplierHandler.UpdateSupplier)
		r.Delete("/{id}", supplierHandler.DeleteSupplier)
	})

	return r, db
}

func TestListSuppliers_Returns200WithBankAccounts(t *testing.T) {
	router, db := setupSupplierTestRouter(t)

	// Create supplier with bank accounts
	supplier := &models.Supplier{
		Name:    "PT Sumber Makmur",
		Address: "Jakarta",
		Active:  true,
		BankAccounts: []models.SupplierBankAccount{
			{AccountName: "BCA", AccountNumber: "1234567890"},
			{AccountName: "Mandiri", AccountNumber: "0987654321"},
		},
	}
	err := db.Create(supplier).Error
	require.NoError(t, err)

	req := httptest.NewRequest("GET", "/api/v1/suppliers", nil)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var response map[string]interface{}
	err = json.Unmarshal(rr.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Contains(t, response, "data")
	assert.Contains(t, response, "meta")

	data := response["data"].([]interface{})
	require.GreaterOrEqual(t, len(data), 1)

	firstSupplier := data[0].(map[string]interface{})
	assert.Equal(t, "PT Sumber Makmur", firstSupplier["name"])
	assert.Contains(t, firstSupplier, "bankAccounts")

	bankAccounts := firstSupplier["bankAccounts"].([]interface{})
	assert.Equal(t, 2, len(bankAccounts))
}

func TestListSuppliers_FilterActive_ReturnsActiveOnly(t *testing.T) {
	router, db := setupSupplierTestRouter(t)

	// Create active and inactive suppliers
	activeSupplier := &models.Supplier{Name: "Active Co", Address: "Jakarta", Active: true}
	require.NoError(t, db.Create(activeSupplier).Error)

	inactiveSupplier := &models.Supplier{Name: "Inactive Co", Address: "Surabaya", Active: false}
	require.NoError(t, db.Exec("INSERT INTO suppliers (name, address, active, created_at, updated_at) VALUES (?, ?, ?, NOW(), NOW())", inactiveSupplier.Name, inactiveSupplier.Address, false).Error)

	req := httptest.NewRequest("GET", "/api/v1/suppliers?active=true", nil)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var response map[string]interface{}
	json.Unmarshal(rr.Body.Bytes(), &response)

	data := response["data"].([]interface{})
	assert.Equal(t, 1, len(data))

	supplierData := data[0].(map[string]interface{})
	assert.Equal(t, "Active Co", supplierData["name"])
}

func TestGetSupplier_Exists_Returns200(t *testing.T) {
	router, db := setupSupplierTestRouter(t)

	supplier := &models.Supplier{
		Name:    "Test Supplier",
		Address: "Test Address",
		Active:  true,
		BankAccounts: []models.SupplierBankAccount{
			{AccountName: "BCA", AccountNumber: "111"},
		},
	}
	require.NoError(t, db.Create(supplier).Error)

	req := httptest.NewRequest("GET", fmt.Sprintf("/api/v1/suppliers/%d", supplier.ID), nil)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var response map[string]interface{}
	json.Unmarshal(rr.Body.Bytes(), &response)

	data := response["data"].(map[string]interface{})
	assert.Equal(t, "Test Supplier", data["name"])
	assert.Contains(t, data, "bankAccounts")
}

func TestGetSupplier_NotFound_Returns404(t *testing.T) {
	router, _ := setupSupplierTestRouter(t)

	req := httptest.NewRequest("GET", "/api/v1/suppliers/99999", nil)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusNotFound, rr.Code)
}

func TestCreateSupplier_WithBankAccounts_Returns201(t *testing.T) {
	router, _ := setupSupplierTestRouter(t)

	body := `{
		"name": "PT Sumber Makmur",
		"address": "Jl. Industri No. 45, Jakarta",
		"phone": "+62-21-5550001",
		"email": "order@sumbermakmur.co.id",
		"website": "sumbermakmur.co.id",
		"bankAccounts": [
			{"accountName": "BCA - Main Account", "accountNumber": "1234567890"},
			{"accountName": "Mandiri - Operations", "accountNumber": "0987654321"}
		]
	}`

	req := httptest.NewRequest("POST", "/api/v1/suppliers", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusCreated, rr.Code)

	var response map[string]interface{}
	json.Unmarshal(rr.Body.Bytes(), &response)

	data := response["data"].(map[string]interface{})
	assert.Equal(t, "PT Sumber Makmur", data["name"])
	assert.Equal(t, "Jl. Industri No. 45, Jakarta", data["address"])

	bankAccounts := data["bankAccounts"].([]interface{})
	assert.Equal(t, 2, len(bankAccounts))
}

func TestCreateSupplier_MissingName_Returns400(t *testing.T) {
	router, _ := setupSupplierTestRouter(t)

	body := `{"address": "Jakarta"}`

	req := httptest.NewRequest("POST", "/api/v1/suppliers", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)

	var response map[string]interface{}
	json.Unmarshal(rr.Body.Bytes(), &response)
	assert.Contains(t, response["error"], "Name is required")
}

func TestCreateSupplier_MissingAddress_Returns400(t *testing.T) {
	router, _ := setupSupplierTestRouter(t)

	body := `{"name": "Test Supplier"}`

	req := httptest.NewRequest("POST", "/api/v1/suppliers", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)

	var response map[string]interface{}
	json.Unmarshal(rr.Body.Bytes(), &response)
	assert.Contains(t, response["error"], "Address is required")
}

func TestCreateSupplier_InvalidEmail_Returns400(t *testing.T) {
	router, _ := setupSupplierTestRouter(t)

	body := `{"name": "Test", "address": "Addr", "email": "not-valid-email"}`

	req := httptest.NewRequest("POST", "/api/v1/suppliers", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)

	var response map[string]interface{}
	json.Unmarshal(rr.Body.Bytes(), &response)
	assert.Contains(t, response["error"], "email")
}

func TestCreateSupplier_BankAccountIncomplete_Returns400(t *testing.T) {
	router, _ := setupSupplierTestRouter(t)

	// Missing accountNumber
	body := `{
		"name": "Test",
		"address": "Addr",
		"bankAccounts": [{"accountName": "BCA"}]
	}`

	req := httptest.NewRequest("POST", "/api/v1/suppliers", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)

	var response map[string]interface{}
	json.Unmarshal(rr.Body.Bytes(), &response)
	assert.Contains(t, response["error"], "accountNumber")
}

func TestUpdateSupplier_ReplacesBankAccounts_Returns200(t *testing.T) {
	router, db := setupSupplierTestRouter(t)

	// Create supplier with initial bank accounts
	supplier := &models.Supplier{
		Name:    "Original Name",
		Address: "Original Address",
		Active:  true,
		BankAccounts: []models.SupplierBankAccount{
			{AccountName: "BCA Old", AccountNumber: "111"},
		},
	}
	require.NoError(t, db.Create(supplier).Error)

	// Update with new bank accounts
	body := fmt.Sprintf(`{
		"name": "Updated Name",
		"address": "Updated Address",
		"bankAccounts": [
			{"accountName": "Mandiri New", "accountNumber": "999"},
			{"accountName": "BNI New", "accountNumber": "888"}
		]
	}`)

	req := httptest.NewRequest("PUT", fmt.Sprintf("/api/v1/suppliers/%d", supplier.ID), strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var response map[string]interface{}
	json.Unmarshal(rr.Body.Bytes(), &response)

	data := response["data"].(map[string]interface{})
	assert.Equal(t, "Updated Name", data["name"])

	bankAccounts := data["bankAccounts"].([]interface{})
	assert.Equal(t, 2, len(bankAccounts))
}

func TestUpdateSupplier_NotFound_Returns404(t *testing.T) {
	router, _ := setupSupplierTestRouter(t)

	body := `{"name": "Updated", "address": "Addr"}`
	req := httptest.NewRequest("PUT", "/api/v1/suppliers/99999", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusNotFound, rr.Code)
}

func TestDeleteSupplier_NoReferences_Returns200(t *testing.T) {
	router, db := setupSupplierTestRouter(t)

	supplier := &models.Supplier{
		Name:    "To Delete",
		Address: "Test Address",
		Active:  true,
	}
	require.NoError(t, db.Create(supplier).Error)

	req := httptest.NewRequest("DELETE", fmt.Sprintf("/api/v1/suppliers/%d", supplier.ID), nil)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var response map[string]interface{}
	json.Unmarshal(rr.Body.Bytes(), &response)
	assert.Equal(t, "Supplier deleted successfully", response["message"])
}

func TestDeleteSupplier_NotFound_Returns404(t *testing.T) {
	router, _ := setupSupplierTestRouter(t)

	req := httptest.NewRequest("DELETE", "/api/v1/suppliers/99999", nil)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusNotFound, rr.Code)
}

func TestDeleteSupplier_ReferencedByPO_Returns409(t *testing.T) {
	router, db := setupSupplierTestRouter(t)

	// Create supplier
	supplier := &models.Supplier{
		Name:    "Referenced Supplier",
		Address: "Test Address",
		Active:  true,
	}
	require.NoError(t, db.Create(supplier).Error)

	// Create a purchase order referencing this supplier
	po := &models.PurchaseOrder{
		PONumber:   "PO-TEST-REF",
		SupplierID: supplier.ID,
		Date:       "2026-01-01",
		Status:     "draft",
	}
	require.NoError(t, db.Create(po).Error)

	req := httptest.NewRequest("DELETE", fmt.Sprintf("/api/v1/suppliers/%d", supplier.ID), nil)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusConflict, rr.Code)

	var response map[string]interface{}
	json.Unmarshal(rr.Body.Bytes(), &response)
	assert.Contains(t, response["error"], "purchase order")
}

func TestCreateSupplier_InvalidBody_Returns400(t *testing.T) {
	router, _ := setupSupplierTestRouter(t)

	req := httptest.NewRequest("POST", "/api/v1/suppliers", strings.NewReader("invalid json"))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestListSuppliers_WithSearch_FiltersResults(t *testing.T) {
	router, db := setupSupplierTestRouter(t)

	s1 := &models.Supplier{Name: "PT Sumber Makmur", Address: "Jakarta", Active: true}
	s2 := &models.Supplier{Name: "CV Jaya Abadi", Address: "Surabaya", Active: true}
	require.NoError(t, db.Create(s1).Error)
	require.NoError(t, db.Create(s2).Error)

	req := httptest.NewRequest("GET", "/api/v1/suppliers?search=sumber", nil)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var response map[string]interface{}
	json.Unmarshal(rr.Body.Bytes(), &response)

	data := response["data"].([]interface{})
	assert.Equal(t, 1, len(data))
	assert.Equal(t, "PT Sumber Makmur", data[0].(map[string]interface{})["name"])
}

func TestCreateSupplier_NoBankAccounts_Returns201(t *testing.T) {
	router, _ := setupSupplierTestRouter(t)

	body := `{
		"name": "Simple Supplier",
		"address": "Simple Address"
	}`

	req := httptest.NewRequest("POST", "/api/v1/suppliers", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusCreated, rr.Code)

	var response map[string]interface{}
	json.Unmarshal(rr.Body.Bytes(), &response)

	data := response["data"].(map[string]interface{})
	assert.Equal(t, "Simple Supplier", data["name"])
}
