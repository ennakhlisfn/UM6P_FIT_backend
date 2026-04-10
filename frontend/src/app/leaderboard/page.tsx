"use client";

import { useEffect, useState } from "react";
import { api } from "@/lib/api";

type LeaderboardRow = {
  id: number;
  name: string;
  totalPoints: number;
  score: number;
};

export default function LeaderboardPage() {
  const [data, setData] = useState<LeaderboardRow[]>([]);
  const [period, setPeriod] = useState("alltime");
  const [error, setError] = useState("");

  useEffect(() => {
    const fetchLB = async () => {
      try {
        const res = await api.request(`/leaderboard?period=${period}`);
        if (res) setData(res);
      } catch (e: any) {
        setError(e.message || "Failed generic leaderboard fetch");
      }
    };
    fetchLB();
  }, [period]);

  return (
    <div className="w-full flex flex-col items-center pt-10">
      <h1 className="text-5xl font-black mb-8 text-gradient tracking-tight">Global Target Rankings</h1>

      <div className="flex gap-4 mb-10">
        {["weekly", "monthly", "yearly", "alltime"].map((p) => (
          <button
            key={p}
            onClick={() => setPeriod(p)}
            className={`px-8 py-3 rounded-xl font-bold transition-all ${
              period === p ? "bg-[var(--primary)] text-white shadow-xl shadow-[var(--primary)]" : "glass-panel hover-lift text-white/80"
            }`}
          >
            {p.toUpperCase()}
          </button>
        ))}
      </div>

      {error ? (
        <div className="bg-red-500/20 text-red-200 border border-red-500/50 p-4 rounded-xl">{error}</div>
      ) : (
        <div className="glass-panel w-full max-w-4xl p-6 overflow-hidden">
          {data.length === 0 ? (
            <p className="text-center italic opacity-60 py-8">No scores logged for this generic constraint period.</p>
          ) : (
            data.map((user, idx) => (
              <div key={user.id} className="flex justify-between items-center p-6 mb-4 glass-panel border-l-[6px] border-l-[var(--accent)] hover-lift">
                <div className="flex gap-6 items-center">
                  <span className="text-3xl font-black opacity-80 w-10">{idx + 1}</span>
                  <span className="text-2xl font-bold text-white/90">{user.name}</span>
                </div>
                <div className="flex flex-col items-end">
                  <span className="text-3xl font-black text-gradient">{user.totalPoints} XP</span>
                  <span className="text-sm opacity-60 font-semibold uppercase tracking-wider">Score Index: {user.score.toFixed(2)}</span>
                </div>
              </div>
            ))
          )}
        </div>
      )}
    </div>
  );
}
