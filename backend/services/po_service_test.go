package services

import (
	"errors"
	"testing"
	"time"

	"github.com/pointofsale/backend/models"
	"github.com/pointofsale/backend/repositories"
	"github.com/pointofsale/backend/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

// --- Mock implementations ---

type mockPORepo struct {
	createFn       func(*models.PurchaseOrder) error
	getByIDFn      func(uint) (*models.PurchaseOrder, error)
	listFn         func(repositories.PaginationParams, string, uint) ([]models.PurchaseOrder, int64, error)
	statusCountsFn func() (map[string]int64, error)
	updateFn       func(*models.PurchaseOrder) error
	deleteFn       func(uint) error
	replaceItemsFn func(uint, []models.PurchaseOrderItem) error
	getProductsFn  func(uint, string) ([]models.Product, error)
}

func (m *mockPORepo) Create(po *models.PurchaseOrder) error {
	if m.createFn != nil {
		return m.createFn(po)
	}
	po.ID = 1
	return nil
}
func (m *mockPORepo) GetByID(id uint) (*models.PurchaseOrder, error) {
	if m.getByIDFn != nil {
		return m.getByIDFn(id)
	}
	return nil, gorm.ErrRecordNotFound
}
func (m *mockPORepo) List(p repositories.PaginationParams, s string, sid uint) ([]models.PurchaseOrder, int64, error) {
	if m.listFn != nil {
		return m.listFn(p, s, sid)
	}
	return nil, 0, nil
}
func (m *mockPORepo) StatusCounts() (map[string]int64, error) {
	if m.statusCountsFn != nil {
		return m.statusCountsFn()
	}
	return map[string]int64{"all": 0}, nil
}
func (m *mockPORepo) Update(po *models.PurchaseOrder) error {
	if m.updateFn != nil {
		return m.updateFn(po)
	}
	return nil
}
func (m *mockPORepo) Delete(id uint) error {
	if m.deleteFn != nil {
		return m.deleteFn(id)
	}
	return nil
}
func (m *mockPORepo) ReplaceItems(poID uint, items []models.PurchaseOrderItem) error {
	if m.replaceItemsFn != nil {
		return m.replaceItemsFn(poID, items)
	}
	return nil
}
func (m *mockPORepo) GetProductsForPO(supplierID uint, search string) ([]models.Product, error) {
	if m.getProductsFn != nil {
		return m.getProductsFn(supplierID, search)
	}
	return nil, nil
}

type mockStockRepo struct {
	createFn        func(*models.StockMovement) error
	getByVariantFn  func(string) ([]models.StockMovement, error)
	getByReferenceFn func(string, uint) ([]models.StockMovement, error)
}

func (m *mockStockRepo) Create(movement *models.StockMovement) error {
	if m.createFn != nil {
		return m.createFn(movement)
	}
	return nil
}
func (m *mockStockRepo) GetByVariant(variantID string) ([]models.StockMovement, error) {
	if m.getByVariantFn != nil {
		return m.getByVariantFn(variantID)
	}
	return nil, nil
}
func (m *mockStockRepo) GetByReference(referenceType string, referenceID uint) ([]models.StockMovement, error) {
	if m.getByReferenceFn != nil {
		return m.getByReferenceFn(referenceType, referenceID)
	}
	return nil, nil
}

// --- Tests ---

func TestCreatePO_Valid_GeneratesPONumber(t *testing.T) {
	db := testutil.SetupTestDB(t)
	poRepo := &mockPORepo{}
	stockRepo := &mockStockRepo{}
	seqSvc := NewSequenceService(db)

	svc := NewPOService(db, poRepo, stockRepo, seqSvc)

	supplier := testutil.CreateTestSupplier(t, db)
	product := testutil.CreateTestProduct(t, db)
	variant := product.Variants[0]
	unit := product.Units[0]

	input := CreatePOInput{
		SupplierID: supplier.ID,
		Date:       "2026-01-15",
		Notes:      "Test order",
		Items: []CreatePOItemInput{
			{
				ProductID:  product.ID,
				VariantID:  variant.ID,
				UnitID:     unit.ID,
				OrderedQty: 10,
				Price:      15000,
			},
		},
	}

	po, err := svc.CreatePO(input)
	require.NoError(t, err)
	assert.NotEmpty(t, po.PONumber)
	assert.Contains(t, po.PONumber, "PO-")
}

func TestCreatePO_DenormalizesItemFields(t *testing.T) {
	db := testutil.SetupTestDB(t)
	poRepo := &mockPORepo{}
	stockRepo := &mockStockRepo{}
	seqSvc := NewSequenceService(db)

	svc := NewPOService(db, poRepo, stockRepo, seqSvc)

	supplier := testutil.CreateTestSupplier(t, db)
	product := testutil.CreateTestProduct(t, db)
	variant := product.Variants[0]
	unit := product.Units[0]

	input := CreatePOInput{
		SupplierID: supplier.ID,
		Date:       "2026-01-15",
		Items: []CreatePOItemInput{
			{
				ProductID:  product.ID,
				VariantID:  variant.ID,
				UnitID:     unit.ID,
				OrderedQty: 5,
				Price:      10000,
			},
		},
	}

	po, err := svc.CreatePO(input)
	require.NoError(t, err)
	require.Len(t, po.Items, 1)
	assert.Equal(t, product.Name, po.Items[0].ProductName)
	assert.Equal(t, unit.Name, po.Items[0].UnitName)
	assert.NotEmpty(t, po.Items[0].VariantLabel)
}

func TestCreatePO_InactiveSupplier_ReturnsError(t *testing.T) {
	db := testutil.SetupTestDB(t)
	poRepo := &mockPORepo{}
	stockRepo := &mockStockRepo{}
	seqSvc := NewSequenceService(db)

	svc := NewPOService(db, poRepo, stockRepo, seqSvc)

	inactiveSupplier := testutil.CreateTestSupplier(t, db, func(s *models.Supplier) {
		s.Active = false
	})

	input := CreatePOInput{
		SupplierID: inactiveSupplier.ID,
		Date:       "2026-01-15",
		Items:      []CreatePOItemInput{{ProductID: 1, VariantID: "uuid", UnitID: 1, OrderedQty: 1}},
	}

	_, err := svc.CreatePO(input)
	require.Error(t, err)
	serviceErr, ok := err.(*ServiceError)
	require.True(t, ok)
	assert.Equal(t, ErrValidation, serviceErr.Err)
}

func TestCreatePO_NoItemsWithQty_ReturnsValidation(t *testing.T) {
	db := testutil.SetupTestDB(t)
	poRepo := &mockPORepo{}
	stockRepo := &mockStockRepo{}
	seqSvc := NewSequenceService(db)

	svc := NewPOService(db, poRepo, stockRepo, seqSvc)

	supplier := testutil.CreateTestSupplier(t, db)

	input := CreatePOInput{
		SupplierID: supplier.ID,
		Date:       "2026-01-15",
		Items:      []CreatePOItemInput{},
	}

	_, err := svc.CreatePO(input)
	require.Error(t, err)
	serviceErr, ok := err.(*ServiceError)
	require.True(t, ok)
	assert.Equal(t, ErrValidation, serviceErr.Err)
}

func TestUpdatePO_NonDraft_ReturnsForbidden(t *testing.T) {
	db := testutil.SetupTestDB(t)
	stockRepo := &mockStockRepo{}
	seqSvc := NewSequenceService(db)

	sentPO := &models.PurchaseOrder{ID: 1, Status: "sent", SupplierID: 1}
	poRepo := &mockPORepo{
		getByIDFn: func(id uint) (*models.PurchaseOrder, error) {
			return sentPO, nil
		},
	}

	svc := NewPOService(db, poRepo, stockRepo, seqSvc)

	_, err := svc.UpdatePO(1, CreatePOInput{SupplierID: 1, Date: "2026-01-15"})
	require.Error(t, err)
	serviceErr, ok := err.(*ServiceError)
	require.True(t, ok)
	assert.Equal(t, ErrForbidden, serviceErr.Err)
}

func TestDeletePO_NonDraft_ReturnsForbidden(t *testing.T) {
	db := testutil.SetupTestDB(t)
	stockRepo := &mockStockRepo{}
	seqSvc := NewSequenceService(db)

	sentPO := &models.PurchaseOrder{ID: 1, Status: "sent"}
	poRepo := &mockPORepo{
		getByIDFn: func(id uint) (*models.PurchaseOrder, error) {
			return sentPO, nil
		},
	}

	svc := NewPOService(db, poRepo, stockRepo, seqSvc)

	err := svc.DeletePO(1)
	require.Error(t, err)
	serviceErr, ok := err.(*ServiceError)
	require.True(t, ok)
	assert.Equal(t, ErrForbidden, serviceErr.Err)
}

func TestUpdatePOStatus_ValidTransition_Succeeds(t *testing.T) {
	db := testutil.SetupTestDB(t)
	stockRepo := &mockStockRepo{}
	seqSvc := NewSequenceService(db)

	draftPO := &models.PurchaseOrder{ID: 1, Status: "draft"}
	var savedPO *models.PurchaseOrder
	poRepo := &mockPORepo{
		getByIDFn: func(id uint) (*models.PurchaseOrder, error) {
			return draftPO, nil
		},
		updateFn: func(po *models.PurchaseOrder) error {
			savedPO = po
			return nil
		},
	}

	svc := NewPOService(db, poRepo, stockRepo, seqSvc)

	updated, err := svc.UpdatePOStatus(1, "sent")
	require.NoError(t, err)
	assert.Equal(t, "sent", updated.Status)
	require.NotNil(t, savedPO)
	assert.Equal(t, "sent", savedPO.Status)
}

func TestUpdatePOStatus_InvalidTransition_ReturnsError(t *testing.T) {
	db := testutil.SetupTestDB(t)
	stockRepo := &mockStockRepo{}
	seqSvc := NewSequenceService(db)

	// received -> draft is not allowed
	receivedPO := &models.PurchaseOrder{ID: 1, Status: "received"}
	poRepo := &mockPORepo{
		getByIDFn: func(id uint) (*models.PurchaseOrder, error) {
			return receivedPO, nil
		},
	}

	svc := NewPOService(db, poRepo, stockRepo, seqSvc)

	_, err := svc.UpdatePOStatus(1, "draft")
	require.Error(t, err)
	serviceErr, ok := err.(*ServiceError)
	require.True(t, ok)
	assert.Equal(t, ErrValidation, serviceErr.Err)
}

func TestDeletePO_NotFound_ReturnsError(t *testing.T) {
	db := testutil.SetupTestDB(t)
	stockRepo := &mockStockRepo{}
	seqSvc := NewSequenceService(db)

	poRepo := &mockPORepo{
		getByIDFn: func(id uint) (*models.PurchaseOrder, error) {
			return nil, gorm.ErrRecordNotFound
		},
	}

	svc := NewPOService(db, poRepo, stockRepo, seqSvc)

	err := svc.DeletePO(999)
	require.Error(t, err)
	serviceErr, ok := err.(*ServiceError)
	require.True(t, ok)
	assert.Equal(t, ErrNotFound, serviceErr.Err)
}

func TestDeletePO_RepoError_ReturnsError(t *testing.T) {
	db := testutil.SetupTestDB(t)
	stockRepo := &mockStockRepo{}
	seqSvc := NewSequenceService(db)

	draftPO := &models.PurchaseOrder{ID: 1, Status: "draft"}
	poRepo := &mockPORepo{
		getByIDFn: func(id uint) (*models.PurchaseOrder, error) {
			return draftPO, nil
		},
		deleteFn: func(id uint) error {
			return errors.New("database error")
		},
	}

	svc := NewPOService(db, poRepo, stockRepo, seqSvc)

	err := svc.DeletePO(1)
	require.Error(t, err)
}

func TestListPOs_ReturnsStatusCounts(t *testing.T) {
	db := testutil.SetupTestDB(t)
	stockRepo := &mockStockRepo{}
	seqSvc := NewSequenceService(db)

	poRepo := &mockPORepo{
		listFn: func(p repositories.PaginationParams, s string, sid uint) ([]models.PurchaseOrder, int64, error) {
			return []models.PurchaseOrder{}, 0, nil
		},
		statusCountsFn: func() (map[string]int64, error) {
			return map[string]int64{"all": 5, "draft": 3, "sent": 2}, nil
		},
	}

	svc := NewPOService(db, poRepo, stockRepo, seqSvc)

	params := repositories.PaginationParams{Page: 1, PageSize: 10}
	_, _, counts, err := svc.ListPOs(params, "", 0)
	require.NoError(t, err)
	assert.Equal(t, int64(5), counts["all"])
	assert.Equal(t, int64(3), counts["draft"])
}

func TestReceivePO_BankTransferNoBankAccount_ReturnsError(t *testing.T) {
	db := testutil.SetupTestDB(t)
	stockRepo := &mockStockRepo{}
	seqSvc := NewSequenceService(db)

	sentPO := &models.PurchaseOrder{
		ID:     1,
		Status: "sent",
		Items: []models.PurchaseOrderItem{
			{ID: "item-1", OrderedQty: 10, Price: 5000},
		},
	}
	poRepo := &mockPORepo{
		getByIDFn: func(id uint) (*models.PurchaseOrder, error) {
			return sentPO, nil
		},
	}

	svc := NewPOService(db, poRepo, stockRepo, seqSvc)

	input := ReceivePOInput{
		ReceivedDate:  time.Now().Format("2006-01-02"),
		PaymentMethod: "bank_transfer",
		// No SupplierBankAccountID
		Items: []ReceivePOItemInput{
			{ItemID: "item-1", ReceivedQty: 10, ReceivedPrice: 5000, IsVerified: true},
		},
	}

	_, err := svc.ReceivePO(1, input)
	require.Error(t, err)
	serviceErr, ok := err.(*ServiceError)
	require.True(t, ok)
	assert.Equal(t, ErrValidation, serviceErr.Err)
}
