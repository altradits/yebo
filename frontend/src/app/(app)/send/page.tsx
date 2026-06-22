"use client";
import { useState } from "react";
import { useRouter } from "next/navigation";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Card } from "@/components/ui/card";
import { formatKES, satsToKES } from "@/lib/format";

const BTC_KES = 9_200_000;
const KES_PRESETS = [100, 500, 1_000, 5_000];

const RECENT_RECIPIENTS = [
  { name: "James Odhiambo",  phone: "+254711000001", initials: "JO" },
  { name: "Grace Wafula",    phone: "+254722000002", initials: "GW" },
  { name: "Peter Kamau",     phone: "+254733000003", initials: "PK" },
  { name: "Faith Akinyi",    phone: "+254744000004", initials: "FA" },
];

export default function SendPage() {
  const router = useRouter();
  const [step, setStep] = useState<"recipient" | "amount" | "confirm">("recipient");
  const [recipient, setRecipient] = useState("");
  const [recipientName, setRecipientName] = useState("");
  const [amountKES, setAmountKES] = useState("");
  const [memo, setMemo] = useState("");
  const [sending, setSending] = useState(false);

  const amountSats = amountKES
    ? Math.round((parseFloat(amountKES) / BTC_KES) * 100_000_000)
    : 0;

  function selectRecipient(phone: string, name: string) {
    setRecipient(phone);
    setRecipientName(name);
    setStep("amount");
  }

  async function handleSend() {
    setSending(true);
    // TODO: call /api/transactions/send
    await new Promise((r) => setTimeout(r, 1500));
    setSending(false);
    router.push("/?sent=1");
  }

  return (
    <div className="flex flex-col px-4 pt-6 pb-4 min-h-full">

      {/* Header */}
      <div className="flex items-center gap-3 mb-6">
        <button onClick={() => step === "recipient" ? router.back() : setStep("recipient")} className="text-ink-soft text-sm">
          ← {step === "recipient" ? "Back" : "Change recipient"}
        </button>
      </div>
      <h1 className="text-2xl font-bold text-ink mb-6">Send money</h1>

      {/* Step 1: Recipient */}
      {step === "recipient" && (
        <div className="flex flex-col gap-5">
          <Input
            label="Phone number or Lightning address"
            placeholder="07xx xxx xxx or name@altradits.com"
            value={recipient}
            onChange={(e) => setRecipient(e.target.value)}
            type="text"
            inputMode="tel"
          />

          {/* QR option */}
          <button className="flex items-center gap-3 p-4 border border-border rounded-btn bg-surface-card text-sm font-medium text-ink">
            <span className="text-xl">📷</span>
            Scan QR code or Lightning invoice
          </button>

          <Button
            fullWidth
            disabled={!recipient}
            onClick={() => { setRecipientName(recipient); setStep("amount"); }}
          >
            Continue
          </Button>

          {/* Recent */}
          {RECENT_RECIPIENTS.length > 0 && (
            <div>
              <p className="text-sm font-medium text-ink-soft mb-3">Recent</p>
              <div className="flex flex-col gap-0 bg-surface-card rounded-card overflow-hidden border border-border">
                {RECENT_RECIPIENTS.map((r, i) => (
                  <button
                    key={r.phone}
                    onClick={() => selectRecipient(r.phone, r.name)}
                    className={[
                      "flex items-center gap-3 px-4 py-3.5 active:bg-green-50 text-left",
                      i < RECENT_RECIPIENTS.length - 1 ? "border-b border-border" : "",
                    ].join(" ")}
                  >
                    <div className="w-10 h-10 rounded-full bg-green-100 flex items-center justify-center text-green-700 font-bold text-sm flex-shrink-0">
                      {r.initials}
                    </div>
                    <div>
                      <p className="text-sm font-medium text-ink">{r.name}</p>
                      <p className="text-xs text-ink-muted">{r.phone}</p>
                    </div>
                  </button>
                ))}
              </div>
            </div>
          )}
        </div>
      )}

      {/* Step 2: Amount */}
      {step === "amount" && (
        <div className="flex flex-col gap-5">
          <Card className="flex items-center gap-3 border border-green-100">
            <div className="w-10 h-10 rounded-full bg-green-100 flex items-center justify-center text-green-700 font-bold flex-shrink-0">
              {recipientName[0]?.toUpperCase()}
            </div>
            <div>
              <p className="font-semibold text-ink text-sm">{recipientName}</p>
              <p className="text-xs text-ink-muted">{recipient !== recipientName ? recipient : ""}</p>
            </div>
          </Card>

          {/* KES presets */}
          <div className="grid grid-cols-4 gap-2">
            {KES_PRESETS.map((kes) => (
              <button
                key={kes}
                onClick={() => setAmountKES(String(kes))}
                className={[
                  "py-2.5 rounded-btn text-sm font-semibold border transition-colors",
                  amountKES === String(kes)
                    ? "bg-green-600 text-white border-green-600"
                    : "bg-surface-card text-ink border-border",
                ].join(" ")}
              >
                {formatKES(kes).replace("KES ", "")}
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

          {amountSats > 0 && (
            <p className="text-sm text-ink-soft text-center">
              ≈ {amountSats.toLocaleString("en-KE")} sats
            </p>
          )}

          <Input
            label="Note (optional)"
            placeholder="What's this for?"
            value={memo}
            onChange={(e) => setMemo(e.target.value)}
          />

          <Button
            fullWidth
            disabled={!amountKES || parseFloat(amountKES) <= 0}
            onClick={() => setStep("confirm")}
          >
            Review
          </Button>
        </div>
      )}

      {/* Step 3: Confirm */}
      {step === "confirm" && (
        <div className="flex flex-col gap-5">
          <Card className="border border-border" padding="lg">
            <div className="flex flex-col items-center gap-2 py-4">
              <div className="w-16 h-16 rounded-full bg-green-100 flex items-center justify-center text-green-700 font-bold text-2xl">
                {recipientName[0]?.toUpperCase()}
              </div>
              <p className="font-semibold text-ink">{recipientName}</p>
              <p className="text-3xl font-bold text-ink mt-2">{formatKES(parseFloat(amountKES))}</p>
              <p className="text-sm text-ink-muted">{amountSats.toLocaleString("en-KE")} sats</p>
              {memo && <p className="text-sm text-ink-soft mt-1 italic">&ldquo;{memo}&rdquo;</p>}
            </div>
          </Card>

          <p className="text-xs text-ink-muted text-center">
            Lightning payment · instant · irreversible
          </p>

          <Button fullWidth loading={sending} onClick={handleSend}>
            Confirm & send
          </Button>
          <Button variant="ghost" fullWidth onClick={() => setStep("amount")}>
            Go back
          </Button>
        </div>
      )}

    </div>
  );
}
