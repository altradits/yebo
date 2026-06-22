"use client";
import { useState, useTransition } from "react";
import { useRouter } from "next/navigation";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { requestOTP } from "@/lib/api";

export default function LoginPage() {
  const router = useRouter();
  const [pending, startTransition] = useTransition();
  const [phone, setPhone] = useState("");
  const [error, setError] = useState("");

  function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    setError("");

    // Normalise: strip spaces/dashes, add +254 if bare 07xx
    const raw = phone.replace(/[\s\-]/g, "");
    const normalised = raw.startsWith("0")
      ? "+254" + raw.slice(1)
      : raw.startsWith("254")
      ? "+" + raw
      : raw;

    if (!/^\+254[0-9]{9}$/.test(normalised)) {
      setError("Enter a valid Kenyan number: 07xx xxx xxx");
      return;
    }

    startTransition(async () => {
      try {
        await requestOTP(normalised);
        router.push(`/verify?phone=${encodeURIComponent(normalised)}`);
      } catch (err: unknown) {
        setError(err instanceof Error ? err.message : "Could not send code. Try again.");
      }
    });
  }

  return (
    <main className="flex flex-col min-h-full bg-surface">
      {/* Header */}
      <div className="flex flex-col items-center pt-20 pb-10 px-6">
        <div className="w-16 h-16 rounded-2xl bg-green-600 flex items-center justify-center mb-6">
          <span className="text-white text-2xl font-bold">A</span>
        </div>
        <h1 className="text-2xl font-bold text-ink">Welcome to Altradits</h1>
        <p className="text-ink-soft mt-2 text-center">
          Your community Bitcoin bank. No ID needed — just your phone.
        </p>
      </div>

      {/* Form */}
      <form onSubmit={handleSubmit} className="flex flex-col gap-4 px-6 pb-10">
        <Input
          label="Phone number"
          type="tel"
          placeholder="0712 345 678"
          value={phone}
          onChange={(e) => setPhone(e.target.value)}
          error={error}
          inputMode="tel"
          autoComplete="tel"
          autoFocus
        />
        <Button type="submit" loading={pending} fullWidth>
          Get verification code
        </Button>
        <p className="text-xs text-ink-muted text-center">
          We&apos;ll send a 6-digit code via SMS. Standard rates apply.
        </p>
      </form>

      {/* Trust badges */}
      <div className="mt-auto px-6 pb-10 flex flex-col gap-3">
        {[
          { icon: "🔒", text: "No ID or documents needed" },
          { icon: "⚡", text: "Instant Lightning payments" },
          { icon: "📱", text: "M-Pesa in, M-Pesa out" },
        ].map((b) => (
          <div key={b.text} className="flex items-center gap-3 text-sm text-ink-soft">
            <span className="text-lg">{b.icon}</span>
            <span>{b.text}</span>
          </div>
        ))}
      </div>
    </main>
  );
}
