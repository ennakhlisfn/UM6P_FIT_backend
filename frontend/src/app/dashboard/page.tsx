"use client";

import { useEffect, useState } from "react";
import { api } from "@/lib/api";

type History = {
  id: number;
  name: string;
  date: string;
  exercises: any[];
};

export default function DashboardPage() {
  const [history, setHistory] = useState<History[]>([]);
  const [loading, setLoading] = useState(true);
  const [success, setSuccess] = useState("");
  const [weightInput, setWeightInput] = useState("");
  const userId = typeof window !== "undefined" ? localStorage.getItem("auth_userid") : null;

  useEffect(() => {
    if (!userId) {
      window.location.href = "/auth";
      return;
    }
    
    api.request(`/users/${userId}/workouts?period=all`).then((data) => {
      setHistory(data || []);
      setLoading(false);
    }).catch(console.error);
  }, [userId]);

  const handleUpdateWeight = async () => {
    if (!userId || !weightInput) return;
    try {
      await api.request(`/users/${userId}/weight`, {
        method: "PUT",
        body: JSON.stringify({ weight: parseFloat(weightInput) })
      });
      setSuccess("Weight natively updated and synced to the generic timeline log!");
      setTimeout(() => setSuccess(""), 4000);
      setWeightInput("");
    } catch(e: any) {
      alert("Failed to securely push internal weight parameters: " + e.message);
    }
  };

  if (loading) return <div className="text-gradient font-black text-2xl pt-20 pulse-glow">Loading Secure Environment...</div>;

  return (
    <div className="w-full flex flex-col pt-10">
      <h1 className="text-5xl font-black mb-10 text-gradient tracking-tight">System Dashboard</h1>

      <div className="grid grid-cols-1 md:grid-cols-2 gap-8 w-full">
        <div className="glass-panel p-8 shadow-xl">
          <h2 className="text-2xl font-bold border-b border-white/10 pb-4 mb-6">Quick Health Commands</h2>
          
          <div className="flex gap-4 mb-4">
            <input 
              className="glass-input p-4 w-full rounded-xl" 
              placeholder="Record new internal Weight (kg) array" 
              type="number"
              step="0.1"
              value={weightInput}
              onChange={e => setWeightInput(e.target.value)}
            />
            <button onClick={handleUpdateWeight} className="btn-primary w-40 hover-lift text-lg font-bold">Sync Row</button>
          </div>
          {success && <p className="text-green-400 text-sm font-bold pulse-glow bg-green-900/40 p-3 rounded">{success}</p>}
        </div>

        <div className="glass-panel p-8 shadow-xl flex flex-col h-[500px]">
          <h2 className="text-2xl font-bold border-b border-white/10 pb-4 mb-6">Internal Workout Logs</h2>
          {history.length === 0 ? (
            <p className="italic opacity-60 my-auto text-center">You have totally zero workouts securely recorded locally.</p>
          ) : (
            <div className="overflow-y-auto flex-1 pr-2">
              {history.map(item => (
                <div key={item.id} className="bg-white/5 border border-white/10 rounded-xl p-5 mb-4 hover-lift">
                  <h3 className="font-bold text-xl mb-1 text-gradient">{item.name}</h3>
                  <p className="text-sm opacity-60">{new Date(item.date).toLocaleString(undefined, { dateStyle: 'full', timeStyle: 'short' })}</p>
                </div>
              ))}
            </div>
          )}
        </div>
      </div>
    </div>
  );
}
