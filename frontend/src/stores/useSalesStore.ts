'use client';

import { create } from 'zustand';
import { apiClient, ApiError } from '@/lib/api';

// --- API shapes from backend ---

export interface ProductImage {
  id?: number;
  imageUrl: string;
  sortOrder: number;
}

export interface ProductUnit {
  id: number;
  name: string;
  isBase: boolean;
  toBaseUnit: number;
}

export interface VariantAttribute {
  attributeName: string;
  attributeValue: string;
}

export interface PricingTier {
  minQty: number;
  value: number;
}

export interface SearchVariant {
  id: string;
  sku: string;
  barcode: string;
  currentStock: number;
  attributes: VariantAttribute[];
  images: { imageUrl: string; sortOrder: number }[];
  pricingTiers: PricingTier[];
}

export interface SearchProduct {
  id: number;
  name: string;
  description: string;
  hasVariants: boolean;
  priceSetting: string;
  markupType: string | null;
  images: ProductImage[];
  units: ProductUnit[];
  variants: SearchVariant[];
}

// --- Cart and session types ---

export interface CartItem {
  productId: number;
  variantId: string;
  quantity: number;
  selectedUnitId: string;
}

export interface SalesSession {
  id: number;
  name: string;
  cart: CartItem[];
  paymentMethod: 'cash' | 'card' | 'qris' | null;
}

// --- Checkout types ---

export interface CheckoutItem {
  productName: string;
  sku: string;
  attributes: Record<string, string>;
  quantity: number;
  unitName: string;
  price: number;
  total: number;
}

export interface CheckoutResult {
  transactionNumber: string;
  date: Date;
  items: CheckoutItem[];
  totalItems: number;
  subtotal: number;
  grandTotal: number;
  paymentMethod: 'cash' | 'card' | 'qris';
}

// --- Backend SalesTransaction shape ---

interface SalesTransactionItem {
  id: string;
  productName: string;
  variantLabel: string;
  sku: string;
  unitName: string;
  quantity: number;
  baseQty: number;
  unitPrice: number;
  totalPrice: number;
}

interface SalesTransaction {
  id: number;
  transactionNumber: string;
  date: string;
  subtotal: number;
  grandTotal: number;
  totalItems: number;
  paymentMethod: 'cash' | 'card' | 'qris';
  items: SalesTransactionItem[];
  createdAt: string;
}

// --- CheckoutInput for the API ---

interface CheckoutInput {
  paymentMethod: 'cash' | 'card' | 'qris';
  items: {
    productId: number;
    variantId: string;
    unitId: number;
    quantity: number;
  }[];
}

// --- Store state ---

interface SalesState {
  sessions: SalesSession[];
  activeSessionId: number;
  nextSessionNumber: number;

  // Product cache: productId → SearchProduct (populated from search results)
  productCache: Record<number, SearchProduct>;

  // Search state
  searchResults: SearchProduct[];
  isSearching: boolean;

  createSession: () => void;
  closeSession: (id: number) => void;
  setActiveSession: (id: number) => void;

  searchProducts: (query: string) => Promise<void>;
  clearSearch: () => void;

  addToCart: (sessionId: number, productId: number, variantId: string) => void;
  updateCartItemQuantity: (sessionId: number, variantId: string, quantity: number) => void;
  updateCartItemUnit: (sessionId: number, variantId: string, unitId: string) => void;
  removeFromCart: (sessionId: number, variantId: string) => void;

  setPaymentMethod: (sessionId: number, method: 'cash' | 'card' | 'qris') => void;
  checkout: (sessionId: number) => Promise<CheckoutResult>;
  resetSession: (sessionId: number) => void;
}

