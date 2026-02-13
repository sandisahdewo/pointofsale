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
import { useSupplierStore, Supplier, BankAccount } from '@/stores/useSupplierStore';
import { useToastStore } from '@/stores/useToastStore';

const DEFAULT_PAGE_SIZE = 10;

export default function MasterSupplierPage() {
  const { suppliers, addSupplier, updateSupplier, deleteSupplier } = useSupplierStore();
  const { addToast } = useToastStore();

  const [search, setSearch] = useState('');
  const [currentPage, setCurrentPage] = useState(1);
  const [pageSize, setPageSize] = useState(DEFAULT_PAGE_SIZE);
  const [sortKey, setSortKey] = useState<string | null>(null);
  const [sortDirection, setSortDirection] = useState<SortDirection>(null);

  // Modal state
  const [isFormOpen, setIsFormOpen] = useState(false);
  const [isDeleteOpen, setIsDeleteOpen] = useState(false);
  const [editingSupplier, setEditingSupplier] = useState<Supplier | null>(null);
  const [deletingSupplier, setDeletingSupplier] = useState<Supplier | null>(null);

  // Form state
  const [formName, setFormName] = useState('');
  const [formAddress, setFormAddress] = useState('');
  const [formPhone, setFormPhone] = useState('');
  const [formEmail, setFormEmail] = useState('');
  const [formWebsite, setFormWebsite] = useState('');
  const [formActive, setFormActive] = useState(true);
  const [formBankAccounts, setFormBankAccounts] = useState<BankAccount[]>([]);
  const [formErrors, setFormErrors] = useState<Record<string, string>>({});

  const filtered = useMemo(() => {
    if (!search) return suppliers;
    const q = search.toLowerCase();
    return suppliers.filter(
      (s) =>
        s.name.toLowerCase().includes(q) ||
        s.address.toLowerCase().includes(q) ||
        s.email.toLowerCase().includes(q)
    );
  }, [suppliers, search]);

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
    setEditingSupplier(null);
    setFormName('');
    setFormAddress('');
    setFormPhone('');
    setFormEmail('');
    setFormWebsite('');
    setFormActive(true);
    setFormBankAccounts([]);
    setFormErrors({});
    setIsFormOpen(true);
  };

  const openEditModal = (supplier: Supplier) => {
    setEditingSupplier(supplier);
    setFormName(supplier.name);
    setFormAddress(supplier.address);
    setFormPhone(supplier.phone);
    setFormEmail(supplier.email);
    setFormWebsite(supplier.website);
    setFormActive(supplier.active);
    setFormBankAccounts([...supplier.bankAccounts]);
    setFormErrors({});
    setIsFormOpen(true);
  };

  const openDeleteModal = (supplier: Supplier) => {
    setDeletingSupplier(supplier);
    setIsDeleteOpen(true);
  };

  const addBankAccount = () => {
    setFormBankAccounts([
      ...formBankAccounts,
      { id: crypto.randomUUID(), accountName: '', accountNumber: '' },
    ]);
  };

  const removeBankAccount = (index: number) => {
    setFormBankAccounts(formBankAccounts.filter((_, i) => i !== index));
  };

  const updateBankAccount = (
    index: number,
    field: 'accountName' | 'accountNumber',
    value: string
  ) => {
    setFormBankAccounts(
      formBankAccounts.map((account, i) =>
        i === index ? { ...account, [field]: value } : account
      )
    );
    // Clear error for this field
    setFormErrors((prev) => {
      const newErrors = { ...prev };
      delete newErrors[`bankAccount_${index}_${field}`];
      return newErrors;
    });
  };

  const validateForm = () => {
    const errors: Record<string, string> = {};

    if (!formName.trim()) errors.name = 'Name is required';
    if (!formAddress.trim()) errors.address = 'Address is required';

    // Email validation
    if (formEmail.trim()) {
      const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
      if (!emailRegex.test(formEmail.trim())) {
        errors.email = 'Invalid email format';
      }
    }

    // Bank account validation
    formBankAccounts.forEach((account, index) => {
      if (account.accountName.trim() && !account.accountNumber.trim()) {
        errors[`bankAccount_${index}_accountNumber`] = 'Account number is required';
      }
      if (!account.accountName.trim() && account.accountNumber.trim()) {
        errors[`bankAccount_${index}_accountName`] = 'Account name is required';
      }
      if (!account.accountName.trim() && !account.accountNumber.trim()) {
        // Empty rows are allowed if both are empty (will be filtered out)
      }
    });

    setFormErrors(errors);
    return Object.keys(errors).length === 0;
  };

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (!validateForm()) return;

    // Filter out empty bank account rows
    const validBankAccounts = formBankAccounts.filter(
      (account) => account.accountName.trim() && account.accountNumber.trim()
    );

    const supplierData = {
      name: formName.trim(),
      address: formAddress.trim(),
      phone: formPhone.trim(),
      email: formEmail.trim(),
      website: formWebsite.trim(),
      bankAccounts: validBankAccounts,
      active: formActive,
    };

    if (editingSupplier) {
      updateSupplier(editingSupplier.id, supplierData);
      addToast('Supplier updated successfully', 'success');
    } else {
      addSupplier(supplierData);
      addToast('Supplier created successfully', 'success');
    }
    setIsFormOpen(false);
  };

  const handleDelete = () => {
    if (deletingSupplier) {
      deleteSupplier(deletingSupplier.id);
      addToast(`Supplier ${deletingSupplier.name} has been deleted`, 'success');
      const newTotal = Math.max(1, Math.ceil((filtered.length - 1) / pageSize));
      if (currentPage > newTotal) setCurrentPage(newTotal);
    }
    setIsDeleteOpen(false);
  };

  const columns = [
    { key: 'id', label: 'ID', sortable: true },
    { key: 'name', label: 'Name', sortable: true },
    { key: 'address', label: 'Address', sortable: false },
    {
      key: 'phone',
      label: 'Phone',
      sortable: false,
      render: (item: Supplier) => (
        <span className="text-gray-700">{item.phone || '—'}</span>
      ),
    },
    {
      key: 'email',
      label: 'Email',
      sortable: false,
      render: (item: Supplier) => (
        <span className="text-gray-700">{item.email || '—'}</span>
      ),
    },
    {
      key: 'active',
      label: 'Status',
      sortable: true,
      render: (item: Supplier) => (
        <Badge variant={item.active ? 'success' : 'secondary'}>
          {item.active ? 'Active' : 'Inactive'}
        </Badge>
      ),
    },
    {
      key: 'actions',
      label: 'Actions',
      render: (item: Supplier) => (
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
          <h1 className="text-2xl font-bold text-gray-900">Master Supplier</h1>
          <Button onClick={openAddModal}>Add Supplier</Button>
        </div>

        <div className="bg-white rounded-lg border border-gray-200 shadow-sm">
          <div className="p-4 border-b border-gray-200">
            <Input
              placeholder="Search suppliers..."
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
        title={editingSupplier ? 'Edit Supplier' : 'Create Supplier'}
      >
        <form onSubmit={handleSubmit} className="space-y-4">
          <Input
            label="Name"
            placeholder="Supplier name"
            value={formName}
            onChange={(e) => {
              setFormName(e.target.value);
              setFormErrors((prev) => ({ ...prev, name: '' }));
            }}
            error={formErrors.name}
            required
          />

          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">
              Address
            </label>
            <textarea
              placeholder="Supplier address"
              value={formAddress}
              onChange={(e) => {
                setFormAddress(e.target.value);
                setFormErrors((prev) => ({ ...prev, address: '' }));
              }}
              className={`w-full rounded-md border px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500 ${
                formErrors.address
                  ? 'border-red-500 focus:ring-red-500 focus:border-red-500'
                  : 'border-gray-300'
              }`}
              rows={3}
              required
            />
            {formErrors.address && (
              <p className="mt-1 text-sm text-red-600">{formErrors.address}</p>
            )}
          </div>

          <Input
            label="Phone"
            placeholder="Phone number"
            value={formPhone}
            onChange={(e) => setFormPhone(e.target.value)}
          />

          <Input
            label="Email"
            type="email"
            placeholder="Email address"
            value={formEmail}
            onChange={(e) => {
              setFormEmail(e.target.value);
              setFormErrors((prev) => ({ ...prev, email: '' }));
            }}
            error={formErrors.email}
          />

          <Input
            label="Website"
            placeholder="Website URL"
            value={formWebsite}
            onChange={(e) => setFormWebsite(e.target.value)}
          />

          {editingSupplier && (
            <Toggle
              label="Active"
              checked={formActive}
              onChange={setFormActive}
            />
          )}

          {/* Bank Accounts Section */}
          <div className="border-t pt-4 mt-4">
            <label className="block text-sm font-medium text-gray-700 mb-2">
              Bank Accounts (optional)
            </label>

            {formBankAccounts.length > 0 && (
              <div className="space-y-2 mb-3">
                {formBankAccounts.map((account, index) => (
                  <div key={account.id} className="flex gap-2 items-start">
                    <div className="flex-1">
                      <Input
                        placeholder="Account Name"
                        value={account.accountName}
                        onChange={(e) =>
                          updateBankAccount(index, 'accountName', e.target.value)
                        }
                        error={formErrors[`bankAccount_${index}_accountName`]}
                      />
                    </div>
                    <div className="flex-1">
                      <Input
                        placeholder="Account Number"
                        value={account.accountNumber}
                        onChange={(e) =>
                          updateBankAccount(index, 'accountNumber', e.target.value)
                        }
                        error={formErrors[`bankAccount_${index}_accountNumber`]}
                      />
                    </div>
                    <Button
                      type="button"
                      variant="danger"
                      size="sm"
                      onClick={() => removeBankAccount(index)}
                      className="mt-0"
                    >
                      Remove
                    </Button>
                  </div>
                ))}
              </div>
            )}

            <Button
              type="button"
              variant="outline"
              size="sm"
              onClick={addBankAccount}
            >
              + Add Bank Account
            </Button>
          </div>

          <div className="flex justify-end gap-2 pt-2 border-t">
            <Button
              type="button"
              variant="outline"
              onClick={() => setIsFormOpen(false)}
            >
              Cancel
            </Button>
            <Button type="submit">
              {editingSupplier ? 'Update' : 'Create'}
            </Button>
          </div>
        </form>
      </Modal>

      {/* Delete Confirmation Modal */}
      <Modal
        isOpen={isDeleteOpen}
        onClose={() => setIsDeleteOpen(false)}
        title="Delete Supplier"
      >
        <p className="text-sm text-gray-600 mb-6">
          Are you sure you want to delete{' '}
          <strong>{deletingSupplier?.name}</strong>? This action cannot be
          undone.
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
