'use client';

import React, { useState, useMemo } from 'react';
import { useRouter } from 'next/navigation';
import AdminLayout from '@/components/layout/AdminLayout';
import Button from '@/components/ui/Button';
import Input from '@/components/ui/Input';
import Table from '@/components/ui/Table';
import type { SortDirection } from '@/components/ui/Table';
import ConfirmModal from '@/components/ui/ConfirmModal';
import RoleFormModal from '@/components/role/RoleFormModal';
import { useRoleStore, Role } from '@/stores/useRoleStore';
import { useUserStore } from '@/stores/useUserStore';
import { useToastStore } from '@/stores/useToastStore';

const DEFAULT_PAGE_SIZE = 10;

export default function RolesPage() {
  const router = useRouter();
  const { roles, deleteRole } = useRoleStore();
  const { users, removeRoleFromUsers } = useUserStore();
  const { addToast } = useToastStore();

  const [search, setSearch] = useState('');
  const [currentPage, setCurrentPage] = useState(1);
  const [pageSize, setPageSize] = useState(DEFAULT_PAGE_SIZE);
  const [sortKey, setSortKey] = useState<string | null>(null);
  const [sortDirection, setSortDirection] = useState<SortDirection>(null);

  // Modal state
  const [isFormOpen, setIsFormOpen] = useState(false);
  const [editingRole, setEditingRole] = useState<Role | null>(null);
  const [isDeleteOpen, setIsDeleteOpen] = useState(false);
  const [deletingRole, setDeletingRole] = useState<Role | null>(null);

  const getUserCount = (roleId: number): number => {
    return users.filter((u) => u.roles.includes(roleId)).length;
  };

  const filtered = useMemo(() => {
    if (!search) return roles;
    const q = search.toLowerCase();
    return roles.filter(
      (r) =>
        r.name.toLowerCase().includes(q) ||
        r.description.toLowerCase().includes(q)
    );
  }, [roles, search]);

  const sorted = useMemo(() => {
    if (!sortKey || !sortDirection) return filtered;
    return [...filtered].sort((a, b) => {
      const aVal = (a as unknown as Record<string, unknown>)[sortKey];
      const bVal = (b as unknown as Record<string, unknown>)[sortKey];
      if (typeof aVal === 'number' && typeof bVal === 'number') {
        return sortDirection === 'asc' ? aVal - bVal : bVal - aVal;
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

  const handleDelete = () => {
    if (deletingRole) {
      deleteRole(deletingRole.id);
      removeRoleFromUsers(deletingRole.id);
      addToast(`Role ${deletingRole.name} has been deleted.`, 'success');
      const newTotal = Math.max(1, Math.ceil((filtered.length - 1) / pageSize));
      if (currentPage > newTotal) setCurrentPage(newTotal);
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
        const count = getUserCount(item.id);
        return (
          <span className="text-gray-700">
            {count} {count === 1 ? 'user' : 'users'}
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

      <RoleFormModal
        isOpen={isFormOpen}
        onClose={() => setIsFormOpen(false)}
        editingRole={editingRole}
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
