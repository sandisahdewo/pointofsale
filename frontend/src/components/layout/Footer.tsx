'use client';

import React from 'react';

export default function Footer() {
  return (
    <footer className="py-4 px-6 text-center text-sm text-gray-500 border-t border-gray-200 bg-white">
      &copy; {new Date().getFullYear()} Point of Sale. All rights reserved.
    </footer>
  );
}
