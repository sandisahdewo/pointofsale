package services

import (
	"fmt"
	"strings"
)

// CalculateToBaseUnit computes a unit's total conversion value to the base unit.
func CalculateToBaseUnit(unitName string, unitsByName map[string]CreateProductUnitInput) (float64, error) {
	visited := make(map[string]bool)
	return calculateToBaseRecursive(unitName, unitsByName, visited)
}

func calculateToBaseRecursive(unitName string, unitsByName map[string]CreateProductUnitInput, visited map[string]bool) (float64, error) {
	key := strings.ToLower(strings.TrimSpace(unitName))
	unit, ok := findUnitByName(unitsByName, key)
	if !ok {
		return 0, fmt.Errorf("unit %q not found", unitName)
	}

	if unit.IsBase {
		return 1, nil
	}

	if visited[key] {
		return 0, fmt.Errorf("circular unit reference detected")
	}
	visited[key] = true

	if strings.TrimSpace(unit.ConvertsToName) == "" {
		return 0, fmt.Errorf("unit %q must define convertsToName", unit.Name)
	}
	if unit.ConversionFactor <= 0 {
		return 0, fmt.Errorf("unit %q conversionFactor must be greater than 0", unit.Name)
	}

	parentValue, err := calculateToBaseRecursive(unit.ConvertsToName, unitsByName, visited)
	if err != nil {
		return 0, err
	}
	return unit.ConversionFactor * parentValue, nil
}

func findUnitByName(unitsByName map[string]CreateProductUnitInput, lookupKey string) (CreateProductUnitInput, bool) {
	if unit, ok := unitsByName[lookupKey]; ok {
		return unit, true
	}

	for key, unit := range unitsByName {
		if strings.ToLower(strings.TrimSpace(key)) == lookupKey {
			return unit, true
		}
		if strings.ToLower(strings.TrimSpace(unit.Name)) == lookupKey {
			return unit, true
		}
	}

	return CreateProductUnitInput{}, false
}

// ValidateUnitCircularRef validates there are no circular conversion chains.
func ValidateUnitCircularRef(units []CreateProductUnitInput) error {
	unitsByName := make(map[string]CreateProductUnitInput, len(units))
	for _, unit := range units {
		unitsByName[strings.ToLower(strings.TrimSpace(unit.Name))] = unit
	}

	for _, unit := range units {
		visited := make(map[string]bool)
		_, err := calculateToBaseRecursive(unit.Name, unitsByName, visited)
		if err != nil && strings.Contains(strings.ToLower(err.Error()), "circular") {
			return err
		}
		if err != nil && strings.Contains(strings.ToLower(err.Error()), "not found") {
			return err
		}
	}
	return nil
}

// ResolveUnitDependencyOrder returns units sorted so each parent appears before dependents.
func ResolveUnitDependencyOrder(units []CreateProductUnitInput) ([]CreateProductUnitInput, error) {
	unitsByName := make(map[string]CreateProductUnitInput, len(units))
	for _, unit := range units {
		unitsByName[strings.ToLower(strings.TrimSpace(unit.Name))] = unit
	}

	visited := make(map[string]bool)
	visiting := make(map[string]bool)
	ordered := make([]CreateProductUnitInput, 0, len(units))

	var visit func(string) error
	visit = func(name string) error {
		key := strings.ToLower(strings.TrimSpace(name))
		if visited[key] {
			return nil
		}
		if visiting[key] {
			return fmt.Errorf("circular unit reference detected")
		}

		unit, ok := unitsByName[key]
		if !ok {
			return fmt.Errorf("unit %q not found", name)
		}

		visiting[key] = true
		if !unit.IsBase && strings.TrimSpace(unit.ConvertsToName) != "" {
			if err := visit(unit.ConvertsToName); err != nil {
				return err
			}
		}
		visiting[key] = false
		visited[key] = true
		ordered = append(ordered, unit)
		return nil
	}

	for _, unit := range units {
		if err := visit(unit.Name); err != nil {
			return nil, err
		}
	}

	return ordered, nil
}
