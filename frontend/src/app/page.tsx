import Link from "next/link";

export default function Home() {
  return (
    <div className="flex flex-col items-center justify-center min-h-[75vh] max-w-4xl text-center">
      <h1 className="text-6xl font-black mb-6 tracking-tighter">
        The Future Of <br />
        <span className="text-gradient">Gamified Fitness.</span>
      </h1>
      
      <p className="text-xl text-white/70 mb-10 max-w-2xl leading-relaxed">
        Engineered heavily on Go and React, UM6P FIT transforms your generic workout metrics into undeniable XP points and competitive global progression arrays.
      </p>

      <div className="flex gap-6">
        <Link href="/auth" className="btn-primary text-xl px-10 py-4 hover-lift">
          Start Training
        </Link>
        <Link href="/leaderboard" className="glass-input flex items-center justify-center font-bold px-10 py-4 hover-lift bg-white/5 border-white/20 rounded-lg">
          View Leaderboard
        </Link>
      </div>
    </div>
  );
}
