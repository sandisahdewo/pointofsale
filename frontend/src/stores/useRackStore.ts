'use client';

import { create } from 'zustand';
import { initialRacks } from '@/data/racks';

export interface Rack {
  id: number;
  name: string;
  code: string;
  location: string;
  capacity: number;
  description: string;
  active: boolean;
  createdAt: string;
}

interface RackState {
  racks: Rack[];
  addRack: (rack: Omit<Rack, 'id' | 'createdAt'>) => void;
  updateRack: (id: number, rack: Partial<Omit<Rack, 'id' | 'createdAt'>>) => void;
  deleteRack: (id: number) => void;
  getActiveRacks: () => Rack[];
  isCodeUnique: (code: string, excludeId?: number) => boolean;
}

export const useRackStore = create<RackState>((set, get) => ({
  racks: initialRacks,

  addRack: (rack) =>
    set((state) => {
      const maxId = state.racks.reduce((max, r) => Math.max(max, r.id), 0);
      const newRack: Rack = {
        ...rack,
        id: maxId + 1,
        createdAt: new Date().toISOString(),
      };
      return { racks: [...state.racks, newRack] };
    }),

  updateRack: (id, data) =>
    set((state) => ({
      racks: state.racks.map((r) => (r.id === id ? { ...r, ...data } : r)),
    })),

  deleteRack: (id) =>
    set((state) => ({
      racks: state.racks.filter((r) => r.id !== id),
    })),

  getActiveRacks: () => get().racks.filter((r) => r.active),

  isCodeUnique: (code: string, excludeId?: number) => {
    const normalizedCode = code.trim().toLowerCase();
    return !get().racks.some(
      (r) => r.code.toLowerCase() === normalizedCode && r.id !== excludeId
    );
  },
}));
