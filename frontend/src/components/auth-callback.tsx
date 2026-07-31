"use client";

import Link from "next/link";
import { CheckCircle2, Loader2, XCircle } from "lucide-react";
import { useEffect, useState } from "react";

import { completeLogin } from "@/lib/auth";

export function AuthCallback() {
  const [status, setStatus] = useState<"loading" | "done" | "error">("loading");
  const [message, setMessage] = useState("ログインを完了しています。");

  useEffect(() => {
    const params = new URLSearchParams(window.location.search);
    const code = params.get("code");
    const state = params.get("state");
    const error = params.get("error_description") ?? params.get("error");
    if (error) {
      queueMicrotask(() => {
        setStatus("error");
        setMessage(error);
      });
      return;
    }
    if (!code || !state) {
      queueMicrotask(() => {
        setStatus("error");
        setMessage("認証コードが見つかりません。");
      });
      return;
    }

    completeLogin(code, state)
      .then(() => {
        setStatus("done");
        setMessage("ログインしました。商品カタログへ戻ります。");
        window.setTimeout(() => window.location.replace("/"), 800);
      })
      .catch((err: unknown) => {
        setStatus("error");
        setMessage(err instanceof Error ? err.message : "ログインに失敗しました。");
      });
  }, []);

  return (
    <main className="flex min-h-screen items-center justify-center bg-[#f7f8f6] px-4 text-ink">
      <div className="w-full max-w-md rounded-md border border-[#dfe5df] bg-white p-6 text-center shadow-sm">
        {status === "loading" ? <Loader2 className="mx-auto mb-4 animate-spin text-moss" /> : null}
        {status === "done" ? <CheckCircle2 className="mx-auto mb-4 text-moss" /> : null}
        {status === "error" ? <XCircle className="mx-auto mb-4 text-red-600" /> : null}
        <p className="text-sm text-[#5d6a63]">{message}</p>
        {status === "error" ? (
          <Link href="/login" className="mt-5 inline-flex h-10 items-center rounded-md bg-ink px-4 text-sm font-semibold text-white">
            ログインへ戻る
          </Link>
        ) : null}
      </div>
    </main>
  );
}
