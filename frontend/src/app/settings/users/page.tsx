'use client';

import React, { useState, useMemo } from 'react';
import AdminLayout from '@/components/layout/AdminLayout';
import Button from '@/components/ui/Button';
import Input from '@/components/ui/Input';
import Table from '@/components/ui/Table';
import type { SortDirection } from '@/components/ui/Table';
import ConfirmModal from '@/components/ui/ConfirmModal';
import Avatar from '@/components/ui/Avatar';
import Badge from '@/components/ui/Badge';
import UserFormModal from '@/components/user/UserFormModal';
import { useUserStore, User } from '@/stores/useUserStore';
import { useRoleStore } from '@/stores/useRoleStore';
import { useToastStore } from '@/stores/useToastStore';

const DEFAULT_PAGE_SIZE = 10;

const statusBadgeVariant: Record<User['status'], 'green' | 'yellow' | 'gray'> = {
  active: 'green',
  pending: 'yellow',
  inactive: 'gray',
};

const statusLabel: Record<User['status'], string> = {
  active: 'Active',
  pending: 'Pending',
  inactive: 'Inactive',
};

export default function SettingsUsersPage() {
  const { users, addUser, updateUser, deleteUser, approveUser } = useUserStore();
  const { roles } = useRoleStore();
  const { addToast } = useToastStore();

  const [search, setSearch] = useState('');
  const [currentPage, setCurrentPage] = useState(1);
  const [pageSize, setPageSize] = useState(DEFAULT_PAGE_SIZE);
  const [sortKey, setSortKey] = useState<string | null>(null);
  const [sortDirection, setSortDirection] = useState<SortDirection>(null);

  // Modal state
  const [isFormOpen, setIsFormOpen] = useState(false);
  const [editingUser, setEditingUser] = useState<User | null>(null);
  const [isDeleteOpen, setIsDeleteOpen] = useState(false);
  const [deletingUser, setDeletingUser] = useState<User | null>(null);
  const [isRejectOpen, setIsRejectOpen] = useState(false);
  const [rejectingUser, setRejectingUser] = useState<User | null>(null);

  // Build role lookup map
  const roleMap = useMemo(() => {
    const map = new Map<number, string>();
    roles.forEach((r) => map.set(r.id, r.name));
    return map;
  }, [roles]);

  const getRoleNames = (roleIds: number[]) => {
    const names = roleIds
      .map((id) => roleMap.get(id))
      .filter(Boolean)
      .join(', ');
    return names || '\u2014';
  };

  const filtered = useMemo(() => {
    if (!search) return users;
    const q = search.toLowerCase();
    return users.filter(
      (u) =>
        u.name.toLowerCase().includes(q) ||
        u.email.toLowerCase().includes(q)
    );
  }, [users, search]);

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
        case 'email':
          aVal = a.email.toLowerCase();
          bVal = b.email.toLowerCase();
          break;
        case 'status':
          aVal = a.status;
          bVal = b.status;
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

  const openCreateModal = () => {
    setEditingUser(null);
    setIsFormOpen(true);
  };

  const openEditModal = (user: User) => {
    setEditingUser(user);
    setIsFormOpen(true);
  };

  const openDeleteModal = (user: User) => {
    setDeletingUser(user);
    setIsDeleteOpen(true);
  };

  const openRejectModal = (user: User) => {
    setRejectingUser(user);
    setIsRejectOpen(true);
  };

  const handleSave = (data: Omit<User, 'id' | 'createdAt'>) => {
    if (editingUser) {
      updateUser(editingUser.id, data);
      addToast('User updated successfully.', 'success');
    } else {
      addUser(data);
      addToast(
        `User created successfully. Credentials have been sent to ${data.email}.`,
        'success'
      );
    }
    setIsFormOpen(false);
  };

  const handleDelete = () => {
    if (deletingUser) {
      deleteUser(deletingUser.id);
      addToast(`User ${deletingUser.name} has been deleted.`, 'success');
      const newTotal = Math.max(1, Math.ceil((filtered.length - 1) / pageSize));
      if (currentPage > newTotal) setCurrentPage(newTotal);
    }
    setIsDeleteOpen(false);
  };

  const handleApprove = (user: User) => {
    approveUser(user.id);
    addToast(`User ${user.name} has been approved.`, 'success');
  };

  const handleReject = () => {
    if (rejectingUser) {
      deleteUser(rejectingUser.id);
      addToast(`User ${rejectingUser.name} has been rejected.`, 'success');
      const newTotal = Math.max(1, Math.ceil((filtered.length - 1) / pageSize));
      if (currentPage > newTotal) setCurrentPage(newTotal);
    }
    setIsRejectOpen(false);
  };

  const columns = [
    { key: 'id', label: 'ID', sortable: true },
    {
      key: 'profilePicture',
      label: 'Profile',
      render: (item: User) => (
        <Avatar
          src={item.profilePicture || undefined}
          name={item.name}
          size="sm"
        />
      ),
    },
    { key: 'name', label: 'Name', sortable: true },
    { key: 'email', label: 'Email', sortable: true },
    { key: 'phone', label: 'Phone' },
    {
      key: 'roles',
      label: 'Roles',
      render: (item: User) => (
        <span className="text-sm text-gray-700">
          {getRoleNames(item.roles)}
        </span>
      ),
    },
    {
      key: 'status',
      label: 'Status',
      sortable: true,
      render: (item: User) => (
        <Badge variant={statusBadgeVariant[item.status]}>
          {statusLabel[item.status]}
        </Badge>
      ),
    },
    {
      key: 'actions',
      label: 'Actions',
      render: (item: User) => (
        <div className="flex gap-2">
          <Button
            size="sm"
            variant="outline"
            onClick={() => openEditModal(item)}
            title="Edit user"
          >
            Edit
          </Button>
          {item.status === 'pending' && (
            <>
              <Button
                size="sm"
                variant="primary"
                onClick={() => handleApprove(item)}
                title="Approve user"
              >
                Approve
              </Button>
              <Button
                size="sm"
                variant="danger"
                onClick={() => openRejectModal(item)}
                title="Reject user"
              >
                Reject
              </Button>
            </>
          )}
          {item.isSuperAdmin ? (
            <Button
              size="sm"
              variant="danger"
              disabled
              title="Super admin cannot be deleted"
            >
              Delete
            </Button>
          ) : (
            <Button
              size="sm"
              variant="danger"
              onClick={() => openDeleteModal(item)}
              title="Delete user"
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
          <h1 className="text-2xl font-bold text-gray-900">Users</h1>
          <Button onClick={openCreateModal}>Create User</Button>
        </div>

        <div className="bg-white rounded-lg border border-gray-200 shadow-sm">
          <div className="p-4 border-b border-gray-200">
            <Input
              placeholder="Search users..."
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

      {/* Create/Edit User Modal */}
      <UserFormModal
        isOpen={isFormOpen}
        onClose={() => setIsFormOpen(false)}
        user={editingUser}
        onSave={handleSave}
      />

      {/* Delete Confirmation Modal */}
      <ConfirmModal
        isOpen={isDeleteOpen}
        onClose={() => setIsDeleteOpen(false)}
        onConfirm={handleDelete}
        title="Delete User"
        message={`Are you sure you want to delete ${deletingUser?.name}? This action cannot be undone.`}
        cancelLabel="Cancel"
        confirmLabel="Delete"
        variant="danger"
      />

      {/* Reject Confirmation Modal */}
      <ConfirmModal
        isOpen={isRejectOpen}
        onClose={() => setIsRejectOpen(false)}
        onConfirm={handleReject}
        title="Reject User"
        message={`Are you sure you want to reject ${rejectingUser?.name}? This will remove their registration.`}
        cancelLabel="Cancel"
        confirmLabel="Reject"
        variant="danger"
      />
    </AdminLayout>
  );
}
