package services

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCalculateTieredPrice_SingleTier_ReturnsBasePrice(t *testing.T) {
	tiers := []PricingTier{{MinQty: 1, Value: 75000}}
	price, err := CalculateTieredPrice(tiers, 5, 1) // 5 pcs, toBaseUnit=1
	require.NoError(t, err)
	assert.Equal(t, 75000.0, price)
}

func TestCalculateTieredPrice_QtyMatchesSecondTier_ReturnsSecondTierPrice(t *testing.T) {
	tiers := []PricingTier{
		{MinQty: 1, Value: 75000},
		{MinQty: 12, Value: 70000},
	}
	price, err := CalculateTieredPrice(tiers, 12, 1) // exactly 12 pcs
	require.NoError(t, err)
	assert.Equal(t, 70000.0, price)
}

func TestCalculateTieredPrice_QtyBetweenTiers_UsesLowerTier(t *testing.T) {
	tiers := []PricingTier{
		{MinQty: 1, Value: 75000},
		{MinQty: 12, Value: 70000},
	}
	price, err := CalculateTieredPrice(tiers, 5, 1) // 5 pcs, below second tier
	require.NoError(t, err)
	assert.Equal(t, 75000.0, price)
}

func TestCalculateTieredPrice_LargeQty_UsesHighestTier(t *testing.T) {
	tiers := []PricingTier{
		{MinQty: 1, Value: 75000},
		{MinQty: 12, Value: 70000},
		{MinQty: 144, Value: 65000},
	}
	price, err := CalculateTieredPrice(tiers, 500, 1) // 500 pcs
	require.NoError(t, err)
	assert.Equal(t, 65000.0, price)
}

func TestCalculateTieredPrice_WithUnitConversion_ConvertsToBaseFirst(t *testing.T) {
	tiers := []PricingTier{
		{MinQty: 1, Value: 75000},
		{MinQty: 12, Value: 70000},
	}
	// 1 Dozen = 12 pcs, so baseQty = 1 * 12 = 12, matches second tier
	price, err := CalculateTieredPrice(tiers, 1, 12) // 1 unit, toBaseUnit=12
	require.NoError(t, err)
	assert.Equal(t, 70000.0, price) // per-base-unit price
}

func TestCalculateTieredPrice_EmptyTiers_ReturnsError(t *testing.T) {
	tiers := []PricingTier{}
	_, err := CalculateTieredPrice(tiers, 5, 1)
	assert.Error(t, err)
}
