"use client";
import { useState, useEffect, useTransition } from "react";
import { useRouter } from "next/navigation";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Card } from "@/components/ui/card";
import { formatKES, formatSats, satsToKES } from "@/lib/format";
import { getMe, withdrawMpesa } from "@/lib/api";

const KES_PRESETS = [500, 1_000, 2_000, 5_000];

export default function WithdrawPage() {
  const router = useRouter();
  const [pending, startTransition] = useTransition();
  const [amountKES, setAmountKES] = useState("");
  const [phone, setPhone] = useState("");
  const [error, setError] = useState("");
  const [success, setSuccess] = useState("");
  const [balanceSats, setBalanceSats] = useState(0);
  const [btcKES, setBtcKES] = useState(0);
  const [step, setStep] = useState<"amount" | "confirm">("amount");

  useEffect(() => {
    getMe().then((u) => {
      setBtcKES(u.btc_kes);
      setBalanceSats(u.balance_sats);
      setPhone(u.phone);
    }).catch(() => {});
  }, []);

  const kes = parseFloat(amountKES) || 0;
  const sats = btcKES > 0 ? Math.round((kes / btcKES) * 1e8) : 0;
  const balanceKES = satsToKES(balanceSats, btcKES);

  function handleReview(e: React.FormEvent) {
    e.preventDefault();
    setError("");
    if (kes < 10) { setError("Minimum withdrawal is KES 10"); return; }
    if (!phone)   { setError("Phone number required"); return; }
    if (sats > balanceSats) { setError(`Insufficient balance (${formatKES(balanceKES)} available)`); return; }
    setStep("confirm");
  }

  function handleConfirm() {
    setError("");
    startTransition(async () => {
      try {
        const res = await withdrawMpesa(kes, phone);
        setSuccess(res.message);
      } catch (err: unknown) {
        setError(err instanceof Error ? err.message : "Withdrawal failed. Try again.");
        setStep("amount");
      }
    });
  }

  if (success) {
    return (
      <div className="flex flex-col px-4 pt-6 pb-4 min-h-full gap-5">
        <button onClick={() => router.back()} className="text-ink-soft text-sm self-start">← Back</button>
        <Card className="border border-green-200 bg-green-50" padding="lg">
          <div className="flex flex-col items-center gap-3 py-4 text-center">
            <span className="text-5xl">✅</span>
            <p className="font-semibold text-ink text-lg">Withdrawal sent</p>
            <p className="text-sm text-ink-muted">{success}</p>
          </div>
        </Card>
        <Button fullWidth onClick={() => router.push("/home")}>Back to home</Button>
      </div>
    );
  }

  return (
    <div className="flex flex-col px-4 pt-6 pb-4 min-h-full">

      <div className="flex items-center gap-3 mb-6">
        <button
          onClick={() => step === "amount" ? router.back() : setStep("amount")}
          className="text-ink-soft text-sm"
        >
          ← {step === "amount" ? "Back" : "Change amount"}
        </button>
      </div>
      <h1 className="text-2xl font-bold text-ink mb-2">Withdraw to M-Pesa</h1>

      {/* Balance */}
      {btcKES > 0 && (
        <p className="text-sm text-ink-muted mb-6">
          Available: <span className="font-semibold text-ink">{formatKES(balanceKES)}</span>
          {" "}({formatSats(balanceSats)})
        </p>
      )}

      {step === "amount" && (
        <form onSubmit={handleReview} className="flex flex-col gap-5">

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
            <Card className="bg-surface-card border border-border">
              <div className="flex items-center justify-between text-sm">
                <span className="text-ink-soft">You will spend</span>
                <span className="font-bold text-debit">−{formatSats(sats)}</span>
              </div>
            </Card>
          )}

          <Input
            label="Send to phone"
            type="tel"
            inputMode="tel"
            placeholder="0712 345 678"
            value={phone}
            onChange={(e) => setPhone(e.target.value)}
            hint="M-Pesa payment sent to this number"
            error={error}
          />

          <Button type="submit" fullWidth disabled={!amountKES || kes < 10}>
            Review withdrawal
          </Button>

        </form>
      )}

      {step === "confirm" && (
        <div className="flex flex-col gap-5">
          <Card className="border border-border" padding="lg">
            <div className="flex flex-col items-center gap-2 py-4 text-center">
              <span className="text-5xl">📤</span>
              <p className="text-3xl font-bold text-ink mt-2">{formatKES(kes)}</p>
              <p className="text-sm text-ink-muted">{formatSats(sats)} deducted from balance</p>
              <p className="text-sm text-ink-soft mt-2">→ {phone}</p>
            </div>
          </Card>

          {error && <p className="text-sm text-red-600 text-center">{error}</p>}

          <p className="text-xs text-ink-muted text-center">
            This will deduct sats from your wallet and send KES via M-Pesa. This action is irreversible.
          </p>

          <Button fullWidth loading={pending} onClick={handleConfirm}>
            Confirm withdrawal
          </Button>
          <Button variant="ghost" fullWidth onClick={() => setStep("amount")}>
            Go back
          </Button>
        </div>
      )}

    </div>
  );
}
