'use client';

import React, { useState, useEffect, useCallback, useRef } from 'react';
import { useRouter } from 'next/navigation';
import AdminLayout from '@/components/layout/AdminLayout';
import Button from '@/components/ui/Button';
import Input from '@/components/ui/Input';
import Textarea from '@/components/ui/Textarea';
import Select from '@/components/ui/Select';
import MultiSelect from '@/components/ui/MultiSelect';
import Toggle from '@/components/ui/Toggle';
import Tabs from '@/components/ui/Tabs';
import ImageUpload from '@/components/ui/ImageUpload';
import ConfirmModal from '@/components/ui/ConfirmModal';
import UnitsTab from '@/components/product/UnitsTab';
import VariantsTab from '@/components/product/VariantsTab';
import VariantPricing from '@/components/product/VariantPricing';
import { useProductStore, Product, PriceSetting, MarkupType, ProductUnit, VariantAttribute, ProductVariant, PricingTier } from '@/stores/useProductStore';
import { useCategoryStore } from '@/stores/useCategoryStore';
import { useSupplierStore } from '@/stores/useSupplierStore';
import { useToastStore } from '@/stores/useToastStore';
import { ApiError } from '@/lib/api';
import Link from 'next/link';

interface ProductFormProps {
  mode: 'add' | 'edit';
  initialProduct?: Product;
}

interface FormState {
  name: string;
  description: string;
  categoryId: number;
  images: string[];
  priceSetting: PriceSetting;
  markupType: MarkupType;
  hasVariants: boolean;
  status: 'active' | 'inactive';
  supplierIds: number[];
}

const EMPTY_FORM: FormState = {
  name: '',
  description: '',
  categoryId: 0,
  images: [],
  priceSetting: 'fixed',
  markupType: 'percentage',
  hasVariants: false,
  status: 'active',
  supplierIds: [],
};

function formFromProduct(product: Product): FormState {
  return {
    name: product.name,
    description: product.description,
    categoryId: product.categoryId,
    images: product.images,
    priceSetting: product.priceSetting,
    markupType: product.markupType ?? 'percentage',
    hasVariants: product.hasVariants,
    status: product.status,
    supplierIds: product.supplierIds ?? [],
  };
}

const TABS = [
  { key: 'price', label: 'Price' },
  { key: 'units', label: 'Units' },
  { key: 'variants', label: 'Variants' },
];

