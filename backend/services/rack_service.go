package services

import (
	"strings"

	"github.com/pointofsale/backend/models"
	"github.com/pointofsale/backend/repositories"
	"gorm.io/gorm"
)

// RackInput is the DTO for creating and updating racks
type RackInput struct {
	Name        string `json:"name"`
	Code        string `json:"code"`
	Location    string `json:"location"`
	Capacity    int    `json:"capacity"`
	Description string `json:"description"`
	Active      *bool  `json:"active"`
}

// RackServiceRepository extends the base RackRepository with cleanup operations
type RackServiceRepository interface {
	repositories.RackRepository
	CleanupVariantRacks(rackID uint) error
}

// RackService handles rack business logic
type RackService struct {
	rackRepo RackServiceRepository
}

// NewRackService creates a new rack service instance
func NewRackService(rackRepo RackServiceRepository) *RackService {
	return &RackService{rackRepo: rackRepo}
}

// ListRacks returns paginated racks
func (s *RackService) ListRacks(page, pageSize int, search, active, sortBy, sortDir string) ([]models.Rack, int64, *ServiceError) {
	racks, total, err := s.rackRepo.List(page, pageSize, search, active, sortBy, sortDir)
	if err != nil {
		return nil, 0, &ServiceError{
			Err:     err,
			Message: "Failed to list racks",
			Code:    "INTERNAL_ERROR",
		}
	}
	return racks, total, nil
}

// GetRack returns a rack by ID
func (s *RackService) GetRack(id uint) (*models.Rack, *ServiceError) {
	rack, err := s.rackRepo.FindByID(id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, &ServiceError{
				Err:     ErrNotFound,
				Message: "Rack not found",
				Code:    "RACK_NOT_FOUND",
			}
		}
		return nil, &ServiceError{
			Err:     err,
			Message: "Failed to get rack",
			Code:    "INTERNAL_ERROR",
		}
	}
	return rack, nil
}

// CreateRack creates a new rack with validation
func (s *RackService) CreateRack(input RackInput) (*models.Rack, *ServiceError) {
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

	// Validate code
	trimmedCode := strings.TrimSpace(input.Code)
	if trimmedCode == "" {
		return nil, &ServiceError{
			Err:     ErrValidation,
			Message: "Code is required",
			Code:    "VALIDATION_ERROR",
		}
	}
	if len(trimmedCode) > 50 {
		return nil, &ServiceError{
			Err:     ErrValidation,
			Message: "Code must be at most 50 characters",
			Code:    "VALIDATION_ERROR",
		}
	}

	// Validate location
	trimmedLocation := strings.TrimSpace(input.Location)
	if trimmedLocation == "" {
		return nil, &ServiceError{
			Err:     ErrValidation,
			Message: "Location is required",
			Code:    "VALIDATION_ERROR",
		}
	}
	if len(trimmedLocation) > 255 {
		return nil, &ServiceError{
			Err:     ErrValidation,
			Message: "Location must be at most 255 characters",
			Code:    "VALIDATION_ERROR",
		}
	}

	// Validate capacity
	if input.Capacity <= 0 {
		return nil, &ServiceError{
			Err:     ErrValidation,
			Message: "Capacity must be greater than 0",
			Code:    "VALIDATION_ERROR",
		}
	}

	// Check code uniqueness
	existing, _ := s.rackRepo.FindByCode(trimmedCode)
	if existing != nil {
		return nil, &ServiceError{
			Err:     ErrConflict,
			Message: "Rack code already exists",
			Code:    "RACK_CODE_EXISTS",
		}
	}

	// Create rack
	rack := &models.Rack{
		Name:        trimmedName,
		Code:        trimmedCode,
		Location:    trimmedLocation,
		Capacity:    input.Capacity,
		Description: strings.TrimSpace(input.Description),
		Active:      true,
	}

	if err := s.rackRepo.Create(rack); err != nil {
		return nil, &ServiceError{
			Err:     err,
			Message: "Failed to create rack",
			Code:    "INTERNAL_ERROR",
		}
	}

	return rack, nil
}

// UpdateRack updates an existing rack
func (s *RackService) UpdateRack(id uint, input RackInput) (*models.Rack, *ServiceError) {
	// Find rack
	rack, err := s.rackRepo.FindByID(id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, &ServiceError{
				Err:     ErrNotFound,
				Message: "Rack not found",
				Code:    "RACK_NOT_FOUND",
			}
		}
		return nil, &ServiceError{
			Err:     err,
			Message: "Failed to get rack",
			Code:    "INTERNAL_ERROR",
		}
	}

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

	// Validate code
	trimmedCode := strings.TrimSpace(input.Code)
	if trimmedCode == "" {
		return nil, &ServiceError{
			Err:     ErrValidation,
			Message: "Code is required",
			Code:    "VALIDATION_ERROR",
		}
	}
	if len(trimmedCode) > 50 {
		return nil, &ServiceError{
			Err:     ErrValidation,
			Message: "Code must be at most 50 characters",
			Code:    "VALIDATION_ERROR",
		}
	}

	// Validate location
	trimmedLocation := strings.TrimSpace(input.Location)
	if trimmedLocation == "" {
		return nil, &ServiceError{
			Err:     ErrValidation,
			Message: "Location is required",
			Code:    "VALIDATION_ERROR",
		}
	}
	if len(trimmedLocation) > 255 {
		return nil, &ServiceError{
			Err:     ErrValidation,
			Message: "Location must be at most 255 characters",
			Code:    "VALIDATION_ERROR",
		}
	}

	// Validate capacity
	if input.Capacity <= 0 {
		return nil, &ServiceError{
			Err:     ErrValidation,
			Message: "Capacity must be greater than 0",
			Code:    "VALIDATION_ERROR",
		}
	}

	// Check code uniqueness excluding self
	existing, _ := s.rackRepo.FindByCodeExcluding(trimmedCode, id)
	if existing != nil {
		return nil, &ServiceError{
			Err:     ErrConflict,
			Message: "Rack code already exists",
			Code:    "RACK_CODE_EXISTS",
		}
	}

	// Update fields
	rack.Name = trimmedName
	rack.Code = trimmedCode
	rack.Location = trimmedLocation
	rack.Capacity = input.Capacity
	rack.Description = strings.TrimSpace(input.Description)
	if input.Active != nil {
		rack.Active = *input.Active
	}

	if err := s.rackRepo.Update(rack); err != nil {
		return nil, &ServiceError{
			Err:     err,
			Message: "Failed to update rack",
			Code:    "INTERNAL_ERROR",
		}
	}

	return rack, nil
}

// DeleteRack deletes a rack by ID, cleaning up variant_racks junction entries
func (s *RackService) DeleteRack(id uint) *ServiceError {
	// Find rack
	_, err := s.rackRepo.FindByID(id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return &ServiceError{
				Err:     ErrNotFound,
				Message: "Rack not found",
				Code:    "RACK_NOT_FOUND",
			}
		}
		return &ServiceError{
			Err:     err,
			Message: "Failed to get rack",
			Code:    "INTERNAL_ERROR",
		}
	}

	// Clean up variant_racks junction entries
	if err := s.rackRepo.CleanupVariantRacks(id); err != nil {
		return &ServiceError{
			Err:     err,
			Message: "Failed to clean up variant rack associations",
			Code:    "INTERNAL_ERROR",
		}
	}

	// Delete rack
	if err := s.rackRepo.Delete(id); err != nil {
		return &ServiceError{
			Err:     err,
			Message: "Failed to delete rack",
			Code:    "INTERNAL_ERROR",
		}
	}

	return nil
}
