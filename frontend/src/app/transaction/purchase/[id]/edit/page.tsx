'use client';

import React, { useState, useEffect } from 'react';
import { useRouter, useParams } from 'next/navigation';
import AdminLayout from '@/components/layout/AdminLayout';
import PurchaseOrderForm from '@/components/purchase/PurchaseOrderForm';
import { usePurchaseOrderStore, PurchaseOrder } from '@/stores/usePurchaseOrderStore';
import { useToastStore } from '@/stores/useToastStore';
import { ApiError } from '@/lib/api';

export default function EditPurchaseOrderPage() {
  const router = useRouter();
  const params = useParams();
  const { fetchPurchaseOrder } = usePurchaseOrderStore();
  const { addToast } = useToastStore();

  const [po, setPo] = useState<PurchaseOrder | null>(null);
  const [isLoading, setIsLoading] = useState(true);

  useEffect(() => {
    const loadPO = async () => {
      const id = Number(params.id);
      if (!id) {
        router.push('/transaction/purchase');
        return;
      }
      try {
        setIsLoading(true);
        const loaded = await fetchPurchaseOrder(id);
        if (loaded.status !== 'draft') {
          router.push(`/transaction/purchase/${id}`);
          return;
        }
        setPo(loaded);
      } catch (error) {
        if (error instanceof ApiError) {
          addToast(error.message, 'error');
        } else {
          addToast('Failed to load purchase order', 'error');
        }
        router.push('/transaction/purchase');
      } finally {
        setIsLoading(false);
      }
    };

    void loadPO();
  }, [params.id, fetchPurchaseOrder, addToast, router]);

  if (isLoading) {
    return (
      <AdminLayout>
        <div className="text-center py-12 text-gray-500">Loading purchase order...</div>
      </AdminLayout>
    );
  }

  if (!po || po.status !== 'draft') {
    return null;
  }

  return <PurchaseOrderForm mode="edit" initialPO={po} />;
}
