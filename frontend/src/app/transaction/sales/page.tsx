'use client';

import React, { useState, useMemo } from 'react';
import AdminLayout from '@/components/layout/AdminLayout';
import SessionTabs from '@/components/sales/SessionTabs';
import ProductSearch from '@/components/sales/ProductSearch';
import Cart from '@/components/sales/Cart';
import CartSummary from '@/components/sales/CartSummary';
import PaymentMethodSelector from '@/components/sales/PaymentMethodSelector';
import Receipt from '@/components/sales/Receipt';
import Button from '@/components/ui/Button';
import { useSalesStore, CheckoutResult } from '@/stores/useSalesStore';
import { useProductStore } from '@/stores/useProductStore';
import { useToastStore } from '@/stores/useToastStore';

export default function SalesPage() {
  const { sessions, activeSessionId, checkout, resetSession } = useSalesStore();
  const { getProduct } = useProductStore();
  const { addToast } = useToastStore();
  const [receipt, setReceipt] = useState<CheckoutResult | null>(null);

  const activeSession = sessions.find((s) => s.id === activeSessionId);

  // Check if checkout button should be disabled
  const checkoutDisabled = useMemo(() => {
    if (!activeSession) return true;
    if (activeSession.cart.length === 0) return true;
    if (!activeSession.paymentMethod) return true;

    // Check for stock errors
    for (const cartItem of activeSession.cart) {
      const product = getProduct(cartItem.productId);
      if (!product) return true;

      const variant = product.variants.find((v) => v.id === cartItem.variantId);
      if (!variant) return true;

      const selectedUnit = product.units.find((u) => u.id === cartItem.selectedUnitId);
      if (!selectedUnit) return true;

      const baseQty = cartItem.quantity * selectedUnit.toBaseUnit;
      if (baseQty > variant.currentStock) return true;
    }

    return false;
  }, [activeSession, getProduct]);

  const handleCheckout = () => {
    if (!activeSession || checkoutDisabled) return;

    try {
      const result = checkout(activeSession.id);
      setReceipt(result);
      addToast('Transaction completed successfully!', 'success');
    } catch (error) {
      addToast('Checkout failed. Please try again.', 'error');
    }
  };

  const handleCloseReceipt = () => {
    if (activeSession) {
      resetSession(activeSession.id);
    }
    setReceipt(null);
  };

  return (
    <AdminLayout>
      <div className="space-y-6">
        <h1 className="text-2xl font-bold text-gray-900">Sales</h1>

        <SessionTabs />

        <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
          <div className="lg:col-span-2 space-y-6">
            <ProductSearch sessionId={activeSessionId} />
            <Cart sessionId={activeSessionId} />
          </div>

          <div className="space-y-6">
            <CartSummary sessionId={activeSessionId} />
            <PaymentMethodSelector sessionId={activeSessionId} />
            <Button
              variant="primary"
              size="lg"
              onClick={handleCheckout}
              disabled={checkoutDisabled}
              className="w-full"
            >
              Checkout
            </Button>
          </div>
        </div>
      </div>

      {receipt && <Receipt receipt={receipt} onClose={handleCloseReceipt} />}
    </AdminLayout>
  );
}
