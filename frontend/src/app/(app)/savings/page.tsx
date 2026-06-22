"use client";
import { useState, useEffect, useTransition } from "react";
import { useRouter } from "next/navigation";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Card } from "@/components/ui/card";
import { formatKES, formatSats, satsToKES, timeAgo } from "@/lib/format";
import { getSavings, lockSavings, getBalance, type SavingsData, type SavingsLock } from "@/lib/api";

export default function SavingsPage() {
  const router = useRouter();
  const [data, setData] = useState<SavingsData | null>(null);
  const [balanceSats, setBalanceSats] = useState(0);
  const [amountKES, setAmountKES] = useState("");
  const [error, setError] = useState("");
  const [success, setSuccess] = useState("");
  const [showLockForm, setShowLockForm] = useState(false);
  const [pending, startTransition] = useTransition();

  useEffect(() => {
    getSavings().then(setData).catch(() => {});
    getBalance().then((b) => setBalanceSats(b.balance_sats)).catch(() => {});
  }, []);

  const btcKES = data?.btc_kes ?? 0;
  const kes = parseFloat(amountKES) || 0;
  const amountSats = btcKES > 0 ? Math.round((kes / btcKES) * 1e8) : 0;
  const ratePct = data ? (data.pool_rate_bps / 100).toFixed(1) : "—";
  const lockDays = data?.pool_lock_days ?? 30;

  function handleLock(e: React.FormEvent) {
    e.preventDefault();
    setError("");
    if (!data) return;
    if (amountSats < data.min_sats) {
      setError(`Minimum lock is ${formatSats(data.min_sats)} (${formatKES(satsToKES(data.min_sats, btcKES))})`);
      return;
    }
    if (amountSats > balanceSats) {
      setError("Insufficient balance");
      return;
    }
    startTransition(async () => {
      try {
        const res = await lockSavings(amountSats);
        setSuccess(`Locked! Unlocks ${new Date(res.unlocks_at).toLocaleDateString("en-KE", { day: "numeric", month: "long", year: "numeric" })}`);
        setShowLockForm(false);
        setAmountKES("");
        getSavings().then(setData).catch(() => {});
        getBalance().then((b) => setBalanceSats(b.balance_sats)).catch(() => {});
      } catch (err: unknown) {
        setError(err instanceof Error ? err.message : "Lock failed. Try again.");
      }
    });
  }

  const activeLocks = data?.locks.filter((l) => l.status === "locked") ?? [];
  const pastLocks   = data?.locks.filter((l) => l.status !== "locked") ?? [];

  return (
    <div className="flex flex-col px-4 pt-6 pb-4 min-h-full gap-5">

      <div className="flex items-center gap-3">
        <button onClick={() => router.back()} className="text-ink-soft text-sm">← Back</button>
      </div>
      <h1 className="text-2xl font-bold text-ink">Savings</h1>

      {/* APY card */}
      <Card className="bg-green-600 text-white" padding="lg">
        <div className="flex items-center justify-between">
          <div>
            <p className="text-green-100 text-sm">Current rate</p>
            <p className="text-4xl font-bold">{ratePct}%</p>
            <p className="text-green-200 text-sm mt-0.5">APY · {lockDays}-day lock</p>
          </div>
          <span className="text-5xl">✨</span>
        </div>
      </Card>

      {success && (
        <Card className="border border-green-200 bg-green-50">
          <p className="text-sm font-medium text-green-700 text-center">{success}</p>
        </Card>
      )}

      {/* Lock new savings */}
      {!showLockForm ? (
        <Button fullWidth onClick={() => setShowLockForm(true)}>
          + Lock savings to earn interest
        </Button>
      ) : (
        <Card className="border border-green-100">
          <p className="font-semibold text-ink mb-4">Lock savings</p>
          <form onSubmit={handleLock} className="flex flex-col gap-4">
            <Input
              label="Amount (KES)"
              prefix="KES"
              type="number"
              inputMode="decimal"
              placeholder="0"
              value={amountKES}
              onChange={(e) => setAmountKES(e.target.value)}
              hint={amountSats > 0 ? `≈ ${formatSats(amountSats)}` : undefined}
              error={error}
            />
            {data && btcKES > 0 && amountSats >= data.min_sats && (
              <Card className="bg-green-50 border border-green-100" padding="sm">
                <div className="flex justify-between text-sm">
                  <span className="text-ink-soft">Monthly interest</span>
                  <span className="font-semibold text-credit">
                    +{formatSats(Math.round(amountSats * data.pool_rate_bps / 12 / 10_000))}
                  </span>
                </div>
                <div className="flex justify-between text-xs text-ink-muted mt-1">
                  <span>Unlocks after</span>
                  <span>{lockDays} days</span>
                </div>
              </Card>
            )}
            <div className="flex gap-3">
              <Button type="button" variant="ghost" fullWidth onClick={() => { setShowLockForm(false); setError(""); }}>
                Cancel
              </Button>
              <Button type="submit" fullWidth loading={pending} disabled={amountSats <= 0}>
                Lock
              </Button>
            </div>
          </form>
        </Card>
      )}

      {/* Active locks */}
      {activeLocks.length > 0 && (
        <div>
          <p className="text-sm font-semibold text-ink mb-3">Active locks</p>
          <div className="flex flex-col gap-3">
            {activeLocks.map((lock) => (
              <LockCard key={lock.id} lock={lock} btcKES={btcKES} />
            ))}
          </div>
        </div>
      )}

      {/* Past locks */}
      {pastLocks.length > 0 && (
        <div>
          <p className="text-sm font-semibold text-ink-muted mb-3">Past locks</p>
          <div className="flex flex-col gap-3">
            {pastLocks.map((lock) => (
              <LockCard key={lock.id} lock={lock} btcKES={btcKES} past />
            ))}
          </div>
        </div>
      )}

      {data?.locks.length === 0 && !showLockForm && (
        <div className="flex flex-col items-center py-12 text-center">
          <span className="text-4xl mb-3">🔒</span>
          <p className="font-semibold text-ink">No active locks</p>
          <p className="text-sm text-ink-muted mt-1">Lock sats to earn {ratePct}% APY</p>
        </div>
      )}

    </div>
  );
}

