package services

import (
	"fmt"
	"strings"

	"github.com/pointofsale/backend/models"
	"github.com/pointofsale/backend/repositories"
	"github.com/pointofsale/backend/utils"
	"gorm.io/gorm"
)

// SupplierRepositoryInterface defines the repository interface needed by SupplierService
type SupplierRepositoryInterface interface {
	Create(supplier *models.Supplier) error
	FindByID(id uint) (*models.Supplier, error)
	List(params repositories.PaginationParams, active *bool) ([]models.Supplier, int64, error)
	Update(supplier *models.Supplier, bankAccounts []models.SupplierBankAccount) error
	Delete(id uint) error
	CountPurchaseOrdersBySupplierID(supplierID uint) (int64, error)
	CleanupProductSuppliers(supplierID uint) error
}

// SupplierService handles supplier business logic
type SupplierService struct {
	supplierRepo SupplierRepositoryInterface
}

// NewSupplierService creates a new supplier service instance
func NewSupplierService(supplierRepo SupplierRepositoryInterface) *SupplierService {
	return &SupplierService{supplierRepo: supplierRepo}
}

// BankAccountInput is the DTO for bank account input
type BankAccountInput struct {
	AccountName   string `json:"accountName"`
	AccountNumber string `json:"accountNumber"`
}

// CreateSupplierInput is the DTO for creating a supplier
type CreateSupplierInput struct {
	Name         string             `json:"name"`
	Address      string             `json:"address"`
	Phone        string             `json:"phone,omitempty"`
	Email        string             `json:"email,omitempty"`
	Website      string             `json:"website,omitempty"`
	BankAccounts []BankAccountInput `json:"bankAccounts,omitempty"`
}

// UpdateSupplierInput is the DTO for updating a supplier
type UpdateSupplierInput struct {
	Name         string              `json:"name"`
	Address      string              `json:"address"`
	Phone        string              `json:"phone,omitempty"`
	Email        string              `json:"email,omitempty"`
	Website      string              `json:"website,omitempty"`
	Active       *bool               `json:"active,omitempty"`
	BankAccounts *[]BankAccountInput `json:"bankAccounts,omitempty"`
}

// ListSuppliers returns paginated suppliers with optional filtering
func (s *SupplierService) ListSuppliers(params repositories.PaginationParams, active *bool) ([]models.Supplier, int64, error) {
	suppliers, total, err := s.supplierRepo.List(params, active)
	if err != nil {
		return nil, 0, &ServiceError{
			Err:     err,
			Message: "Failed to list suppliers",
			Code:    "INTERNAL_ERROR",
		}
	}
	return suppliers, total, nil
}

// GetSupplier returns a single supplier by ID
func (s *SupplierService) GetSupplier(id uint) (*models.Supplier, error) {
	supplier, err := s.supplierRepo.FindByID(id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, &ServiceError{
				Err:     ErrNotFound,
				Message: "Supplier not found",
				Code:    "SUPPLIER_NOT_FOUND",
			}
		}
		return nil, &ServiceError{
			Err:     ErrNotFound,
			Message: "Supplier not found",
			Code:    "SUPPLIER_NOT_FOUND",
		}
	}
	return supplier, nil
}

// CreateSupplier creates a new supplier with validation
func (s *SupplierService) CreateSupplier(input CreateSupplierInput) (*models.Supplier, error) {
	// Validate name
	trimmedName := strings.TrimSpace(input.Name)
	if trimmedName == "" {
		return nil, &ServiceError{
			Err:     ErrValidation,
			Message: "Name is required",
			Code:    "VALIDATION_ERROR",
		}
	}
	if len(trimmedName) > 255 {
		return nil, &ServiceError{
			Err:     ErrValidation,
			Message: "Name must be at most 255 characters",
			Code:    "VALIDATION_ERROR",
		}
	}

	// Validate address
	trimmedAddress := strings.TrimSpace(input.Address)
	if trimmedAddress == "" {
		return nil, &ServiceError{
			Err:     ErrValidation,
			Message: "Address is required",
			Code:    "VALIDATION_ERROR",
		}
	}

	// Validate email (optional, but if provided must be valid)
	if input.Email != "" && !utils.ValidateEmail(input.Email) {
		return nil, &ServiceError{
			Err:     ErrValidation,
			Message: "Invalid email format",
			Code:    "VALIDATION_ERROR",
		}
	}

	// Validate bank accounts
	if err := validateBankAccounts(input.BankAccounts); err != nil {
		return nil, err
	}

	// Build model
	supplier := &models.Supplier{
		Name:    trimmedName,
		Address: trimmedAddress,
		Phone:   strings.TrimSpace(input.Phone),
		Email:   strings.TrimSpace(input.Email),
		Website: strings.TrimSpace(input.Website),
		Active:  true,
	}

	// Convert bank account inputs to models
	for _, ba := range input.BankAccounts {
		supplier.BankAccounts = append(supplier.BankAccounts, models.SupplierBankAccount{
			AccountName:   strings.TrimSpace(ba.AccountName),
			AccountNumber: strings.TrimSpace(ba.AccountNumber),
		})
	}

	if err := s.supplierRepo.Create(supplier); err != nil {
		return nil, &ServiceError{
			Err:     err,
			Message: "Failed to create supplier",
			Code:    "INTERNAL_ERROR",
		}
	}

	// Reload with bank accounts
	created, err := s.supplierRepo.FindByID(supplier.ID)
	if err != nil {
		return supplier, nil
	}

	return created, nil
}

