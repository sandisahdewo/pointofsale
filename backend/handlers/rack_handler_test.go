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

// rackRepoAdapter wraps RackRepositoryImpl to satisfy RackServiceRepository
type rackRepoAdapter struct {
	*repositories.RackRepositoryImpl
	db *gorm.DB
}

func (a *rackRepoAdapter) CleanupVariantRacks(rackID uint) error {
	if a.db.Migrator().HasTable("variant_racks") {
		return a.db.Exec("DELETE FROM variant_racks WHERE rack_id = ?", rackID).Error
	}
	return nil
}

// setupRackTestRouter creates a test router with rack endpoints
func setupRackTestRouter(t *testing.T) (chi.Router, *gorm.DB) {
	t.Helper()

	db := testutil.SetupTestDB(t)

	// Initialize layers
	rackRepo := repositories.NewRackRepository(db)
	adapter := &rackRepoAdapter{RackRepositoryImpl: rackRepo, db: db}
	rackService := services.NewRackService(adapter)
	rackHandler := NewRackHandler(rackService)

	// Setup router
	r := chi.NewRouter()
	r.Route("/api/v1/racks", func(r chi.Router) {
		r.Get("/", rackHandler.ListRacks)
		r.Get("/{id}", rackHandler.GetRack)
		r.Post("/", rackHandler.CreateRack)
		r.Put("/{id}", rackHandler.UpdateRack)
		r.Delete("/{id}", rackHandler.DeleteRack)
	})

	return r, db
}

// createTestRackInDB creates a rack directly in the DB for handler tests
func createTestRackInDB(t *testing.T, db *gorm.DB, rack *models.Rack) *models.Rack {
	t.Helper()
	err := db.Create(rack).Error
	require.NoError(t, err)
	return rack
}

// TestListRacks_Returns200 verifies list endpoint with pagination
func TestListRacks_Returns200(t *testing.T) {
	router, db := setupRackTestRouter(t)
	defer testutil.CleanupTestDB(t, db)

	createTestRackInDB(t, db, &models.Rack{
		Name: "Main Display", Code: "R-001", Location: "Store Front", Capacity: 100, Active: true,
	})
	createTestRackInDB(t, db, &models.Rack{
		Name: "Electronics Shelf", Code: "R-002", Location: "Store Front", Capacity: 50, Active: true,
	})

	req := httptest.NewRequest("GET", "/api/v1/racks", nil)
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

	meta := response["meta"].(map[string]interface{})
	assert.NotNil(t, meta["page"])
	assert.NotNil(t, meta["pageSize"])
	assert.NotNil(t, meta["totalItems"])
	assert.NotNil(t, meta["totalPages"])
}

// TestListRacks_WithSearch_FiltersResults verifies search functionality
func TestListRacks_WithSearch_FiltersResults(t *testing.T) {
	router, db := setupRackTestRouter(t)
	defer testutil.CleanupTestDB(t, db)

	createTestRackInDB(t, db, &models.Rack{
		Name: "Main Display", Code: "MD-001", Location: "Store Front", Capacity: 100, Active: true,
	})
	createTestRackInDB(t, db, &models.Rack{
		Name: "Cold Storage", Code: "CS-001", Location: "Warehouse", Capacity: 200, Active: true,
	})

	req := httptest.NewRequest("GET", "/api/v1/racks?search=Cold", nil)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var response map[string]interface{}
	json.Unmarshal(rr.Body.Bytes(), &response)

	data := response["data"].([]interface{})
	assert.Equal(t, 1, len(data))
	rack := data[0].(map[string]interface{})
	assert.Equal(t, "Cold Storage", rack["name"])
}

