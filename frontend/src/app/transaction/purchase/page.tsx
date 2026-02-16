'use client';

import React, { useState, useMemo } from 'react';
import { useRouter } from 'next/navigation';
import AdminLayout from '@/components/layout/AdminLayout';
import Button from '@/components/ui/Button';
import StatusBadge from '@/components/ui/StatusBadge';
import ConfirmModal from '@/components/ui/ConfirmModal';
import { usePurchaseOrderStore, POStatus, PurchaseOrder } from '@/stores/usePurchaseOrderStore';
import { useToastStore } from '@/stores/useToastStore';

const STATUS_COLORS: Record<POStatus, 'green' | 'blue' | 'yellow' | 'amber' | 'gray' | 'red'> = {
  draft: 'gray',
  sent: 'blue',
  received: 'amber',
  completed: 'green',
  cancelled: 'red',
};

type StatusFilter = 'all' | POStatus;

export default function PurchaseOrderListPage() {
  const router = useRouter();
  const { purchaseOrders, deletePurchaseOrder } = usePurchaseOrderStore();
  const { addToast } = useToastStore();

  const [statusFilter, setStatusFilter] = useState<StatusFilter>('all');
  const [searchQuery, setSearchQuery] = useState('');
  const [currentPage, setCurrentPage] = useState(1);
  const [itemsPerPage] = useState(10);
  const [deleteModal, setDeleteModal] = useState<{ isOpen: boolean; poId: number | null; poNumber: string }>({
    isOpen: false,
    poId: null,
    poNumber: '',
  });

  // Filter and search
  const filteredOrders = useMemo(() => {
    let filtered = purchaseOrders;

    // Status filter
    if (statusFilter !== 'all') {
      filtered = filtered.filter((po) => po.status === statusFilter);
    }

    // Search filter
    if (searchQuery.trim()) {
      const query = searchQuery.toLowerCase();
      filtered = filtered.filter(
        (po) =>
          po.poNumber.toLowerCase().includes(query) ||
          po.supplierName.toLowerCase().includes(query)
      );
    }

    // Sort by date (newest first)
    return filtered.sort((a, b) => new Date(b.date).getTime() - new Date(a.date).getTime());
  }, [purchaseOrders, statusFilter, searchQuery]);

  // Pagination
  const totalPages = Math.ceil(filteredOrders.length / itemsPerPage);
  const paginatedOrders = filteredOrders.slice(
    (currentPage - 1) * itemsPerPage,
    currentPage * itemsPerPage
  );

  // Status counts
  const statusCounts = useMemo(() => {
    const counts: Record<StatusFilter, number> = {
      all: purchaseOrders.length,
      draft: 0,
      sent: 0,
      received: 0,
      completed: 0,
      cancelled: 0,
    };

    purchaseOrders.forEach((po) => {
      counts[po.status]++;
    });

    return counts;
  }, [purchaseOrders]);

  const handleDelete = () => {
    if (deleteModal.poId) {
      deletePurchaseOrder(deleteModal.poId);
      addToast(`Purchase order ${deleteModal.poNumber} has been deleted.`, 'success');
      setDeleteModal({ isOpen: false, poId: null, poNumber: '' });
    }
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

  const getActionButtons = (po: PurchaseOrder) => {
    const actions = [];

    actions.push(
      <Button
        key="view"
        size="sm"
        variant="outline"
        onClick={() => router.push(`/transaction/purchase/${po.id}`)}
      >
        View
      </Button>
    );

    if (po.status === 'draft') {
      actions.push(
        <Button
          key="edit"
          size="sm"
          variant="outline"
          onClick={() => router.push(`/transaction/purchase/${po.id}/edit`)}
        >
          Edit
        </Button>,
        <Button
          key="delete"
          size="sm"
          variant="danger"
          onClick={() => setDeleteModal({ isOpen: true, poId: po.id, poNumber: po.poNumber })}
        >
          Delete
        </Button>
      );
    }

    if (po.status === 'sent') {
      actions.push(
        <Button
          key="receive"
          size="sm"
          variant="primary"
          onClick={() => router.push(`/transaction/purchase/${po.id}/receive`)}
        >
          Receive
        </Button>
      );
    }

    if (po.status === 'received') {
      actions.push(
        <Button
          key="complete"
          size="sm"
          variant="primary"
          onClick={() => router.push(`/transaction/purchase/${po.id}`)}
        >
          Complete
        </Button>
      );
    }

    return actions;
  };

  return (
    <AdminLayout>
      <div className="space-y-6">
        {/* Header */}
        <div className="flex items-center justify-between">
          <h1 className="text-2xl font-bold text-gray-900">Purchase Orders</h1>
          <Button onClick={() => router.push('/transaction/purchase/add')}>
            + New Order
          </Button>
        </div>

        {/* Status Tabs */}
        <div className="flex gap-2 border-b border-gray-200">
          {(['all', 'draft', 'sent', 'received', 'completed', 'cancelled'] as StatusFilter[]).map(
            (status) => (
              <button
                key={status}
                onClick={() => {
                  setStatusFilter(status);
                  setCurrentPage(1);
                }}
                className={`px-4 py-2 text-sm font-medium capitalize transition-colors ${
                  statusFilter === status
                    ? 'border-b-2 border-blue-600 text-blue-600'
                    : 'text-gray-600 hover:text-gray-900'
                }`}
              >
                {status} ({statusCounts[status]})
              </button>
            )
          )}
        </div>

        {/* Search */}
        <div>
          <input
            type="text"
            placeholder="🔍 Search by PO number or supplier..."
            value={searchQuery}
            onChange={(e) => {
              setSearchQuery(e.target.value);
              setCurrentPage(1);
            }}
            className="w-full max-w-md rounded-md border border-gray-300 px-4 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
          />
        </div>

        {/* Cards */}
        <div className="space-y-4">
          {paginatedOrders.length === 0 ? (
            <div className="text-center py-12 text-gray-500">
              {searchQuery ? 'No purchase orders found matching your search.' : 'No purchase orders yet.'}
            </div>
          ) : (
            paginatedOrders.map((po) => {
              const itemCount = po.items.length;
              const total = po.subtotal || po.items.reduce((sum, item) => sum + item.orderedQty * item.price, 0);

              return (
                <div
                  key={po.id}
                  className="bg-white rounded-lg border border-gray-200 shadow-sm p-6"
                >
                  <div className="flex items-start justify-between mb-3">
                    <h3 className="text-lg font-bold text-gray-900">{po.poNumber}</h3>
                    <StatusBadge status={po.status} colorMap={STATUS_COLORS} />
                  </div>
                  <p className="text-gray-700 font-medium mb-2">{po.supplierName}</p>
                  <p className="text-sm text-gray-600 mb-4">
                    {formatDate(po.date)} · {itemCount} item{itemCount !== 1 ? 's' : ''} ·{' '}
                    {formatCurrency(total)}
                  </p>
                  <div className="flex justify-end gap-2">
                    {getActionButtons(po)}
                  </div>
                </div>
              );
            })
          )}
        </div>

        {/* Pagination */}
        {filteredOrders.length > 0 && (
          <div className="flex items-center justify-between">
            <p className="text-sm text-gray-600">
              Showing {(currentPage - 1) * itemsPerPage + 1}-
              {Math.min(currentPage * itemsPerPage, filteredOrders.length)} of{' '}
              {filteredOrders.length} order{filteredOrders.length !== 1 ? 's' : ''}
            </p>
            <div className="flex gap-2">
              <Button
                size="sm"
                variant="outline"
                onClick={() => setCurrentPage((p) => Math.max(1, p - 1))}
                disabled={currentPage === 1}
              >
                &lt; Prev
              </Button>
              <Button
                size="sm"
                variant="outline"
                onClick={() => setCurrentPage((p) => Math.min(totalPages, p + 1))}
                disabled={currentPage === totalPages}
              >
                Next &gt;
              </Button>
            </div>
          </div>
        )}
      </div>

      {/* Delete Confirmation Modal */}
      <ConfirmModal
        isOpen={deleteModal.isOpen}
        onClose={() => setDeleteModal({ isOpen: false, poId: null, poNumber: '' })}
        onConfirm={handleDelete}
        title="Delete Purchase Order"
        message={`Are you sure you want to delete ${deleteModal.poNumber}? This action cannot be undone.`}
        confirmLabel="Delete"
        variant="danger"
      />
    </AdminLayout>
  );
}
