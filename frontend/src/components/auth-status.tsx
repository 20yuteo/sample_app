"use client";

import Link from "next/link";
import { LogIn, LogOut, ShieldCheck, User } from "lucide-react";
import { useState } from "react";

import { AuthSession, clearSession, getStoredSession, logoutURL } from "@/lib/auth";

export function AuthStatus() {
  const [session] = useState<AuthSession | null>(() => getStoredSession());

  if (!session) {
    return (
      <Link
        href="/login"
        className="inline-flex h-10 items-center gap-2 rounded-md border border-[#d6ddd7] bg-white px-3 text-sm font-semibold text-ink shadow-sm"
      >
        <LogIn size={17} />
        ログイン
      </Link>
    );
  }

  const roles = session.profile.realm_access?.roles ?? [];
  const isAdmin = roles.includes("ec_admin");

  return (
    <div className="flex items-center gap-2">
      <div className="hidden items-center gap-2 rounded-md border border-[#d6ddd7] bg-white px-3 py-2 text-sm sm:flex">
        {isAdmin ? <ShieldCheck size={16} className="text-skyline" /> : <User size={16} className="text-moss" />}
        <span className="max-w-44 truncate">{session.profile.email ?? session.profile.preferred_username}</span>
      </div>
      <button
        onClick={() => {
          clearSession();
          window.location.assign(logoutURL());
        }}
        className="inline-flex h-10 w-10 items-center justify-center rounded-md border border-[#d6ddd7] bg-white text-ink shadow-sm"
      >
        <LogOut size={17} aria-label="Logout" />
      </button>
    </div>
  );
}
