'use client';

import { create } from 'zustand';
import { useProductStore } from './useProductStore';

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
  transactionId: number;
  date: Date;
  items: CheckoutItem[];
  totalItems: number;
  subtotal: number;
  grandTotal: number;
  paymentMethod: 'cash' | 'card' | 'qris';
}

interface SalesState {
  sessions: SalesSession[];
  activeSessionId: number;
  nextSessionNumber: number;
  transactionCounter: number;

  createSession: () => void;
  closeSession: (id: number) => void;
  setActiveSession: (id: number) => void;

  addToCart: (sessionId: number, productId: number, variantId: string) => void;
  updateCartItemQuantity: (sessionId: number, variantId: string, quantity: number) => void;
  updateCartItemUnit: (sessionId: number, variantId: string, unitId: string) => void;
  removeFromCart: (sessionId: number, variantId: string) => void;

  setPaymentMethod: (sessionId: number, method: 'cash' | 'card' | 'qris') => void;
  checkout: (sessionId: number) => CheckoutResult;
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
  transactionCounter: 1,

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
      // Create a new session if this was the last one
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
      // Update active session if the closed one was active
      const newActiveId = state.activeSessionId === id
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

  addToCart: (sessionId, productId, variantId) => {
    const state = get();
    const productStore = useProductStore.getState();
    const product = productStore.getProduct(productId);

    if (!product) return;

    const baseUnit = product.units.find((u) => u.isBase);
    if (!baseUnit) return;

    set({
      sessions: state.sessions.map((session) => {
        if (session.id !== sessionId) return session;

        const existingItem = session.cart.find((item) => item.variantId === variantId);

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
                selectedUnitId: baseUnit.id,
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

  checkout: (sessionId) => {
    const state = get();
    const session = state.sessions.find((s) => s.id === sessionId);

    if (!session || !session.paymentMethod) {
      throw new Error('Invalid session or payment method not selected');
    }

    const productStore = useProductStore.getState();
    const checkoutItems: CheckoutItem[] = [];
    let subtotal = 0;

    // Process each cart item
    for (const cartItem of session.cart) {
      const product = productStore.getProduct(cartItem.productId);
      if (!product) continue;

      const variant = product.variants.find((v) => v.id === cartItem.variantId);
      if (!variant) continue;

      const selectedUnit = product.units.find((u) => u.id === cartItem.selectedUnitId);
      if (!selectedUnit) continue;

      // Calculate base quantity for pricing
      const baseQty = cartItem.quantity * selectedUnit.toBaseUnit;

      // Find the highest tier where baseQty >= tier.minQty
      let applicableTier = variant.pricingTiers[0];
      for (const tier of variant.pricingTiers) {
        if (baseQty >= tier.minQty) {
          applicableTier = tier;
        }
      }

      // Calculate per-unit price and total
      const perUnitPrice = applicableTier.value * selectedUnit.toBaseUnit;
      const itemTotal = cartItem.quantity * perUnitPrice;

      checkoutItems.push({
        productName: product.name,
        sku: variant.sku,
        attributes: variant.attributes,
        quantity: cartItem.quantity,
        unitName: selectedUnit.name,
        price: perUnitPrice,
        total: itemTotal,
      });

      subtotal += itemTotal;

      // Deduct stock from product store
      productStore.updateProduct(product.id, {
        variants: product.variants.map((v) =>
          v.id === variant.id
            ? { ...v, currentStock: v.currentStock - baseQty }
            : v
        ),
      });
    }

    const result: CheckoutResult = {
      transactionId: state.transactionCounter,
      date: new Date(),
      items: checkoutItems,
      totalItems: checkoutItems.reduce((sum, item) => sum + item.quantity, 0),
      subtotal,
      grandTotal: subtotal,
      paymentMethod: session.paymentMethod,
    };

    // Increment transaction counter
    set({ transactionCounter: state.transactionCounter + 1 });

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
