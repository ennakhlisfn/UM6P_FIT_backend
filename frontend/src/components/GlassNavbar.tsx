"use client";

import Link from "next/link";
import { useEffect, useState } from "react";
import { api } from "@/lib/api";

type Notification = {
  id: number;
  message: string;
  isRead: boolean;
  createdAt: string;
};

export default function GlassNavbar() {
  const [notifications, setNotifications] = useState<Notification[]>([]);
  const [isClient, setIsClient] = useState(false);
  
  useEffect(() => {
    setIsClient(true);
    
    // If the browser registers an authenticated user, natively execute background polling
    if (typeof window !== "undefined" && localStorage.getItem("auth_token")) {
      const fetchNotifs = async () => {
        try {
          const userId = localStorage.getItem("auth_userid");
          if (!userId) return;
          const data = await api.request(`/users/${userId}/notifications`);
          setNotifications(data || []);
        } catch (e) {
          console.error("Failed executing generic UI poll for notifications.");
        }
      };
      
      fetchNotifs();
      // Periodically ping for notifications every 30 seconds alongside Cron
      const timer = setInterval(fetchNotifs, 30000);
      return () => clearInterval(timer);
    }
  }, []);

  const unreadCount = notifications.filter((n) => !n.isRead).length;

  return (
    <nav className="glass-nav flex items-center justify-between px-8 py-4">
      <Link href="/" className="text-2xl font-bold text-gradient">
        UM6P FIT
      </Link>
      
      {isClient && typeof window !== "undefined" && localStorage.getItem("auth_token") ? (
        <div className="flex text-lg gap-6 items-center">
          <Link href="/dashboard" className="hover:text-[var(--primary)] transition-colors font-semibold">Dashboard</Link>
          <Link href="/leaderboard" className="hover:text-[var(--primary)] transition-colors font-semibold">Leaderboard</Link>
          
          <div className="relative group cursor-pointer p-2 glass-panel hover-lift text-sm font-semibold">
            <span className="text-lg">🔔</span> Inbox
            {unreadCount > 0 && (
              <span className="absolute -top-2 -right-2 bg-red-600 text-white text-xs px-2 py-1 rounded-full pulse-glow">
                {unreadCount}
              </span>
            )}
            
            {/* The Hidden Dropdown Interactor Menu */}
            <div className="absolute right-0 top-12 w-64 glass-panel p-4 hidden group-hover:block transition-all shadow-lg rounded-lg border-white/20 z-50 overflow-y-auto max-h-60">
                {notifications.length === 0 ? (
                  <p className="text-white/60 text-sm italic">Inbox is completely empty...</p>
                ) : (
                  notifications.map(n => (
                    <div key={n.id} className={`text-sm mb-2 p-3 rounded ${n.isRead ? 'glass-input opacity-40' : 'bg-red-500/20 shadow-inner text-white'}`}>
                      {n.message}
                    </div>
                  ))
                )}
            </div>
          </div>

          <button onClick={() => { localStorage.clear(); window.location.href="/"; }} className="text-sm border border-white/20 px-4 py-2 rounded-lg hover:bg-white/10 transition">
            Sign Out
          </button>
        </div>
      ) : (
        <div className="flex gap-4">
          <Link href="/auth" className="btn-primary hover-lift text-sm">Sign In / Register</Link>
        </div>
      )}
    </nav>
  );
}
