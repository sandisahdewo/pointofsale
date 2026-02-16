'use client';

import React, { useState, useEffect } from 'react';
import { useRouter, useParams } from 'next/navigation';
import Link from 'next/link';
import Button from '@/components/ui/Button';
import Select from '@/components/ui/Select';
import DatePicker from '@/components/ui/DatePicker';
import Checkbox from '@/components/ui/Checkbox';
import { usePurchaseOrderStore, PaymentMethod, PurchaseOrder } from '@/stores/usePurchaseOrderStore';
import { useSupplierStore } from '@/stores/useSupplierStore';
import { useToastStore } from '@/stores/useToastStore';
import { ApiError } from '@/lib/api';

interface ItemReceiveData {
  id: string;
  orderedQty: number;
  receivedQty: number;
  price: number;
  receivedPrice: number;
  isVerified: boolean;
}

const WARNING_DISMISS_KEY = 'po_receive_match_warning_dismissed';

function getWarningDismissedPreference(): boolean {
  if (typeof window === 'undefined') return false;
  return localStorage.getItem(WARNING_DISMISS_KEY) === 'true';
}

function buildInitialItemsData(po?: PurchaseOrder): Record<string, ItemReceiveData> {
  if (!po) return {};

  const data: Record<string, ItemReceiveData> = {};
  po.items.forEach((item) => {
    data[item.id] = {
      id: item.id,
      orderedQty: item.orderedQty,
      receivedQty: item.orderedQty,
      price: item.price,
      receivedPrice: item.price,
      isVerified: true,
    };
  });

  return data;
}