export default function ProductForm({ mode, initialProduct }: ProductFormProps) {
  const router = useRouter();
  const { createProduct, updateProductRemote } = useProductStore();
  const { categories, fetchAllCategories } = useCategoryStore();
  const { getActiveSuppliers, fetchAllSuppliers } = useSupplierStore();
  const { addToast } = useToastStore();

  const [form, setForm] = useState<FormState>(
    initialProduct ? formFromProduct(initialProduct) : EMPTY_FORM
  );
  const [errors, setErrors] = useState<Record<string, string>>({});
  const [activeTab, setActiveTab] = useState('price');
  const [isDirty, setIsDirty] = useState(false);
  const [isSaving, setIsSaving] = useState(false);
  const [units, setUnits] = useState<ProductUnit[]>(initialProduct?.units ?? []);
  const [variantAttributes, setVariantAttributes] = useState<VariantAttribute[]>(
    initialProduct?.variantAttributes ?? []
  );
  const [variants, setVariants] = useState<ProductVariant[]>(
    initialProduct?.variants ?? []
  );
  const [productLevelPricing, setProductLevelPricing] = useState<PricingTier[]>([]);
  const [isPricingEditMode, setIsPricingEditMode] = useState(false);
  const [pricingSnapshot, setPricingSnapshot] = useState<PricingTier[]>([]);
  const initialSnapshot = useRef(JSON.stringify({
    form: initialProduct ? formFromProduct(initialProduct) : EMPTY_FORM,
    units: initialProduct?.units ?? [],
    variantAttributes: initialProduct?.variantAttributes ?? [],
    variants: initialProduct?.variants ?? [],
  }));

  // Confirmation modal state
  const [confirmModal, setConfirmModal] = useState<{
    isOpen: boolean;
    title: string;
    message: string;
    cancelLabel: string;
    confirmLabel: string;
    variant: 'primary' | 'danger';
    onConfirm: () => void;
  }>({
    isOpen: false,
    title: '',
    message: '',
    cancelLabel: 'Cancel',
    confirmLabel: 'Continue',
    variant: 'primary',
    onConfirm: () => {},
  });

  const closeConfirmModal = useCallback(() => {
    setConfirmModal((prev) => ({ ...prev, isOpen: false }));
  }, []);

  const showConfirm = useCallback(
    (opts: {
      title: string;
      message: string;
      cancelLabel?: string;
      confirmLabel?: string;
      variant?: 'primary' | 'danger';
      onConfirm: () => void;
    }) => {
      setConfirmModal({
        isOpen: true,
        title: opts.title,
        message: opts.message,
        cancelLabel: opts.cancelLabel ?? 'Cancel',
        confirmLabel: opts.confirmLabel ?? 'Continue',
        variant: opts.variant ?? 'primary',
        onConfirm: opts.onConfirm,
      });
    },
    [],
  );

  const updateField = useCallback(<K extends keyof FormState>(key: K, value: FormState[K]) => {
    setForm((prev) => ({ ...prev, [key]: value }));
    setIsDirty(true);
    if (errors[key]) {
      setErrors((prev) => {
        const next = { ...prev };
        delete next[key];
        return next;
      });
    }
  }, [errors]);

  // Unsaved changes warning via beforeunload (browser fallback)
  useEffect(() => {
    const handler = (e: BeforeUnloadEvent) => {
      if (isDirty) {
        e.preventDefault();
      }
    };
    window.addEventListener('beforeunload', handler);
    return () => window.removeEventListener('beforeunload', handler);
  }, [isDirty]);

  // Track dirty state by comparing against initial snapshot
  useEffect(() => {
    const currentSnapshot = JSON.stringify({ form, units, variantAttributes, variants });
    setIsDirty(currentSnapshot !== initialSnapshot.current);
  }, [form, units, variantAttributes, variants]);

  useEffect(() => {
    const loadMasterData = async () => {
      try {
        await Promise.all([
          fetchAllCategories(),
          fetchAllSuppliers({ active: true }),
        ]);
      } catch (error) {
        if (error instanceof ApiError) {
          addToast(error.message, 'error');
        } else {
          addToast('Failed to load categories and suppliers', 'error');
        }
      }
    };

    void loadMasterData();
  }, [fetchAllCategories, fetchAllSuppliers, addToast]);

  const handlePriceSettingChange = (value: PriceSetting) => {
    if (form.priceSetting === value) return;
    if (variants.length > 0) {
      showConfirm({
        title: 'Change Price Setting',
        message: 'Changing price setting will reset variant pricing data. Are you sure you want to continue?',
        onConfirm: () => {
          updateField('priceSetting', value);
          closeConfirmModal();
        },
      });
      return;
    }
    updateField('priceSetting', value);
  };

  const handleHasVariantsChange = (value: boolean) => {
    if (form.hasVariants === value) return;
    if (variants.length > 0) {
      showConfirm({
        title: 'Reset Variant Data',
        message: 'Changing this will reset variant data. Are you sure you want to continue?',
        onConfirm: () => {
          updateField('hasVariants', value);
          closeConfirmModal();
        },
      });
      return;
    }
    updateField('hasVariants', value);
  };

  const handleProductLevelPricingChange = (newTiers: PricingTier[]) => {
    // Just update local state - modal will be shown on Save Pricing button
    setProductLevelPricing(newTiers);
    setIsDirty(true);
  };

  const handlePricingEdit = () => {
    setPricingSnapshot([...productLevelPricing]);
    setIsPricingEditMode(true);
  };

  const handlePricingCancel = () => {
    setProductLevelPricing(pricingSnapshot);
    setIsPricingEditMode(false);
  };

  const handlePricingSave = () => {
    // Check if variants already have pricing data
    const variantsHavePricing = variants.some(v => v.pricingTiers && v.pricingTiers.length > 0);

    if (variantsHavePricing && productLevelPricing.length > 0) {
      showConfirm({
        title: 'Update All Variant Pricing',
        message: 'This will replace pricing on all variants, including any custom values. Are you sure you want to continue?',
        confirmLabel: 'Update All',
        onConfirm: () => {
          // Apply to all variants
          setVariants(variants.map(v => ({ ...v, pricingTiers: productLevelPricing })));
          setIsPricingEditMode(false);
          setIsDirty(true);
          closeConfirmModal();
        },
      });
      return;
    }

    // No variants have pricing or no tiers to apply - apply directly
    if (productLevelPricing.length > 0) {
      setVariants(variants.map(v => ({ ...v, pricingTiers: productLevelPricing })));
    }
    setIsPricingEditMode(false);
    setIsDirty(true);
  };

  const validate = (): boolean => {
    const newErrors: Record<string, string> = {};
    if (!form.name.trim()) newErrors.name = 'Name is required';
    if (!form.categoryId) newErrors.categoryId = 'Category is required';
    if (!form.priceSetting) newErrors.priceSetting = 'Price setting is required';
    if (units.length === 0) newErrors.units = 'At least one unit (base unit) must be defined';
    setErrors(newErrors);
    if (newErrors.priceSetting) {
      setActiveTab('price');
    } else if (newErrors.units) {
      setActiveTab('units');
    }
    return Object.keys(newErrors).length === 0;
  };

  const handleSave = async () => {
    if (!validate()) return;

    const productData = {
      name: form.name.trim(),
      description: form.description.trim(),
      categoryId: form.categoryId,
      images: form.images,
      priceSetting: form.priceSetting,
      markupType: form.priceSetting === 'markup' ? form.markupType : undefined,
      hasVariants: form.hasVariants,
      status: form.status,
      supplierIds: form.supplierIds,
      units,
      variantAttributes,
      variants,
    };

    setIsSaving(true);
    try {
      if (mode === 'edit' && initialProduct) {
        await updateProductRemote(initialProduct.id, productData);
        addToast('Product updated successfully', 'success');
      } else {
        await createProduct(productData);
        addToast('Product added successfully', 'success');
      }

      setIsDirty(false);
      router.push('/master/product');
    } catch (error) {
      if (error instanceof ApiError) {
        addToast(error.message, 'error');
      } else {
        addToast('Failed to save product', 'error');
      }
    } finally {
      setIsSaving(false);
    }
  };

  const handleCancel = () => {
    if (isDirty) {
      showConfirm({
        title: 'Unsaved Changes',
        message: 'You have unsaved changes. Are you sure you want to leave?',
        cancelLabel: 'Stay',
        confirmLabel: 'Leave',
        variant: 'danger',
        onConfirm: () => {
          setIsDirty(false);
          closeConfirmModal();
          router.push('/master/product');
        },
      });
      return;
    }
    router.push('/master/product');
  };

  const categoryOptions = categories.map((c) => ({
    value: String(c.id),
    label: c.name,
  }));

  const supplierOptions = getActiveSuppliers().map((s) => ({
    value: String(s.id),
    label: s.name,
  }));

  return (
    <AdminLayout>
      <div className="space-y-6">
        {/* Sticky header */}
        <div className="sticky top-16 z-10 bg-gray-50 -mx-6 -mt-6 px-6 py-4 border-b border-gray-200">
          <div className="flex items-center justify-between">
            <Link
              href="/master/product"
              className="inline-flex items-center gap-1 text-sm text-gray-600 hover:text-gray-900"
            >
              <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 19l-7-7 7-7" />
              </svg>
              Back to Product List
            </Link>
            <div className="flex gap-2">
              <Button variant="outline" onClick={handleCancel}>
                Cancel
              </Button>
              <Button onClick={() => void handleSave()} disabled={isSaving}>
                {isSaving ? 'Saving...' : mode === 'edit' ? 'Update Product' : 'Save Product'}
              </Button>
            </div>
          </div>
        </div>

        {/* Page title */}
        <h1 className="text-2xl font-bold text-gray-900">
          {mode === 'edit' ? 'Edit Product' : 'Add Product'}
        </h1>

        {/* General fields */}
        <div className="bg-white rounded-lg border border-gray-200 shadow-sm p-6 space-y-6">
          <Input
            label="Name"
            placeholder="Product name"
            value={form.name}
            onChange={(e) => updateField('name', e.target.value)}
            error={errors.name}
          />

          <Textarea
            label="Description"
            placeholder="Product description"
            value={form.description}
            onChange={(e) => updateField('description', e.target.value)}
          />

          <Select
            label="Category"
            placeholder="Select a category"
            options={categoryOptions}
            value={String(form.categoryId)}
            onChange={(e) => updateField('categoryId', Number(e.target.value))}
            error={errors.categoryId}
          />

          <MultiSelect
            label="Suppliers"
            placeholder="Select suppliers (optional)"
            options={supplierOptions}
            value={form.supplierIds.map(String)}
            onChange={(values) => updateField('supplierIds', values.map(Number))}
          />

          <ImageUpload
            label="Images"
            images={form.images}
            onChange={(images) => updateField('images', images)}
          />

          {/* Has Variants - radio group */}
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-2">
              Has Variants
            </label>
            <div className="flex gap-2">
              {([
                { value: false, label: 'No' },
                { value: true, label: 'Yes' },
              ]).map((option) => (
                <button
                  key={String(option.value)}
                  type="button"
                  onClick={() => handleHasVariantsChange(option.value)}
                  className={`px-4 py-2 text-sm font-medium rounded-md border transition-colors cursor-pointer ${
                    form.hasVariants === option.value
                      ? 'bg-blue-600 text-white border-blue-600'
                      : 'bg-white text-gray-700 border-gray-300 hover:bg-gray-50'
                  }`}
                >
                  {option.label}
                </button>
              ))}
            </div>
          </div>

          {/* Status */}
          <Toggle
            label={form.status === 'active' ? 'Active' : 'Inactive'}
            checked={form.status === 'active'}
            onChange={(checked) => updateField('status', checked ? 'active' : 'inactive')}
          />
        </div>

        {/* Tab section */}
        <div className="bg-white rounded-lg border border-gray-200 shadow-sm">
          <Tabs tabs={TABS} activeTab={activeTab} onTabChange={setActiveTab} />
          <div className="p-6">
            {activeTab === 'price' && (
              <div className="space-y-6">
                {/* Price Setting - radio group */}
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-2">
                    Price Setting
                  </label>
                  <div className="flex gap-2">
                    {([
                      { value: 'fixed' as PriceSetting, label: 'Fixed Price' },
                      { value: 'markup' as PriceSetting, label: 'Markup Price' },
                    ]).map((option) => (
                      <button
                        key={option.value}
                        type="button"
                        onClick={() => handlePriceSettingChange(option.value)}
                        className={`px-4 py-2 text-sm font-medium rounded-md border transition-colors cursor-pointer ${
                          form.priceSetting === option.value
                            ? 'bg-blue-600 text-white border-blue-600'
                            : 'bg-white text-gray-700 border-gray-300 hover:bg-gray-50'
                        }`}
                      >
                        {option.label}
                      </button>
                    ))}
                  </div>
                  {errors.priceSetting && (
                    <p className="mt-1 text-sm text-red-600">{errors.priceSetting}</p>
                  )}
                </div>

                {/* Markup Type - conditional */}
                {form.priceSetting === 'markup' && (
                  <div>
                    <label className="block text-sm font-medium text-gray-700 mb-2">
                      Markup Type
                    </label>
                    <div className="flex gap-2">
                      {([
                        { value: 'percentage' as MarkupType, label: 'Percentage' },
                        { value: 'fixed_amount' as MarkupType, label: 'Fixed Amount' },
                      ]).map((option) => (
                        <button
                          key={option.value}
                          type="button"
                          onClick={() => updateField('markupType', option.value)}
                          className={`px-4 py-2 text-sm font-medium rounded-md border transition-colors cursor-pointer ${
                            form.markupType === option.value
                              ? 'bg-blue-600 text-white border-blue-600'
                              : 'bg-white text-gray-700 border-gray-300 hover:bg-gray-50'
                          }`}
                        >
                          {option.label}
                        </button>
                      ))}
                    </div>
                  </div>
                )}

                <hr className="border-gray-200" />

                {/* Default Pricing */}
                <div>
                  <div className="flex items-center justify-between mb-2">
                    <h3 className="text-sm font-semibold text-gray-800">Default Pricing (optional)</h3>
                    {!isPricingEditMode && (
                      <Button type="button" variant="outline" size="sm" onClick={handlePricingEdit}>
                        Edit
                      </Button>
                    )}
                  </div>
                  <div className="text-xs text-gray-600 bg-gray-50 border border-gray-200 rounded-md px-3 py-2 mb-4">
                    Pricing set here will be applied to ALL variants. Variants can override these values individually in the Variants tab. However, saving these values will REPLACE all variant pricing, including any overrides.
                  </div>

                  <VariantPricing
                    priceSetting={form.priceSetting}
                    markupType={form.priceSetting === 'markup' ? form.markupType : undefined}
                    pricingTiers={productLevelPricing.length > 0 ? productLevelPricing : [{ minQty: 1, value: 0 }]}
                    disabled={!isPricingEditMode}
                    onChange={(fields) => {
                      if (fields.pricingTiers !== undefined) {
                        handleProductLevelPricingChange(fields.pricingTiers);
                      }
                    }}
                  />

                  {isPricingEditMode && (
                    <div className="flex gap-2 mt-4">
                      <Button type="button" variant="outline" size="sm" onClick={handlePricingCancel}>
                        Cancel
                      </Button>
                      <Button type="button" size="sm" onClick={handlePricingSave}>
                        Save Pricing
                      </Button>
                    </div>
                  )}
                </div>
              </div>
            )}
            {activeTab === 'units' && (
              <UnitsTab units={units} onChange={setUnits} />
            )}
            {activeTab === 'variants' && (
              <VariantsTab
                hasVariants={form.hasVariants}
                priceSetting={form.priceSetting}
                markupType={form.priceSetting === 'markup' ? form.markupType : undefined}
                variantAttributes={variantAttributes}
                variants={variants}
                onAttributesChange={setVariantAttributes}
                onVariantsChange={setVariants}
              />
            )}
          </div>
        </div>
      </div>

      {/* Confirmation modal */}
      <ConfirmModal
        isOpen={confirmModal.isOpen}
        onClose={closeConfirmModal}
        onConfirm={confirmModal.onConfirm}
        title={confirmModal.title}
        message={confirmModal.message}
        cancelLabel={confirmModal.cancelLabel}
        confirmLabel={confirmModal.confirmLabel}
        variant={confirmModal.variant}
      />
    </AdminLayout>
  );
}
