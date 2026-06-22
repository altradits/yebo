"use client";
import { useState, useEffect, useCallback } from "react";
import { Card } from "@/components/ui/card";
import { formatSats, formatKES, satsToKES, timeAgo } from "@/lib/format";
import { getTransactions, getBalance, type Transaction } from "@/lib/api";

type TxType = "all" | "deposit" | "withdrawal" | "send" | "receive" | "savings_interest";

const FILTERS: { key: TxType; label: string }[] = [
  { key: "all",              label: "All"         },
  { key: "deposit",          label: "Deposits"    },
  { key: "withdrawal",       label: "Withdrawals" },
  { key: "send",             label: "Sent"        },
  { key: "receive",          label: "Received"    },
  { key: "savings_interest", label: "Interest"    },
];

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

export default function ActivityPage() {
  const [filter, setFilter] = useState<TxType>("all");
  const [search, setSearch] = useState("");
  const [transactions, setTransactions] = useState<Transaction[]>([]);
  const [btcKES, setBtcKES] = useState(0);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    getBalance().then((b) => setBtcKES(b.btc_kes)).catch(() => {});
    getTransactions(100).then((txs) => {
      setTransactions(txs);
      setLoading(false);
    }).catch(() => setLoading(false));
  }, []);

  const filtered = transactions.filter((tx) => {
    if (filter !== "all" && tx.type !== filter) return false;
    if (search && !tx.note.toLowerCase().includes(search.toLowerCase())) return false;
    return true;
  });

  const totalIn  = filtered.filter((t) => t.is_credit).reduce((s, t) => s + t.amount_sats, 0);
  const totalOut = filtered.filter((t) => !t.is_credit).reduce((s, t) => s + t.amount_sats, 0);

  return (
    <div className="flex flex-col pt-6 pb-4 min-h-full">

      <div className="px-4 mb-4">
        <h1 className="text-2xl font-bold text-ink">Activity</h1>
      </div>

      {/* Search */}
      <div className="px-4 mb-3">
        <div className="relative">
          <span className="absolute left-3.5 top-1/2 -translate-y-1/2 text-ink-muted">🔍</span>
          <input
            type="search"
            placeholder="Search transactions…"
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            className="w-full h-11 rounded-btn border border-border bg-surface-card pl-10 pr-4 text-sm text-ink placeholder:text-ink-muted focus:outline-none focus:ring-2 focus:ring-green-600"
          />
        </div>
      </div>

      {/* Filter chips */}
      <div className="flex gap-2 overflow-x-auto px-4 pb-1 scrollbar-none">
        {FILTERS.map(({ key, label }) => (
          <button
            key={key}
            onClick={() => setFilter(key)}
            className={[
              "flex-shrink-0 px-3 py-1.5 rounded-full text-sm font-medium border transition-colors",
              filter === key
                ? "bg-green-600 text-white border-green-600"
                : "bg-surface-card text-ink-soft border-border",
            ].join(" ")}
          >
            {label}
          </button>
        ))}
      </div>

      {/* Summary row */}
      {(filter !== "all" || search) && filtered.length > 0 && (
        <div className="flex gap-3 px-4 mt-3">
          <div className="flex-1 bg-green-50 rounded-btn p-3 text-center">
            <p className="text-xs text-ink-muted">In</p>
            <p className="text-sm font-bold text-credit">+{formatSats(totalIn)}</p>
            {btcKES > 0 && <p className="text-xs text-ink-muted">{formatKES(satsToKES(totalIn, btcKES))}</p>}
          </div>
          <div className="flex-1 bg-red-50 rounded-btn p-3 text-center">
            <p className="text-xs text-ink-muted">Out</p>
            <p className="text-sm font-bold text-debit">−{formatSats(totalOut)}</p>
            {btcKES > 0 && <p className="text-xs text-ink-muted">{formatKES(satsToKES(totalOut, btcKES))}</p>}
          </div>
        </div>
      )}

      {/* List */}
      <div className="px-4 mt-4">
        {loading ? (
          <div className="flex flex-col items-center justify-center py-16 text-center">
            <p className="text-ink-muted text-sm">Loading…</p>
          </div>
        ) : filtered.length === 0 ? (
          <div className="flex flex-col items-center justify-center py-16 text-center">
            <span className="text-4xl mb-3">📭</span>
            <p className="font-semibold text-ink">No transactions found</p>
            <p className="text-sm text-ink-muted mt-1">Try a different filter</p>
          </div>
        ) : (
          <Card padding="none">
            <ul className="divide-y divide-border">
              {filtered.map((tx) => (
                <li key={tx.id} className="flex items-center gap-3 px-4 py-4">
                  <span className="text-xl w-8 text-center flex-shrink-0">
                    {TX_ICON[tx.type] ?? "💸"}
                  </span>
                  <div className="flex-1 min-w-0">
                    <p className="text-sm font-medium text-ink">{tx.note || tx.type}</p>
                    <p className="text-xs text-ink-muted mt-0.5">{timeAgo(tx.created_at)}</p>
                  </div>
                  <div className="text-right flex-shrink-0">
                    <p className={["text-sm font-semibold tabular-nums", tx.is_credit ? "text-credit" : "text-debit"].join(" ")}>
                      {tx.is_credit ? "+" : "−"}{formatSats(tx.amount_sats)}
                    </p>
                    {btcKES > 0 && (
                      <p className="text-xs text-ink-muted">
                        {formatKES(satsToKES(tx.amount_sats, btcKES))}
                      </p>
                    )}
                  </div>
                </li>
              ))}
            </ul>
          </Card>
        )}
      </div>

    </div>
  );
}