export default function ReceivePurchaseOrderPage() {
  const router = useRouter();
  const params = useParams();
  const { getPurchaseOrder, receivePurchaseOrder } = usePurchaseOrderStore();
  const { suppliers, fetchAllSuppliers } = useSupplierStore();
  const { addToast } = useToastStore();

  const po = getPurchaseOrder(Number(params.id));

  const [receivedDate, setReceivedDate] = useState(() => {
    const now = new Date();
    return now.toISOString().slice(0, 16); // YYYY-MM-DDTHH:mm format for datetime-local
  });
  const [paymentMethod, setPaymentMethod] = useState<PaymentMethod>('cash');
  const [bankAccountId, setBankAccountId] = useState('');
  const [itemsData, setItemsData] = useState<Record<string, ItemReceiveData>>(() => buildInitialItemsData(po));
  const [errors, setErrors] = useState<Record<string, string>>({});
  const [dontShowAgain, setDontShowAgain] = useState(() => getWarningDismissedPreference());
  const [showWarning, setShowWarning] = useState(() => !getWarningDismissedPreference());

  useEffect(() => {
    const loadSuppliers = async () => {
      try {
        await fetchAllSuppliers();
      } catch (error) {
        if (error instanceof ApiError) {
          addToast(error.message, 'error');
        } else {
          addToast('Failed to load suppliers', 'error');
        }
      }
    };

    void loadSuppliers();
  }, [fetchAllSuppliers, addToast]);

  useEffect(() => {
    if (!po) {
      router.push('/transaction/purchase');
      return;
    }

    if (po.status !== 'sent') {
      router.push(`/transaction/purchase/${params.id}`);
    }
  }, [po, router, params.id]);

  if (!po || po.status !== 'sent') {
    return null;
  }

  const supplier = suppliers.find((s) => s.id === po.supplierId);

  const handleQtyChange = (itemId: string, qty: number) => {
    setItemsData((prev) => {
      const item = prev[itemId];
      const orderedQty = item.orderedQty;
      const receivedPrice = item.receivedPrice;
      const isMatch = qty === orderedQty && receivedPrice === item.price;

      return {
        ...prev,
        [itemId]: {
          ...item,
          receivedQty: Math.max(0, qty),
          isVerified: isMatch,
        },
      };
    });
  };

  const handlePriceChange = (itemId: string, price: number) => {
    setItemsData((prev) => {
      const item = prev[itemId];
      const orderedQty = item.orderedQty;
      const receivedQty = item.receivedQty;
      const isMatch = receivedQty === orderedQty && price === item.price;

      return {
        ...prev,
        [itemId]: {
          ...item,
          receivedPrice: Math.max(0, price),
          isVerified: isMatch,
        },
      };
    });
  };

  const handleVerifyToggle = (itemId: string) => {
    setItemsData((prev) => ({
      ...prev,
      [itemId]: {
        ...prev[itemId],
        isVerified: !prev[itemId].isVerified,
      },
    }));
  };

  const handleDismissWarning = () => {
    if (dontShowAgain) {
      localStorage.setItem(WARNING_DISMISS_KEY, 'true');
    }
    setShowWarning(false);
  };

  const validate = (): boolean => {
    const newErrors: Record<string, string> = {};

    if (!receivedDate) {
      newErrors.receivedDate = 'Received date is required';
    }
    if (!paymentMethod) {
      newErrors.paymentMethod = 'Payment method is required';
    }
    if (paymentMethod !== 'cash' && !bankAccountId) {
      newErrors.bankAccountId = 'Bank account is required';
    }

    Object.values(itemsData).forEach((item) => {
      if (item.receivedQty < 0) {
        newErrors.items = 'All received quantities must be >= 0';
      }
      if (item.receivedPrice < 0) {
        newErrors.items = 'All prices must be >= 0';
      }
    });

    setErrors(newErrors);
    return Object.keys(newErrors).length === 0;
  };

  const handleSave = () => {
    if (!validate()) {
      addToast('Please fix the errors before saving', 'error');
      return;
    }

    receivePurchaseOrder(po.id, {
      receivedDate,
      paymentMethod,
      supplierBankAccountId: paymentMethod !== 'cash' ? bankAccountId : undefined,
      items: Object.values(itemsData).map((item) => ({
        id: item.id,
        receivedQty: item.receivedQty,
        receivedPrice: item.receivedPrice,
        isVerified: item.isVerified,
      })),
    });

    addToast('Purchase order received successfully. Stock has been updated.', 'success');
    router.push(`/transaction/purchase/${po.id}`);
  };

  // Calculate summary
  const summary = {
    subtotal: Object.values(itemsData).reduce(
      (sum, item) => sum + item.receivedQty * item.receivedPrice,
      0,
    ),
    totalItems: Object.values(itemsData).reduce((sum, item) => sum + item.receivedQty, 0),
  };

  // Check if all items match
  const allItemsMatch = Object.values(itemsData).every(
    (item) => item.receivedQty === item.orderedQty && item.receivedPrice === item.price,
  );

  // Group items by product
  const itemsByProduct = po.items.reduce((acc, item) => {
    if (!acc[item.productId]) {
      acc[item.productId] = [];
    }
    acc[item.productId].push(item);
    return acc;
  }, {} as Record<number, typeof po.items>);

  const formatCurrency = (amount: number) => {
    return new Intl.NumberFormat('id-ID', {
      style: 'currency',
      currency: 'IDR',
      minimumFractionDigits: 0,
    }).format(amount);
  };

  return (
    <div className="min-h-screen flex flex-col">
      {/* Sticky Header */}
      <div className="sticky top-0 z-10 bg-white border-b border-gray-200 px-6 py-4">
        <div className="flex items-center justify-between">
          <Link
            href={`/transaction/purchase/${po.id}`}
            className="text-sm text-gray-600 hover:text-gray-900 flex items-center gap-1"
          >
            ← Back to {po.poNumber}
          </Link>
          <Button onClick={handleSave}>Save Receive</Button>
        </div>
      </div>

      {/* Form Content */}
      <div className="flex-1 p-6 space-y-6">
        <div>
          <h1 className="text-2xl font-bold text-gray-900">Receive — {po.poNumber}</h1>
          <p className="text-sm text-gray-600">Supplier: {po.supplierName}</p>
        </div>

        {/* Top Section */}
        <div className="bg-white rounded-lg border border-gray-200 p-6 space-y-4">
          <DatePicker
            label="Received Date"
            type="datetime"
            value={receivedDate}
            onChange={setReceivedDate}
            error={errors.receivedDate}
            required
          />

          <Select
            label="Payment Method"
            value={paymentMethod}
            onChange={(e) => {
              setPaymentMethod(e.target.value as PaymentMethod);
              if (e.target.value === 'cash') {
                setBankAccountId('');
              }
            }}
            options={[
              { value: 'cash', label: 'Cash' },
              { value: 'credit_card', label: 'Credit Card' },
              { value: 'bank_transfer', label: 'Bank Transfer' },
            ]}
            error={errors.paymentMethod}
            required
          />

          {paymentMethod !== 'cash' && (
            <Select
              label="Bank Account"
              value={bankAccountId}
              onChange={(e) => setBankAccountId(e.target.value)}
              options={[
                { value: '', label: 'Select account' },
                ...(supplier?.bankAccounts || []).map((acc) => ({
                  value: acc.id,
                  label: `${acc.accountName} - ${acc.accountNumber}`,
                })),
              ]}
              error={errors.bankAccountId}
              required
            />
          )}
        </div>

        {/* Warning Banner */}
        {showWarning && allItemsMatch && (
          <div className="bg-amber-50 border border-amber-200 rounded-lg p-4">
            <div className="flex items-start gap-3">
              <span className="text-amber-600 text-xl">⚠️</span>
              <div className="flex-1">
                <p className="text-sm text-amber-800 mb-2">
                  Received quantity matches ordered quantity.
                </p>
                <label className="flex items-center gap-2 text-sm text-amber-700 cursor-pointer">
                  <Checkbox
                    checked={dontShowAgain}
                    onChange={(checked) => setDontShowAgain(checked)}
                  />
                  I understand, don&apos;t show this message again.
                </label>
              </div>
              <button
                onClick={handleDismissWarning}
                className="text-amber-600 hover:text-amber-800"
              >
                ✕
              </button>
            </div>
          </div>
        )}

        {/* Item Verification */}
        <div className="bg-white rounded-lg border border-gray-200 p-6 space-y-4">
          <h2 className="text-lg font-semibold text-gray-900">Item Verification</h2>

          {errors.items && <p className="text-sm text-red-600">{errors.items}</p>}

          <div className="space-y-4">
            {Object.entries(itemsByProduct).map(([productId, items]) => {
              const firstItem = items[0];

              return (
                <div key={productId} className="border border-gray-200 rounded-lg">
                  <div className="bg-gray-50 px-4 py-2 border-b border-gray-200">
                    <h3 className="font-medium text-gray-900">{firstItem.productName}</h3>
                  </div>

                  <table className="w-full">
                    <thead className="bg-gray-50 border-b border-gray-200">
                      <tr>
                        <th className="px-4 py-2 text-left text-xs font-medium text-gray-600">
                          Variant
                        </th>
                        <th className="px-4 py-2 text-left text-xs font-medium text-gray-600">
                          Unit
                        </th>
                        <th className="px-4 py-2 text-left text-xs font-medium text-gray-600">
                          Ordered
                        </th>
                        <th className="px-4 py-2 text-left text-xs font-medium text-gray-600">
                          Received
                        </th>
                        <th className="px-4 py-2 text-left text-xs font-medium text-gray-600">
                          Price
                        </th>
                        <th className="px-4 py-2 text-left text-xs font-medium text-gray-600">
                          Status
                        </th>
                      </tr>
                    </thead>
                    <tbody>
                      {items.map((item) => {
                        const data = itemsData[item.id];
                        if (!data) return null;

                        const isMatch =
                          data.receivedQty === data.orderedQty &&
                          data.receivedPrice === data.price;
                        const isDisabled =
                          !showWarning && allItemsMatch && dontShowAgain && isMatch;

                        return (
                          <tr key={item.id} className="border-t border-gray-100">
                            <td className="px-4 py-2 text-sm text-gray-700">
                              {item.variantLabel}
                            </td>
                            <td className="px-4 py-2 text-sm text-gray-600">{item.unitName || '—'}</td>
                            <td className="px-4 py-2 text-sm text-gray-600">
                              {data.orderedQty}
                            </td>
                            <td className="px-4 py-2">
                              <input
                                type="number"
                                min="0"
                                value={data.receivedQty}
                                onChange={(e) =>
                                  handleQtyChange(item.id, parseInt(e.target.value) || 0)
                                }
                                disabled={isDisabled}
                                className="w-24 rounded-md border border-gray-300 px-2 py-1 text-sm disabled:bg-gray-100 disabled:cursor-not-allowed"
                              />
                            </td>
                            <td className="px-4 py-2">
                              <input
                                type="number"
                                min="0"
                                value={data.receivedPrice}
                                onChange={(e) =>
                                  handlePriceChange(item.id, parseInt(e.target.value) || 0)
                                }
                                disabled={isDisabled}
                                className="w-28 rounded-md border border-gray-300 px-2 py-1 text-sm disabled:bg-gray-100 disabled:cursor-not-allowed"
                              />
                            </td>
                            <td className="px-4 py-2">
                              {isMatch ? (
                                <label className="flex items-center gap-1 text-sm text-green-700 cursor-pointer">
                                  <Checkbox
                                    checked={data.isVerified}
                                    onChange={() => handleVerifyToggle(item.id)}
                                  />
                                  OK
                                </label>
                              ) : (
                                <span className="text-sm text-amber-600">⚠ Mismatch</span>
                              )}
                            </td>
                          </tr>
                        );
                      })}
                    </tbody>
                  </table>
                </div>
              );
            })}
          </div>
        </div>

        {/* Summary */}
        <div className="bg-white rounded-lg border border-gray-200 p-6">
          <div className="space-y-2 text-sm">
            <div className="flex justify-between">
              <span className="text-gray-600">Subtotal:</span>
              <span className="font-medium text-gray-900">{formatCurrency(summary.subtotal)}</span>
            </div>
            <div className="flex justify-between">
              <span className="text-gray-600">Total Items:</span>
              <span className="font-medium text-gray-900">{summary.totalItems}</span>
            </div>
            <div className="flex justify-between border-t border-gray-200 pt-2">
              <span className="text-gray-900 font-semibold">Total Price:</span>
              <span className="font-semibold text-gray-900">{formatCurrency(summary.subtotal)}</span>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}
