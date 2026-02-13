'use client';

import React, { useEffect } from 'react';
import { useRouter, useParams } from 'next/navigation';
import PurchaseOrderForm from '@/components/purchase/PurchaseOrderForm';
import { usePurchaseOrderStore } from '@/stores/usePurchaseOrderStore';

export default function EditPurchaseOrderPage() {
  const router = useRouter();
  const params = useParams();
  const { getPurchaseOrder } = usePurchaseOrderStore();
  const po = getPurchaseOrder(Number(params.id));

  useEffect(() => {
    if (!po) {
      router.push('/transaction/purchase');
      return;
    }

    if (po.status !== 'draft') {
      router.push(`/transaction/purchase/${params.id}`);
    }
  }, [po, router, params.id]);

  if (!po || po.status !== 'draft') {
    return null;
  }

  return <PurchaseOrderForm mode="edit" initialPO={po} />;
}
