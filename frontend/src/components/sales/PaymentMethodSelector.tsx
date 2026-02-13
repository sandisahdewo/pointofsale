'use client';

import React from 'react';
import { useSalesStore } from '@/stores/useSalesStore';

interface PaymentMethodSelectorProps {
  sessionId: number;
}

export default function PaymentMethodSelector({ sessionId }: PaymentMethodSelectorProps) {
  const sessions = useSalesStore((state) => state.sessions);
  const setPaymentMethod = useSalesStore((state) => state.setPaymentMethod);

  const session = sessions.find((s) => s.id === sessionId);

  if (!session) {
    return null;
  }

  const paymentMethods = [
    {
      value: 'cash' as const,
      label: 'Cash',
      icon: (
        <svg className="w-8 h-8" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path
            strokeLinecap="round"
            strokeLinejoin="round"
            strokeWidth={2}
            d="M17 9V7a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2m2 4h10a2 2 0 002-2v-6a2 2 0 00-2-2H9a2 2 0 00-2 2v6a2 2 0 002 2zm7-5a2 2 0 11-4 0 2 2 0 014 0z"
          />
        </svg>
      ),
    },
    {
      value: 'card' as const,
      label: 'Card',
      icon: (
        <svg className="w-8 h-8" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path
            strokeLinecap="round"
            strokeLinejoin="round"
            strokeWidth={2}
            d="M3 10h18M7 15h1m4 0h1m-7 4h12a3 3 0 003-3V8a3 3 0 00-3-3H6a3 3 0 00-3 3v8a3 3 0 003 3z"
          />
        </svg>
      ),
    },
    {
      value: 'qris' as const,
      label: 'QRIS',
      icon: (
        <svg className="w-8 h-8" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path
            strokeLinecap="round"
            strokeLinejoin="round"
            strokeWidth={2}
            d="M12 4v1m6 11h2m-6 0h-2v4m0-11v3m0 0h.01M12 12h4.01M16 20h4M4 12h4m12 0h.01M5 8h2a1 1 0 001-1V5a1 1 0 00-1-1H5a1 1 0 00-1 1v2a1 1 0 001 1zm12 0h2a1 1 0 001-1V5a1 1 0 00-1-1h-2a1 1 0 00-1 1v2a1 1 0 001 1zM5 20h2a1 1 0 001-1v-2a1 1 0 00-1-1H5a1 1 0 00-1 1v2a1 1 0 001 1z"
          />
        </svg>
      ),
    },
  ];

  return (
    <div className="space-y-3">
      <h3 className="text-sm font-medium text-gray-700">Payment Method</h3>
      <div className="grid grid-cols-3 gap-4">
        {paymentMethods.map((method) => {
          const isSelected = session.paymentMethod === method.value;
          return (
            <button
              key={method.value}
              onClick={() => setPaymentMethod(sessionId, method.value)}
              className={`flex flex-col items-center justify-center p-4 rounded-lg border-2 transition-all ${
                isSelected
                  ? 'border-blue-500 bg-blue-50 text-blue-700'
                  : 'border-gray-200 bg-white text-gray-600 hover:border-gray-300 hover:bg-gray-50'
              }`}
            >
              <div className="mb-2">{method.icon}</div>
              <span className="text-sm font-medium">{method.label}</span>
            </button>
          );
        })}
      </div>
    </div>
  );
}
