'use client';

import React, { useEffect } from 'react';
import { useRouter } from 'next/navigation';
import Header from './Header';
import Sidebar from './Sidebar';
import Footer from './Footer';
import ToastContainer from '@/components/ui/Toast';
import { useSidebarStore } from '@/stores/useSidebarStore';
import { useAuthStore } from '@/stores/useAuthStore';

interface AdminLayoutProps {
  children: React.ReactNode;
}

export default function AdminLayout({ children }: AdminLayoutProps) {
  const { isOpen } = useSidebarStore();
  const { isAuthenticated, isInitialized, initialize } = useAuthStore();
  const router = useRouter();

  useEffect(() => {
    initialize();
  }, [initialize]);

  useEffect(() => {
    if (isInitialized && !isAuthenticated) {
      router.replace('/login');
    }
  }, [isInitialized, isAuthenticated, router]);

  if (!isInitialized) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-gray-50">
        <div className="text-center">
          <div className="w-12 h-12 border-4 border-blue-600 border-t-transparent rounded-full animate-spin mx-auto mb-4"></div>
          <p className="text-gray-600">Loading...</p>
        </div>
      </div>
    );
  }

  if (!isAuthenticated) {
    return null;
  }

  return (
    <div className="min-h-screen flex flex-col">
      <Header />
      <div className="flex flex-1 pt-16">
        <Sidebar />
        <main
          className={`flex-1 transition-all duration-300 ${
            isOpen ? 'ml-60' : 'ml-0'
          }`}
        >
          <div className="p-6 min-h-[calc(100vh-8rem)]">{children}</div>
          <Footer />
        </main>
      </div>
      <ToastContainer />
    </div>
  );
}
