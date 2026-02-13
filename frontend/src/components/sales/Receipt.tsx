'use client';

import React, { useEffect } from 'react';
import Button from '@/components/ui/Button';
import { CheckoutResult } from '@/stores/useSalesStore';
import { formatCurrency } from '@/utils/currency';

interface ReceiptProps {
  receipt: CheckoutResult;
  onClose: () => void;
}

export default function Receipt({ receipt, onClose }: ReceiptProps) {
  useEffect(() => {
    const handleEsc = (e: KeyboardEvent) => {
      if (e.key === 'Escape') onClose();
    };
    document.addEventListener('keydown', handleEsc);
    document.body.style.overflow = 'hidden';

    return () => {
      document.removeEventListener('keydown', handleEsc);
      document.body.style.overflow = '';
    };
  }, [onClose]);

  const formatDateTime = (date: Date) => {
    return new Intl.DateTimeFormat('id-ID', {
      dateStyle: 'medium',
      timeStyle: 'short',
    }).format(date);
  };

  const handlePrint = () => {
    window.print();
  };

  const formatPaymentMethod = (method: string) => {
    return method.charAt(0).toUpperCase() + method.slice(1);
  };

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center print:bg-white">
      <div
        className="fixed inset-0 bg-black/50 print:hidden"
        onClick={onClose}
      />
      <div className="relative z-10 bg-white rounded-lg shadow-xl w-full max-w-md mx-4 print:shadow-none print:max-w-full print:mx-0">
        <div className="p-8" id="receipt-print">
          <div className="font-mono text-sm">
            <div className="text-center mb-4">
              <div className="border-t-2 border-b-2 border-gray-800 py-1 mb-2">
                ================================
              </div>
              <div className="font-bold text-lg mb-1">SALES RECEIPT</div>
              <div className="border-t-2 border-b-2 border-gray-800 py-1 mt-2">
                ================================
              </div>
            </div>

            <div className="mb-4">
              <div className="flex justify-between">
                <span>Date:</span>
                <span>{formatDateTime(receipt.date)}</span>
              </div>
              <div className="flex justify-between">
                <span>Transaction:</span>
                <span>#{receipt.transactionId}</span>
              </div>
            </div>

            <div className="border-t border-gray-400 my-2">
              --------------------------------
            </div>

            <div className="space-y-3">
              {receipt.items.map((item, index) => (
                <div key={index} className="text-xs">
                  <div className="flex justify-between font-bold">
                    <span className="flex-1 truncate">{item.productName}</span>
                    <span className="ml-2">{item.quantity}</span>
                    <span className="ml-2 w-20 text-right">{formatCurrency(item.price)}</span>
                    <span className="ml-2 w-24 text-right font-bold">{formatCurrency(item.total)}</span>
                  </div>
                  <div className="text-gray-600 ml-2">
                    {item.sku}
                    {Object.keys(item.attributes).length > 0 && (
                      <span> | {Object.entries(item.attributes).map(([k, v]) => `${k}: ${v}`).join(', ')}</span>
                    )}
                  </div>
                  <div className="text-gray-600 ml-2">
                    Unit: {item.unitName}
                  </div>
                </div>
              ))}
            </div>

            <div className="border-t border-gray-400 my-2">
              --------------------------------
            </div>

            <div className="space-y-1">
              <div className="flex justify-between">
                <span>Total Items:</span>
                <span className="font-bold">{receipt.totalItems}</span>
              </div>
              <div className="flex justify-between">
                <span>Subtotal:</span>
                <span className="font-bold">{formatCurrency(receipt.subtotal)}</span>
              </div>
              <div className="flex justify-between text-lg font-bold mt-2">
                <span>Grand Total:</span>
                <span>{formatCurrency(receipt.grandTotal)}</span>
              </div>
              <div className="flex justify-between mt-2">
                <span>Payment:</span>
                <span className="font-bold">{formatPaymentMethod(receipt.paymentMethod)}</span>
              </div>
            </div>

            <div className="border-t-2 border-b-2 border-gray-800 py-1 mt-4">
              ================================
            </div>

            <div className="text-center mt-4 text-xs text-gray-600">
              Thank you for your purchase!
            </div>
          </div>
        </div>

        <div className="flex gap-3 px-8 pb-8 print:hidden">
          <Button variant="primary" onClick={handlePrint} className="flex-1">
            Print
          </Button>
          <Button variant="outline" onClick={onClose} className="flex-1">
            Close
          </Button>
        </div>
      </div>
    </div>
  );
}
