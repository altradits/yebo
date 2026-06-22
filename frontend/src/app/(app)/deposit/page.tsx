"use client";
import { useState, useEffect, useTransition } from "react";
import { useRouter } from "next/navigation";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Card } from "@/components/ui/card";
import { formatKES, formatSats, satsToKES } from "@/lib/format";
import { getMe, depositMpesa } from "@/lib/api";

const KES_PRESETS = [500, 1_000, 2_000, 5_000];

export default function DepositPage() {
  const router = useRouter();
  const [pending, startTransition] = useTransition();
  const [amountKES, setAmountKES] = useState("");
  const [phone, setPhone] = useState("");
  const [error, setError] = useState("");
  const [success, setSuccess] = useState("");
  const [btcKES, setBtcKES] = useState(0);

  useEffect(() => {
    getMe().then((u) => {
      setBtcKES(u.btc_kes);
      setPhone(u.phone);
    }).catch(() => {});
  }, []);

  const kes = parseFloat(amountKES) || 0;
  const sats = btcKES > 0 ? Math.round((kes / btcKES) * 1e8) : 0;

  function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    setError("");
    if (kes < 10) { setError("Minimum deposit is KES 10"); return; }
    if (!phone)   { setError("Phone number required"); return; }

    startTransition(async () => {
      try {
        const res = await depositMpesa(kes, phone);
        setSuccess(res.message);
        setAmountKES("");
      } catch (err: unknown) {
        setError(err instanceof Error ? err.message : "Deposit failed. Try again.");
      }
    });
  }

  return (
    <div className="flex flex-col px-4 pt-6 pb-4 min-h-full">

      {/* Header */}
      <div className="flex items-center gap-3 mb-6">
        <button onClick={() => router.back()} className="text-ink-soft text-sm">← Back</button>
      </div>
      <h1 className="text-2xl font-bold text-ink mb-2">Deposit via M-Pesa</h1>
      <p className="text-sm text-ink-muted mb-6">
        Enter amount — we&apos;ll send an M-Pesa prompt to your phone.
      </p>

      {success ? (
        <div className="flex flex-col gap-5">
          <Card className="border border-green-200 bg-green-50" padding="lg">
            <div className="flex flex-col items-center gap-3 py-4 text-center">
              <span className="text-5xl">📱</span>
              <p className="font-semibold text-ink text-lg">Check your phone</p>
              <p className="text-sm text-ink-muted">{success}</p>
            </div>
          </Card>
          <Button fullWidth onClick={() => { setSuccess(""); }}>
            Deposit again
          </Button>
          <Button variant="ghost" fullWidth onClick={() => router.push("/home")}>
            Back to home
          </Button>
        </div>
      ) : (
        <form onSubmit={handleSubmit} className="flex flex-col gap-5">

          {/* Amount presets */}
          <div className="grid grid-cols-4 gap-2">
            {KES_PRESETS.map((k) => (
              <button
                key={k}
                type="button"
                onClick={() => setAmountKES(String(k))}
                className={[
                  "py-2.5 rounded-btn text-sm font-semibold border transition-colors",
                  amountKES === String(k)
                    ? "bg-green-600 text-white border-green-600"
                    : "bg-surface-card text-ink border-border",
                ].join(" ")}
              >
                {formatKES(k).replace("KES ", "")}
              </button>
            ))}
          </div>

          <Input
            label="Amount (KES)"
            prefix="KES"
            type="number"
            inputMode="decimal"
            placeholder="0"
            value={amountKES}
            onChange={(e) => setAmountKES(e.target.value)}
          />

          {sats > 0 && (
            <Card className="bg-green-50 border border-green-100">
              <div className="flex items-center justify-between text-sm">
                <span className="text-ink-soft">You will receive</span>
                <span className="font-bold text-ink">{formatSats(sats)}</span>
              </div>
              <div className="flex items-center justify-between text-xs text-ink-muted mt-1">
                <span>Rate</span>
                <span>1 BTC = {formatKES(btcKES)}</span>
              </div>
            </Card>
          )}

          <Input
            label="M-Pesa phone"
            type="tel"
            inputMode="tel"
            placeholder="0712 345 678"
            value={phone}
            onChange={(e) => setPhone(e.target.value)}
            hint="STK Push will be sent to this number"
            error={error}
          />

          <Button type="submit" fullWidth loading={pending} disabled={!amountKES || kes < 10}>
            Request M-Pesa payment
          </Button>

          {/* How it works */}
          <Card className="border border-border bg-surface-card">
            <p className="text-xs font-semibold text-ink-muted mb-2">How it works</p>
            <ol className="flex flex-col gap-1.5">
              {[
                "Enter amount and tap Request",
                "An M-Pesa prompt appears on your phone",
                "Enter your PIN to confirm",
                "Sats are credited instantly",
              ].map((step, i) => (
                <li key={i} className="flex gap-2 text-xs text-ink-soft">
                  <span className="w-4 h-4 rounded-full bg-green-100 text-green-700 flex items-center justify-center font-bold flex-shrink-0 text-[10px]">
                    {i + 1}
                  </span>
                  {step}
                </li>
              ))}
            </ol>
          </Card>

        </form>
      )}

    </div>
  );
}
