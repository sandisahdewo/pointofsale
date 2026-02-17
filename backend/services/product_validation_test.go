package services

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func validProductInput() CreateProductInput {
	return CreateProductInput{
		Name:         "Rice",
		Description:  "Premium rice",
		CategoryID:   1,
		PriceSetting: "fixed",
		HasVariants:  false,
		Status:       "active",
		Units: []CreateProductUnitInput{
			{Name: "Kg", IsBase: true},
		},
		Variants: []CreateProductVariantInput{
			{
				SKU:        "RC-001",
				Attributes: []CreateVariantAttributeInput{},
				PricingTiers: []CreateVariantPricingTierInput{
					{MinQty: 1, Value: 15000},
				},
			},
		},
	}
}

func TestValidateProduct_ValidMinimal_ReturnsNil(t *testing.T) {
	input := validProductInput()

	err := ValidateProductInput(input)
	require.NoError(t, err)
}

func TestValidateProduct_MissingName_ReturnsError(t *testing.T) {
	input := validProductInput()
	input.Name = ""

	err := ValidateProductInput(input)
	require.Error(t, err)
	assert.ErrorContains(t, err, "name is required")
}

func TestValidateProduct_MissingCategory_ReturnsError(t *testing.T) {
	input := validProductInput()
	input.CategoryID = 0

	err := ValidateProductInput(input)
	require.Error(t, err)
	assert.ErrorContains(t, err, "categoryId is required")
}

func TestValidateProduct_MarkupWithoutMarkupType_ReturnsError(t *testing.T) {
	input := validProductInput()
	input.PriceSetting = "markup"
	input.MarkupType = nil

	err := ValidateProductInput(input)
	require.Error(t, err)
	assert.ErrorContains(t, err, "markupType is required when priceSetting is markup")
}

func TestValidateProduct_FixedWithMarkupType_ReturnsError(t *testing.T) {
	input := validProductInput()
	input.PriceSetting = "fixed"
	markupType := "percentage"
	input.MarkupType = &markupType

	err := ValidateProductInput(input)
	require.Error(t, err)
	assert.ErrorContains(t, err, "markupType must be empty when priceSetting is fixed")
}

func TestValidateProduct_NoBaseUnit_ReturnsError(t *testing.T) {
	input := validProductInput()
	input.Units = []CreateProductUnitInput{
		{Name: "Kg", IsBase: false, ConversionFactor: 1, ConvertsToName: "Gram"},
	}

	err := ValidateProductInput(input)
	require.Error(t, err)
	assert.ErrorContains(t, err, "exactly one base unit is required")
}

func TestValidateProduct_MultipleBaseUnits_ReturnsError(t *testing.T) {
	input := validProductInput()
	input.Units = []CreateProductUnitInput{
		{Name: "Kg", IsBase: true},
		{Name: "Liter", IsBase: true},
	}

	err := ValidateProductInput(input)
	require.Error(t, err)
	assert.ErrorContains(t, err, "exactly one base unit is required")
}

func TestValidateProduct_DuplicateUnitNames_ReturnsError(t *testing.T) {
	input := validProductInput()
	input.Units = []CreateProductUnitInput{
		{Name: "Kg", IsBase: true},
		{Name: "kg", IsBase: false, ConversionFactor: 2, ConvertsToName: "Kg"},
	}

	err := ValidateProductInput(input)
	require.Error(t, err)
	assert.ErrorContains(t, err, "unit name must be unique")
}

func TestValidateProduct_NoVariants_ReturnsError(t *testing.T) {
	input := validProductInput()
	input.Variants = nil

	err := ValidateProductInput(input)
	require.Error(t, err)
	assert.ErrorContains(t, err, "at least one variant is required")
}

func TestValidateProduct_HasVariantsFalseMultipleVariants_ReturnsError(t *testing.T) {
	input := validProductInput()
	input.HasVariants = false
	input.Variants = []CreateProductVariantInput{
		{
			SKU:          "RC-001",
			PricingTiers: []CreateVariantPricingTierInput{{MinQty: 1, Value: 15000}},
		},
		{
			SKU:          "RC-002",
			PricingTiers: []CreateVariantPricingTierInput{{MinQty: 1, Value: 16000}},
		},
	}

	err := ValidateProductInput(input)
	require.Error(t, err)
	assert.ErrorContains(t, err, "hasVariants=false requires exactly one variant")
}

func TestValidateProduct_DuplicateSKU_ReturnsError(t *testing.T) {
	input := validProductInput()
	input.HasVariants = true
	input.Variants = []CreateProductVariantInput{
		{
			SKU:          "TS-RED-S",
			Attributes:   []CreateVariantAttributeInput{{AttributeName: "Color", AttributeValue: "Red"}},
			PricingTiers: []CreateVariantPricingTierInput{{MinQty: 1, Value: 75000}},
		},
		{
			SKU:          "TS-RED-S",
			Attributes:   []CreateVariantAttributeInput{{AttributeName: "Color", AttributeValue: "Blue"}},
			PricingTiers: []CreateVariantPricingTierInput{{MinQty: 1, Value: 76000}},
		},
	}

	err := ValidateProductInput(input)
	require.Error(t, err)
	assert.ErrorContains(t, err, "duplicate sku")
}

func TestValidateProduct_PricingTiersMissingMinQty1_ReturnsError(t *testing.T) {
	input := validProductInput()
	input.Variants[0].PricingTiers = []CreateVariantPricingTierInput{
		{MinQty: 2, Value: 15000},
	}

	err := ValidateProductInput(input)
	require.Error(t, err)
	assert.ErrorContains(t, err, "first pricing tier minQty must be 1")
}

func TestValidateProduct_PricingTiersNotAscending_ReturnsError(t *testing.T) {
	input := validProductInput()
	input.Variants[0].PricingTiers = []CreateVariantPricingTierInput{
		{MinQty: 1, Value: 15000},
		{MinQty: 10, Value: 14000},
		{MinQty: 5, Value: 13000},
	}

	err := ValidateProductInput(input)
	require.Error(t, err)
	assert.ErrorContains(t, err, "pricing tiers must be sorted by minQty ascending")
}
