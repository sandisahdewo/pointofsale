'use client';

import { create } from 'zustand';
import { apiClient, PaginationMeta } from '@/lib/api';

export interface PurchaseOrderItem {
  id: string;
  purchaseOrderId?: number;
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

export interface ReceiveData {
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

export interface POStatusCounts {
  all: number;
  draft: number;
  sent: number;
  received: number;
  completed: number;
  cancelled: number;
}

export interface FetchPurchaseOrdersParams {
  page?: number;
  pageSize?: number;
  status?: POStatus | 'all';
  supplierId?: number;
  search?: string;
}

export interface FetchPurchaseOrdersResult {
  data: PurchaseOrder[];
  meta: PaginationMeta;
  statusCounts: POStatusCounts;
}

// API shape for a PO item from the backend
interface PurchaseOrderItemApi {
  id: string;
  purchaseOrderId?: number;
  productId: number;
  variantId: string;
  unitId: number;
  unitName: string;
  productName: string;
  variantLabel: string;
  sku: string;
  currentStock: number;
  orderedQty: number;
  price: number;
  receivedQty?: number;
  receivedPrice?: number;
  isVerified?: boolean;
}

// API shape for a PO from the backend
interface PurchaseOrderApi {
  id: number;
  poNumber: string;
  supplierId: number;
  supplier?: { id: number; name: string };
  date: string;
  status: POStatus;
  notes: string;
  receivedDate?: string;
  paymentMethod?: PaymentMethod;
  supplierBankAccountId?: string;
  subtotal?: number;
  totalItems?: number;
  items?: PurchaseOrderItemApi[];
  createdAt: string;
  updatedAt: string;
}

interface PurchaseOrderListResponse {
  data: PurchaseOrderApi[];
  meta: PaginationMeta;
  statusCounts: POStatusCounts;
}

// API shape for products returned by /purchase-orders/products endpoint
interface POProductVariantApi {
  id: string;
  sku?: string;
  barcode?: string;
  currentStock?: number;
  attributes?: Array<{ attributeName: string; attributeValue: string }>;
  pricingTiers?: Array<{ minQty: number; value: number }>;
}

interface POProductUnitApi {
  id: number;
  name: string;
  isBase: boolean;
  conversionFactor?: number;
}

export interface POProductApi {
  id: number;
  name: string;
  categoryId: number;
  units?: POProductUnitApi[];
  variants?: POProductVariantApi[];
}

interface CreatePurchaseOrderInput {
  supplierId: number;
  date: string;
  notes: string;
  items: Array<{
    productId: number;
    variantId: string;
    unitId: number;
    orderedQty: number;
    price: number;
  }>;
}

interface PurchaseOrderState {
  purchaseOrders: PurchaseOrder[];

  // Async API methods
  fetchPurchaseOrders: (params?: FetchPurchaseOrdersParams) => Promise<FetchPurchaseOrdersResult>;
  fetchPurchaseOrder: (id: number) => Promise<PurchaseOrder>;
  createPurchaseOrderRemote: (input: CreatePurchaseOrderInput) => Promise<PurchaseOrder>;
  updatePurchaseOrderRemote: (id: number, input: CreatePurchaseOrderInput) => Promise<PurchaseOrder>;
  deletePurchaseOrderRemote: (id: number) => Promise<void>;
  updateStatusRemote: (id: number, status: POStatus) => Promise<PurchaseOrder>;
  receivePurchaseOrderRemote: (id: number, receiveData: ReceiveData) => Promise<PurchaseOrder>;
  fetchProductsForPO: (supplierId?: number, search?: string) => Promise<POProductApi[]>;

  // Local helpers
  getNextPoNumber: () => string;
  getPurchaseOrder: (id: number) => PurchaseOrder | undefined;