// TestListRacks_FilterActive_ReturnsActiveOnly verifies active filter
func TestListRacks_FilterActive_ReturnsActiveOnly(t *testing.T) {
	router, db := setupRackTestRouter(t)
	defer testutil.CleanupTestDB(t, db)

	createTestRackInDB(t, db, &models.Rack{
		Name: "Active Rack", Code: "AR-001", Location: "Location 1", Capacity: 50, Active: true,
	})
	inactiveRack := createTestRackInDB(t, db, &models.Rack{
		Name: "Inactive Rack", Code: "IR-001", Location: "Location 2", Capacity: 30, Active: true,
	})
	db.Model(inactiveRack).Update("active", false)

	req := httptest.NewRequest("GET", "/api/v1/racks?active=true", nil)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var response map[string]interface{}
	json.Unmarshal(rr.Body.Bytes(), &response)

	data := response["data"].([]interface{})
	assert.Equal(t, 1, len(data))
	rack := data[0].(map[string]interface{})
	assert.Equal(t, true, rack["active"])
}

// TestGetRack_Exists_Returns200 verifies getting a rack by ID
func TestGetRack_Exists_Returns200(t *testing.T) {
	router, db := setupRackTestRouter(t)
	defer testutil.CleanupTestDB(t, db)

	rack := createTestRackInDB(t, db, &models.Rack{
		Name: "Main Display", Code: "R-001", Location: "Store Front", Capacity: 100,
		Description: "Primary display", Active: true,
	})

	req := httptest.NewRequest("GET", fmt.Sprintf("/api/v1/racks/%d", rack.ID), nil)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var response map[string]interface{}
	json.Unmarshal(rr.Body.Bytes(), &response)

	data := response["data"].(map[string]interface{})
	assert.Equal(t, "Main Display", data["name"])
	assert.Equal(t, "R-001", data["code"])
	assert.Equal(t, "Store Front", data["location"])
	assert.Equal(t, float64(100), data["capacity"])
	assert.Equal(t, "Primary display", data["description"])
}

// TestGetRack_NotFound_Returns404 verifies 404 for non-existent rack
func TestGetRack_NotFound_Returns404(t *testing.T) {
	router, db := setupRackTestRouter(t)
	defer testutil.CleanupTestDB(t, db)

	req := httptest.NewRequest("GET", "/api/v1/racks/99999", nil)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusNotFound, rr.Code)

	var response map[string]interface{}
	json.Unmarshal(rr.Body.Bytes(), &response)
	assert.Contains(t, response, "error")
}

// TestCreateRack_ValidBody_Returns201 verifies rack creation
func TestCreateRack_ValidBody_Returns201(t *testing.T) {
	router, db := setupRackTestRouter(t)
	defer testutil.CleanupTestDB(t, db)

	body := `{
		"name": "Main Display",
		"code": "R-001",
		"location": "Store Front",
		"capacity": 100,
		"description": "Primary display shelf"
	}`

	req := httptest.NewRequest("POST", "/api/v1/racks", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusCreated, rr.Code)

	var response map[string]interface{}
	json.Unmarshal(rr.Body.Bytes(), &response)

	data := response["data"].(map[string]interface{})
	assert.Equal(t, "Main Display", data["name"])
	assert.Equal(t, "R-001", data["code"])
	assert.Equal(t, "Store Front", data["location"])
	assert.Equal(t, float64(100), data["capacity"])
	assert.Equal(t, "Primary display shelf", data["description"])
	assert.Equal(t, true, data["active"])
}

// TestCreateRack_DuplicateCode_Returns409 verifies conflict on duplicate code
func TestCreateRack_DuplicateCode_Returns409(t *testing.T) {
	router, db := setupRackTestRouter(t)
	defer testutil.CleanupTestDB(t, db)

	createTestRackInDB(t, db, &models.Rack{
		Name: "Existing", Code: "R-001", Location: "Location", Capacity: 50, Active: true,
	})

	body := `{
		"name": "New Rack",
		"code": "R-001",
		"location": "Other Location",
		"capacity": 100
	}`

	req := httptest.NewRequest("POST", "/api/v1/racks", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusConflict, rr.Code)

	var response map[string]interface{}
	json.Unmarshal(rr.Body.Bytes(), &response)
	assert.Contains(t, response["error"], "already exists")
}

