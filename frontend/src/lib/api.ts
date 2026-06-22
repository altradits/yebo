const BASE = process.env.NEXT_PUBLIC_API_URL ?? "http://localhost:8080";

async function request<T>(path: string, init?: RequestInit): Promise<T> {
  const res = await fetch(`${BASE}${path}`, {
    credentials: "include",
    headers: { "Content-Type": "application/json", ...init?.headers },
    ...init,
  });
  if (!res.ok) {
    const body = await res.json().catch(() => ({}));
    throw new Error(body.error ?? `HTTP ${res.status}`);
  }
  return res.json() as T;
}

// ── Auth ──────────────────────────────────────────────────────────────────────

export function requestOTP(phone: string) {
  return request<{ message: string }>("/api/auth/request-otp", {
    method: "POST",
    body: JSON.stringify({ phone }),
  });
}

export function verifyOTP(phone: string, otp: string) {
  return request<{ token: string; user: User }>("/api/auth/verify-otp", {
    method: "POST",
    body: JSON.stringify({ phone, otp }),
  });
}

export function logout() {
  return request<void>("/api/auth/logout", { method: "POST" });
}

// ── User ──────────────────────────────────────────────────────────────────────

export function getMe() {
  return request<User>("/api/user");
}

export function getBalance() {
  return request<{ balance_sats: number; btc_kes: number }>("/api/user/balance");
}

export function getTransactions(limit = 20, offset = 0) {
  return request<Transaction[]>(`/api/user/transactions?limit=${limit}&offset=${offset}`);
}

// ── Community ─────────────────────────────────────────────────────────────────

export function getCommunityStats() {
  return request<CommunityStats>("/api/community/stats");
}

export function getChamas() {
  return request<Chama[]>("/api/chamas");
}

export function getChama(id: number) {
  return request<ChamaDetail>(`/api/chamas/${id}`);
}

export function depositMpesa(amountKES: number, phone: string) {
  return request<{ message: string; checkout_request_id: string }>("/api/deposit/mpesa", {
    method: "POST",
    body: JSON.stringify({ amount_kes: amountKES, phone }),
  });
}

export function withdrawMpesa(amountKES: number, phone: string) {
  return request<{ message: string; amount_sats: number }>("/api/withdraw/mpesa", {
    method: "POST",
    body: JSON.stringify({ amount_kes: amountKES, phone }),
  });
}

export function getSavings() {
  return request<SavingsData>("/api/savings");
}

export function lockSavings(amountSats: number) {
  return request<{ message: string; lock_id: number; unlocks_at: string; rate_bps: number }>("/api/savings/lock", {
    method: "POST",
    body: JSON.stringify({ amount_sats: amountSats }),
  });
}

// ── Types ─────────────────────────────────────────────────────────────────────

export interface User {
  id: number;
  phone: string;
  full_name: string;
  balance_sats: number;
  btc_kes: number;
  total_interest_earned: number;
  created_at: string;
}

export interface Transaction {
  id: number;
  type: "deposit" | "withdrawal" | "send" | "receive" | "savings_interest" | "chama_contribution" | "chama_payout";
  amount_sats: number;
  note: string;
  ref_id: string;
  created_at: string;
  is_credit: boolean;
}

export interface CommunityStats {
  member_count: number;
  total_savings_sats: number;
  total_interest_paid_sats: number;
  btc_kes: number;
}

export interface Chama {
  id: number;
  name: string;
  description: string;
  balance_sats: number;
  balance_kes: number;
  member_count: number;
  status: "active" | "paused" | "closed";
  role: "admin" | "member";
  contribution_sats: number;
  cycle_days: number;
  btc_kes: number;
}

export interface ChamaMember {
  id: number;
  name: string;
  role: "admin" | "member";
  joined_at: string;
  is_me: boolean;
}

export interface ChamaDetail extends Chama {
  description: string;
  created_at: string;
  members: ChamaMember[];
}

export interface SavingsLock {
  id: number;
  amount_sats: number;
  amount_kes: number;
  lock_days: number;
  rate_bps: number;
  locked_at: string;
  unlocks_at: string;
  status: "locked" | "unlocked" | "early_exit";
  interest_earned_sats: number;
}

export interface SavingsData {
  locks: SavingsLock[];
  pool_rate_bps: number;
  pool_lock_days: number;
  min_sats: number;
  max_sats: number;
  btc_kes: number;
}
