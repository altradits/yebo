"use client";
import { useState, useTransition, useRef, useEffect, Suspense } from "react";
import { useRouter, useSearchParams } from "next/navigation";
import { Button } from "@/components/ui/button";
import { verifyOTP } from "@/lib/api";

export default function VerifyPage() {
  return (
    <Suspense>
      <VerifyForm />
    </Suspense>
  );
}

function VerifyForm() {
  const router = useRouter();
  const params = useSearchParams();
  const phone = params.get("phone") ?? "";
  const [otp, setOtp] = useState(["", "", "", "", "", ""]);
  const [error, setError] = useState("");
  const [pending, startTransition] = useTransition();
  const [resendSeconds, setResendSeconds] = useState(30);
  const inputs = useRef<(HTMLInputElement | null)[]>([]);

  // Countdown for resend
  useEffect(() => {
    if (resendSeconds <= 0) return;
    const t = setTimeout(() => setResendSeconds((s) => s - 1), 1000);
    return () => clearTimeout(t);
  }, [resendSeconds]);

  function handleChange(i: number, val: string) {
    const digit = val.replace(/\D/g, "").slice(-1);
    const next = [...otp];
    next[i] = digit;
    setOtp(next);
    setError("");
    if (digit && i < 5) inputs.current[i + 1]?.focus();
    if (next.every((d) => d)) submitOTP(next.join(""));
  }

  function handleKeyDown(i: number, e: React.KeyboardEvent) {
    if (e.key === "Backspace" && !otp[i] && i > 0) {
      inputs.current[i - 1]?.focus();
    }
  }

  function handlePaste(e: React.ClipboardEvent) {
    const text = e.clipboardData.getData("text").replace(/\D/g, "").slice(0, 6);
    if (text.length === 6) {
      setOtp(text.split(""));
      submitOTP(text);
    }
  }

  function submitOTP(code: string) {
    setError("");
    startTransition(async () => {
      try {
        await verifyOTP(phone, code);
        router.push("/");
      } catch (err: unknown) {
        setError(err instanceof Error ? err.message : "Wrong code. Try again.");
        setOtp(["", "", "", "", "", ""]);
        inputs.current[0]?.focus();
      }
    });
  }

  return (
    <main className="flex flex-col min-h-full bg-surface px-6">
      <button
        onClick={() => router.back()}
        className="mt-6 self-start text-ink-soft flex items-center gap-1 text-sm"
      >
        ← Back
      </button>

      <div className="mt-10 mb-8">
        <h1 className="text-2xl font-bold text-ink">Enter your code</h1>
        <p className="text-ink-soft mt-2">
          We sent a 6-digit code to <span className="font-medium text-ink">{phone}</span>
        </p>
      </div>

      {/* OTP boxes */}
      <div className="flex gap-3 justify-between" onPaste={handlePaste}>
        {otp.map((digit, i) => (
          <input
            key={i}
            ref={(el) => { inputs.current[i] = el; }}
            type="text"
            inputMode="numeric"
            autoComplete="one-time-code"
            maxLength={1}
            value={digit}
            onChange={(e) => handleChange(i, e.target.value)}
            onKeyDown={(e) => handleKeyDown(i, e)}
            className={[
              "w-full aspect-square text-center text-2xl font-bold rounded-btn border",
              "focus:outline-none focus:ring-2 focus:ring-green-600 focus:border-transparent",
              error ? "border-red-400 bg-red-50" : "border-border bg-surface-card",
            ].join(" ")}
          />
        ))}
      </div>

      {error && <p className="mt-4 text-sm text-red-600 text-center">{error}</p>}

      <Button
        className="mt-6"
        fullWidth
        loading={pending}
        disabled={otp.some((d) => !d)}
        onClick={() => submitOTP(otp.join(""))}
      >
        Verify
      </Button>

      <div className="mt-6 text-center">
        {resendSeconds > 0 ? (
          <p className="text-sm text-ink-muted">Resend code in {resendSeconds}s</p>
        ) : (
          <button
            className="text-sm font-medium text-green-600"
            onClick={() => {
              setResendSeconds(30);
              setOtp(["", "", "", "", "", ""]);
              inputs.current[0]?.focus();
            }}
          >
            Resend code
          </button>
        )}
      </div>
    </main>
  );
}