// TestCreateRack_MissingCode_Returns400 verifies validation
func TestCreateRack_MissingCode_Returns400(t *testing.T) {
	router, db := setupRackTestRouter(t)
	defer testutil.CleanupTestDB(t, db)

	body := `{
		"name": "Rack",
		"location": "Location",
		"capacity": 100
	}`

	req := httptest.NewRequest("POST", "/api/v1/racks", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)

	var response map[string]interface{}
	json.Unmarshal(rr.Body.Bytes(), &response)
	assert.Contains(t, response["error"], "Code is required")
}

// TestCreateRack_MissingName_Returns400 verifies name validation
func TestCreateRack_MissingName_Returns400(t *testing.T) {
	router, db := setupRackTestRouter(t)
	defer testutil.CleanupTestDB(t, db)

	body := `{
		"code": "R-001",
		"location": "Location",
		"capacity": 100
	}`

	req := httptest.NewRequest("POST", "/api/v1/racks", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)

	var response map[string]interface{}
	json.Unmarshal(rr.Body.Bytes(), &response)
	assert.Contains(t, response["error"], "Name is required")
}

// TestCreateRack_ZeroCapacity_Returns400 verifies capacity validation
func TestCreateRack_ZeroCapacity_Returns400(t *testing.T) {
	router, db := setupRackTestRouter(t)
	defer testutil.CleanupTestDB(t, db)

	body := `{
		"name": "Rack",
		"code": "R-001",
		"location": "Location",
		"capacity": 0
	}`

	req := httptest.NewRequest("POST", "/api/v1/racks", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)

	var response map[string]interface{}
	json.Unmarshal(rr.Body.Bytes(), &response)
	assert.Contains(t, response["error"], "Capacity must be greater than 0")
}

