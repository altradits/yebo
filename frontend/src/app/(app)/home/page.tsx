"use client";
import { useState, useEffect } from "react";
import Link from "next/link";
import { Card } from "@/components/ui/card";
import { formatSats, formatKES, satsToKES, timeAgo } from "@/lib/format";
import { getMe, getTransactions, type User, type Transaction } from "@/lib/api";

const TX_ICON: Record<string, string> = {
  deposit:             "📥",
  withdrawal:          "📤",
  send:                "↗️",
  receive:             "↙️",
  savings_interest:    "✨",
  savings_lock:        "🔒",
  savings_unlock:      "🔓",
  chama_contribution:  "🤝",
  chama_payout:        "🏦",
};

export default function HomePage() {
  const [user, setUser] = useState<User | null>(null);
  const [transactions, setTransactions] = useState<Transaction[]>([]);
  const [showSats, setShowSats] = useState(false);

  useEffect(() => {
    getMe().then(setUser).catch(() => {});
    getTransactions(5).then(setTransactions).catch(() => {});
  }, []);

  const hour = new Date().getHours();
  const greeting = hour < 12 ? "Good morning" : hour < 17 ? "Good afternoon" : "Good evening";
  const firstName = user?.full_name?.split(" ")[0] ?? "…";
  const balanceSats = user?.balance_sats ?? 0;
  const btcKES = user?.btc_kes ?? 0;
  const kesBalance = satsToKES(balanceSats, btcKES);
  const apyBps = 500;
  const monthlyInterest = Math.round((balanceSats * apyBps) / 12 / 10_000);

  return (
    <div className="flex flex-col gap-4 px-4 pt-6 pb-4">

      {/* Greeting */}
      <div className="flex items-center justify-between">
        <div>
          <p className="text-ink-soft text-sm">{greeting}</p>
          <h1 className="text-xl font-bold text-ink">{firstName} 👋</h1>
        </div>
        <div className="w-10 h-10 rounded-full bg-green-100 flex items-center justify-center text-green-700 font-bold text-lg">
          {firstName[0]}
        </div>
      </div>

      {/* Balance card */}
      <Card className="bg-green-600 text-white" padding="lg">
        <div className="flex items-start justify-between mb-1">
          <p className="text-green-100 text-sm font-medium">Total balance</p>
          <button
            onClick={() => setShowSats((s) => !s)}
            className="text-xs bg-white/20 text-white px-2 py-1 rounded-full"
          >
            {showSats ? "Show KES" : "Show sats"}
          </button>
        </div>

        {showSats ? (
          <>
            <p className="text-4xl font-bold tracking-tight">{balanceSats.toLocaleString("en-KE")}</p>
            <p className="text-green-200 text-sm mt-0.5">sats  ≈ {formatKES(kesBalance)}</p>
          </>
        ) : (
          <>
            <p className="text-4xl font-bold tracking-tight">{btcKES > 0 ? formatKES(kesBalance) : "—"}</p>
            <p className="text-green-200 text-sm mt-0.5">{formatSats(balanceSats)}</p>
          </>
        )}

        <div className="mt-4 pt-4 border-t border-white/20 flex items-center gap-2">
          <span className="text-lg">✨</span>
          <p className="text-sm text-green-100">
            Earning <span className="text-white font-semibold">{formatSats(monthlyInterest)}</span>/month
            {" "}at <span className="text-white font-semibold">{(apyBps / 100).toFixed(1)}% APY</span>
          </p>
        </div>
      </Card>

      {/* Quick actions */}
      <div className="grid grid-cols-4 gap-3">
        {[
          { label: "Deposit",  icon: "📥", href: "/deposit"  },
          { label: "Send",     icon: "↗️", href: "/send"     },
          { label: "Withdraw", icon: "📤", href: "/withdraw" },
          { label: "Save",     icon: "🔒", href: "/savings"  },
        ].map(({ label, icon, href }) => (
          <Link
            key={label}
            href={href}
            className="flex flex-col items-center gap-1.5 bg-surface-card rounded-card p-3 shadow-sm active:scale-95 transition-transform"
          >
            <span className="text-2xl">{icon}</span>
            <span className="text-xs font-medium text-ink-soft">{label}</span>
          </Link>
        ))}
      </div>

      {/* Interest earned card */}
      {(user?.total_interest_earned ?? 0) > 0 && (
        <Card className="border border-gold-100 bg-gold-100/30">
          <div className="flex items-center gap-3">
            <span className="text-2xl">✨</span>
            <div>
              <p className="text-sm font-semibold text-ink">
                {formatSats(user!.total_interest_earned)} earned
              </p>
              <p className="text-xs text-ink-soft">Total interest since you joined</p>
            </div>
          </div>
        </Card>
      )}

      {/* Recent transactions */}
      <div>
        <div className="flex items-center justify-between mb-3">
          <h2 className="font-semibold text-ink">Recent activity</h2>
          <Link href="/activity" className="text-sm text-green-600 font-medium">View all</Link>
        </div>
        {transactions.length === 0 ? (
          <Card>
            <p className="text-sm text-ink-muted text-center py-4">No transactions yet</p>
          </Card>
        ) : (
          <Card padding="none">
            <ul className="divide-y divide-border">
              {transactions.map((tx) => (
                <li key={tx.id} className="flex items-center gap-3 px-4 py-3.5">
                  <span className="text-xl w-8 text-center flex-shrink-0">
                    {TX_ICON[tx.type] ?? "💸"}
                  </span>
                  <div className="flex-1 min-w-0">
                    <p className="text-sm font-medium text-ink truncate">{tx.note || tx.type}</p>
                    <p className="text-xs text-ink-muted">{timeAgo(tx.created_at)}</p>
                  </div>
                  <p className={["text-sm font-semibold tabular-nums", tx.is_credit ? "text-credit" : "text-debit"].join(" ")}>
                    {tx.is_credit ? "+" : "−"}{formatSats(tx.amount_sats)}
                  </p>
                </li>
              ))}
            </ul>
          </Card>
        )}
      </div>

      {/* Community card */}
      <Link href="/community">
        <Card className="border border-green-100 active:scale-[0.99] transition-transform">
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-3">
              <span className="text-2xl">👥</span>
              <div>
                <p className="font-semibold text-ink text-sm">Your community</p>
                <p className="text-xs text-ink-muted">See who&apos;s saving with you</p>
              </div>
            </div>
            <span className="text-ink-muted text-lg">›</span>
          </div>
        </Card>
      </Link>

    </div>
  );
}
