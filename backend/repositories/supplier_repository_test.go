package repositories

import (
	"testing"

	"github.com/pointofsale/backend/models"
	"github.com/pointofsale/backend/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateSupplier_WithBankAccounts_CreatesAll(t *testing.T) {
	db := testutil.SetupTestDB(t)

	repo := NewSupplierRepository(db)

	supplier := &models.Supplier{
		Name:    "PT Sumber Makmur",
		Address: "Jl. Industri No. 45, Jakarta",
		Phone:   "+62-21-5550001",
		Email:   "order@sumbermakmur.co.id",
		Website: "sumbermakmur.co.id",
		Active:  true,
		BankAccounts: []models.SupplierBankAccount{
			{AccountName: "BCA - Main Account", AccountNumber: "1234567890"},
			{AccountName: "Mandiri - Operations", AccountNumber: "0987654321"},
		},
	}

	err := repo.Create(supplier)
	require.NoError(t, err)
	assert.NotZero(t, supplier.ID)

	// Verify bank accounts were created
	found, err := repo.FindByID(supplier.ID)
	require.NoError(t, err)
	assert.Equal(t, 2, len(found.BankAccounts))
	assert.Equal(t, "BCA - Main Account", found.BankAccounts[0].AccountName)
}

func TestGetSupplier_EagerLoadsBankAccounts(t *testing.T) {
	db := testutil.SetupTestDB(t)

	repo := NewSupplierRepository(db)

	// Create supplier with bank accounts
	supplier := &models.Supplier{
		Name:    "Test Supplier",
		Address: "Test Address",
		Active:  true,
		BankAccounts: []models.SupplierBankAccount{
			{AccountName: "BCA", AccountNumber: "111"},
			{AccountName: "BNI", AccountNumber: "222"},
		},
	}
	err := repo.Create(supplier)
	require.NoError(t, err)

	// Fetch by ID
	found, err := repo.FindByID(supplier.ID)
	require.NoError(t, err)
	assert.Equal(t, supplier.ID, found.ID)
	assert.Equal(t, "Test Supplier", found.Name)
	assert.Equal(t, 2, len(found.BankAccounts))
}

func TestGetSupplier_NotFound_ReturnsError(t *testing.T) {
	db := testutil.SetupTestDB(t)

	repo := NewSupplierRepository(db)

	found, err := repo.FindByID(99999)
	assert.Error(t, err)
	assert.Nil(t, found)
}

func TestListSuppliers_FilterByActive_Works(t *testing.T) {
	db := testutil.SetupTestDB(t)

	repo := NewSupplierRepository(db)

	// Create active supplier
	active := &models.Supplier{Name: "Active Supplier", Address: "Addr1", Active: true}
	err := repo.Create(active)
	require.NoError(t, err)

	// Create inactive supplier
	inactive := &models.Supplier{Name: "Inactive Supplier", Address: "Addr2", Active: false}
	err = repo.Create(inactive)
	require.NoError(t, err)

	// Filter active only
	activeFilter := true
	suppliers, total, err := repo.List(PaginationParams{Page: 1, PageSize: 10, SortBy: "id", SortDir: "asc"}, &activeFilter)
	require.NoError(t, err)
	assert.Equal(t, int64(1), total)
	assert.Equal(t, "Active Supplier", suppliers[0].Name)

	// Filter inactive only
	inactiveFilter := false
	suppliers, total, err = repo.List(PaginationParams{Page: 1, PageSize: 10, SortBy: "id", SortDir: "asc"}, &inactiveFilter)
	require.NoError(t, err)
	assert.Equal(t, int64(1), total)
	assert.Equal(t, "Inactive Supplier", suppliers[0].Name)

	// No filter - all results
	suppliers, total, err = repo.List(PaginationParams{Page: 1, PageSize: 10, SortBy: "id", SortDir: "asc"}, nil)
	require.NoError(t, err)
	assert.Equal(t, int64(2), total)
}

