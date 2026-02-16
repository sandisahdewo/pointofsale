'use client';

import React, { useState, useEffect, useCallback } from 'react';
import AdminLayout from '@/components/layout/AdminLayout';
import Button from '@/components/ui/Button';
import Input from '@/components/ui/Input';
import Table from '@/components/ui/Table';
import type { SortDirection } from '@/components/ui/Table';
import Modal from '@/components/ui/Modal';
import Toggle from '@/components/ui/Toggle';
import Badge from '@/components/ui/Badge';
import { useRackStore, Rack } from '@/stores/useRackStore';
import { useToastStore } from '@/stores/useToastStore';
import { ApiError } from '@/lib/api';

const DEFAULT_PAGE_SIZE = 10;

export default function MasterRackPage() {
  const { fetchRacks, createRack, updateRack, deleteRack } = useRackStore();
  const { addToast } = useToastStore();

  const [racks, setRacks] = useState<Rack[]>([]);
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
  const [editingRack, setEditingRack] = useState<Rack | null>(null);
  const [deletingRack, setDeletingRack] = useState<Rack | null>(null);

  // Form state
  const [formName, setFormName] = useState('');
  const [formCode, setFormCode] = useState('');
  const [formLocation, setFormLocation] = useState('');
  const [formCapacity, setFormCapacity] = useState('');
  const [formDescription, setFormDescription] = useState('');
  const [formActive, setFormActive] = useState(true);
  const [formErrors, setFormErrors] = useState<Record<string, string>>({});

  useEffect(() => {
    const timer = setTimeout(() => {
      setDebouncedSearch(search);
      setCurrentPage(1);
    }, 300);

    return () => clearTimeout(timer);
  }, [search]);

  const loadRacks = useCallback(async () => {
    setLoading(true);
    try {
      const response = await fetchRacks({
        page: currentPage,
        pageSize,
        search: debouncedSearch || undefined,
        sortBy: sortKey || undefined,
        sortDir: sortDirection || undefined,
      });

      setRacks(response.data);
      setTotalItems(response.meta.totalItems);
      setTotalPages(response.meta.totalPages || 1);
    } catch (error) {
      if (error instanceof ApiError) {
        addToast(error.message, 'error');
      } else {
        addToast('Failed to load racks', 'error');
      }
    } finally {
      setLoading(false);
    }
  }, [currentPage, pageSize, debouncedSearch, sortKey, sortDirection, fetchRacks, addToast]);

  useEffect(() => {
    loadRacks();
  }, [loadRacks]);

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
    setEditingRack(null);
    setFormName('');
    setFormCode('');
    setFormLocation('');
    setFormCapacity('');
    setFormDescription('');
    setFormActive(true);
    setFormErrors({});
    setIsFormOpen(true);
  };

  const openEditModal = (rack: Rack) => {
    setEditingRack(rack);
    setFormName(rack.name);
    setFormCode(rack.code);
    setFormLocation(rack.location);
    setFormCapacity(String(rack.capacity));
    setFormDescription(rack.description);
    setFormActive(rack.active);
    setFormErrors({});
    setIsFormOpen(true);
  };

  const openDeleteModal = (rack: Rack) => {
    setDeletingRack(rack);
    setIsDeleteOpen(true);
  };

  const validateForm = () => {
    const errors: Record<string, string> = {};

    if (!formName.trim()) errors.name = 'Name is required';
    if (!formCode.trim()) errors.code = 'Code is required';
    if (!formLocation.trim()) errors.location = 'Location is required';

    const capacity = Number(formCapacity);
    if (!formCapacity.trim()) {
      errors.capacity = 'Capacity is required';
    } else if (Number.isNaN(capacity) || capacity <= 0) {
      errors.capacity = 'Capacity must be a positive number';
    }

    setFormErrors(errors);
    return Object.keys(errors).length === 0;
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!validateForm()) return;

    setIsSubmitting(true);

    try {
      const payload = {
        name: formName.trim(),
        code: formCode.trim(),
        location: formLocation.trim(),
        capacity: Number(formCapacity),
        description: formDescription.trim(),
      };

      if (editingRack) {
        await updateRack(editingRack.id, {
          ...payload,
          active: formActive,
        });
        addToast('Rack updated successfully', 'success');
      } else {
        await createRack(payload);
        addToast('Rack created successfully', 'success');
      }

      setIsFormOpen(false);
      await loadRacks();
    } catch (error) {
      if (error instanceof ApiError) {
        addToast(error.message, 'error');
      } else {
        addToast('Failed to save rack', 'error');
      }
    } finally {
      setIsSubmitting(false);
    }
  };

  const handleDelete = async () => {
    if (!deletingRack) return;

    setIsSubmitting(true);

    try {
      await deleteRack(deletingRack.id);
      addToast(`Rack ${deletingRack.name} has been deleted`, 'success');

      if (currentPage > 1 && racks.length === 1) {
        setCurrentPage((prev) => prev - 1);
      } else {
        await loadRacks();
      }
    } catch (error) {
      if (error instanceof ApiError) {
        addToast(error.message, 'error');
      } else {
        addToast('Failed to delete rack', 'error');
      }
    } finally {
      setIsSubmitting(false);
      setIsDeleteOpen(false);
    }
  };

  const columns = [
    { key: 'id', label: 'ID', sortable: true },
    { key: 'name', label: 'Name', sortable: true },
    { key: 'code', label: 'Code', sortable: true },
    { key: 'location', label: 'Location', sortable: true },
    {
      key: 'capacity',
      label: 'Capacity',
      sortable: false,
      render: (item: Rack) => <span className="text-gray-700">{item.capacity}</span>,
    },
    {
      key: 'active',
      label: 'Status',
      sortable: true,
      render: (item: Rack) => (
        <Badge variant={item.active ? 'success' : 'secondary'}>
          {item.active ? 'Active' : 'Inactive'}
        </Badge>
      ),
    },
    {
      key: 'actions',
      label: 'Actions',
      render: (item: Rack) => (
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
          <h1 className="text-2xl font-bold text-gray-900">Master Rack</h1>
          <Button onClick={openAddModal}>Add Rack</Button>
        </div>

        <div className="bg-white rounded-lg border border-gray-200 shadow-sm">
          <div className="p-4 border-b border-gray-200">
            <Input
              placeholder="Search racks..."
              value={search}
              onChange={(e) => setSearch(e.target.value)}
              className="max-w-sm"
            />
          </div>
          <Table
            columns={columns}
            data={racks}
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
        title={editingRack ? 'Edit Rack' : 'Create Rack'}
      >
        <form onSubmit={handleSubmit} className="space-y-4">
          <Input
            label="Name"
            placeholder="Rack name"
            value={formName}
            onChange={(e) => {
              setFormName(e.target.value);
              setFormErrors((prev) => ({ ...prev, name: '' }));
            }}
            error={formErrors.name}
            required
          />

          <Input
            label="Code"
            placeholder="Rack code (e.g., R-001)"
            value={formCode}
            onChange={(e) => {
              setFormCode(e.target.value);
              setFormErrors((prev) => ({ ...prev, code: '' }));
            }}
            error={formErrors.code}
            required
          />

          <Input
            label="Location"
            placeholder="Physical location"
            value={formLocation}
            onChange={(e) => {
              setFormLocation(e.target.value);
              setFormErrors((prev) => ({ ...prev, location: '' }));
            }}
            error={formErrors.location}
            required
          />

          <Input
            label="Capacity"
            type="number"
            placeholder="Capacity (must be > 0)"
            value={formCapacity}
            onChange={(e) => {
              setFormCapacity(e.target.value);
              setFormErrors((prev) => ({ ...prev, capacity: '' }));
            }}
            error={formErrors.capacity}
            required
            min="1"
          />

          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">Description</label>
            <textarea
              placeholder="Optional description"
              value={formDescription}
              onChange={(e) => setFormDescription(e.target.value)}
              className="w-full rounded-md border border-gray-300 px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
              rows={3}
            />
          </div>

          {editingRack && (
            <Toggle label="Active" checked={formActive} onChange={setFormActive} />
          )}

          <div className="flex justify-end gap-2 pt-2 border-t">
            <Button
              type="button"
              variant="outline"
              onClick={() => setIsFormOpen(false)}
              disabled={isSubmitting}
            >
              Cancel
            </Button>
            <Button type="submit" disabled={isSubmitting}>
              {editingRack ? 'Update' : 'Create'}
            </Button>
          </div>
        </form>
      </Modal>

      <Modal
        isOpen={isDeleteOpen}
        onClose={() => setIsDeleteOpen(false)}
        title="Delete Rack"
      >
        <p className="text-sm text-gray-600 mb-6">
          Are you sure you want to delete rack{' '}
          <strong>
            {deletingRack?.name} ({deletingRack?.code})
          </strong>
          ? This action cannot be undone.
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
