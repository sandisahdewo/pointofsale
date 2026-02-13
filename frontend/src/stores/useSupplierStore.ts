'use client';

import { create } from 'zustand';
import { initialSuppliers } from '@/data/suppliers';

export interface BankAccount {
  id: string;
  accountName: string;
  accountNumber: string;
}

export interface Supplier {
  id: number;
  name: string;
  address: string;
  phone: string;
  email: string;
  website: string;
  bankAccounts: BankAccount[];
  active: boolean;
  createdAt: string;
}

interface SupplierState {
  suppliers: Supplier[];
  addSupplier: (supplier: Omit<Supplier, 'id' | 'createdAt'>) => void;
  updateSupplier: (id: number, supplier: Partial<Omit<Supplier, 'id' | 'createdAt'>>) => void;
  deleteSupplier: (id: number) => void;
  getActiveSuppliers: () => Supplier[];
}

export const useSupplierStore = create<SupplierState>((set, get) => ({
  suppliers: initialSuppliers,

  addSupplier: (supplier) =>
    set((state) => {
      const maxId = state.suppliers.reduce((max, s) => Math.max(max, s.id), 0);
      const newSupplier: Supplier = {
        ...supplier,
        id: maxId + 1,
        createdAt: new Date().toISOString(),
        bankAccounts: supplier.bankAccounts.map((account) => ({
          ...account,
          id: account.id || crypto.randomUUID(),
        })),
      };
      return { suppliers: [...state.suppliers, newSupplier] };
    }),

  updateSupplier: (id, data) =>
    set((state) => ({
      suppliers: state.suppliers.map((s) =>
        s.id === id
          ? {
              ...s,
              ...data,
              bankAccounts: data.bankAccounts
                ? data.bankAccounts.map((account) => ({
                    ...account,
                    id: account.id || crypto.randomUUID(),
                  }))
                : s.bankAccounts,
            }
          : s
      ),
    })),

  deleteSupplier: (id) =>
    set((state) => ({
      suppliers: state.suppliers.filter((s) => s.id !== id),
    })),

  getActiveSuppliers: () => get().suppliers.filter((s) => s.active),
}));
