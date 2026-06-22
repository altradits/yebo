"use client";
import { useState, useEffect } from "react";
import { useRouter, useParams } from "next/navigation";
import { Card } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { formatKES, formatSats, timeAgo } from "@/lib/format";
import { getChama, type ChamaDetail } from "@/lib/api";

export default function ChamaDetailPage() {
  const router = useRouter();
  const params = useParams();
  const id = Number(params.id);
  const [chama, setChama] = useState<ChamaDetail | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");

  useEffect(() => {
    if (!id) return;
    getChama(id)
      .then(setChama)
      .catch(() => setError("Could not load chama"))
      .finally(() => setLoading(false));
  }, [id]);

  if (loading) {
    return (
      <div className="flex flex-col px-4 pt-6 min-h-full">
        <button onClick={() => router.back()} className="text-ink-soft text-sm self-start mb-6">← Back</button>
        <p className="text-ink-muted text-sm">Loading…</p>
      </div>
    );
  }

  if (error || !chama) {
    return (
      <div className="flex flex-col px-4 pt-6 min-h-full">
        <button onClick={() => router.back()} className="text-ink-soft text-sm self-start mb-6">← Back</button>
        <div className="flex flex-col items-center py-16 text-center">
          <span className="text-4xl mb-3">😕</span>
          <p className="font-semibold text-ink">{error || "Chama not found"}</p>
        </div>
      </div>
    );
  }

  const btcKES = chama.btc_kes;

  return (
    <div className="flex flex-col px-4 pt-6 pb-4 min-h-full gap-5">

      <div className="flex items-center gap-3">
        <button onClick={() => router.back()} className="text-ink-soft text-sm">← Back</button>
      </div>

      {/* Header */}
      <div className="flex items-start justify-between">
        <div>
          <h1 className="text-2xl font-bold text-ink">{chama.name}</h1>
          {chama.description && <p className="text-sm text-ink-muted mt-1">{chama.description}</p>}
        </div>
        {chama.role === "admin" && (
          <span className="text-xs bg-gold-100 text-gold-500 px-2 py-0.5 rounded font-medium mt-1">Admin</span>
        )}
      </div>

      {/* Balance card */}
      <Card className="bg-green-600 text-white" padding="lg">
        <p className="text-green-100 text-sm font-medium">Group balance</p>
        <p className="text-3xl font-bold mt-1">
          {btcKES > 0 ? formatKES(Math.round(chama.balance_kes)) : formatSats(chama.balance_sats)}
        </p>
        {btcKES > 0 && <p className="text-green-200 text-sm mt-0.5">{formatSats(chama.balance_sats)}</p>}
        <div className="mt-4 pt-4 border-t border-white/20 grid grid-cols-3 gap-2 text-center">
          <div>
            <p className="text-white font-semibold">{chama.member_count}</p>
            <p className="text-green-200 text-xs">Members</p>
          </div>
          <div>
            <p className="text-white font-semibold">{chama.cycle_days}d</p>
            <p className="text-green-200 text-xs">Cycle</p>
          </div>
          <div>
            <p className={["font-semibold text-sm", chama.status === "active" ? "text-white" : "text-green-200"].join(" ")}>
              {chama.status}
            </p>
            <p className="text-green-200 text-xs">Status</p>
          </div>
        </div>
      </Card>

      {/* Contribution info */}
      {chama.contribution_sats > 0 && (
        <Card className="border border-border">
          <div className="flex items-center justify-between">
            <div>
              <p className="text-sm font-semibold text-ink">Contribution per cycle</p>
              <p className="text-xs text-ink-muted">{chama.cycle_days}-day cycle</p>
            </div>
            <p className="font-bold text-ink">{formatSats(chama.contribution_sats)}</p>
          </div>
        </Card>
      )}

      {/* Actions */}
      <div className="flex gap-3">
        <Button variant="secondary" fullWidth>Contribute</Button>
        {chama.role === "admin" && (
          <Button variant="secondary" fullWidth>Withdraw</Button>
        )}
      </div>

      {/* Members */}
      <div>
        <p className="text-sm font-semibold text-ink mb-3">Members ({chama.member_count})</p>
        <Card padding="none">
          <ul className="divide-y divide-border">
            {chama.members.map((m) => (
              <li key={m.id} className={["flex items-center gap-3 px-4 py-3.5", m.is_me ? "bg-green-50" : ""].join(" ")}>
                <div className="w-9 h-9 rounded-full bg-green-100 flex items-center justify-center text-green-700 font-bold text-sm flex-shrink-0">
                  {m.name[0]?.toUpperCase()}
                </div>
                <div className="flex-1">
                  <p className={["text-sm font-medium", m.is_me ? "text-green-700" : "text-ink"].join(" ")}>
                    {m.name} {m.is_me && <span className="text-xs">(you)</span>}
                  </p>
                  <p className="text-xs text-ink-muted">Joined {timeAgo(m.joined_at)}</p>
                </div>
                {m.role === "admin" && (
                  <span className="text-xs bg-gold-100 text-gold-500 px-1.5 py-0.5 rounded font-medium">Admin</span>
                )}
              </li>
            ))}
          </ul>
        </Card>
      </div>

    </div>
  );
}
