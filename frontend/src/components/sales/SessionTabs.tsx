'use client';

import React, { useState, useEffect, useRef } from 'react';
import { useSalesStore } from '@/stores/useSalesStore';

export default function SessionTabs() {
  const { sessions, activeSessionId, createSession, closeSession, setActiveSession } = useSalesStore();
  const [confirmingCloseId, setConfirmingCloseId] = useState<number | null>(null);
  const tabsRef = useRef<HTMLDivElement>(null);

  // Reset confirmation on clicks outside the tabs container
  useEffect(() => {
    const handleClickOutside = (e: MouseEvent) => {
      if (tabsRef.current && !tabsRef.current.contains(e.target as Node)) {
        setConfirmingCloseId(null);
      }
    };

    document.addEventListener('mousedown', handleClickOutside);
    return () => document.removeEventListener('mousedown', handleClickOutside);
  }, []);

  const handleCreateSession = () => {
    if (sessions.length < 10) {
      createSession();
    }
  };

  const handleTabClick = (sessionId: number) => {
    setConfirmingCloseId(null);
    setActiveSession(sessionId);
  };

  const handleCloseClick = (e: React.MouseEvent, sessionId: number) => {
    e.stopPropagation();

    if (confirmingCloseId === sessionId) {
      // Second click - confirm close
      closeSession(sessionId);
      setConfirmingCloseId(null);
    } else {
      // First click - show confirmation
      setConfirmingCloseId(sessionId);
    }
  };

  return (
    <div className="bg-white border-b border-gray-200 mb-6">
      <div ref={tabsRef} className="flex gap-0 overflow-x-auto">
        {sessions.map((session) => {
          const isActive = session.id === activeSessionId;
          const isConfirming = confirmingCloseId === session.id;

          return (
            <div
              key={session.id}
              role="tab"
              tabIndex={0}
              onClick={() => handleTabClick(session.id)}
              onKeyDown={(e) => { if (e.key === 'Enter' || e.key === ' ') handleTabClick(session.id); }}
              className={`flex items-center gap-2 px-4 py-3 text-sm font-medium transition-colors whitespace-nowrap cursor-pointer ${
                isActive
                  ? 'border-b-2 border-blue-600 text-blue-600'
                  : 'text-gray-600 hover:text-gray-900 hover:bg-gray-50'
              }`}
            >
              <span>{session.name}</span>
              <button
                onClick={(e) => handleCloseClick(e, session.id)}
                className="ml-1 text-gray-400 hover:text-gray-600 transition-colors cursor-pointer"
              >
                {isConfirming ? (
                  <svg className="w-4 h-4 text-green-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 13l4 4L19 7" />
                  </svg>
                ) : (
                  <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
                  </svg>
                )}
              </button>
            </div>
          );
        })}

        <button
          onClick={handleCreateSession}
          disabled={sessions.length >= 10}
          className="flex items-center justify-center px-4 py-3 text-gray-600 hover:text-gray-900 hover:bg-gray-50 transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
        >
          <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 4v16m8-8H4" />
          </svg>
        </button>
      </div>
    </div>
  );
}