  // Deprecated synchronous methods kept for backward compatibility
  addPurchaseOrder: (po: Omit<PurchaseOrder, 'id' | 'poNumber' | 'createdAt' | 'updatedAt'>) => void;
  updatePurchaseOrder: (id: number, po: Partial<PurchaseOrder>) => void;
  deletePurchaseOrder: (id: number) => void;
  updateStatus: (id: number, status: POStatus) => void;
  receivePurchaseOrder: (id: number, receiveData: ReceiveData) => void;
  completePurchaseOrder: (id: number) => void;
  cancelPurchaseOrder: (id: number) => void;
}

function normalizePurchaseOrderItem(item: PurchaseOrderItemApi): PurchaseOrderItem {
  return {
    id: item.id,
    purchaseOrderId: item.purchaseOrderId,
    productId: item.productId,
    productName: item.productName,
    variantId: item.variantId,
    variantLabel: item.variantLabel,
    sku: item.sku,
    unitId: String(item.unitId),
    unitName: item.unitName,
    currentStock: Number(item.currentStock ?? 0),
    orderedQty: Number(item.orderedQty),
    price: Number(item.price),
    receivedQty: item.receivedQty != null ? Number(item.receivedQty) : undefined,
    receivedPrice: item.receivedPrice != null ? Number(item.receivedPrice) : undefined,
    isVerified: item.isVerified,
  };
}

function normalizePurchaseOrder(api: PurchaseOrderApi): PurchaseOrder {
  return {
    id: api.id,
    poNumber: api.poNumber,
    supplierId: api.supplierId,
    supplierName: api.supplier?.name ?? '',
    date: api.date,
    status: api.status,
    notes: api.notes ?? '',
    receivedDate: api.receivedDate,
    paymentMethod: api.paymentMethod,
    supplierBankAccountId: api.supplierBankAccountId,
    subtotal: api.subtotal != null ? Number(api.subtotal) : undefined,
    totalItems: api.totalItems != null ? Number(api.totalItems) : undefined,
    items: (api.items ?? []).map(normalizePurchaseOrderItem),
    createdAt: api.createdAt,
    updatedAt: api.updatedAt,
  };
}

function buildPOQuery(params: FetchPurchaseOrdersParams = {}): string {
  const query = new URLSearchParams();
  if (params.page) query.set('page', String(params.page));
  if (params.pageSize) query.set('pageSize', String(params.pageSize));
  if (params.status && params.status !== 'all') query.set('status', params.status);
  if (params.supplierId) query.set('supplierId', String(params.supplierId));
  if (params.search) query.set('search', params.search);
  const qs = query.toString();
  return qs ? `?${qs}` : '';
}

export const usePurchaseOrderStore = create<PurchaseOrderState>((set, get) => ({
  purchaseOrders: [],

  fetchPurchaseOrders: async (params = {}) => {
    const response = await apiClient<PurchaseOrderApi[]>(
      `/api/v1/purchase-orders${buildPOQuery(params)}`
    );
    const raw = response as unknown as PurchaseOrderListResponse;
    const data = raw.data.map(normalizePurchaseOrder);

    // Merge fetched page into local store (upsert)
    set((state) => {
      const existingIds = new Set(data.map((po) => po.id));
      const kept = state.purchaseOrders.filter((po) => !existingIds.has(po.id));
      return { purchaseOrders: [...kept, ...data] };
    });

    return {
      data,
      meta: raw.meta,
      statusCounts: raw.statusCounts ?? {
        all: raw.meta.totalItems,
        draft: 0,
        sent: 0,
        received: 0,
        completed: 0,
        cancelled: 0,
      },
    };
  },

  fetchPurchaseOrder: async (id) => {
    const response = await apiClient<PurchaseOrderApi>(`/api/v1/purchase-orders/${id}`);
    const po = normalizePurchaseOrder(response.data);

    set((state) => {
      const exists = state.purchaseOrders.some((item) => item.id === id);
      if (exists) {
        return {
          purchaseOrders: state.purchaseOrders.map((item) => (item.id === id ? po : item)),
        };
      }
      return { purchaseOrders: [...state.purchaseOrders, po] };
    });

    return po;
  },

  createPurchaseOrderRemote: async (input) => {
    const response = await apiClient<PurchaseOrderApi>('/api/v1/purchase-orders', {
      method: 'POST',
      body: JSON.stringify(input),
    });

    const po = normalizePurchaseOrder(response.data);
    set((state) => ({ purchaseOrders: [...state.purchaseOrders, po] }));
    return po;
  },

  updatePurchaseOrderRemote: async (id, input) => {
    const response = await apiClient<PurchaseOrderApi>(`/api/v1/purchase-orders/${id}`, {
      method: 'PUT',
      body: JSON.stringify(input),
    });

    const po = normalizePurchaseOrder(response.data);
    set((state) => ({
      purchaseOrders: state.purchaseOrders.map((item) => (item.id === id ? po : item)),
    }));
    return po;
  },

  deletePurchaseOrderRemote: async (id) => {
    await apiClient(`/api/v1/purchase-orders/${id}`, {
      method: 'DELETE',
    });

    set((state) => ({
      purchaseOrders: state.purchaseOrders.filter((item) => item.id !== id),
    }));
  },

  updateStatusRemote: async (id, status) => {
    const response = await apiClient<PurchaseOrderApi>(
      `/api/v1/purchase-orders/${id}/status`,
      {
        method: 'PATCH',
        body: JSON.stringify({ status }),
      }
    );

    const po = normalizePurchaseOrder(response.data);
    set((state) => ({
      purchaseOrders: state.purchaseOrders.map((item) => (item.id === id ? po : item)),
    }));
    return po;
  },

  receivePurchaseOrderRemote: async (id, receiveData) => {
    const response = await apiClient<PurchaseOrderApi>(
      `/api/v1/purchase-orders/${id}/receive`,
      {
        method: 'POST',
        body: JSON.stringify({
          receivedDate: receiveData.receivedDate,
          paymentMethod: receiveData.paymentMethod,
          supplierBankAccountId: receiveData.supplierBankAccountId,
          items: receiveData.items.map((item) => ({
            itemId: item.id,
            receivedQty: item.receivedQty,
            receivedPrice: item.receivedPrice,
            isVerified: item.isVerified,
          })),
        }),
      }
    );

    const po = normalizePurchaseOrder(response.data);
    set((state) => ({
      purchaseOrders: state.purchaseOrders.map((item) => (item.id === id ? po : item)),
    }));
    return po;
  },

  fetchProductsForPO: async (supplierId?, search?) => {
    const query = new URLSearchParams();
    if (supplierId) query.set('supplierId', String(supplierId));
    if (search) query.set('search', search);
    const qs = query.toString();
    const endpoint = `/api/v1/purchase-orders/products${qs ? `?${qs}` : ''}`;
    const response = await apiClient<POProductApi[]>(endpoint);
    const raw = response as unknown as { data: POProductApi[] };
    return raw.data ?? (response.data as unknown as POProductApi[]);
  },

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

  // Deprecated synchronous methods kept for backward compatibility
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

      const subtotal = updatedItems.reduce(
        (sum, item) => sum + (item.receivedQty || 0) * (item.receivedPrice || 0),
        0
      );
      const totalItems = updatedItems.reduce(
        (sum, item) => sum + (item.receivedQty || 0),
        0
      );

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
}));
