'use client';

import React, { useState, useMemo, useEffect, useCallback } from 'react';
import { useRouter } from 'next/navigation';
import AdminLayout from '@/components/layout/AdminLayout';
import Button from '@/components/ui/Button';
import Input from '@/components/ui/Input';
import Table from '@/components/ui/Table';
import type { SortDirection } from '@/components/ui/Table';
import Modal from '@/components/ui/Modal';
import { useProductStore, Product } from '@/stores/useProductStore';
import { useCategoryStore } from '@/stores/useCategoryStore';
import { useToastStore } from '@/stores/useToastStore';
import { ApiError } from '@/lib/api';

const DEFAULT_PAGE_SIZE = 10;

export default function MasterProductPage() {
  const router = useRouter();
  const { products, deleteProduct } = useProductStore();
  const { categories, fetchAllCategories } = useCategoryStore();
  const { addToast } = useToastStore();

  const [search, setSearch] = useState('');
  const [currentPage, setCurrentPage] = useState(1);
  const [pageSize, setPageSize] = useState(DEFAULT_PAGE_SIZE);
  const [sortKey, setSortKey] = useState<string | null>(null);
  const [sortDirection, setSortDirection] = useState<SortDirection>(null);

  // Delete modal state
  const [isDeleteOpen, setIsDeleteOpen] = useState(false);
  const [deletingProduct, setDeletingProduct] = useState<Product | null>(null);

  useEffect(() => {
    const loadCategories = async () => {
      try {
        await fetchAllCategories();
      } catch (error) {
        if (error instanceof ApiError) {
          addToast(error.message, 'error');
        } else {
          addToast('Failed to load categories', 'error');
        }
      }
    };

    void loadCategories();
  }, [fetchAllCategories, addToast]);

  // Build a lookup map for category names
  const categoryMap = useMemo(() => {
    const map = new Map<number, string>();
    categories.forEach((c) => map.set(c.id, c.name));
    return map;
  }, [categories]);

  const getCategoryName = useCallback(
    (categoryId: number) => categoryMap.get(categoryId) ?? 'Unknown',
    [categoryMap],
  );

  const filtered = useMemo(() => {
    if (!search) return products;
    const q = search.toLowerCase();
    return products.filter((p) => p.name.toLowerCase().includes(q));
  }, [products, search]);

  const sorted = useMemo(() => {
    if (!sortKey || !sortDirection) return filtered;
    return [...filtered].sort((a, b) => {
      let aVal: string | number;
      let bVal: string | number;

      switch (sortKey) {
        case 'id':
          aVal = a.id;
          bVal = b.id;
          break;
        case 'name':
          aVal = a.name.toLowerCase();
          bVal = b.name.toLowerCase();
          break;
        case 'category':
          aVal = getCategoryName(a.categoryId).toLowerCase();
          bVal = getCategoryName(b.categoryId).toLowerCase();
          break;
        default:
          return 0;
      }

      if (typeof aVal === 'number' && typeof bVal === 'number') {
        return sortDirection === 'asc' ? aVal - bVal : bVal - aVal;
      }
      if (aVal < bVal) return sortDirection === 'asc' ? -1 : 1;
      if (aVal > bVal) return sortDirection === 'asc' ? 1 : -1;
      return 0;
    });
  }, [filtered, sortKey, sortDirection, getCategoryName]);

  const totalPages = Math.max(1, Math.ceil(sorted.length / pageSize));
  const paginated = sorted.slice(
    (currentPage - 1) * pageSize,
    currentPage * pageSize
  );

  const handleSearch = (value: string) => {
    setSearch(value);
    setCurrentPage(1);
  };

  const handleSort = (key: string, direction: SortDirection) => {
    setSortKey(direction === null ? null : key);
    setSortDirection(direction);
    setCurrentPage(1);
  };

  const handlePageSizeChange = (size: number) => {
    setPageSize(size);
    setCurrentPage(1);
  };

  const openDeleteModal = (product: Product) => {
    setDeletingProduct(product);
    setIsDeleteOpen(true);
  };

  const handleDelete = () => {
    if (deletingProduct) {
      deleteProduct(deletingProduct.id);
      addToast('Product deleted successfully', 'success');
      const newTotal = Math.max(1, Math.ceil((filtered.length - 1) / pageSize));
      if (currentPage > newTotal) setCurrentPage(newTotal);
    }
    setIsDeleteOpen(false);
  };

  const columns = [
    { key: 'id', label: 'ID', sortable: true },
    {
      key: 'image',
      label: 'Image',
      render: (item: Product) => {
        const src = item.images?.[0];
        return src ? (
          <img
            src={src}
            alt={item.name}
            className="w-10 h-10 rounded object-cover"
          />
        ) : (
          <div className="w-10 h-10 rounded bg-gray-200 flex items-center justify-center">
            <span className="text-gray-400 text-xs">N/A</span>
          </div>
        );
      },
    },
    { key: 'name', label: 'Name', sortable: true },
    {
      key: 'category',
      label: 'Category',
      sortable: true,
      render: (item: Product) => getCategoryName(item.categoryId),
    },
    {
      key: 'status',
      label: 'Status',
      render: (item: Product) => (
        <span
          className={`inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium ${
            item.status === 'active'
              ? 'bg-green-100 text-green-800'
              : 'bg-gray-100 text-gray-800'
          }`}
        >
          {item.status === 'active' ? 'Active' : 'Inactive'}
        </span>
      ),
    },
    {
      key: 'actions',
      label: 'Actions',
      render: (item: Product) => (
        <div className="flex gap-2">
          <Button
            size="sm"
            variant="outline"
            onClick={() => router.push(`/master/product/edit/${item.id}`)}
          >
            Edit
          </Button>
          <Button
            size="sm"
            variant="danger"
            onClick={() => openDeleteModal(item)}
          >
            Delete
          </Button>
        </div>
      ),
    },
  ];

  return (
    <AdminLayout>
      <div className="space-y-6">
        <div className="flex items-center justify-between">
          <h1 className="text-2xl font-bold text-gray-900">Master Product</h1>
          <Button onClick={() => router.push('/master/product/add')}>
            Add Product
          </Button>
        </div>

        <div className="bg-white rounded-lg border border-gray-200 shadow-sm">
          <div className="p-4 border-b border-gray-200">
            <Input
              placeholder="Search products..."
              value={search}
              onChange={(e) => handleSearch(e.target.value)}
              className="max-w-sm"
            />
          </div>
          <Table
            columns={columns}
            data={paginated}
            currentPage={currentPage}
            totalPages={totalPages}
            onPageChange={setCurrentPage}
            sortKey={sortKey}
            sortDirection={sortDirection}
            onSort={handleSort}
            pageSize={pageSize}
            onPageSizeChange={handlePageSizeChange}
            totalItems={sorted.length}
          />
        </div>
      </div>

      {/* Delete Confirmation Modal */}
      <Modal
        isOpen={isDeleteOpen}
        onClose={() => setIsDeleteOpen(false)}
        title="Delete Product"
      >
        <p className="text-sm text-gray-600 mb-6">
          Are you sure you want to delete{' '}
          <strong>{deletingProduct?.name}</strong>? This action cannot be undone.
        </p>
        <div className="flex justify-end gap-2">
          <Button variant="outline" onClick={() => setIsDeleteOpen(false)}>
            Cancel
          </Button>
          <Button variant="danger" onClick={handleDelete}>
            Delete
          </Button>
        </div>
      </Modal>
    </AdminLayout>
  );
}
