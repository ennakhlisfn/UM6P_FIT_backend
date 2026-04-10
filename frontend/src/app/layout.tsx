import type { Metadata } from "next";
import "./globals.css";
import GlassNavbar from "@/components/GlassNavbar";

export const metadata: Metadata = {
  title: "UM6P FIT | Next.js Engine",
  description: "Advanced Glassmorphic Workout Analytics Platform",
};

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html lang="en">
      <body className="antialiased bg-gradient-mesh min-h-screen flex flex-col text-white">
          <GlassNavbar />
          
          <main className="flex-1 flex flex-col p-8 sm:p-20 relative z-10 w-full max-w-6xl mx-auto">
            {children}
          </main>
      </body>
    </html>
  );
}
