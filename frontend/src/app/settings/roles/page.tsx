'use client';

import React, { useState, useCallback, useEffect } from 'react';
import { useRouter } from 'next/navigation';
import AdminLayout from '@/components/layout/AdminLayout';
import Button from '@/components/ui/Button';
import Input from '@/components/ui/Input';
import Table from '@/components/ui/Table';
import type { SortDirection } from '@/components/ui/Table';
import ConfirmModal from '@/components/ui/ConfirmModal';
import RoleFormModal from '@/components/role/RoleFormModal';
import { useRoleStore, Role } from '@/stores/useRoleStore';
import { useToastStore } from '@/stores/useToastStore';
import { ApiError } from '@/lib/api';

const DEFAULT_PAGE_SIZE = 10;

export default function RolesPage() {
  const router = useRouter();
  const { fetchRoles, createRole, updateRole, deleteRole } = useRoleStore();
  const { addToast } = useToastStore();

  const [roles, setRoles] = useState<Role[]>([]);
  const [loading, setLoading] = useState(true);
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
  const [editingRole, setEditingRole] = useState<Role | null>(null);
  const [isDeleteOpen, setIsDeleteOpen] = useState(false);
  const [deletingRole, setDeletingRole] = useState<Role | null>(null);

  // Debounce search
  useEffect(() => {
    const timer = setTimeout(() => {
      setDebouncedSearch(search);
      setCurrentPage(1);
    }, 300);
    return () => clearTimeout(timer);
  }, [search]);

  const loadRoles = useCallback(async () => {
    setLoading(true);
    try {
      const response = await fetchRoles({
        page: currentPage,
        pageSize,
        search: debouncedSearch || undefined,
        sortBy: sortKey || undefined,
        sortDir: sortDirection || undefined,
      });
      setRoles(response.data);
      setTotalItems(response.meta.totalItems);
      setTotalPages(response.meta.totalPages);
    } catch (error) {
      if (error instanceof ApiError) {
        addToast(error.message, 'error');
      } else {
        addToast('Failed to load roles', 'error');
      }
    } finally {
      setLoading(false);
    }
  }, [currentPage, pageSize, debouncedSearch, sortKey, sortDirection, fetchRoles, addToast]);

  useEffect(() => {
    loadRoles();
  }, [loadRoles]);

  const handleSearch = (value: string) => {
    setSearch(value);
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
    setEditingRole(null);
    setIsFormOpen(true);
  };

  const openEditModal = (role: Role) => {
    setEditingRole(role);
    setIsFormOpen(true);
  };

  const openDeleteModal = (role: Role) => {
    setDeletingRole(role);
    setIsDeleteOpen(true);
  };

  const handleFormSave = async (input: { name: string; description: string }) => {
    try {
      if (editingRole) {
        await updateRole(editingRole.id, input);
        addToast(`Role ${input.name} updated successfully.`, 'success');
      } else {
        await createRole(input);
        addToast(`Role ${input.name} created successfully.`, 'success');
      }
      setIsFormOpen(false);
      loadRoles();
    } catch (error) {
      if (error instanceof ApiError) {
        addToast(error.message, 'error');
      } else {
        addToast('An error occurred', 'error');
      }
      throw error;
    }
  };

  const handleDelete = async () => {
    if (!deletingRole) return;
    try {
      await deleteRole(deletingRole.id);
      addToast(`Role ${deletingRole.name} has been deleted.`, 'success');
      loadRoles();
    } catch (error) {
      if (error instanceof ApiError) {
        addToast(error.message, 'error');
      } else {
        addToast('An error occurred', 'error');
      }
    }
    setIsDeleteOpen(false);
  };

  const columns = [
    { key: 'id', label: 'ID', sortable: true },
    { key: 'name', label: 'Name', sortable: true },
    { key: 'description', label: 'Description', sortable: true },
    {
      key: 'users',
      label: 'Users',
      render: (item: Role) => {
        return (
          <span className="text-gray-700">
            {item.userCount} {item.userCount === 1 ? 'user' : 'users'}
          </span>
        );
      },
    },
    {
      key: 'actions',
      label: 'Actions',
      render: (item: Role) => (
        <div className="flex gap-2">
          <Button
            size="sm"
            variant="outline"
            onClick={() =>
              router.push(`/settings/roles/${item.id}/permissions`)
            }
          >
            Permissions
          </Button>
          {item.isSystem ? (
            <span
              className="inline-flex items-center px-3 py-1.5 text-sm font-medium text-gray-400 cursor-not-allowed"
              title="System role cannot be edited."
            >
              Edit
            </span>
          ) : (
            <Button
              size="sm"
              variant="outline"
              onClick={() => openEditModal(item)}
            >
              Edit
            </Button>
          )}
          {item.isSystem ? (
            <span
              className="inline-flex items-center px-3 py-1.5 text-sm font-medium text-gray-400 cursor-not-allowed"
              title="System role cannot be deleted."
            >
              Delete
            </span>
          ) : (
            <Button
              size="sm"
              variant="danger"
              onClick={() => openDeleteModal(item)}
            >
              Delete
            </Button>
          )}
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
          <h1 className="text-2xl font-bold text-gray-900">
            Roles & Permissions
          </h1>
          <Button onClick={openAddModal}>Create Role</Button>
        </div>

        <div className="bg-white rounded-lg border border-gray-200 shadow-sm">
          <div className="p-4 border-b border-gray-200">
            <Input
              placeholder="Search roles..."
              value={search}
              onChange={(e) => handleSearch(e.target.value)}
              className="max-w-sm"
            />
          </div>
          <Table
            columns={columns}
            data={roles}
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

      <RoleFormModal
        isOpen={isFormOpen}
        onClose={() => setIsFormOpen(false)}
        editingRole={editingRole}
        onSave={handleFormSave}
      />

      <ConfirmModal
        isOpen={isDeleteOpen}
        onClose={() => setIsDeleteOpen(false)}
        onConfirm={handleDelete}
        title="Delete Role"
        message={`Are you sure you want to delete the role "${deletingRole?.name}"? Users assigned to this role will lose these permissions.`}
        cancelLabel="Cancel"
        confirmLabel="Delete"
        variant="danger"
      />
    </AdminLayout>
  );
}
