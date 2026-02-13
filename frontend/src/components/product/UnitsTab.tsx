'use client';

import React, { useState, useCallback, useMemo } from 'react';
import Button from '@/components/ui/Button';
import Input from '@/components/ui/Input';
import Select from '@/components/ui/Select';
import type { ProductUnit } from '@/stores/useProductStore';

interface UnitsTabProps {
  units: ProductUnit[];
  onChange: (units: ProductUnit[]) => void;
  locked?: boolean;
}

interface NewUnitForm {
  name: string;
  conversionFactor: string;
  convertsTo: string;
}

interface EditState {
  unitId: string;
  name: string;
  conversionFactor: string;
  convertsTo: string;
}

const EMPTY_FORM: NewUnitForm = { name: '', conversionFactor: '', convertsTo: '' };

function generateId(): string {
  return `unit_${Date.now()}_${Math.random().toString(36).slice(2, 9)}`;
}

/**
 * Recursively compute the toBaseUnit value for a unit by following
 * its conversion chain. Returns null if a cycle is detected.
 */
function computeToBaseUnit(
  unitId: string,
  unitsMap: Map<string, ProductUnit>,
  visited: Set<string> = new Set(),
): number | null {
  const unit = unitsMap.get(unitId);
  if (!unit) return null;
  if (unit.isBase) return 1;
  if (unit.convertsTo === null) return unit.conversionFactor;
  if (visited.has(unitId)) return null; // cycle detected
  visited.add(unitId);
  const parentValue = computeToBaseUnit(unit.convertsTo, unitsMap, visited);
  if (parentValue === null) return null;
  return unit.conversionFactor * parentValue;
}

/**
 * Check if adding a reference from `fromId` -> `toId` would create a cycle.
 */
function wouldCreateCycle(
  fromId: string,
  toId: string,
  unitsMap: Map<string, ProductUnit>,
): boolean {
  const visited = new Set<string>();
  let current: string | null = toId;
  while (current !== null) {
    if (current === fromId) return true;
    if (visited.has(current)) return false;
    visited.add(current);
    const unit = unitsMap.get(current);
    if (!unit) return false;
    current = unit.convertsTo;
  }
  return false;
}

/**
 * Recalculate toBaseUnit for all units after a change.
 */
function recalculateAll(units: ProductUnit[]): ProductUnit[] {
  const unitsMap = new Map(units.map((u) => [u.id, u]));
  return units.map((u) => {
    if (u.isBase) return { ...u, toBaseUnit: 1 };
    const value = computeToBaseUnit(u.id, unitsMap);
    return { ...u, toBaseUnit: value ?? u.toBaseUnit };
  });
}

/**
 * Get all unit ids that depend (directly or transitively) on the given unit id.
 */
function getDependents(unitId: string, units: ProductUnit[]): string[] {
  const directDeps = units.filter((u) => u.convertsTo === unitId).map((u) => u.id);
  const allDeps: string[] = [...directDeps];
  for (const dep of directDeps) {
    allDeps.push(...getDependents(dep, units));
  }
  return allDeps;
}