function LockCard({ lock, btcKES, past }: { lock: SavingsLock; btcKES: number; past?: boolean }) {
  const unlockDate = new Date(lock.unlocks_at).toLocaleDateString("en-KE", {
    day: "numeric", month: "short", year: "numeric",
  });
  return (
    <Card className={["border", past ? "border-border opacity-70" : "border-green-200"].join(" ")}>
      <div className="flex items-start justify-between">
        <div>
          <p className="font-semibold text-ink">{formatSats(lock.amount_sats)}</p>
          {btcKES > 0 && <p className="text-xs text-ink-muted">{formatKES(lock.amount_kes)}</p>}
          <p className="text-xs text-ink-soft mt-1">{(lock.rate_bps / 100).toFixed(1)}% APY · {lock.lock_days}d</p>
        </div>
        <div className="text-right">
          <span className={[
            "text-xs font-semibold px-2 py-0.5 rounded-full",
            lock.status === "locked"   ? "bg-green-100 text-green-700" :
            lock.status === "unlocked" ? "bg-surface-card text-ink-muted border border-border" :
            "bg-red-50 text-red-600",
          ].join(" ")}>
            {lock.status}
          </span>
          <p className="text-xs text-ink-muted mt-1">
            {lock.status === "locked" ? `Unlocks ${unlockDate}` : `Unlocked ${unlockDate}`}
          </p>
          {lock.interest_earned_sats > 0 && (
            <p className="text-xs text-credit mt-0.5">+{formatSats(lock.interest_earned_sats)} earned</p>
          )}
        </div>
      </div>
    </Card>
  );
}
