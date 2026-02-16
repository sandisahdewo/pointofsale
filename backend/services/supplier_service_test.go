package services

import (
	"errors"
	"testing"

	"github.com/pointofsale/backend/models"
	"github.com/pointofsale/backend/repositories"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockSupplierRepo implements SupplierRepositoryInterface for testing
type mockSupplierRepo struct {
	createFn                          func(*models.Supplier) error
	findByIDFn                        func(uint) (*models.Supplier, error)
	listFn                            func(repositories.PaginationParams, *bool) ([]models.Supplier, int64, error)
	updateFn                          func(*models.Supplier, []models.SupplierBankAccount) error
	deleteFn                          func(uint) error
	countPurchaseOrdersBySupplierIDFn func(uint) (int64, error)
	cleanupProductSuppliersFn         func(uint) error
}

func (m *mockSupplierRepo) Create(supplier *models.Supplier) error {
	if m.createFn != nil {
		return m.createFn(supplier)
	}
	return nil
}

func (m *mockSupplierRepo) FindByID(id uint) (*models.Supplier, error) {
	if m.findByIDFn != nil {
		return m.findByIDFn(id)
	}
	return nil, nil
}

func (m *mockSupplierRepo) List(params repositories.PaginationParams, active *bool) ([]models.Supplier, int64, error) {
	if m.listFn != nil {
		return m.listFn(params, active)
	}
	return nil, 0, nil
}

func (m *mockSupplierRepo) Update(supplier *models.Supplier, bankAccounts []models.SupplierBankAccount) error {
	if m.updateFn != nil {
		return m.updateFn(supplier, bankAccounts)
	}
	return nil
}

func (m *mockSupplierRepo) Delete(id uint) error {
	if m.deleteFn != nil {
		return m.deleteFn(id)
	}
	return nil
}

func (m *mockSupplierRepo) CountPurchaseOrdersBySupplierID(supplierID uint) (int64, error) {
	if m.countPurchaseOrdersBySupplierIDFn != nil {
		return m.countPurchaseOrdersBySupplierIDFn(supplierID)
	}
	return 0, nil
}

func (m *mockSupplierRepo) CleanupProductSuppliers(supplierID uint) error {
	if m.cleanupProductSuppliersFn != nil {
		return m.cleanupProductSuppliersFn(supplierID)
	}
	return nil
}

func TestCreateSupplier_Valid_Succeeds(t *testing.T) {
	repo := &mockSupplierRepo{
		createFn: func(s *models.Supplier) error {
			s.ID = 1
			return nil
		},
		findByIDFn: func(id uint) (*models.Supplier, error) {
			return &models.Supplier{
				ID:      1,
				Name:    "PT Sumber Makmur",
				Address: "Jakarta",
				Active:  true,
			}, nil
		},
	}
	svc := NewSupplierService(repo)

	input := CreateSupplierInput{
		Name:    "PT Sumber Makmur",
		Address: "Jakarta",
		BankAccounts: []BankAccountInput{
			{AccountName: "BCA", AccountNumber: "123"},
		},
	}

	supplier, err := svc.CreateSupplier(input)
	require.NoError(t, err)
	assert.NotNil(t, supplier)
	assert.Equal(t, "PT Sumber Makmur", supplier.Name)
}

func TestCreateSupplier_MissingName_ReturnsValidation(t *testing.T) {
	repo := &mockSupplierRepo{}
	svc := NewSupplierService(repo)

	input := CreateSupplierInput{
		Name:    "",
		Address: "Jakarta",
	}

	supplier, err := svc.CreateSupplier(input)
	assert.Nil(t, supplier)
	require.Error(t, err)

	serviceErr, ok := err.(*ServiceError)
	require.True(t, ok)
	assert.Equal(t, ErrValidation, serviceErr.Err)
}

func TestCreateSupplier_MissingAddress_ReturnsValidation(t *testing.T) {
	repo := &mockSupplierRepo{}
	svc := NewSupplierService(repo)

	input := CreateSupplierInput{
		Name:    "Test",
		Address: "",
	}

	supplier, err := svc.CreateSupplier(input)
	assert.Nil(t, supplier)
	require.Error(t, err)

	serviceErr, ok := err.(*ServiceError)
	require.True(t, ok)
	assert.Equal(t, ErrValidation, serviceErr.Err)
}

func TestCreateSupplier_InvalidEmail_ReturnsValidation(t *testing.T) {
	repo := &mockSupplierRepo{}
	svc := NewSupplierService(repo)

	input := CreateSupplierInput{
		Name:    "Test",
		Address: "Addr",
		Email:   "not-an-email",
	}

	supplier, err := svc.CreateSupplier(input)
	assert.Nil(t, supplier)
	require.Error(t, err)

	serviceErr, ok := err.(*ServiceError)
	require.True(t, ok)
	assert.Equal(t, ErrValidation, serviceErr.Err)
	assert.Contains(t, serviceErr.Message, "email")
}

func TestCreateSupplier_BankAccountMissingFields_ReturnsValidation(t *testing.T) {
	repo := &mockSupplierRepo{}
	svc := NewSupplierService(repo)

	// Missing accountNumber
	input := CreateSupplierInput{
		Name:    "Test",
		Address: "Addr",
		BankAccounts: []BankAccountInput{
			{AccountName: "BCA", AccountNumber: ""},
		},
	}

	supplier, err := svc.CreateSupplier(input)
	assert.Nil(t, supplier)
	require.Error(t, err)

	serviceErr, ok := err.(*ServiceError)
	require.True(t, ok)
	assert.Equal(t, ErrValidation, serviceErr.Err)

	// Missing accountName
	input2 := CreateSupplierInput{
		Name:    "Test",
		Address: "Addr",
		BankAccounts: []BankAccountInput{
			{AccountName: "", AccountNumber: "123"},
		},
	}

	supplier, err = svc.CreateSupplier(input2)
	assert.Nil(t, supplier)
	require.Error(t, err)

	serviceErr, ok = err.(*ServiceError)
	require.True(t, ok)
	assert.Equal(t, ErrValidation, serviceErr.Err)
}

func TestUpdateSupplier_SyncsBankAccountsAtomically(t *testing.T) {
	existingSupplier := &models.Supplier{
		ID:      1,
		Name:    "Old Name",
		Address: "Old Addr",
		Active:  true,
	}

	var updatedSupplier *models.Supplier
	var updatedBankAccounts []models.SupplierBankAccount

	repo := &mockSupplierRepo{
		findByIDFn: func(id uint) (*models.Supplier, error) {
			if updatedSupplier != nil {
				return updatedSupplier, nil
			}
			return existingSupplier, nil
		},
		updateFn: func(s *models.Supplier, ba []models.SupplierBankAccount) error {
			updatedSupplier = s
			updatedSupplier.BankAccounts = []models.SupplierBankAccount{
				{ID: "new-uuid", AccountName: "New Bank", AccountNumber: "999"},
			}
			updatedBankAccounts = ba
			return nil
		},
	}
	svc := NewSupplierService(repo)

	input := UpdateSupplierInput{
		Name:    "New Name",
		Address: "New Addr",
		BankAccounts: &[]BankAccountInput{
			{AccountName: "New Bank", AccountNumber: "999"},
		},
	}

	result, err := svc.UpdateSupplier(1, input)
	require.NoError(t, err)
	assert.Equal(t, "New Name", result.Name)
	assert.Equal(t, 1, len(updatedBankAccounts))
	assert.Equal(t, "New Bank", updatedBankAccounts[0].AccountName)
}

func TestDeleteSupplier_ReferencedByPO_ReturnsConflict(t *testing.T) {
	repo := &mockSupplierRepo{
		findByIDFn: func(id uint) (*models.Supplier, error) {
			return &models.Supplier{ID: 1, Name: "Test", Address: "Addr"}, nil
		},
		countPurchaseOrdersBySupplierIDFn: func(supplierID uint) (int64, error) {
			return 3, nil
		},
	}
	svc := NewSupplierService(repo)

	err := svc.DeleteSupplier(1)
	require.Error(t, err)

	serviceErr, ok := err.(*ServiceError)
	require.True(t, ok)
	assert.Equal(t, ErrConflict, serviceErr.Err)
	assert.Contains(t, serviceErr.Message, "3 purchase order(s)")
}

func TestDeleteSupplier_NotFound_ReturnsNotFound(t *testing.T) {
	repo := &mockSupplierRepo{
		findByIDFn: func(id uint) (*models.Supplier, error) {
			return nil, errors.New("not found")
		},
	}
	svc := NewSupplierService(repo)

	err := svc.DeleteSupplier(999)
	require.Error(t, err)

	serviceErr, ok := err.(*ServiceError)
	require.True(t, ok)
	assert.Equal(t, ErrNotFound, serviceErr.Err)
}

func TestDeleteSupplier_ReferencedByProductsOnly_CleansUpAndDeletes(t *testing.T) {
	cleanedUp := false
	deleted := false

	repo := &mockSupplierRepo{
		findByIDFn: func(id uint) (*models.Supplier, error) {
			return &models.Supplier{ID: 1, Name: "Test", Address: "Addr"}, nil
		},
		countPurchaseOrdersBySupplierIDFn: func(supplierID uint) (int64, error) {
			return 0, nil // No PO references
		},
		cleanupProductSuppliersFn: func(supplierID uint) error {
			cleanedUp = true
			return nil
		},
		deleteFn: func(id uint) error {
			deleted = true
			return nil
		},
	}
	svc := NewSupplierService(repo)

	err := svc.DeleteSupplier(1)
	require.NoError(t, err)
	assert.True(t, cleanedUp, "should cleanup product_suppliers")
	assert.True(t, deleted, "should delete supplier")
}

func TestDeleteSupplier_NoReferences_Deletes(t *testing.T) {
	deleted := false

	repo := &mockSupplierRepo{
		findByIDFn: func(id uint) (*models.Supplier, error) {
			return &models.Supplier{ID: 1, Name: "Test", Address: "Addr"}, nil
		},
		countPurchaseOrdersBySupplierIDFn: func(supplierID uint) (int64, error) {
			return 0, nil
		},
		cleanupProductSuppliersFn: func(supplierID uint) error {
			return nil
		},
		deleteFn: func(id uint) error {
			deleted = true
			return nil
		},
	}
	svc := NewSupplierService(repo)

	err := svc.DeleteSupplier(1)
	require.NoError(t, err)
	assert.True(t, deleted)
}