// TestCreateRack_InvalidJSON_Returns400 verifies JSON parsing
func TestCreateRack_InvalidJSON_Returns400(t *testing.T) {
	router, db := setupRackTestRouter(t)
	defer testutil.CleanupTestDB(t, db)

	body := `{invalid json`

	req := httptest.NewRequest("POST", "/api/v1/racks", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

// TestUpdateRack_ValidBody_Returns200 verifies rack update
func TestUpdateRack_ValidBody_Returns200(t *testing.T) {
	router, db := setupRackTestRouter(t)
	defer testutil.CleanupTestDB(t, db)

	rack := createTestRackInDB(t, db, &models.Rack{
		Name: "OldName", Code: "R-001", Location: "Old Loc", Capacity: 50, Active: true,
	})

	body := `{
		"name": "NewName",
		"code": "R-001",
		"location": "New Loc",
		"capacity": 100,
		"active": true
	}`

	req := httptest.NewRequest("PUT", fmt.Sprintf("/api/v1/racks/%d", rack.ID), strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var response map[string]interface{}
	json.Unmarshal(rr.Body.Bytes(), &response)

	data := response["data"].(map[string]interface{})
	assert.Equal(t, "NewName", data["name"])
	assert.Equal(t, "New Loc", data["location"])
	assert.Equal(t, float64(100), data["capacity"])
}

// TestUpdateRack_DuplicateCodeOtherRack_Returns409 verifies conflict on update
func TestUpdateRack_DuplicateCodeOtherRack_Returns409(t *testing.T) {
	router, db := setupRackTestRouter(t)
	defer testutil.CleanupTestDB(t, db)

	rack1 := createTestRackInDB(t, db, &models.Rack{
		Name: "Rack 1", Code: "R-001", Location: "Location 1", Capacity: 50, Active: true,
	})
	createTestRackInDB(t, db, &models.Rack{
		Name: "Rack 2", Code: "R-002", Location: "Location 2", Capacity: 75, Active: true,
	})

	body := `{
		"name": "Rack 1 Updated",
		"code": "R-002",
		"location": "Location 1",
		"capacity": 50,
		"active": true
	}`

	req := httptest.NewRequest("PUT", fmt.Sprintf("/api/v1/racks/%d", rack1.ID), strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusConflict, rr.Code)

	var response map[string]interface{}
	json.Unmarshal(rr.Body.Bytes(), &response)
	assert.Contains(t, response["error"], "already exists")
}

// TestUpdateRack_SameCodeSelf_Returns200 verifies updating with own code works
func TestUpdateRack_SameCodeSelf_Returns200(t *testing.T) {
	router, db := setupRackTestRouter(t)
	defer testutil.CleanupTestDB(t, db)

	rack := createTestRackInDB(t, db, &models.Rack{
		Name: "Rack 1", Code: "R-001", Location: "Location 1", Capacity: 50, Active: true,
	})

	body := `{
		"name": "Rack 1 Updated",
		"code": "R-001",
		"location": "New Location",
		"capacity": 200,
		"active": true
	}`

	req := httptest.NewRequest("PUT", fmt.Sprintf("/api/v1/racks/%d", rack.ID), strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var response map[string]interface{}
	json.Unmarshal(rr.Body.Bytes(), &response)

	data := response["data"].(map[string]interface{})
	assert.Equal(t, "Rack 1 Updated", data["name"])
	assert.Equal(t, "R-001", data["code"])
	assert.Equal(t, float64(200), data["capacity"])
}

// TestUpdateRack_NotFound_Returns404 verifies 404 for non-existent rack
func TestUpdateRack_NotFound_Returns404(t *testing.T) {
	router, db := setupRackTestRouter(t)
	defer testutil.CleanupTestDB(t, db)

	body := `{
		"name": "Rack",
		"code": "R-001",
		"location": "Location",
		"capacity": 50
	}`

	req := httptest.NewRequest("PUT", "/api/v1/racks/99999", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusNotFound, rr.Code)
}

// TestDeleteRack_Returns200 verifies rack deletion
func TestDeleteRack_Returns200(t *testing.T) {
	router, db := setupRackTestRouter(t)
	defer testutil.CleanupTestDB(t, db)

	rack := createTestRackInDB(t, db, &models.Rack{
		Name: "ToDelete", Code: "DEL-001", Location: "Location", Capacity: 50, Active: true,
	})

	req := httptest.NewRequest("DELETE", fmt.Sprintf("/api/v1/racks/%d", rack.ID), nil)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var response map[string]interface{}
	json.Unmarshal(rr.Body.Bytes(), &response)
	assert.Contains(t, response["message"], "deleted successfully")

	// Verify rack is actually deleted
	getReq := httptest.NewRequest("GET", fmt.Sprintf("/api/v1/racks/%d", rack.ID), nil)
	getRr := httptest.NewRecorder()
	router.ServeHTTP(getRr, getReq)
	assert.Equal(t, http.StatusNotFound, getRr.Code)
}

// TestDeleteRack_NotFound_Returns404 verifies 404 for non-existent rack
func TestDeleteRack_NotFound_Returns404(t *testing.T) {
	router, db := setupRackTestRouter(t)
	defer testutil.CleanupTestDB(t, db)

	req := httptest.NewRequest("DELETE", "/api/v1/racks/99999", nil)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusNotFound, rr.Code)
}

// TestDeleteRack_InvalidID_Returns400 verifies invalid ID format
func TestDeleteRack_InvalidID_Returns400(t *testing.T) {
	router, db := setupRackTestRouter(t)
	defer testutil.CleanupTestDB(t, db)

	req := httptest.NewRequest("DELETE", "/api/v1/racks/invalid", nil)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

// TestGetRack_InvalidID_Returns400 verifies invalid ID format
func TestGetRack_InvalidID_Returns400(t *testing.T) {
	router, db := setupRackTestRouter(t)
	defer testutil.CleanupTestDB(t, db)

	req := httptest.NewRequest("GET", "/api/v1/racks/invalid", nil)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}
