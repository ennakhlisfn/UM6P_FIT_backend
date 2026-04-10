"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import { api } from "@/lib/api";

export default function AuthPage() {
  const [isLogin, setIsLogin] = useState(true);
  const router = useRouter();

  // Controlled Inputs
  const [email, setEmail] = useState("core@test.com"); // Pre-filled debug runner
  const [password, setPassword] = useState("pass");
  const [name, setName] = useState("");
  const [age, setAge] = useState(25);
  const [height, setHeight] = useState(180);
  const [weight, setWeight] = useState(70);
  const [error, setError] = useState("");

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError("");

    try {
      if (!isLogin) {
        // Register Hook
        await api.request("/users", {
          method: "POST",
          body: JSON.stringify({ name, email, password, age, height, weight })
        });
      }
      
      // Auto-Login flow natively minting tokens
      const data = await api.request("/login", {
        method: "POST",
        body: JSON.stringify({ email, password })
      });

      if (data && data.token) {
        localStorage.setItem("auth_token", data.token);
        localStorage.setItem("auth_userid", data.user.id.toString());
        // Force refresh mapping the Navbar globals natively
        window.location.href = "/dashboard";
      }
    } catch (err: any) {
      setError(err.message || "Failed generic authentication lock");
    }
  };

  return (
    <div className="flex flex-col items-center justify-center min-h-[75vh]">
      <div className="glass-panel p-8 w-full max-w-md shadow-2xl">
        <h1 className="text-4xl font-black mb-8 text-center text-gradient tracking-tight">
          {isLogin ? "Welcome Back" : "Join UM6P FIT"}
        </h1>

        {error && <div className="bg-red-500/20 border border-red-500/50 text-red-100 p-3 rounded-lg mb-6 text-sm">{error}</div>}

        <form onSubmit={handleSubmit} className="flex flex-col gap-5">
          {!isLogin && (
            <>
              <input className="glass-input p-3 rounded-lg w-full" placeholder="Full Name" value={name} onChange={e=>setName(e.target.value)} required />
              <div className="flex gap-3">
                <input type="number" className="glass-input p-3 rounded-lg w-1/3 text-center" placeholder="Age" value={age} onChange={e=>setAge(Number(e.target.value))} required />
                <input type="number" className="glass-input p-3 rounded-lg w-1/3 text-center" placeholder="Height" value={height} onChange={e=>setHeight(Number(e.target.value))} required />
                <input type="number" step="0.1" className="glass-input p-3 rounded-lg w-1/3 text-center" placeholder="Weight" value={weight} onChange={e=>setWeight(Number(e.target.value))} required />
              </div>
            </>
          )}

          <input type="email" className="glass-input p-3 rounded-lg w-full" placeholder="Email Address" value={email} onChange={e=>setEmail(e.target.value)} required />
          <input type="password" className="glass-input p-3 rounded-lg w-full" placeholder="Secure Password" value={password} onChange={e=>setPassword(e.target.value)} required />

          <button type="submit" className="btn-primary mt-4 py-4 text-lg font-bold shadow-lg hover:shadow-xl hover-lift">
            {isLogin ? "Access Dashboard" : "Create Account Engine"}
          </button>
        </form>

        <p className="mt-8 text-center text-sm text-gray-300">
          {isLogin ? "Don't have an account?" : "Already joined?"}{" "}
          <button onClick={() => setIsLogin(!isLogin)} className="text-[var(--primary)] font-bold hover:underline">
            {isLogin ? "Register now" : "Sign in securely"}
          </button>
        </p>
      </div>
    </div>
  );
}
