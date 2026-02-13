'use client';

import React from 'react';
import Header from './Header';
import Sidebar from './Sidebar';
import Footer from './Footer';
import ToastContainer from '@/components/ui/Toast';
import { useSidebarStore } from '@/stores/useSidebarStore';

interface AdminLayoutProps {
  children: React.ReactNode;
}

export default function AdminLayout({ children }: AdminLayoutProps) {
  const { isOpen } = useSidebarStore();

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