export default function UnitsTab({ units, onChange, locked = false }: UnitsTabProps) {
  const [showAddForm, setShowAddForm] = useState(false);
  const [form, setForm] = useState<NewUnitForm>(EMPTY_FORM);
  const [formErrors, setFormErrors] = useState<Partial<Record<keyof NewUnitForm, string>>>({});
  const [deleteError, setDeleteError] = useState<string | null>(null);
  const [editState, setEditState] = useState<EditState | null>(null);
  const [editErrors, setEditErrors] = useState<Partial<Record<string, string>>>({});

  // Base unit prompt state (when no units exist)
  const [baseUnitName, setBaseUnitName] = useState('');
  const [baseUnitError, setBaseUnitError] = useState('');

  const unitsMap = useMemo(() => new Map(units.map((u) => [u.id, u])), [units]);

  const unitOptions = useMemo(
    () => units.map((u) => ({ value: u.id, label: u.name })),
    [units],
  );

  const validateForm = useCallback((): boolean => {
    const errors: Partial<Record<keyof NewUnitForm, string>> = {};
    const trimmedName = form.name.trim();

    if (!trimmedName) {
      errors.name = 'Unit name is required';
    } else if (units.some((u) => u.name.toLowerCase() === trimmedName.toLowerCase())) {
      errors.name = 'Unit name must be unique';
    }

    const factor = Number(form.conversionFactor);
    if (!form.conversionFactor || isNaN(factor)) {
      errors.conversionFactor = 'Conversion factor is required';
    } else if (factor <= 0) {
      errors.conversionFactor = 'Conversion factor must be greater than 0';
    }

    if (!form.convertsTo) {
      errors.convertsTo = 'Select a unit to convert to';
    }

    setFormErrors(errors);
    return Object.keys(errors).length === 0;
  }, [form, units]);

  const handleAdd = useCallback(() => {
    if (!validateForm()) return;

    const newUnit: ProductUnit = {
      id: generateId(),
      name: form.name.trim(),
      conversionFactor: Number(form.conversionFactor),
      convertsTo: form.convertsTo,
      toBaseUnit: 0,
      isBase: false,
    };

    const updated = recalculateAll([...units, newUnit]);
    onChange(updated);
    setForm(EMPTY_FORM);
    setFormErrors({});
    setShowAddForm(false);
  }, [form, units, onChange, validateForm]);

  const handleAddBaseUnit = useCallback(() => {
    const trimmed = baseUnitName.trim();
    if (!trimmed) {
      setBaseUnitError('Unit name is required');
      return;
    }
    const baseUnit: ProductUnit = {
      id: generateId(),
      name: trimmed,
      conversionFactor: 1,
      convertsTo: null,
      toBaseUnit: 1,
      isBase: true,
    };
    onChange([baseUnit]);
    setBaseUnitName('');
    setBaseUnitError('');
  }, [baseUnitName, onChange]);

  const handleDelete = useCallback(
    (unitId: string) => {
      const unit = unitsMap.get(unitId);
      if (!unit || unit.isBase) return;

      const dependents = getDependents(unitId, units);
      if (dependents.length > 0) {
        const depNames = dependents
          .map((id) => unitsMap.get(id)?.name)
          .filter(Boolean)
          .join(', ');
        setDeleteError(
          `Cannot delete "${unit.name}" because other units depend on it: ${depNames}. Remove those units first.`,
        );
        return;
      }

      setDeleteError(null);
      const updated = recalculateAll(units.filter((u) => u.id !== unitId));
      onChange(updated);
    },
    [units, unitsMap, onChange],
  );

  const handleCancelAdd = useCallback(() => {
    setShowAddForm(false);
    setForm(EMPTY_FORM);
    setFormErrors({});
  }, []);

  // --- Edit handlers ---

  const startEdit = useCallback((unit: ProductUnit) => {
    setEditState({
      unitId: unit.id,
      name: unit.name,
      conversionFactor: String(unit.conversionFactor),
      convertsTo: unit.convertsTo ?? '',
    });
    setEditErrors({});
    setDeleteError(null);
  }, []);

  const cancelEdit = useCallback(() => {
    setEditState(null);
    setEditErrors({});
  }, []);

  const saveEdit = useCallback(() => {
    if (!editState) return;
    const unit = unitsMap.get(editState.unitId);
    if (!unit) return;

    const errors: Record<string, string> = {};
    const trimmedName = editState.name.trim();

    if (!trimmedName) {
      errors.name = 'Unit name is required';
    } else if (
      units.some(
        (u) => u.id !== editState.unitId && u.name.toLowerCase() === trimmedName.toLowerCase(),
      )
    ) {
      errors.name = 'Unit name must be unique';
    }

    if (!unit.isBase) {
      const factor = Number(editState.conversionFactor);
      if (!editState.conversionFactor || isNaN(factor)) {
        errors.conversionFactor = 'Conversion factor is required';
      } else if (factor <= 0) {
        errors.conversionFactor = 'Must be greater than 0';
      }

      if (!editState.convertsTo) {
        errors.convertsTo = 'Select a unit to convert to';
      } else if (editState.convertsTo === editState.unitId) {
        errors.convertsTo = 'Cannot reference itself';
      } else if (
        wouldCreateCycle(
          editState.unitId,
          editState.convertsTo,
          new Map(
            units.map((u) =>
              u.id === editState.unitId
                ? [u.id, { ...u, convertsTo: editState.convertsTo }]
                : [u.id, u],
            ),
          ),
        )
      ) {
        errors.convertsTo = 'Would create a circular reference';
      }
    }

    if (Object.keys(errors).length > 0) {
      setEditErrors(errors);
      return;
    }

    const updated = units.map((u) => {
      if (u.id !== editState.unitId) return u;
      if (u.isBase) {
        return { ...u, name: trimmedName };
      }
      return {
        ...u,
        name: trimmedName,
        conversionFactor: Number(editState.conversionFactor),
        convertsTo: editState.convertsTo,
      };
    });

    onChange(recalculateAll(updated));
    setEditState(null);
    setEditErrors({});
  }, [editState, units, unitsMap, onChange]);

  const getConversionDisplay = (unit: ProductUnit): string => {
    if (unit.isBase) return 'Base Unit';
    const target = unit.convertsTo ? unitsMap.get(unit.convertsTo) : null;
    if (!target) return `${unit.conversionFactor}`;
    return `1 = ${unit.conversionFactor} × ${target.name}`;
  };

  // Preview text for the add form
  const formPreview = useMemo(() => {
    const target = form.convertsTo ? unitsMap.get(form.convertsTo) : null;
    const factor = form.conversionFactor;
    const name = form.name.trim() || '?';
    const targetName = target?.name || '?';
    if (!factor || !target) return null;
    return `1 ${name} = ${factor} × ${targetName}`;
  }, [form, unitsMap]);

  // Edit options: exclude the unit being edited
  const editUnitOptions = useMemo(() => {
    if (!editState) return [];
    return units
      .filter((u) => u.id !== editState.unitId)
      .map((u) => ({ value: u.id, label: u.name }));
  }, [units, editState]);

  // Pencil icon SVG
  const PencilIcon = () => (
    <svg className="h-4 w-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
      <path
        strokeLinecap="round"
        strokeLinejoin="round"
        strokeWidth={2}
        d="M15.232 5.232l3.536 3.536m-2.036-5.036a2.5 2.5 0 113.536 3.536L6.5 21.036H3v-3.572L16.732 3.732z"
      />
    </svg>
  );

  // Trash icon SVG
  const TrashIcon = () => (
    <svg className="h-4 w-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
      <path
        strokeLinecap="round"
        strokeLinejoin="round"
        strokeWidth={2}
        d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16"
      />
    </svg>
  );

  // --- No units: show "Define base unit" prompt ---
  if (units.length === 0 && !locked) {
    return (
      <div className="space-y-4">
        <div className="rounded-lg border-2 border-dashed border-gray-300 p-8 text-center">
          <p className="text-sm text-gray-500 mb-4">
            No units defined yet. Define the base unit for this product (e.g., Pcs, Kg, Liter, Meter).
          </p>
          <div className="max-w-xs mx-auto space-y-3">
            <Input
              label="Base Unit Name"
              placeholder="e.g. Pcs, Kg, Liter"
              value={baseUnitName}
              onChange={(e) => {
                setBaseUnitName(e.target.value);
                if (baseUnitError) setBaseUnitError('');
              }}
              error={baseUnitError}
            />
            <Button onClick={handleAddBaseUnit}>Add Base Unit</Button>
          </div>
        </div>
      </div>
    );
  }

  return (
    <div className="space-y-4">
      {/* Lock banner */}
      {locked && (
        <div className="flex items-center gap-2 rounded-md border border-amber-300 bg-amber-50 px-4 py-3 text-sm text-amber-800">
          <svg
            className="h-4 w-4 flex-shrink-0"
            fill="none"
            stroke="currentColor"
            viewBox="0 0 24 24"
          >
            <path
              strokeLinecap="round"
              strokeLinejoin="round"
              strokeWidth={2}
              d="M12 15v2m-6 4h12a2 2 0 002-2v-6a2 2 0 00-2-2H6a2 2 0 00-2 2v6a2 2 0 002 2zm10-10V7a4 4 0 00-8 0v4h8z"
            />
          </svg>
          Units cannot be modified while stock exists.
        </div>
      )}

      {/* Delete error */}
      {deleteError && (
        <div className="flex items-start gap-2 rounded-md border border-red-300 bg-red-50 px-4 py-3 text-sm text-red-800">
          <svg
            className="mt-0.5 h-4 w-4 flex-shrink-0"
            fill="none"
            stroke="currentColor"
            viewBox="0 0 24 24"
          >
            <path
              strokeLinecap="round"
              strokeLinejoin="round"
              strokeWidth={2}
              d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"
            />
          </svg>
          <span>{deleteError}</span>
          <button
            onClick={() => setDeleteError(null)}
            className="ml-auto text-red-600 hover:text-red-800 cursor-pointer"
          >
            <svg className="h-4 w-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                strokeWidth={2}
                d="M6 18L18 6M6 6l12 12"
              />
            </svg>
          </button>
        </div>
      )}

      {/* Units table */}
      <div className="overflow-x-auto">
        <table className="w-full text-sm text-left">
          <thead className="bg-gray-50 text-gray-600 uppercase text-xs">
            <tr>
              <th className="px-4 py-3 font-medium">Unit</th>
              <th className="px-4 py-3 font-medium">Conversion</th>
              <th className="px-4 py-3 font-medium">= Base Unit</th>
              <th className="px-4 py-3 font-medium w-28">Actions</th>
            </tr>
          </thead>
          <tbody className="divide-y divide-gray-200">
            {units.map((unit) => (
              <React.Fragment key={unit.id}>
                {editState?.unitId === unit.id ? (
                  /* --- Inline edit row --- */
                  <tr className="bg-blue-50">
                    <td className="px-4 py-3" colSpan={unit.isBase ? 2 : 1}>
                      <Input
                        placeholder="Unit name"
                        value={editState.name}
                        onChange={(e) =>
                          setEditState((s) => s && { ...s, name: e.target.value })
                        }
                        error={editErrors.name}
                      />
                    </td>
                    {!unit.isBase && (
                      <td className="px-4 py-3" colSpan={1}>
                        <div className="flex items-center gap-2">
                          <span className="text-gray-500 text-xs whitespace-nowrap">1 =</span>
                          <Input
                            type="number"
                            placeholder="Factor"
                            min="0"
                            step="any"
                            value={editState.conversionFactor}
                            onChange={(e) =>
                              setEditState((s) => s && { ...s, conversionFactor: e.target.value })
                            }
                            error={editErrors.conversionFactor}
                          />
                          <span className="text-gray-500 text-xs">×</span>
                          <Select
                            placeholder="Unit..."
                            options={editUnitOptions}
                            value={editState.convertsTo}
                            onChange={(e) =>
                              setEditState((s) => s && { ...s, convertsTo: e.target.value })
                            }
                            error={editErrors.convertsTo}
                          />
                        </div>
                      </td>
                    )}
                    <td className="px-4 py-3 text-gray-400 font-mono text-center">—</td>
                    <td className="px-4 py-3">
                      <div className="flex gap-1">
                        <Button size="sm" onClick={saveEdit}>
                          Save
                        </Button>
                        <Button variant="outline" size="sm" onClick={cancelEdit}>
                          Cancel
                        </Button>
                      </div>
                    </td>
                  </tr>
                ) : (
                  /* --- Display row --- */
                  <tr className="hover:bg-gray-50">
                    <td className="px-4 py-3 text-gray-700">
                      <span className="font-medium">{unit.name}</span>
                      {unit.isBase && (
                        <span className="ml-2 inline-flex items-center rounded-full bg-blue-100 px-2 py-0.5 text-xs font-medium text-blue-700">
                          Base Unit
                        </span>
                      )}
                    </td>
                    <td className="px-4 py-3 text-gray-600">{getConversionDisplay(unit)}</td>
                    <td className="px-4 py-3 text-gray-700 font-mono">{unit.toBaseUnit}</td>
                    <td className="px-4 py-3">
                      {!locked && (
                        <div className="flex gap-1">
                          <button
                            onClick={() => startEdit(unit)}
                            className="text-gray-500 hover:text-blue-600 cursor-pointer p-1 rounded hover:bg-blue-50"
                            title={unit.isBase ? 'Rename base unit' : 'Edit unit'}
                          >
                            <PencilIcon />
                          </button>
                          {!unit.isBase && (
                            <button
                              onClick={() => handleDelete(unit.id)}
                              className="text-red-500 hover:text-red-700 cursor-pointer p-1 rounded hover:bg-red-50"
                              title="Delete unit"
                            >
                              <TrashIcon />
                            </button>
                          )}
                        </div>
                      )}
                    </td>
                  </tr>
                )}
              </React.Fragment>
            ))}
          </tbody>
        </table>
      </div>

      {/* Inline add form */}
      {showAddForm && !locked && (
        <div className="rounded-md border border-gray-200 bg-gray-50 p-4 space-y-3">
          <h4 className="text-sm font-medium text-gray-700">Add Unit</h4>
          <div className="grid grid-cols-1 sm:grid-cols-3 gap-3">
            <Input
              label="Unit Name"
              placeholder="e.g. Dozen, Box, Case"
              value={form.name}
              onChange={(e) => {
                setForm((f) => ({ ...f, name: e.target.value }));
                if (formErrors.name) setFormErrors((err) => ({ ...err, name: undefined }));
              }}
              error={formErrors.name}
            />
            <Input
              label="Conversion Factor"
              type="number"
              placeholder="e.g. 12"
              min="0"
              step="any"
              value={form.conversionFactor}
              onChange={(e) => {
                setForm((f) => ({ ...f, conversionFactor: e.target.value }));
                if (formErrors.conversionFactor)
                  setFormErrors((err) => ({ ...err, conversionFactor: undefined }));
              }}
              error={formErrors.conversionFactor}
            />
            <Select
              label="Converts To"
              placeholder="Select unit..."
              options={unitOptions}
              value={form.convertsTo}
              onChange={(e) => {
                setForm((f) => ({ ...f, convertsTo: e.target.value }));
                if (formErrors.convertsTo)
                  setFormErrors((err) => ({ ...err, convertsTo: undefined }));
              }}
              error={formErrors.convertsTo}
            />
          </div>

          {/* Preview */}
          {formPreview && (
            <p className="text-sm text-gray-500 italic">{formPreview}</p>
          )}

          <div className="flex gap-2">
            <Button size="sm" onClick={handleAdd}>
              Add Unit
            </Button>
            <Button variant="outline" size="sm" onClick={handleCancelAdd}>
              Cancel
            </Button>
          </div>
        </div>
      )}

      {/* Add button */}
      {!showAddForm && !locked && units.length > 0 && (
        <Button
          variant="outline"
          size="sm"
          onClick={() => {
            setShowAddForm(true);
            setDeleteError(null);
            setEditState(null);
          }}
        >
          + Add Unit
        </Button>
      )}
    </div>
  );
}