export const useSalesStore = create<SalesState>((set, get) => ({
  sessions: [
    {
      id: 1,
      name: 'Session 1',
      cart: [],
      paymentMethod: null,
    },
  ],
  activeSessionId: 1,
  nextSessionNumber: 2,
  productCache: {},
  searchResults: [],
  isSearching: false,

  createSession: () => {
    const state = get();
    if (state.sessions.length >= 10) {
      return;
    }
    const newSession: SalesSession = {
      id: state.nextSessionNumber,
      name: `Session ${state.nextSessionNumber}`,
      cart: [],
      paymentMethod: null,
    };
    set({
      sessions: [...state.sessions, newSession],
      nextSessionNumber: state.nextSessionNumber + 1,
      activeSessionId: newSession.id,
    });
  },

  closeSession: (id) => {
    const state = get();
    const remainingSessions = state.sessions.filter((s) => s.id !== id);

    if (remainingSessions.length === 0) {
      const newSession: SalesSession = {
        id: state.nextSessionNumber,
        name: `Session ${state.nextSessionNumber}`,
        cart: [],
        paymentMethod: null,
      };
      set({
        sessions: [newSession],
        nextSessionNumber: state.nextSessionNumber + 1,
        activeSessionId: newSession.id,
      });
    } else {
      const newActiveId =
        state.activeSessionId === id
          ? remainingSessions[0].id
          : state.activeSessionId;
      set({
        sessions: remainingSessions,
        activeSessionId: newActiveId,
      });
    }
  },

  setActiveSession: (id) => {
    set({ activeSessionId: id });
  },

  searchProducts: async (query: string) => {
    set({ isSearching: true });
    try {
      const response = await apiClient<SearchProduct[]>(
        `/api/v1/sales/products/search?q=${encodeURIComponent(query)}`
      );
      const results = response.data;

      // Populate the product cache with the new results
      const newCache: Record<number, SearchProduct> = { ...get().productCache };
      for (const product of results) {
        newCache[product.id] = product;
      }

      set({ searchResults: results, productCache: newCache, isSearching: false });
    } catch (err) {
      set({ searchResults: [], isSearching: false });
      if (err instanceof ApiError) {
        throw err;
      }
      throw err;
    }
  },

  clearSearch: () => {
    set({ searchResults: [] });
  },

  addToCart: (sessionId, productId, variantId) => {
    const state = get();
    const product = state.productCache[productId];
    if (!product) return;

    const baseUnit = product.units.find((u) => u.isBase);
    if (!baseUnit) return;

    set({
      sessions: state.sessions.map((session) => {
        if (session.id !== sessionId) return session;

        const existingItem = session.cart.find(
          (item) => item.variantId === variantId
        );

        if (existingItem) {
          return {
            ...session,
            cart: session.cart.map((item) =>
              item.variantId === variantId
                ? { ...item, quantity: item.quantity + 1 }
                : item
            ),
          };
        } else {
          return {
            ...session,
            cart: [
              ...session.cart,
              {
                productId,
                variantId,
                quantity: 1,
                selectedUnitId: String(baseUnit.id),
              },
            ],
          };
        }
      }),
    });
  },

  updateCartItemQuantity: (sessionId, variantId, quantity) => {
    const state = get();
    set({
      sessions: state.sessions.map((session) => {
        if (session.id !== sessionId) return session;
        return {
          ...session,
          cart: session.cart.map((item) =>
            item.variantId === variantId
              ? { ...item, quantity: Math.max(1, quantity) }
              : item
          ),
        };
      }),
    });
  },

  updateCartItemUnit: (sessionId, variantId, unitId) => {
    const state = get();
    set({
      sessions: state.sessions.map((session) => {
        if (session.id !== sessionId) return session;
        return {
          ...session,
          cart: session.cart.map((item) =>
            item.variantId === variantId
              ? { ...item, selectedUnitId: unitId }
              : item
          ),
        };
      }),
    });
  },

  removeFromCart: (sessionId, variantId) => {
    const state = get();
    set({
      sessions: state.sessions.map((session) => {
        if (session.id !== sessionId) return session;
        return {
          ...session,
          cart: session.cart.filter((item) => item.variantId !== variantId),
        };
      }),
    });
  },

  setPaymentMethod: (sessionId, method) => {
    const state = get();
    set({
      sessions: state.sessions.map((session) =>
        session.id === sessionId
          ? { ...session, paymentMethod: method }
          : session
      ),
    });
  },

  checkout: async (sessionId) => {
    const state = get();
    const session = state.sessions.find((s) => s.id === sessionId);

    if (!session || !session.paymentMethod) {
      throw new Error('Invalid session or payment method not selected');
    }

    const checkoutInput: CheckoutInput = {
      paymentMethod: session.paymentMethod,
      items: session.cart.map((cartItem) => ({
        productId: cartItem.productId,
        variantId: cartItem.variantId,
        unitId: Number(cartItem.selectedUnitId),
        quantity: cartItem.quantity,
      })),
    };

    const response = await apiClient<SalesTransaction>(
      '/api/v1/sales/checkout',
      {
        method: 'POST',
        body: JSON.stringify(checkoutInput),
      }
    );

    const tx = response.data;

    // Map backend SalesTransaction → frontend CheckoutResult
    const checkoutItems: CheckoutItem[] = tx.items.map((item) => {
      // variantLabel from backend is "Red / S" — convert to Record<string,string>
      // We store it as a single entry for display consistency
      const attributes: Record<string, string> = {};
      if (item.variantLabel) {
        attributes['Variant'] = item.variantLabel;
      }

      return {
        productName: item.productName,
        sku: item.sku,
        attributes,
        quantity: item.quantity,
        unitName: item.unitName,
        price: item.unitPrice,
        total: item.totalPrice,
      };
    });

    const result: CheckoutResult = {
      transactionNumber: tx.transactionNumber,
      date: new Date(tx.date),
      items: checkoutItems,
      totalItems: tx.totalItems,
      subtotal: tx.subtotal,
      grandTotal: tx.grandTotal,
      paymentMethod: tx.paymentMethod,
    };

    return result;
  },

  resetSession: (sessionId) => {
    const state = get();
    set({
      sessions: state.sessions.map((session) =>
        session.id === sessionId
          ? { ...session, cart: [], paymentMethod: null }
          : session
      ),
    });
  },
}));
