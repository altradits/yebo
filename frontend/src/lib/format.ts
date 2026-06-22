/** Format satoshis as a human-readable string: 1,234,567 sats */
export function formatSats(sats: number): string {
  return sats.toLocaleString("en-KE") + " sats";
}

/** Convert sats to KES at current rate */
export function satsToKES(sats: number, btcKES: number): number {
  return Math.round((sats / 100_000_000) * btcKES);
}

/** Format KES amount */
export function formatKES(amount: number): string {
  return "KES " + amount.toLocaleString("en-KE");
}

/** Short relative time: "2m ago", "3h ago", "2d ago" */
export function timeAgo(date: Date | string): string {
  const d = typeof date === "string" ? new Date(date) : date;
  const diff = Math.floor((Date.now() - d.getTime()) / 1000);
  if (diff < 60) return "just now";
  if (diff < 3600) return `${Math.floor(diff / 60)}m ago`;
  if (diff < 86400) return `${Math.floor(diff / 3600)}h ago`;
  if (diff < 604800) return `${Math.floor(diff / 86400)}d ago`;
  return d.toLocaleDateString("en-KE", { day: "numeric", month: "short" });
}

/** Mask phone number for display: +254 7** *** 678 */
export function maskPhone(phone: string): string {
  if (phone.length < 8) return phone;
  return phone.slice(0, 5) + "*** ***" + phone.slice(-3);
}
