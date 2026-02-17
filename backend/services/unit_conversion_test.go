package services

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCalculateToBaseUnit_BaseUnit_Returns1(t *testing.T) {
	units := map[string]CreateProductUnitInput{
		"Pcs": {Name: "Pcs", IsBase: true},
	}

	value, err := CalculateToBaseUnit("Pcs", units)
	require.NoError(t, err)
	assert.Equal(t, float64(1), value)
}

func TestCalculateToBaseUnit_DirectReference_ReturnsCorrectValue(t *testing.T) {
	units := map[string]CreateProductUnitInput{
		"Pcs":   {Name: "Pcs", IsBase: true},
		"Dozen": {Name: "Dozen", ConversionFactor: 12, ConvertsToName: "Pcs"},
	}

	value, err := CalculateToBaseUnit("Dozen", units)
	require.NoError(t, err)
	assert.Equal(t, float64(12), value)
}

func TestCalculateToBaseUnit_ChainedReference_MultipliesCorrectly(t *testing.T) {
	units := map[string]CreateProductUnitInput{
		"Pcs":   {Name: "Pcs", IsBase: true},
		"Dozen": {Name: "Dozen", ConversionFactor: 12, ConvertsToName: "Pcs"},
		"Box":   {Name: "Box", ConversionFactor: 12, ConvertsToName: "Dozen"},
	}

	value, err := CalculateToBaseUnit("Box", units)
	require.NoError(t, err)
	assert.Equal(t, float64(144), value)
}

func TestCalculateToBaseUnit_BranchingStructure_CalculatesIndependently(t *testing.T) {
	units := map[string]CreateProductUnitInput{
		"Kg":     {Name: "Kg", IsBase: true},
		"Karung": {Name: "Karung", ConversionFactor: 50, ConvertsToName: "Kg"},
		"Bag":    {Name: "Bag", ConversionFactor: 25, ConvertsToName: "Kg"},
	}

	karungValue, err := CalculateToBaseUnit("Karung", units)
	require.NoError(t, err)
	assert.Equal(t, float64(50), karungValue)

	bagValue, err := CalculateToBaseUnit("Bag", units)
	require.NoError(t, err)
	assert.Equal(t, float64(25), bagValue)
}

func TestValidateUnitCircularRef_NoCircle_ReturnsNil(t *testing.T) {
	units := []CreateProductUnitInput{
		{Name: "Pcs", IsBase: true},
		{Name: "Dozen", ConversionFactor: 12, ConvertsToName: "Pcs"},
		{Name: "Box", ConversionFactor: 12, ConvertsToName: "Dozen"},
	}

	err := ValidateUnitCircularRef(units)
	require.NoError(t, err)
}

func TestValidateUnitCircularRef_SelfReference_ReturnsError(t *testing.T) {
	units := []CreateProductUnitInput{
		{Name: "Pcs", IsBase: true},
		{Name: "Dozen", ConversionFactor: 12, ConvertsToName: "Dozen"},
	}

	err := ValidateUnitCircularRef(units)
	require.Error(t, err)
	assert.ErrorContains(t, err, "circular")
}

func TestValidateUnitCircularRef_IndirectCircle_ReturnsError(t *testing.T) {
	units := []CreateProductUnitInput{
		{Name: "Pcs", IsBase: true},
		{Name: "Dozen", ConversionFactor: 12, ConvertsToName: "Box"},
		{Name: "Box", ConversionFactor: 12, ConvertsToName: "Dozen"},
	}

	err := ValidateUnitCircularRef(units)
	require.Error(t, err)
	assert.ErrorContains(t, err, "circular")
}

func TestResolveUnitDependencyOrder_ReturnsBaseFirst(t *testing.T) {
	units := []CreateProductUnitInput{
		{Name: "Box", ConversionFactor: 12, ConvertsToName: "Dozen"},
		{Name: "Pcs", IsBase: true},
		{Name: "Dozen", ConversionFactor: 12, ConvertsToName: "Pcs"},
	}

	ordered, err := ResolveUnitDependencyOrder(units)
	require.NoError(t, err)
	require.Len(t, ordered, 3)

	assert.Equal(t, "Pcs", ordered[0].Name)
	assert.Equal(t, "Dozen", ordered[1].Name)
	assert.Equal(t, "Box", ordered[2].Name)
}
