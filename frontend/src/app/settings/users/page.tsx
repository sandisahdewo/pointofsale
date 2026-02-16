'use client';

import React, { useState, useEffect, useCallback } from 'react';
import AdminLayout from '@/components/layout/AdminLayout';
import Button from '@/components/ui/Button';
import Input from '@/components/ui/Input';
import Table from '@/components/ui/Table';
import type { SortDirection } from '@/components/ui/Table';
import ConfirmModal from '@/components/ui/ConfirmModal';
import Avatar from '@/components/ui/Avatar';
import Badge from '@/components/ui/Badge';
import UserFormModal from '@/components/user/UserFormModal';
import { useUserStore, User, UserRole, CreateUserInput, UpdateUserInput } from '@/stores/useUserStore';
import { useToastStore } from '@/stores/useToastStore';
import { ApiError } from '@/lib/api';

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
  const { fetchUsers, createUser, updateUser, deleteUser, approveUser, rejectUser } = useUserStore();
  const { addToast } = useToastStore();

  const [users, setUsers] = useState<User[]>([]);
  const [loading, setLoading] = useState(true);
  const [totalItems, setTotalItems] = useState(0);
  const [totalPages, setTotalPages] = useState(1);

  const [search, setSearch] = useState('');
  const [searchDebounced, setSearchDebounced] = useState('');
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

  const getRoleNames = (roles: UserRole[] | undefined | null) => {
    if (!roles || roles.length === 0) return '—';
    return roles.map(r => r.name).join(', ');
  };

  // Debounce search input
  useEffect(() => {
    const timer = setTimeout(() => {
      setSearchDebounced(search);
      setCurrentPage(1);
    }, 300);
    return () => clearTimeout(timer);
  }, [search]);

  const loadUsers = useCallback(async () => {
    setLoading(true);
    try {
      const response = await fetchUsers({
        page: currentPage,
        pageSize,
        search: searchDebounced || undefined,
        sortBy: sortKey || undefined,
        sortDir: sortDirection || undefined,
      });
      setUsers(response.data);
      setTotalItems(response.meta.totalItems);
      setTotalPages(response.meta.totalPages);
    } catch (error) {
      if (error instanceof ApiError) {
        addToast(error.message, 'error');
      } else {
        addToast('Failed to load users', 'error');
      }
    } finally {
      setLoading(false);
    }
  }, [currentPage, pageSize, searchDebounced, sortKey, sortDirection, fetchUsers, addToast]);

  useEffect(() => {
    loadUsers();
  }, [loadUsers]);

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

  const handleSave = async (data: CreateUserInput | UpdateUserInput, isEdit: boolean) => {
    try {
      if (isEdit && editingUser) {
        await updateUser(editingUser.id, data as UpdateUserInput);
        addToast('User updated successfully.', 'success');
      } else {
        await createUser(data as CreateUserInput);
        addToast('User created successfully. Credentials sent to email.', 'success');
      }
      setIsFormOpen(false);
      await loadUsers();
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
    if (deletingUser) {
      try {
        await deleteUser(deletingUser.id);
        addToast(`User ${deletingUser.name} has been deleted.`, 'success');
        setIsDeleteOpen(false);
        await loadUsers();
      } catch (error) {
        if (error instanceof ApiError) {
          addToast(error.message, 'error');
        } else {
          addToast('Failed to delete user', 'error');
        }
      }
    }
  };

  const handleApprove = async (user: User) => {
    try {
      await approveUser(user.id);
      addToast(`User ${user.name} has been approved.`, 'success');
      await loadUsers();
    } catch (error) {
      if (error instanceof ApiError) {
        addToast(error.message, 'error');
      } else {
        addToast('Failed to approve user', 'error');
      }
    }
  };

  const handleReject = async () => {
    if (rejectingUser) {
      try {
        await rejectUser(rejectingUser.id);
        addToast(`User ${rejectingUser.name} has been rejected.`, 'success');
        setIsRejectOpen(false);
        await loadUsers();
      } catch (error) {
        if (error instanceof ApiError) {
          addToast(error.message, 'error');
        } else {
          addToast('Failed to reject user', 'error');
        }
      }
    }
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
            data={users}
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
