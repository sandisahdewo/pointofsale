'use client';

import { create } from 'zustand';
import { initialPurchaseOrders } from '@/data/purchaseOrders';
import { useProductStore } from './useProductStore';

export interface PurchaseOrderItem {
  id: string;
  productId: number;
  productName: string;
  variantId: string;
  variantLabel: string;
  sku: string;
  unitId: string;         // selected unit ID from the product's units
  unitName: string;       // denormalized unit label, e.g., "Pcs", "Kg", "Box"
  currentStock: number;
  orderedQty: number;
  price: number;
  receivedQty?: number;
  receivedPrice?: number;
  isVerified?: boolean;
}

export type POStatus = 'draft' | 'sent' | 'received' | 'completed' | 'cancelled';
export type PaymentMethod = 'cash' | 'credit_card' | 'bank_transfer';

export interface PurchaseOrder {
  id: number;
  poNumber: string;
  supplierId: number;
  supplierName: string;
  date: string;
  status: POStatus;
  items: PurchaseOrderItem[];
  notes: string;
  receivedDate?: string;
  paymentMethod?: PaymentMethod;
  supplierBankAccountId?: string;
  subtotal?: number;
  totalItems?: number;
  createdAt: string;
  updatedAt: string;
}

interface ReceiveData {
  receivedDate: string;
  paymentMethod: PaymentMethod;
  supplierBankAccountId?: string;
  items: Array<{
    id: string;
    receivedQty: number;
    receivedPrice: number;
    isVerified: boolean;
  }>;
}

interface PurchaseOrderState {
  purchaseOrders: PurchaseOrder[];
  addPurchaseOrder: (po: Omit<PurchaseOrder, 'id' | 'poNumber' | 'createdAt' | 'updatedAt'>) => void;
  updatePurchaseOrder: (id: number, po: Partial<PurchaseOrder>) => void;
  deletePurchaseOrder: (id: number) => void;
  updateStatus: (id: number, status: POStatus) => void;
  receivePurchaseOrder: (id: number, receiveData: ReceiveData) => void;
  completePurchaseOrder: (id: number) => void;
  cancelPurchaseOrder: (id: number) => void;
  getNextPoNumber: () => string;
  getPurchaseOrder: (id: number) => PurchaseOrder | undefined;
}

export const usePurchaseOrderStore = create<PurchaseOrderState>((set, get) => ({
  purchaseOrders: initialPurchaseOrders,

  addPurchaseOrder: (po) =>
    set((state) => {
      const maxId = state.purchaseOrders.reduce((max, p) => Math.max(max, p.id), 0);
      const poNumber = get().getNextPoNumber();
      const now = new Date().toISOString();

      const newPO: PurchaseOrder = {
        ...po,
        id: maxId + 1,
        poNumber,
        createdAt: now,
        updatedAt: now,
      };

      return { purchaseOrders: [...state.purchaseOrders, newPO] };
    }),

  updatePurchaseOrder: (id, data) =>
    set((state) => ({
      purchaseOrders: state.purchaseOrders.map((po) =>
        po.id === id
          ? { ...po, ...data, updatedAt: new Date().toISOString() }
          : po
      ),
    })),

  deletePurchaseOrder: (id) =>
    set((state) => {
      const po = state.purchaseOrders.find((p) => p.id === id);
      if (po && po.status !== 'draft') {
        console.error('Cannot delete PO with status:', po.status);
        return state;
      }
      return {
        purchaseOrders: state.purchaseOrders.filter((p) => p.id !== id),
      };
    }),

  updateStatus: (id, status) =>
    set((state) => ({
      purchaseOrders: state.purchaseOrders.map((po) =>
        po.id === id
          ? { ...po, status, updatedAt: new Date().toISOString() }
          : po
      ),
    })),

  receivePurchaseOrder: (id, receiveData) =>
    set((state) => {
      const po = state.purchaseOrders.find((p) => p.id === id);
      if (!po || po.status !== 'sent') {
        console.error('Cannot receive PO with status:', po?.status);
        return state;
      }

      // Update items with received data
      const updatedItems = po.items.map((item) => {
        const receivedItem = receiveData.items.find((r) => r.id === item.id);
        if (receivedItem) {
          return {
            ...item,
            receivedQty: receivedItem.receivedQty,
            receivedPrice: receivedItem.receivedPrice,
            isVerified: receivedItem.isVerified,
          };
        }
        return item;
      });

      // Calculate totals
      const subtotal = updatedItems.reduce(
        (sum, item) => sum + (item.receivedQty || 0) * (item.receivedPrice || 0),
        0
      );
      const totalItems = updatedItems.reduce(
        (sum, item) => sum + (item.receivedQty || 0),
        0
      );

      // Update variant stock in product store
      const productStore = useProductStore.getState();
      updatedItems.forEach((item) => {
        if (item.receivedQty) {
          const product = productStore.getProduct(item.productId);
          if (product) {
            const updatedVariants = product.variants.map((v) =>
              v.id === item.variantId
                ? { ...v, currentStock: v.currentStock + item.receivedQty! }
                : v
            );
            productStore.updateProduct(item.productId, { variants: updatedVariants });
          }
        }
      });

      // Update PO with received data
      return {
        purchaseOrders: state.purchaseOrders.map((p) =>
          p.id === id
            ? {
                ...p,
                status: 'received' as POStatus,
                items: updatedItems,
                receivedDate: receiveData.receivedDate,
                paymentMethod: receiveData.paymentMethod,
                supplierBankAccountId: receiveData.supplierBankAccountId,
                subtotal,
                totalItems,
                updatedAt: new Date().toISOString(),
              }
            : p
        ),
      };
    }),

  completePurchaseOrder: (id) =>
    set((state) => {
      const po = state.purchaseOrders.find((p) => p.id === id);
      if (!po || po.status !== 'received') {
        console.error('Cannot complete PO with status:', po?.status);
        return state;
      }
      return {
        purchaseOrders: state.purchaseOrders.map((p) =>
          p.id === id
            ? { ...p, status: 'completed' as POStatus, updatedAt: new Date().toISOString() }
            : p
        ),
      };
    }),

  cancelPurchaseOrder: (id) =>
    set((state) => {
      const po = state.purchaseOrders.find((p) => p.id === id);
      if (!po || (po.status !== 'draft' && po.status !== 'sent')) {
        console.error('Cannot cancel PO with status:', po?.status);
        return state;
      }
      return {
        purchaseOrders: state.purchaseOrders.map((p) =>
          p.id === id
            ? { ...p, status: 'cancelled' as POStatus, updatedAt: new Date().toISOString() }
            : p
        ),
      };
    }),

  getNextPoNumber: () => {
    const state = get();
    const currentYear = new Date().getFullYear();
    const yearPrefix = `PO-${currentYear}-`;

    const existingNumbers = state.purchaseOrders
      .map((po) => po.poNumber)
      .filter((num) => num.startsWith(yearPrefix))
      .map((num) => parseInt(num.split('-')[2], 10))
      .filter((num) => !isNaN(num));

    const maxNumber = existingNumbers.length > 0 ? Math.max(...existingNumbers) : 0;
    const nextNumber = maxNumber + 1;

    return `${yearPrefix}${nextNumber.toString().padStart(4, '0')}`;
  },

  getPurchaseOrder: (id) => get().purchaseOrders.find((po) => po.id === id),
}));
