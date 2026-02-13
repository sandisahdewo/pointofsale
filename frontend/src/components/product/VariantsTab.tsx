'use client';

import React, { useState, useCallback } from 'react';
import Input from '@/components/ui/Input';
import Button from '@/components/ui/Button';
import Select from '@/components/ui/Select';
import MultiSelect from '@/components/ui/MultiSelect';
import TagInput from '@/components/ui/TagInput';
import ImageUpload from '@/components/ui/ImageUpload';
import ConfirmModal from '@/components/ui/ConfirmModal';
import VariantPricing from '@/components/product/VariantPricing';
import { useRackStore } from '@/stores/useRackStore';
import type {
  PriceSetting,
  MarkupType,
  VariantAttribute,
  ProductVariant,
} from '@/stores/useProductStore';

interface VariantsTabProps {
  hasVariants: boolean;
  priceSetting: PriceSetting;
  markupType?: MarkupType;
  variantAttributes: VariantAttribute[];
  variants: ProductVariant[];
  onAttributesChange: (attrs: VariantAttribute[]) => void;
  onVariantsChange: (variants: ProductVariant[]) => void;
}

// ---- helpers ----

let variantIdCounter = 0;
function nextVariantId(): string {
  variantIdCounter += 1;
  return `var_${Date.now()}_${variantIdCounter}`;
}

function cartesian(attrs: VariantAttribute[]): Record<string, string>[] {
  if (attrs.length === 0) return [];
  const filtered = attrs.filter((a) => a.name.trim() && a.values.length > 0);
  if (filtered.length === 0) return [];

  return filtered.reduce<Record<string, string>[]>(
    (acc, attr) => {
      const result: Record<string, string>[] = [];
      for (const combo of acc) {
        for (const val of attr.values) {
          result.push({ ...combo, [attr.name]: val });
        }
      }
      return result;
    },
    [{}]
  );
}

function duplicateValues(variants: ProductVariant[], field: 'sku' | 'barcode'): Set<string> {
  const seen = new Map<string, number>();
  const dupes = new Set<string>();
  for (const v of variants) {
    const val = v[field].trim();
    if (!val) continue;
    seen.set(val, (seen.get(val) ?? 0) + 1);
    if ((seen.get(val) ?? 0) > 1) dupes.add(val);
  }
  return dupes;
}

// ---- Component ----