func TestListSuppliers_SearchByNameAddressEmail_Works(t *testing.T) {
	db := testutil.SetupTestDB(t)

	repo := NewSupplierRepository(db)

	s1 := &models.Supplier{Name: "PT Sumber Makmur", Address: "Jakarta", Email: "order@sumber.co.id", Active: true}
	s2 := &models.Supplier{Name: "CV Jaya Abadi", Address: "Surabaya", Email: "sales@jaya.com", Active: true}
	s3 := &models.Supplier{Name: "UD Berkah", Address: "Bandung Indah", Active: true}
	require.NoError(t, repo.Create(s1))
	require.NoError(t, repo.Create(s2))
	require.NoError(t, repo.Create(s3))

	// Search by name
	suppliers, total, err := repo.List(PaginationParams{Page: 1, PageSize: 10, Search: "sumber", SortBy: "id", SortDir: "asc"}, nil)
	require.NoError(t, err)
	assert.Equal(t, int64(1), total)
	assert.Equal(t, "PT Sumber Makmur", suppliers[0].Name)

	// Search by address
	suppliers, total, err = repo.List(PaginationParams{Page: 1, PageSize: 10, Search: "bandung", SortBy: "id", SortDir: "asc"}, nil)
	require.NoError(t, err)
	assert.Equal(t, int64(1), total)
	assert.Equal(t, "UD Berkah", suppliers[0].Name)

	// Search by email
	suppliers, total, err = repo.List(PaginationParams{Page: 1, PageSize: 10, Search: "jaya", SortBy: "id", SortDir: "asc"}, nil)
	require.NoError(t, err)
	assert.Equal(t, int64(1), total)
	assert.Equal(t, "CV Jaya Abadi", suppliers[0].Name)
}

func TestListSuppliers_EagerLoadsBankAccounts(t *testing.T) {
	db := testutil.SetupTestDB(t)

	repo := NewSupplierRepository(db)

	supplier := &models.Supplier{
		Name:    "Supplier With Banks",
		Address: "Test Address",
		Active:  true,
		BankAccounts: []models.SupplierBankAccount{
			{AccountName: "BCA", AccountNumber: "111"},
		},
	}
	require.NoError(t, repo.Create(supplier))

	suppliers, _, err := repo.List(PaginationParams{Page: 1, PageSize: 10, SortBy: "id", SortDir: "asc"}, nil)
	require.NoError(t, err)
	require.Equal(t, 1, len(suppliers))
	assert.Equal(t, 1, len(suppliers[0].BankAccounts))
}

func TestUpdateSupplier_SyncBankAccounts_ReplacesAll(t *testing.T) {
	db := testutil.SetupTestDB(t)

	repo := NewSupplierRepository(db)

	// Create supplier with 2 bank accounts
	supplier := &models.Supplier{
		Name:    "Test Supplier",
		Address: "Test Address",
		Active:  true,
		BankAccounts: []models.SupplierBankAccount{
			{AccountName: "BCA Old", AccountNumber: "111"},
			{AccountName: "BNI Old", AccountNumber: "222"},
		},
	}
	require.NoError(t, repo.Create(supplier))

	// Update supplier and replace bank accounts with new ones
	supplier.Name = "Updated Supplier"
	newBankAccounts := []models.SupplierBankAccount{
		{AccountName: "Mandiri New", AccountNumber: "333"},
	}

	err := repo.Update(supplier, newBankAccounts)
	require.NoError(t, err)

	// Verify
	found, err := repo.FindByID(supplier.ID)
	require.NoError(t, err)
	assert.Equal(t, "Updated Supplier", found.Name)
	assert.Equal(t, 1, len(found.BankAccounts))
	assert.Equal(t, "Mandiri New", found.BankAccounts[0].AccountName)
	assert.Equal(t, "333", found.BankAccounts[0].AccountNumber)
}

func TestDeleteSupplier_CascadesBankAccounts(t *testing.T) {
	db := testutil.SetupTestDB(t)

	repo := NewSupplierRepository(db)

	// Create supplier with bank accounts
	supplier := &models.Supplier{
		Name:    "To Delete",
		Address: "Test Address",
		Active:  true,
		BankAccounts: []models.SupplierBankAccount{
			{AccountName: "BCA", AccountNumber: "111"},
		},
	}
	require.NoError(t, repo.Create(supplier))

	// Delete supplier
	err := repo.Delete(supplier.ID)
	require.NoError(t, err)

	// Verify supplier is gone
	found, err := repo.FindByID(supplier.ID)
	assert.Error(t, err)
	assert.Nil(t, found)

	// Verify bank accounts are gone (CASCADE)
	var count int64
	db.Model(&models.SupplierBankAccount{}).Where("supplier_id = ?", supplier.ID).Count(&count)
	assert.Equal(t, int64(0), count)
}
