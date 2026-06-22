"use client";
import { useState, useEffect } from "react";
import Link from "next/link";
import { Card } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { formatSats, formatKES, satsToKES } from "@/lib/format";
import { getCommunityStats, getChamas, type CommunityStats, type Chama } from "@/lib/api";

type Tab = "chamas" | "leaderboard";

export default function CommunityPage() {
  const [tab, setTab] = useState<Tab>("chamas");
  const [stats, setStats] = useState<CommunityStats | null>(null);
  const [chamas, setChamas] = useState<Chama[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    getCommunityStats().then(setStats).catch(() => {});
    getChamas().then((c) => { setChamas(c); setLoading(false); }).catch(() => setLoading(false));
  }, []);

  const btcKES = stats?.btc_kes ?? 0;

  return (
    <div className="flex flex-col pt-6 pb-4 min-h-full">
      <div className="px-4 mb-5">
        <h1 className="text-2xl font-bold text-ink">Community</h1>
      </div>

      {/* Community stats */}
      <div className="px-4 mb-5">
        <Card className="bg-green-600 text-white" padding="md">
          <div className="grid grid-cols-3 gap-2 text-center">
            <div>
              <p className="text-2xl font-bold">{stats?.member_count.toLocaleString() ?? "—"}</p>
              <p className="text-green-200 text-xs mt-0.5">Members</p>
            </div>
            <div className="border-x border-white/20">
              <p className="text-2xl font-bold">
                {stats && btcKES > 0
                  ? formatKES(satsToKES(stats.total_savings_sats, btcKES)).replace("KES ", "")
                  : "—"}
              </p>
              <p className="text-green-200 text-xs mt-0.5">Total saved</p>
            </div>
            <div>
              <p className="text-2xl font-bold">
                {stats ? formatSats(stats.total_interest_paid_sats).replace(" sats", "") : "—"}
              </p>
              <p className="text-green-200 text-xs mt-0.5">Interest paid</p>
            </div>
          </div>
        </Card>
      </div>

      {/* Tabs */}
      <div className="flex px-4 mb-4 border-b border-border">
        {(["chamas", "leaderboard"] as Tab[]).map((t) => (
          <button
            key={t}
            onClick={() => setTab(t)}
            className={[
              "flex-1 pb-3 text-sm font-semibold capitalize border-b-2 -mb-px transition-colors",
              tab === t ? "border-green-600 text-green-600" : "border-transparent text-ink-muted",
            ].join(" ")}
          >
            {t}
          </button>
        ))}
      </div>

      {/* Chamas tab */}
      {tab === "chamas" && (
        <div className="px-4 flex flex-col gap-4">
          <Button variant="secondary" fullWidth>
            + Create a Chama
          </Button>

          {loading ? (
            <p className="text-sm text-ink-muted text-center py-8">Loading…</p>
          ) : chamas.length === 0 ? (
            <div className="flex flex-col items-center py-16 text-center">
              <span className="text-4xl mb-3">🤝</span>
              <p className="font-semibold text-ink">No Chamas yet</p>
              <p className="text-sm text-ink-muted mt-1">Create or join a savings group</p>
            </div>
          ) : (
            <div className="flex flex-col gap-3">
              {chamas.map((chama) => (
                <Link key={chama.id} href={`/community/chamas/${chama.id}`}>
                  <Card className="border border-border active:scale-[0.99] transition-transform">
                    <div className="flex items-start justify-between">
                      <div className="flex-1 min-w-0">
                        <div className="flex items-center gap-2">
                          <p className="font-semibold text-ink">{chama.name}</p>
                          {chama.role === "admin" && (
                            <span className="text-xs bg-gold-100 text-gold-500 px-1.5 py-0.5 rounded font-medium">Admin</span>
                          )}
                        </div>
                        <p className="text-xs text-ink-muted mt-0.5">{chama.description}</p>
                        <div className="flex items-center gap-3 mt-3">
                          <div>
                            <p className="text-xs text-ink-muted">Balance</p>
                            <p className="text-sm font-semibold text-ink">
                              {btcKES > 0 ? formatKES(satsToKES(chama.balance_sats, btcKES)) : formatSats(chama.balance_sats)}
                            </p>
                          </div>
                          <div className="w-px h-8 bg-border" />
                          <div>
                            <p className="text-xs text-ink-muted">Members</p>
                            <p className="text-sm font-semibold text-ink">{chama.member_count}</p>
                          </div>
                          <div className="w-px h-8 bg-border" />
                          <div>
                            <p className="text-xs text-ink-muted">Status</p>
                            <span className={["text-xs font-semibold", chama.status === "active" ? "text-credit" : "text-ink-muted"].join(" ")}>
                              {chama.status}
                            </span>
                          </div>
                        </div>
                      </div>
                      <span className="text-ink-muted text-lg ml-2">›</span>
                    </div>
                  </Card>
                </Link>
              ))}
            </div>
          )}
        </div>
      )}

      {/* Leaderboard tab — placeholder until backend endpoint exists */}
      {tab === "leaderboard" && (
        <div className="px-4 flex flex-col items-center py-16 text-center">
          <span className="text-4xl mb-3">🏆</span>
          <p className="font-semibold text-ink">Leaderboard coming soon</p>
          <p className="text-sm text-ink-muted mt-1">Top savers will appear here</p>
        </div>
      )}

    </div>
  );
}