// UpdateSupplier updates an existing supplier
func (s *SupplierService) UpdateSupplier(id uint, input UpdateSupplierInput) (*models.Supplier, error) {
	// Find existing supplier
	supplier, err := s.supplierRepo.FindByID(id)
	if err != nil {
		return nil, &ServiceError{
			Err:     ErrNotFound,
			Message: "Supplier not found",
			Code:    "SUPPLIER_NOT_FOUND",
		}
	}

	// Validate name
	if input.Name != "" {
		trimmedName := strings.TrimSpace(input.Name)
		if len(trimmedName) > 255 {
			return nil, &ServiceError{
				Err:     ErrValidation,
				Message: "Name must be at most 255 characters",
				Code:    "VALIDATION_ERROR",
			}
		}
		supplier.Name = trimmedName
	}

	// Validate address
	if input.Address != "" {
		supplier.Address = strings.TrimSpace(input.Address)
	}

	// Validate email
	if input.Email != "" && !utils.ValidateEmail(input.Email) {
		return nil, &ServiceError{
			Err:     ErrValidation,
			Message: "Invalid email format",
			Code:    "VALIDATION_ERROR",
		}
	}
	if input.Email != "" {
		supplier.Email = strings.TrimSpace(input.Email)
	}

	// Update optional fields
	if input.Phone != "" {
		supplier.Phone = strings.TrimSpace(input.Phone)
	}
	if input.Website != "" {
		supplier.Website = strings.TrimSpace(input.Website)
	}
	if input.Active != nil {
		supplier.Active = *input.Active
	}

	// Handle bank accounts sync
	var bankAccounts []models.SupplierBankAccount
	if input.BankAccounts != nil {
		// Validate bank accounts
		if err := validateBankAccounts(*input.BankAccounts); err != nil {
			return nil, err
		}
		for _, ba := range *input.BankAccounts {
			bankAccounts = append(bankAccounts, models.SupplierBankAccount{
				AccountName:   strings.TrimSpace(ba.AccountName),
				AccountNumber: strings.TrimSpace(ba.AccountNumber),
			})
		}
	}

	if err := s.supplierRepo.Update(supplier, bankAccounts); err != nil {
		return nil, &ServiceError{
			Err:     err,
			Message: "Failed to update supplier",
			Code:    "INTERNAL_ERROR",
		}
	}

	// Reload
	updated, err := s.supplierRepo.FindByID(id)
	if err != nil {
		return supplier, nil
	}

	return updated, nil
}

// DeleteSupplier deletes a supplier with reference checking
func (s *SupplierService) DeleteSupplier(id uint) error {
	// Find supplier
	_, err := s.supplierRepo.FindByID(id)
	if err != nil {
		return &ServiceError{
			Err:     ErrNotFound,
			Message: "Supplier not found",
			Code:    "SUPPLIER_NOT_FOUND",
		}
	}

	// Check purchase order references
	poCount, err := s.supplierRepo.CountPurchaseOrdersBySupplierID(id)
	if err != nil {
		return &ServiceError{
			Err:     err,
			Message: "Failed to check supplier references",
			Code:    "INTERNAL_ERROR",
		}
	}
	if poCount > 0 {
		return &ServiceError{
			Err:     ErrConflict,
			Message: fmt.Sprintf("Cannot delete supplier. It is referenced by %d purchase order(s).", poCount),
			Code:    "SUPPLIER_REFERENCED",
		}
	}

	// Clean up product_suppliers junction (if any)
	if err := s.supplierRepo.CleanupProductSuppliers(id); err != nil {
		return &ServiceError{
			Err:     err,
			Message: "Failed to clean up supplier references",
			Code:    "INTERNAL_ERROR",
		}
	}

	// Delete supplier (CASCADE deletes bank accounts)
	if err := s.supplierRepo.Delete(id); err != nil {
		return &ServiceError{
			Err:     err,
			Message: "Failed to delete supplier",
			Code:    "INTERNAL_ERROR",
		}
	}

	return nil
}

// validateBankAccounts validates bank account inputs
func validateBankAccounts(accounts []BankAccountInput) *ServiceError {
	for i, ba := range accounts {
		if strings.TrimSpace(ba.AccountName) == "" {
			return &ServiceError{
				Err:     ErrValidation,
				Message: fmt.Sprintf("Bank account %d: accountName is required", i+1),
				Code:    "VALIDATION_ERROR",
			}
		}
		if strings.TrimSpace(ba.AccountNumber) == "" {
			return &ServiceError{
				Err:     ErrValidation,
				Message: fmt.Sprintf("Bank account %d: accountNumber is required", i+1),
				Code:    "VALIDATION_ERROR",
			}
		}
	}
	return nil
}
