'use client';

import React, { useState } from 'react';
import { useRouter, useParams } from 'next/navigation';
import Link from 'next/link';
import AdminLayout from '@/components/layout/AdminLayout';
import Button from '@/components/ui/Button';
import StatusBadge from '@/components/ui/StatusBadge';
import ConfirmModal from '@/components/ui/ConfirmModal';
import { usePurchaseOrderStore, POStatus } from '@/stores/usePurchaseOrderStore';
import { useToastStore } from '@/stores/useToastStore';

const STATUS_COLORS: Record<POStatus, 'green' | 'blue' | 'yellow' | 'amber' | 'gray' | 'red'> = {
  draft: 'gray',
  sent: 'blue',
  received: 'amber',
  completed: 'green',
  cancelled: 'red',
};

type ConfirmAction = 'send' | 'complete' | 'cancel' | null;

export default function PurchaseOrderDetailPage() {
  const router = useRouter();
  const params = useParams();
  const { getPurchaseOrder, updateStatus, completePurchaseOrder, cancelPurchaseOrder } = usePurchaseOrderStore();
  const { addToast } = useToastStore();

  const po = getPurchaseOrder(Number(params.id));
  const [expandedProducts, setExpandedProducts] = useState<Set<number>>(new Set());
  const [confirmAction, setConfirmAction] = useState<ConfirmAction>(null);

  if (!po) {
    return (
      <AdminLayout>
        <div className="text-center py-12">
          <p className="text-gray-500">Purchase order not found.</p>
          <Link href="/transaction/purchase" className="text-blue-600 hover:underline mt-4 inline-block">
            ← Back to Purchase Orders
          </Link>
        </div>
      </AdminLayout>
    );
  }

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

  const handleMarkAsSent = () => {
    updateStatus(po.id, 'sent');
    addToast('Purchase order marked as sent', 'success');
    setConfirmAction(null);
  };

  const handleComplete = () => {
    completePurchaseOrder(po.id);
    addToast('Purchase order marked as completed', 'success');
    setConfirmAction(null);
  };

  const handleCancel = () => {
    cancelPurchaseOrder(po.id);
    addToast('Purchase order has been cancelled', 'success');
    setConfirmAction(null);
  };

  const formatCurrency = (amount: number) => {
    return new Intl.NumberFormat('id-ID', {
      style: 'currency',
      currency: 'IDR',
      minimumFractionDigits: 0,
    }).format(amount);
  };

  const formatDate = (dateString: string) => {
    const date = new Date(dateString);
    return new Intl.DateTimeFormat('id-ID', {
      day: 'numeric',
      month: 'short',
      year: 'numeric',
    }).format(date);
  };

  const formatDateTime = (dateString: string) => {
    const date = new Date(dateString);
    return new Intl.DateTimeFormat('id-ID', {
      day: 'numeric',
      month: 'short',
      year: 'numeric',
      hour: '2-digit',
      minute: '2-digit',
    }).format(date);
  };

  // Group items by product
  const itemsByProduct = po.items.reduce((acc, item) => {
    if (!acc[item.productId]) {
      acc[item.productId] = [];
    }
    acc[item.productId].push(item);
    return acc;
  }, {} as Record<number, typeof po.items>);

  const totalItems = po.items.reduce((sum, item) => sum + item.orderedQty, 0);
  const totalPrice = po.items.reduce((sum, item) => sum + item.orderedQty * item.price, 0);

  return (
    <AdminLayout>
      <div className="space-y-6">
        {/* Header */}
        <div className="flex items-center justify-between">
          <Link
            href="/transaction/purchase"
            className="text-sm text-gray-600 hover:text-gray-900 flex items-center gap-1"
          >
            ← Back to Purchase Orders
          </Link>
        </div>

        {/* PO Info */}
        <div className="bg-white rounded-lg border border-gray-200 p-6">
          <div className="flex items-start justify-between mb-4">
            <h1 className="text-2xl font-bold text-gray-900">{po.poNumber}</h1>
            <StatusBadge status={po.status} colorMap={STATUS_COLORS} size="md" />
          </div>

          <div className="space-y-2 text-sm">
            <p>
              <span className="text-gray-600">Supplier:</span>{' '}
              <span className="font-medium text-gray-900">{po.supplierName}</span>
            </p>
            <p>
              <span className="text-gray-600">Date:</span>{' '}
              <span className="text-gray-900">{formatDate(po.date)}</span>
            </p>
            {po.notes && (
              <p>
                <span className="text-gray-600">Notes:</span>{' '}
                <span className="text-gray-900">{po.notes}</span>
              </p>
            )}
          </div>
        </div>

        {/* Order Items */}
        <div className="bg-white rounded-lg border border-gray-200 p-6">
          <h2 className="text-lg font-semibold text-gray-900 mb-4">Order Items</h2>

          <div className="space-y-4">
            {Object.entries(itemsByProduct).map(([productId, items]) => {
              const isExpanded = expandedProducts.has(Number(productId));
              const firstItem = items[0];

              return (
                <div key={productId} className="border border-gray-200 rounded-lg">
                  <div
                    className="flex items-center justify-between p-4 cursor-pointer hover:bg-gray-50"
                    onClick={() => toggleProduct(Number(productId))}
                  >
                    <div className="flex items-center gap-2">
                      <span className="text-lg">{isExpanded ? '▼' : '▶'}</span>
                      <h3 className="font-medium text-gray-900">{firstItem.productName}</h3>
                    </div>
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
                              Unit
                            </th>
                            <th className="px-4 py-2 text-left text-xs font-medium text-gray-600">
                              Ordered
                            </th>
                            <th className="px-4 py-2 text-left text-xs font-medium text-gray-600">
                              Price
                            </th>
                            {po.status === 'received' || po.status === 'completed' ? (
                              <>
                                <th className="px-4 py-2 text-left text-xs font-medium text-gray-600">
                                  Received
                                </th>
                                <th className="px-4 py-2 text-left text-xs font-medium text-gray-600">
                                  Actual Price
                                </th>
                              </>
                            ) : null}
                          </tr>
                        </thead>
                        <tbody>
                          {items.map((item) => (
                            <tr key={item.id} className="border-t border-gray-100">
                              <td className="px-4 py-2 text-sm text-gray-700">{item.variantLabel}</td>
                              <td className="px-4 py-2 text-sm text-gray-600">{item.sku}</td>
                              <td className="px-4 py-2 text-sm text-gray-600">{item.unitName || '—'}</td>
                              <td className="px-4 py-2 text-sm text-gray-600">{item.orderedQty}</td>
                              <td className="px-4 py-2 text-sm text-gray-600">
                                {formatCurrency(item.price)}
                              </td>
                              {po.status === 'received' || po.status === 'completed' ? (
                                <>
                                  <td className="px-4 py-2 text-sm text-gray-600">
                                    {item.receivedQty}{' '}
                                    {item.receivedQty !== item.orderedQty && (
                                      <span className="text-amber-600">⚠</span>
                                    )}
                                  </td>
                                  <td className="px-4 py-2 text-sm text-gray-600">
                                    {formatCurrency(item.receivedPrice || item.price)}
                                  </td>
                                </>
                              ) : null}
                            </tr>
                          ))}
                        </tbody>
                      </table>
                    </div>
                  )}
                </div>
              );
            })}
          </div>

          <div className="border-t border-gray-200 mt-4 pt-4">
            <p className="text-sm font-medium text-gray-900 text-right">
              Total: {totalItems} items · {formatCurrency(totalPrice)}
            </p>
          </div>
        </div>

        {/* Receive Information (only shown when status >= received) */}
        {(po.status === 'received' || po.status === 'completed') && (
          <div className="bg-white rounded-lg border border-gray-200 p-6">
            <h2 className="text-lg font-semibold text-gray-900 mb-4">Receive Information</h2>
            <div className="space-y-2 text-sm">
              <p>
                <span className="text-gray-600">Received:</span>{' '}
                <span className="text-gray-900">{po.receivedDate ? formatDateTime(po.receivedDate) : '—'}</span>
              </p>
              <p>
                <span className="text-gray-600">Payment:</span>{' '}
                <span className="text-gray-900">
                  {po.paymentMethod === 'cash' && 'Cash'}
                  {po.paymentMethod === 'credit_card' && 'Credit Card'}
                  {po.paymentMethod === 'bank_transfer' && `Bank Transfer${po.supplierBankAccountId ? ' → ' + po.supplierBankAccountId : ''}`}
                </span>
              </p>
              <p className="font-medium">
                <span className="text-gray-600">Total Received:</span>{' '}
                <span className="text-gray-900">
                  {po.totalItems} items · {formatCurrency(po.subtotal || 0)}
                </span>
              </p>
            </div>
          </div>
        )}

        {/* Actions */}
        <div className="bg-white rounded-lg border border-gray-200 p-6">
          <h2 className="text-lg font-semibold text-gray-900 mb-4">Actions</h2>
          <div className="flex gap-2">
            {po.status === 'draft' && (
              <>
                <Button onClick={() => router.push(`/transaction/purchase/${po.id}/edit`)}>
                  Edit
                </Button>
                <Button variant="primary" onClick={() => setConfirmAction('send')}>
                  Mark as Sent
                </Button>
                <Button variant="danger" onClick={() => setConfirmAction('cancel')}>
                  Cancel Order
                </Button>
              </>
            )}

            {po.status === 'sent' && (
              <>
                <Button onClick={() => router.push(`/transaction/purchase/${po.id}/receive`)}>
                  Receive
                </Button>
                <Button variant="danger" onClick={() => setConfirmAction('cancel')}>
                  Cancel Order
                </Button>
              </>
            )}

            {po.status === 'received' && (
              <Button variant="primary" onClick={() => setConfirmAction('complete')}>
                Mark as Completed
              </Button>
            )}

            {(po.status === 'completed' || po.status === 'cancelled') && (
              <p className="text-sm text-gray-500">No actions available for this status.</p>
            )}
          </div>
        </div>
      </div>

      {/* Confirmation Modals */}
      <ConfirmModal
        isOpen={confirmAction === 'send'}
        onClose={() => setConfirmAction(null)}
        onConfirm={handleMarkAsSent}
        title="Send Purchase Order"
        message="Mark this PO as sent to the supplier?"
        confirmLabel="Send"
      />

      <ConfirmModal
        isOpen={confirmAction === 'complete'}
        onClose={() => setConfirmAction(null)}
        onConfirm={handleComplete}
        title="Complete Purchase Order"
        message="Mark this PO as completed? This action cannot be undone."
        confirmLabel="Complete"
      />

      <ConfirmModal
        isOpen={confirmAction === 'cancel'}
        onClose={() => setConfirmAction(null)}
        onConfirm={handleCancel}
        title="Cancel Purchase Order"
        message={`Are you sure you want to cancel ${po.poNumber}? This action cannot be undone.`}
        confirmLabel="Cancel Order"
        variant="danger"
      />
    </AdminLayout>
  );
}
