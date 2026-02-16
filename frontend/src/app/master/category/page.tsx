'use client';

import React, { useState, useEffect, useCallback } from 'react';
import AdminLayout from '@/components/layout/AdminLayout';
import Button from '@/components/ui/Button';
import Input from '@/components/ui/Input';
import Table from '@/components/ui/Table';
import type { SortDirection } from '@/components/ui/Table';
import Modal from '@/components/ui/Modal';
import { useCategoryStore, Category } from '@/stores/useCategoryStore';
import { useToastStore } from '@/stores/useToastStore';
import { ApiError } from '@/lib/api';

const DEFAULT_PAGE_SIZE = 10;

export default function MasterCategoryPage() {
  const { fetchCategories, createCategory, updateCategory, deleteCategory } = useCategoryStore();
  const { addToast } = useToastStore();

  const [categories, setCategories] = useState<Category[]>([]);
  const [loading, setLoading] = useState(true);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [totalItems, setTotalItems] = useState(0);
  const [totalPages, setTotalPages] = useState(1);

  const [search, setSearch] = useState('');
  const [debouncedSearch, setDebouncedSearch] = useState('');
  const [currentPage, setCurrentPage] = useState(1);
  const [pageSize, setPageSize] = useState(DEFAULT_PAGE_SIZE);
  const [sortKey, setSortKey] = useState<string | null>(null);
  const [sortDirection, setSortDirection] = useState<SortDirection>(null);

  // Modal state
  const [isFormOpen, setIsFormOpen] = useState(false);
  const [isDeleteOpen, setIsDeleteOpen] = useState(false);
  const [editingCategory, setEditingCategory] = useState<Category | null>(null);
  const [deletingCategory, setDeletingCategory] = useState<Category | null>(null);

  // Form state
  const [formName, setFormName] = useState('');
  const [formDescription, setFormDescription] = useState('');
  const [formErrors, setFormErrors] = useState<Record<string, string>>({});

  useEffect(() => {
    const timer = setTimeout(() => {
      setDebouncedSearch(search);
      setCurrentPage(1);
    }, 300);

    return () => clearTimeout(timer);
  }, [search]);

  const loadCategories = useCallback(async () => {
    setLoading(true);
    try {
      const response = await fetchCategories({
        page: currentPage,
        pageSize,
        search: debouncedSearch || undefined,
        sortBy: sortKey || undefined,
        sortDir: sortDirection || undefined,
      });

      setCategories(response.data);
      setTotalItems(response.meta.totalItems);
      setTotalPages(response.meta.totalPages || 1);
    } catch (error) {
      if (error instanceof ApiError) {
        addToast(error.message, 'error');
      } else {
        addToast('Failed to load categories', 'error');
      }
    } finally {
      setLoading(false);
    }
  }, [currentPage, pageSize, debouncedSearch, sortKey, sortDirection, fetchCategories, addToast]);

  useEffect(() => {
    loadCategories();
  }, [loadCategories]);

  const handleSort = (key: string, direction: SortDirection) => {
    setSortKey(direction === null ? null : key);
    setSortDirection(direction);
    setCurrentPage(1);
  };

  const handlePageSizeChange = (size: number) => {
    setPageSize(size);
    setCurrentPage(1);
  };

  const openAddModal = () => {
    setEditingCategory(null);
    setFormName('');
    setFormDescription('');
    setFormErrors({});
    setIsFormOpen(true);
  };

  const openEditModal = (category: Category) => {
    setEditingCategory(category);
    setFormName(category.name);
    setFormDescription(category.description);
    setFormErrors({});
    setIsFormOpen(true);
  };

  const openDeleteModal = (category: Category) => {
    setDeletingCategory(category);
    setIsDeleteOpen(true);
  };

  const validateForm = () => {
    const errors: Record<string, string> = {};
    if (!formName.trim()) errors.name = 'Name is required';
    setFormErrors(errors);
    return Object.keys(errors).length === 0;
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!validateForm()) return;

    setIsSubmitting(true);
    try {
      if (editingCategory) {
        await updateCategory(editingCategory.id, {
          name: formName.trim(),
          description: formDescription.trim(),
        });
        addToast('Category updated successfully', 'success');
      } else {
        await createCategory({
          name: formName.trim(),
          description: formDescription.trim(),
        });
        addToast('Category created successfully', 'success');
      }

      setIsFormOpen(false);
      await loadCategories();
    } catch (error) {
      if (error instanceof ApiError) {
        addToast(error.message, 'error');
      } else {
        addToast('Failed to save category', 'error');
      }
    } finally {
      setIsSubmitting(false);
    }
  };

  const handleDelete = async () => {
    if (!deletingCategory) return;

    setIsSubmitting(true);
    try {
      await deleteCategory(deletingCategory.id);
      addToast('Category deleted successfully', 'success');

      if (currentPage > 1 && categories.length === 1) {
        setCurrentPage((prev) => prev - 1);
      } else {
        await loadCategories();
      }
    } catch (error) {
      if (error instanceof ApiError) {
        addToast(error.message, 'error');
      } else {
        addToast('Failed to delete category', 'error');
      }
    } finally {
      setIsSubmitting(false);
      setIsDeleteOpen(false);
    }
  };

  const columns = [
    { key: 'id', label: 'ID', sortable: true },
    { key: 'name', label: 'Name', sortable: true },
    { key: 'description', label: 'Description', sortable: true },
    {
      key: 'actions',
      label: 'Actions',
      render: (item: Category) => (
        <div className="flex gap-2">
          <Button size="sm" variant="outline" onClick={() => openEditModal(item)}>
            Edit
          </Button>
          <Button size="sm" variant="danger" onClick={() => openDeleteModal(item)}>
            Delete
          </Button>
        </div>
      ),
    },
  ];

  if (loading) {
    return (
      <AdminLayout>
        <div className="flex items-center justify-center py-20">
          <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600"></div>
        </div>
      </AdminLayout>
    );
  }

  return (
    <AdminLayout>
      <div className="space-y-6">
        <div className="flex items-center justify-between">
          <h1 className="text-2xl font-bold text-gray-900">Master Category</h1>
          <Button onClick={openAddModal}>Add Category</Button>
        </div>

        <div className="bg-white rounded-lg border border-gray-200 shadow-sm">
          <div className="p-4 border-b border-gray-200">
            <Input
              placeholder="Search categories..."
              value={search}
              onChange={(e) => setSearch(e.target.value)}
              className="max-w-sm"
            />
          </div>
          <Table
            columns={columns}
            data={categories}
            currentPage={currentPage}
            totalPages={totalPages}
            onPageChange={setCurrentPage}
            sortKey={sortKey}
            sortDirection={sortDirection}
            onSort={handleSort}
            pageSize={pageSize}
            onPageSizeChange={handlePageSizeChange}
            totalItems={totalItems}
          />
        </div>
      </div>

      <Modal
        isOpen={isFormOpen}
        onClose={() => setIsFormOpen(false)}
        title={editingCategory ? 'Edit Category' : 'Add Category'}
      >
        <form onSubmit={handleSubmit} className="space-y-4">
          <Input
            label="Name"
            placeholder="Category name"
            value={formName}
            onChange={(e) => {
              setFormName(e.target.value);
              setFormErrors((prev) => ({ ...prev, name: '' }));
            }}
            error={formErrors.name}
            required
          />
          <Input
            label="Description"
            placeholder="Category description (optional)"
            value={formDescription}
            onChange={(e) => setFormDescription(e.target.value)}
            error={formErrors.description}
          />
          <div className="flex justify-end gap-2 pt-2">
            <Button
              type="button"
              variant="outline"
              onClick={() => setIsFormOpen(false)}
              disabled={isSubmitting}
            >
              Cancel
            </Button>
            <Button type="submit" disabled={isSubmitting}>
              {editingCategory ? 'Update' : 'Add'}
            </Button>
          </div>
        </form>
      </Modal>

      <Modal
        isOpen={isDeleteOpen}
        onClose={() => setIsDeleteOpen(false)}
        title="Delete Category"
      >
        <p className="text-sm text-gray-600 mb-6">
          Are you sure you want to delete{' '}
          <strong>{deletingCategory?.name}</strong>? This action cannot be
          undone.
        </p>
        <div className="flex justify-end gap-2">
          <Button variant="outline" onClick={() => setIsDeleteOpen(false)} disabled={isSubmitting}>
            Cancel
          </Button>
          <Button variant="danger" onClick={handleDelete} disabled={isSubmitting}>
            Delete
          </Button>
        </div>
      </Modal>
    </AdminLayout>
  );
}
