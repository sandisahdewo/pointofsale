'use client';

import React, { useState, useEffect } from 'react';
import { useRouter } from 'next/navigation';
import Button from '@/components/ui/Button';
import Select from '@/components/ui/Select';
import DatePicker from '@/components/ui/DatePicker';
import Textarea from '@/components/ui/Textarea';
import Input from '@/components/ui/Input';
import ConfirmModal from '@/components/ui/ConfirmModal';
import {
  usePurchaseOrderStore,
  PurchaseOrder,
  PurchaseOrderItem,
  POProductApi,
} from '@/stores/usePurchaseOrderStore';
import { useSupplierStore } from '@/stores/useSupplierStore';
import { useToastStore } from '@/stores/useToastStore';
import { useCategoryStore } from '@/stores/useCategoryStore';
import { ApiError } from '@/lib/api';
import Link from 'next/link';

interface PurchaseOrderFormProps {
  mode: 'add' | 'edit';
  initialPO?: PurchaseOrder;
}

interface FormState {
  supplierId: number;
  supplierName: string;
  date: string;
  notes: string;
  items: PurchaseOrderItem[];
}

export default function PurchaseOrderForm({ mode, initialPO }: PurchaseOrderFormProps) {
  const router = useRouter();
  const {
    createPurchaseOrderRemote,
    updatePurchaseOrderRemote,
    fetchProductsForPO,
  } = usePurchaseOrderStore();
  const { suppliers, getActiveSuppliers, fetchAllSuppliers } = useSupplierStore();
  const { addToast } = useToastStore();
  const { categories, fetchAllCategories } = useCategoryStore();

  const [form, setForm] = useState<FormState>({
    supplierId: initialPO?.supplierId || 0,
    supplierName: initialPO?.supplierName || '',
    date: initialPO?.date || new Date().toISOString().split('T')[0],
    notes: initialPO?.notes || '',
    items: initialPO?.items || [],
  });

  // Products loaded from the PO products endpoint
  const [poProducts, setPoProducts] = useState<POProductApi[]>([]);
  const [isLoadingProducts, setIsLoadingProducts] = useState(false);

  const [errors, setErrors] = useState<Record<string, string>>({});
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [expandedProducts, setExpandedProducts] = useState<Set<number>>(new Set());
  const [searchQuery, setSearchQuery] = useState('');
  const [confirmModal, setConfirmModal] = useState<{
    isOpen: boolean;
    title: string;
    message: string;
    onConfirm: () => void;
  }>({
    isOpen: false,
    title: '',
    message: '',
    onConfirm: () => {},
  });

  const activeSuppliers = getActiveSuppliers();

  useEffect(() => {
    const loadMasterData = async () => {
      try {
        await Promise.all([
          fetchAllSuppliers(),
          fetchAllCategories(),
        ]);
      } catch (error) {
        if (error instanceof ApiError) {
          addToast(error.message, 'error');
        } else {
          addToast('Failed to load suppliers and categories', 'error');
        }
      }
    };

    void loadMasterData();
  }, [fetchAllSuppliers, fetchAllCategories, addToast]);

  // Load products for initial PO (edit mode)
  useEffect(() => {
    if (mode === 'edit' && initialPO?.supplierId) {
      void loadProductsForSupplier(initialPO.supplierId);
    }
  // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [mode, initialPO?.supplierId]);

  const loadProductsForSupplier = async (supplierId: number) => {
    if (!supplierId) {
      setPoProducts([]);
      return;
    }
    try {
      setIsLoadingProducts(true);
      const products = await fetchProductsForPO(supplierId);
      setPoProducts(products);
      return products;
    } catch (error) {
      if (error instanceof ApiError) {
        addToast(error.message, 'error');
      } else {
        addToast('Failed to load products for supplier', 'error');
      }
      return [];
    } finally {
      setIsLoadingProducts(false);
    }
  };

  const buildItemsForSupplier = (supplierId: number, products: POProductApi[]) => {
    const items: PurchaseOrderItem[] = [];
    const expandedProductIds: number[] = [];

    products.forEach((product) => {
      expandedProductIds.push(product.id);
      (product.variants ?? []).forEach((variant) => {
        const variantLabel =
          (variant.attributes && variant.attributes.length > 0)
            ? variant.attributes.map((a) => a.attributeValue).join(' / ')
            : 'Default';

        const baseUnit = (product.units ?? []).find((u) => u.isBase);

        items.push({
          id: crypto.randomUUID(),
          productId: product.id,
          productName: product.name,
          variantId: variant.id,
          variantLabel,
          sku: variant.sku ?? '',
          currentStock: Number(variant.currentStock ?? 0),
          orderedQty: 0,
          price: 0,
          unitId: baseUnit ? String(baseUnit.id) : '',
          unitName: baseUnit?.name ?? '',
        });
      });
    });

    return { items, expandedProductIds };
  };

  const applySupplierSelection = async (supplierId: number) => {
    const supplier = suppliers.find((item) => item.id === supplierId);

    if (mode === 'add' && supplierId > 0) {
      const products = await loadProductsForSupplier(supplierId) ?? [];
      const { items, expandedProductIds } = buildItemsForSupplier(supplierId, products);

      setForm((prev) => ({
        ...prev,
        supplierId,
        supplierName: supplier?.name || '',
        items,
      }));
      setExpandedProducts(new Set(expandedProductIds));
      return;
    }

    setForm((prev) => ({
      ...prev,
      supplierId,
      supplierName: supplier?.name || '',
      items: mode === 'add' ? [] : prev.items,
    }));

    if (mode === 'add') {
      setExpandedProducts(new Set());
      setPoProducts([]);
    }
  };

  const handleSupplierChange = (supplierId: number) => {
    if (mode === 'add' && form.items.length > 0 && supplierId !== form.supplierId) {
      setConfirmModal({
        isOpen: true,
        title: 'Change Supplier',
        message: 'Changing the supplier will reset the product list. Continue?',
        onConfirm: () => {
          void applySupplierSelection(supplierId);
          setConfirmModal((prev) => ({ ...prev, isOpen: false }));
        },
      });
    } else {
      void applySupplierSelection(supplierId);
    }
  };

  const handleQtyChange = (itemId: string, qty: number) => {
    setForm({
      ...form,
      items: form.items.map((item) =>
        item.id === itemId ? { ...item, orderedQty: Math.max(0, qty) } : item
      ),
    });
  };

  const handlePriceChange = (itemId: string, price: number) => {
    setForm({
      ...form,
      items: form.items.map((item) =>
        item.id === itemId ? { ...item, price: Math.max(0, price) } : item
      ),
    });
  };

  const handleUnitChange = (itemId: string, unitId: string) => {
    const item = form.items.find((i) => i.id === itemId);
    if (!item) return;
    const product = poProducts.find((p) => p.id === item.productId);
    const unit = (product?.units ?? []).find((u) => String(u.id) === unitId);
    if (!unit) return;
    setForm({
      ...form,
      items: form.items.map((i) =>
        i.id === itemId ? { ...i, unitId: String(unit.id), unitName: unit.name } : i
      ),
    });
  };

  const removeProduct = (productId: number) => {
    setForm({
      ...form,
      items: form.items.filter((item) => item.productId !== productId),
    });
    setExpandedProducts((prev) => {
      const next = new Set(prev);
      next.delete(productId);
      return next;
    });
  };

  const toggleProduct = (productId: number) => {
    setExpandedProducts((prev) => {
      const next = new Set(prev);
      if (next.has(productId)) {
        next.delete(productId);
      } else {
        next.add(productId);
      }
      return next;
    });
  };

  const validate = (): boolean => {
    const newErrors: Record<string, string> = {};

    if (!form.supplierId) {
      newErrors.supplier = 'Supplier is required';
    }
    if (!form.date) {
      newErrors.date = 'Date is required';
    }

    const hasItems = form.items.some((item) => item.orderedQty > 0);
    if (!hasItems) {
      newErrors.items = 'At least one item must have quantity greater than 0';
    }

    setErrors(newErrors);
    return Object.keys(newErrors).length === 0;
  };

  const handleSave = async () => {
    if (!validate()) {
      addToast('Please fix the errors before saving', 'error');
      return;
    }

    const orderedItems = form.items.filter((item) => item.orderedQty > 0);

    const input = {
      supplierId: form.supplierId,
      date: form.date,
      notes: form.notes,
      items: orderedItems.map((item) => ({
        productId: item.productId,
        variantId: item.variantId,
        unitId: Number(item.unitId),
        orderedQty: item.orderedQty,
        price: item.price,
      })),
    };

    try {
      setIsSubmitting(true);
      if (mode === 'add') {
        await createPurchaseOrderRemote(input);
        addToast('Purchase order created successfully', 'success');
      } else if (initialPO) {
        await updatePurchaseOrderRemote(initialPO.id, input);
        addToast('Purchase order updated successfully', 'success');
      }
      router.push('/transaction/purchase');
    } catch (error) {
      if (error instanceof ApiError) {
        addToast(error.message, 'error');
      } else {
        addToast('Failed to save purchase order', 'error');
      }
    } finally {
      setIsSubmitting(false);
    }
  };

  const handleCancel = () => {
    router.push('/transaction/purchase');
  };

  // Group items by product
  const itemsByProduct = form.items.reduce((acc, item) => {
    if (!acc[item.productId]) {
      acc[item.productId] = [];
    }
    acc[item.productId].push(item);
    return acc;
  }, {} as Record<number, PurchaseOrderItem[]>);

  // Filter products for search
  const filteredProductIds = Object.keys(itemsByProduct)
    .map(Number)
    .filter((productId) => {
      if (!searchQuery.trim()) return true;
      const product = poProducts.find((p) => p.id === productId);
      // Fall back to productName from item if product not found in poProducts
      const productName = product?.name ?? itemsByProduct[productId]?.[0]?.productName ?? '';
      return productName.toLowerCase().includes(searchQuery.toLowerCase());
    });

  // Calculate summary
  const summary = {
    variantCount: form.items.filter((item) => item.orderedQty > 0).length,
    totalQty: form.items.reduce((sum, item) => sum + item.orderedQty, 0),
    estimatedTotal: form.items.reduce((sum, item) => sum + item.orderedQty * item.price, 0),
  };

  return (
    <div className="min-h-screen flex flex-col">
      {/* Sticky Header */}
      <div className="sticky top-0 z-10 bg-white border-b border-gray-200 px-6 py-4">
        <div className="flex items-center justify-between">
          <Link
            href="/transaction/purchase"
            className="text-sm text-gray-600 hover:text-gray-900 flex items-center gap-1"
          >
            Back to Purchase Orders
          </Link>
          <div className="flex gap-2">
            <Button variant="outline" onClick={handleCancel} disabled={isSubmitting}>
              Cancel
            </Button>
            <Button onClick={() => void handleSave()} disabled={isSubmitting}>
              {isSubmitting ? 'Saving...' : 'Save'}
            </Button>
          </div>
        </div>
      </div>

      {/* Form Content */}
      <div className="flex-1 p-6 space-y-6">
        <h1 className="text-2xl font-bold text-gray-900">
          {mode === 'add' ? 'New Purchase Order' : `Edit Purchase Order — ${initialPO?.poNumber}`}
        </h1>

        {/* Basic Info */}
        <div className="bg-white rounded-lg border border-gray-200 p-6 space-y-4">
          <Select
            label="Supplier"
            value={form.supplierId.toString()}
            onChange={(e) => handleSupplierChange(Number(e.target.value))}
            options={[
              { value: '0', label: 'Select supplier' },
              ...activeSuppliers.map((s) => ({
                value: s.id.toString(),
                label: s.name,
              })),
            ]}
            error={errors.supplier}
            required
          />

          <DatePicker
            label="Date"
            value={form.date}
            onChange={(value) => setForm({ ...form, date: value })}
            error={errors.date}
            required
          />

          <Textarea
            label="Notes"
            value={form.notes}
            onChange={(e) => setForm({ ...form, notes: e.target.value })}
            placeholder="Optional internal notes"
            rows={3}
          />
        </div>

        {/* Order Items */}
        <div className="bg-white rounded-lg border border-gray-200 p-6 space-y-4">
          <h2 className="text-lg font-semibold text-gray-900">Order Items</h2>

          {form.supplierId ? (
            <>
              <p className="text-sm text-gray-600">
                Showing products linked to the selected supplier.
              </p>

              <Input
                placeholder="Search products..."
                value={searchQuery}
                onChange={(e) => setSearchQuery(e.target.value)}
              />

              {errors.items && (
                <p className="text-sm text-red-600">{errors.items}</p>
              )}

              {isLoadingProducts ? (
                <p className="text-sm text-gray-500 text-center py-8">Loading products...</p>
              ) : filteredProductIds.length === 0 ? (
                <p className="text-sm text-gray-500 text-center py-8">
                  {searchQuery ? 'No products found matching your search.' : 'No products available for this supplier.'}
                </p>
              ) : (
                <div className="space-y-4">
                  {filteredProductIds.map((productId) => {
                    const product = poProducts.find((p) => p.id === productId);
                    const productName = product?.name ?? itemsByProduct[productId]?.[0]?.productName ?? '';
                    const productItems = itemsByProduct[productId];
                    const isExpanded = expandedProducts.has(productId);
                    const categoryName = product
                      ? (categories.find((c) => c.id === product.categoryId)?.name ?? 'Uncategorized')
                      : '';

                    return (
                      <div key={productId} className="border border-gray-200 rounded-lg">
                        <div
                          className="flex items-center justify-between p-4 cursor-pointer hover:bg-gray-50"
                          onClick={() => toggleProduct(productId)}
                        >
                          <div className="flex items-center gap-2">
                            <span className="text-lg">{isExpanded ? '▼' : '▶'}</span>
                            <h3 className="font-medium text-gray-900">
                              {productName}{categoryName ? ` (${categoryName})` : ''}
                            </h3>
                          </div>
                          <Button
                            size="sm"
                            variant="outline"
                            onClick={(e) => {
                              e.stopPropagation();
                              removeProduct(productId);
                            }}
                          >
                            Remove
                          </Button>
                        </div>

                        {isExpanded && (
                          <div className="border-t border-gray-200">
                            <table className="w-full">
                              <thead className="bg-gray-50">
                                <tr>
                                  <th className="px-4 py-2 text-left text-xs font-medium text-gray-600">
                                    Variant
                                  </th>
                                  <th className="px-4 py-2 text-left text-xs font-medium text-gray-600">
                                    SKU
                                  </th>
                                  <th className="px-4 py-2 text-left text-xs font-medium text-gray-600">
                                    Stock
                                  </th>
                                  <th className="px-4 py-2 text-left text-xs font-medium text-gray-600">
                                    Unit
                                  </th>
                                  <th className="px-4 py-2 text-left text-xs font-medium text-gray-600">
                                    Order Qty
                                  </th>
                                  <th className="px-4 py-2 text-left text-xs font-medium text-gray-600">
                                    Price
                                  </th>
                                </tr>
                              </thead>
                              <tbody>
                                {productItems.map((item) => {
                                  const productUnits = product?.units ?? [];
                                  return (
                                    <tr key={item.id} className="border-t border-gray-100">
                                      <td className="px-4 py-2 text-sm text-gray-700">
                                        {item.variantLabel}
                                      </td>
                                      <td className="px-4 py-2 text-sm text-gray-600">{item.sku}</td>
                                      <td className="px-4 py-2 text-sm text-gray-600">
                                        {item.currentStock}
                                      </td>
                                      <td className="px-4 py-2">
                                        <select
                                          value={item.unitId || ''}
                                          onChange={(e) => handleUnitChange(item.id, e.target.value)}
                                          className="w-24 rounded-md border border-gray-300 px-2 py-1 text-sm"
                                        >
                                          {productUnits.map((unit) => (
                                            <option key={unit.id} value={String(unit.id)}>
                                              {unit.name}
                                            </option>
                                          ))}
                                        </select>
                                      </td>
                                      <td className="px-4 py-2">
                                        <input
                                          type="number"
                                          min="0"
                                          value={item.orderedQty}
                                          onChange={(e) =>
                                            handleQtyChange(item.id, parseInt(e.target.value) || 0)
                                          }
                                          className="w-24 rounded-md border border-gray-300 px-2 py-1 text-sm"
                                        />
                                      </td>
                                      <td className="px-4 py-2">
                                        <input
                                          type="number"
                                          min="0"
                                          value={item.price}
                                          onChange={(e) =>
                                            handlePriceChange(item.id, parseInt(e.target.value) || 0)
                                          }
                                          className="w-28 rounded-md border border-gray-300 px-2 py-1 text-sm"
                                        />
                                      </td>
                                    </tr>
                                  );
                                })}
                              </tbody>
                            </table>
                          </div>
                        )}
                      </div>
                    );
                  })}
                </div>
              )}

              {/* Summary */}
              <div className="border-t border-gray-200 pt-4 mt-4">
                <div className="flex justify-end text-sm">
                  <div className="space-y-1">
                    <p className="text-gray-600">
                      <span className="font-medium">Summary:</span> {summary.variantCount} variant
                      {summary.variantCount !== 1 ? 's' : ''} · Total Qty: {summary.totalQty} · Est.
                      Total: Rp {summary.estimatedTotal.toLocaleString('id-ID')}
                    </p>
                  </div>
                </div>
              </div>
            </>
          ) : (
            <p className="text-sm text-gray-500 text-center py-8">
              Please select a supplier to view products.
            </p>
          )}
        </div>
      </div>

      {/* Confirm Modal */}
      <ConfirmModal
        isOpen={confirmModal.isOpen}
        onClose={() => setConfirmModal({ ...confirmModal, isOpen: false })}
        onConfirm={confirmModal.onConfirm}
        title={confirmModal.title}
        message={confirmModal.message}
        confirmLabel="Continue"
      />
    </div>
  );
}