export default function VariantsTab({
  hasVariants,
  priceSetting,
  markupType,
  variantAttributes,
  variants,
  onAttributesChange,
  onVariantsChange,
}: VariantsTabProps) {
  const { getActiveRacks } = useRackStore();
  const [expandedIds, setExpandedIds] = useState<Set<string>>(new Set());
  const [showRegenerateConfirm, setShowRegenerateConfirm] = useState(false);

  const rackOptions = getActiveRacks().map((r) => ({
    value: String(r.id),
    label: r.name,
  }));

  const toggleExpanded = useCallback((id: string) => {
    setExpandedIds((prev) => {
      const next = new Set(prev);
      if (next.has(id)) next.delete(id);
      else next.add(id);
      return next;
    });
  }, []);

  const dupSkus = duplicateValues(variants, 'sku');
  const dupBarcodes = duplicateValues(variants, 'barcode');

  // ---- Mode A: simple single variant ----
  if (!hasVariants) {
    const variant: ProductVariant = variants[0] ?? {
      id: nextVariantId(),
      sku: '',
      barcode: '',
      attributes: {},
      pricingTiers: [],
      images: [],
      rackIds: [],
      currentStock: 0,
    };

    const updateVariant = (patch: Partial<ProductVariant>) => {
      onVariantsChange([{ ...variant, ...patch }]);
    };

    return (
      <div className="space-y-6">
        <h3 className="text-sm font-semibold text-gray-800">Variant Details</h3>

        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          <Input
            label="SKU"
            value={variant.sku}
            placeholder="e.g. PROD-001"
            onChange={(e) => updateVariant({ sku: e.target.value })}
          />
          <Input
            label="Barcode"
            value={variant.barcode}
            placeholder="e.g. 4901234567890"
            onChange={(e) => updateVariant({ barcode: e.target.value })}
          />
        </div>

        <MultiSelect
          label="Racks"
          placeholder="Select racks (optional)"
          options={rackOptions}
          value={variant.rackIds.map(String)}
          onChange={(values) => updateVariant({ rackIds: values.map(Number) })}
        />

        <div className="text-sm text-gray-600">
          <span className="font-medium">Stock:</span> {variant.currentStock}
        </div>

        <VariantPricing
          priceSetting={priceSetting}
          markupType={markupType}
          pricingTiers={variant.pricingTiers}
          onChange={(fields) => updateVariant(fields)}
        />

        <ImageUpload
          label="Variant Images"
          images={variant.images}
          onChange={(imgs) => updateVariant({ images: imgs })}
        />
      </div>
    );
  }

  // ---- Mode B: full variant system ----

  // -- Step 1: Attribute definitions --
  const addAttribute = () => {
    onAttributesChange([...variantAttributes, { name: '', values: [] }]);
  };

  const updateAttribute = (index: number, patch: Partial<VariantAttribute>) => {
    const updated = variantAttributes.map((a, i) => (i === index ? { ...a, ...patch } : a));
    onAttributesChange(updated);
  };

  const removeAttribute = (index: number) => {
    onAttributesChange(variantAttributes.filter((_, i) => i !== index));
  };

  // -- Step 2: Generate variants --
  const doGenerate = () => {
    const combos = cartesian(variantAttributes);
    const newVariants: ProductVariant[] = combos.map((attrs) => ({
      id: nextVariantId(),
      sku: '',
      barcode: '',
      attributes: attrs,
      pricingTiers: [],
      images: [],
      rackIds: [],
      currentStock: 0,
    }));
    onVariantsChange(newVariants);
    setExpandedIds(new Set());
  };

  const generateVariants = () => {
    if (variants.length > 0) {
      setShowRegenerateConfirm(true);
      return;
    }
    doGenerate();
  };

  // -- Variant row helpers --
  const updateVariantField = (id: string, patch: Partial<ProductVariant>) => {
    onVariantsChange(variants.map((v) => (v.id === id ? { ...v, ...patch } : v)));
  };

  const removeVariant = (id: string) => {
    onVariantsChange(variants.filter((v) => v.id !== id));
    setExpandedIds((prev) => {
      const next = new Set(prev);
      next.delete(id);
      return next;
    });
  };

  const addManualVariant = () => {
    const attrs: Record<string, string> = {};
    for (const a of variantAttributes) {
      attrs[a.name] = '';
    }
    const newVariant: ProductVariant = {
      id: nextVariantId(),
      sku: '',
      barcode: '',
      attributes: attrs,
      pricingTiers: [],
      images: [],
      rackIds: [],
      currentStock: 0,
    };
    onVariantsChange([...variants, newVariant]);
    setExpandedIds((prev) => new Set(prev).add(newVariant.id));
  };

  // Collect attribute column names
  const attrNames = variantAttributes.filter((a) => a.name.trim()).map((a) => a.name);

  return (
    <div className="space-y-8">
      {/* Step 1: Attributes */}
      <section className="space-y-4">
        <div className="flex items-center justify-between">
          <h3 className="text-sm font-semibold text-gray-800">Variant Attributes</h3>
          <Button type="button" variant="outline" size="sm" onClick={addAttribute}>
            + Add Attribute
          </Button>
        </div>

        {variantAttributes.length === 0 && (
          <p className="text-sm text-gray-500">
            No attributes defined yet. Add attributes like Size, Color, etc.
          </p>
        )}

        {variantAttributes.length > 0 && (
          <table className="w-full text-sm border border-gray-200 rounded-md overflow-hidden">
            <thead>
              <tr className="bg-gray-50">
                <th className="text-left px-3 py-2 font-medium text-gray-600 w-48">
                  Attribute Name
                </th>
                <th className="text-left px-3 py-2 font-medium text-gray-600">Values</th>
                <th className="w-16 px-3 py-2" />
              </tr>
            </thead>
            <tbody>
              {variantAttributes.map((attr, idx) => (
                <tr key={idx} className="border-t border-gray-200">
                  <td className="px-3 py-2 align-top">
                    <input
                      type="text"
                      className="w-full rounded border border-gray-300 px-2 py-1 text-sm focus:outline-none focus:ring-1 focus:ring-blue-500"
                      value={attr.name}
                      placeholder="e.g. Size"
                      onChange={(e) => updateAttribute(idx, { name: e.target.value })}
                    />
                  </td>
                  <td className="px-3 py-2">
                    <TagInput
                      tags={attr.values}
                      onAddTag={(tag) =>
                        updateAttribute(idx, { values: [...attr.values, tag] })
                      }
                      onRemoveTag={(tag) =>
                        updateAttribute(idx, {
                          values: attr.values.filter((v) => v !== tag),
                        })
                      }
                      placeholder="Type value and press Enter"
                    />
                  </td>
                  <td className="px-3 py-2 text-center align-top">
                    <button
                      type="button"
                      onClick={() => removeAttribute(idx)}
                      className="text-red-500 hover:text-red-700 cursor-pointer"
                      title="Remove attribute"
                    >
                      <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
                      </svg>
                    </button>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        )}
      </section>

      {/* Step 2: Generate + Variant list */}
      <section className="space-y-4">
        <div className="flex items-center justify-between">
          <h3 className="text-sm font-semibold text-gray-800">
            Variants{variants.length > 0 ? ` (${variants.length})` : ''}
          </h3>
          <div className="flex gap-2">
            <Button
              type="button"
              variant="primary"
              size="sm"
              onClick={generateVariants}
              disabled={variantAttributes.every((a) => !a.name.trim() || a.values.length === 0)}
            >
              Generate Variants
            </Button>
            <Button type="button" variant="outline" size="sm" onClick={addManualVariant}>
              + Add Variant
            </Button>
          </div>
        </div>

        {variants.length === 0 && (
          <p className="text-sm text-gray-500">
            No variants yet. Define attributes above and click Generate Variants, or add manually.
          </p>
        )}

        {variants.length > 0 && (
          <div className="border border-gray-200 rounded-md overflow-hidden">
            <table className="w-full text-sm">
              <thead>
                <tr className="bg-gray-50 text-left">
                  <th className="px-3 py-2 font-medium text-gray-600 w-10">#</th>
                  <th className="px-3 py-2 font-medium text-gray-600">SKU</th>
                  <th className="px-3 py-2 font-medium text-gray-600">Barcode</th>
                  {attrNames.map((name) => (
                    <th key={name} className="px-3 py-2 font-medium text-gray-600">
                      {name}
                    </th>
                  ))}
                  <th className="px-3 py-2 font-medium text-gray-600 w-24">Actions</th>
                </tr>
              </thead>
              <tbody>
                {variants.map((variant, idx) => {
                  const isExpanded = expandedIds.has(variant.id);
                  const skuError =
                    variant.sku.trim() && dupSkus.has(variant.sku.trim())
                      ? 'Duplicate SKU'
                      : undefined;
                  const barcodeError =
                    variant.barcode.trim() && dupBarcodes.has(variant.barcode.trim())
                      ? 'Duplicate barcode'
                      : undefined;

                  return (
                    <React.Fragment key={variant.id}>
                      {/* Summary row */}
                      <tr className="border-t border-gray-200 hover:bg-gray-50">
                        <td className="px-3 py-2 text-gray-500">{idx + 1}</td>
                        <td className="px-3 py-2">
                          <input
                            type="text"
                            className={`w-full rounded border px-2 py-1 text-sm focus:outline-none focus:ring-1 focus:ring-blue-500 ${
                              skuError ? 'border-red-400' : 'border-gray-300'
                            }`}
                            value={variant.sku}
                            placeholder="SKU"
                            onChange={(e) =>
                              updateVariantField(variant.id, { sku: e.target.value })
                            }
                          />
                          {skuError && (
                            <p className="text-xs text-red-500 mt-0.5">{skuError}</p>
                          )}
                        </td>
                        <td className="px-3 py-2">
                          <input
                            type="text"
                            className={`w-full rounded border px-2 py-1 text-sm focus:outline-none focus:ring-1 focus:ring-blue-500 ${
                              barcodeError ? 'border-red-400' : 'border-gray-300'
                            }`}
                            value={variant.barcode}
                            placeholder="Barcode"
                            onChange={(e) =>
                              updateVariantField(variant.id, { barcode: e.target.value })
                            }
                          />
                          {barcodeError && (
                            <p className="text-xs text-red-500 mt-0.5">{barcodeError}</p>
                          )}
                        </td>
                        {attrNames.map((name) => (
                          <td key={name} className="px-3 py-2">
                            <span className="inline-block rounded bg-gray-100 px-2 py-0.5 text-xs text-gray-700">
                              {variant.attributes[name] || '-'}
                            </span>
                          </td>
                        ))}
                        <td className="px-3 py-2">
                          <div className="flex items-center gap-1">
                            <button
                              type="button"
                              onClick={() => toggleExpanded(variant.id)}
                              className="text-gray-500 hover:text-blue-600 cursor-pointer p-1"
                              title={isExpanded ? 'Collapse' : 'Expand'}
                            >
                              <svg
                                className={`w-4 h-4 transition-transform duration-200 ${
                                  isExpanded ? 'rotate-180' : ''
                                }`}
                                fill="none"
                                stroke="currentColor"
                                viewBox="0 0 24 24"
                              >
                                <path
                                  strokeLinecap="round"
                                  strokeLinejoin="round"
                                  strokeWidth={2}
                                  d="M19 9l-7 7-7-7"
                                />
                              </svg>
                            </button>
                            <button
                              type="button"
                              onClick={() => removeVariant(variant.id)}
                              className="text-red-500 hover:text-red-700 cursor-pointer p-1"
                              title="Delete variant"
                            >
                              <svg
                                className="w-4 h-4"
                                fill="none"
                                stroke="currentColor"
                                viewBox="0 0 24 24"
                              >
                                <path
                                  strokeLinecap="round"
                                  strokeLinejoin="round"
                                  strokeWidth={2}
                                  d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16"
                                />
                              </svg>
                            </button>
                          </div>
                        </td>
                      </tr>

                      {/* Expanded detail row */}
                      {isExpanded && (
                        <tr className="border-t border-gray-100 bg-gray-50/50">
                          <td
                            colSpan={4 + attrNames.length + 1}
                            className="px-6 py-4"
                          >
                            <div className="space-y-4 max-w-xl">
                              <MultiSelect
                                label="Racks"
                                placeholder="Select racks (optional)"
                                options={rackOptions}
                                value={variant.rackIds.map(String)}
                                onChange={(values) =>
                                  updateVariantField(variant.id, { rackIds: values.map(Number) })
                                }
                              />

                              <div className="text-sm text-gray-600">
                                <span className="font-medium">Stock:</span> {variant.currentStock}
                              </div>

                              <VariantPricing
                                priceSetting={priceSetting}
                                markupType={markupType}
                                pricingTiers={variant.pricingTiers}
                                onChange={(fields) =>
                                  updateVariantField(variant.id, fields)
                                }
                              />
                              <ImageUpload
                                label="Variant Images"
                                images={variant.images}
                                onChange={(imgs) =>
                                  updateVariantField(variant.id, { images: imgs })
                                }
                              />
                            </div>
                          </td>
                        </tr>
                      )}
                    </React.Fragment>
                  );
                })}
              </tbody>
            </table>
          </div>
        )}
      </section>

      {/* Regenerate confirmation modal */}
      <ConfirmModal
        isOpen={showRegenerateConfirm}
        onClose={() => setShowRegenerateConfirm(false)}
        onConfirm={() => {
          setShowRegenerateConfirm(false);
          doGenerate();
        }}
        title="Regenerate Variants"
        message="This will reset existing variant data. Are you sure you want to continue?"
        confirmLabel="Regenerate"
      />
    </div>
  );
}
