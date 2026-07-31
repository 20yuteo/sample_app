"use client";

import Link from "next/link";
import { ShieldCheck, ShoppingBag, UserRound } from "lucide-react";

import { startLogin } from "@/lib/auth";

export function LoginPanel() {
  return (
    <main className="min-h-screen bg-[#f7f8f6] text-ink">
      <header className="border-b border-[#dfe5df] bg-white">
        <div className="mx-auto flex max-w-5xl items-center justify-between px-4 py-4 sm:px-6 lg:px-8">
          <Link href="/" className="flex items-center gap-2 font-semibold">
            <ShoppingBag size={20} />
            Commerce Lab
          </Link>
          <a
            href="http://localhost:18080/admin"
            className="inline-flex h-10 items-center gap-2 rounded-md border border-[#d6ddd7] bg-white px-3 text-sm font-semibold shadow-sm"
          >
            <ShieldCheck size={17} />
            Keycloak
          </a>
        </div>
      </header>

      <section className="mx-auto grid max-w-5xl gap-5 px-4 py-8 sm:px-6 lg:grid-cols-2 lg:px-8">
        <button
          onClick={() => void startLogin("commerce-frontend")}
          className="rounded-md border border-[#dfe5df] bg-white p-6 text-left shadow-sm transition hover:border-moss"
        >
          <UserRound className="mb-5 text-moss" size={32} />
          <h1 className="text-2xl font-semibold">顧客ログイン</h1>
          <p className="mt-3 text-sm leading-6 text-[#5d6a63]">
            デモ: customer@example.test / customer123
          </p>
        </button>

        <button
          onClick={() => void startLogin("commerce-admin")}
          className="rounded-md border border-[#dfe5df] bg-white p-6 text-left shadow-sm transition hover:border-skyline"
        >
          <ShieldCheck className="mb-5 text-skyline" size={32} />
          <h2 className="text-2xl font-semibold">EC管理者ログイン</h2>
          <p className="mt-3 text-sm leading-6 text-[#5d6a63]">
            デモ: admin@example.test / admin123
          </p>
        </button>
      </section>
    </main>
  );
}

