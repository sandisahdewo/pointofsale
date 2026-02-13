'use client';

import React, { useState, useMemo } from 'react';
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

const DEFAULT_PAGE_SIZE = 10;

export default function MasterRackPage() {
  const { racks, addRack, updateRack, deleteRack, isCodeUnique } = useRackStore();
  const { addToast } = useToastStore();

  const [search, setSearch] = useState('');
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

  const filtered = useMemo(() => {
    if (!search) return racks;
    const q = search.toLowerCase();
    return racks.filter(
      (r) =>
        r.name.toLowerCase().includes(q) ||
        r.code.toLowerCase().includes(q) ||
        r.location.toLowerCase().includes(q)
    );
  }, [racks, search]);

  const sorted = useMemo(() => {
    if (!sortKey || !sortDirection) return filtered;
    return [...filtered].sort((a, b) => {
      const aVal = (a as unknown as Record<string, unknown>)[sortKey];
      const bVal = (b as unknown as Record<string, unknown>)[sortKey];
      if (typeof aVal === 'number' && typeof bVal === 'number') {
        return sortDirection === 'asc' ? aVal - bVal : bVal - aVal;
      }
      if (typeof aVal === 'boolean' && typeof bVal === 'boolean') {
        return sortDirection === 'asc'
          ? (aVal ? 1 : 0) - (bVal ? 1 : 0)
          : (bVal ? 1 : 0) - (aVal ? 1 : 0);
      }
      const aStr = String(aVal).toLowerCase();
      const bStr = String(bVal).toLowerCase();
      if (aStr < bStr) return sortDirection === 'asc' ? -1 : 1;
      if (aStr > bStr) return sortDirection === 'asc' ? 1 : -1;
      return 0;
    });
  }, [filtered, sortKey, sortDirection]);

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
    if (!formCode.trim()) {
      errors.code = 'Code is required';
    } else if (!isCodeUnique(formCode.trim(), editingRack?.id)) {
      errors.code = 'Rack code already exists';
    }
    if (!formLocation.trim()) errors.location = 'Location is required';

    const capacity = Number(formCapacity);
    if (!formCapacity.trim()) {
      errors.capacity = 'Capacity is required';
    } else if (isNaN(capacity) || capacity <= 0) {
      errors.capacity = 'Capacity must be a positive number';
    }

    setFormErrors(errors);
    return Object.keys(errors).length === 0;
  };

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (!validateForm()) return;

    const rackData = {
      name: formName.trim(),
      code: formCode.trim(),
      location: formLocation.trim(),
      capacity: Number(formCapacity),
      description: formDescription.trim(),
      active: formActive,
    };

    if (editingRack) {
      updateRack(editingRack.id, rackData);
      addToast('Rack updated successfully', 'success');
    } else {
      addRack(rackData);
      addToast('Rack created successfully', 'success');
    }
    setIsFormOpen(false);
  };

  const handleDelete = () => {
    if (deletingRack) {
      deleteRack(deletingRack.id);
      addToast(`Rack ${deletingRack.name} has been deleted`, 'success');
      const newTotal = Math.max(1, Math.ceil((filtered.length - 1) / pageSize));
      if (currentPage > newTotal) setCurrentPage(newTotal);
    }
    setIsDeleteOpen(false);
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
      render: (item: Rack) => (
        <span className="text-gray-700">{item.capacity}</span>
      ),
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

      {/* Add/Edit Modal */}
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
            <label className="block text-sm font-medium text-gray-700 mb-1">
              Description
            </label>
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
            >
              Cancel
            </Button>
            <Button type="submit">{editingRack ? 'Update' : 'Create'}</Button>
          </div>
        </form>
      </Modal>

      {/* Delete Confirmation Modal */}
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
