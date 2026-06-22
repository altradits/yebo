"use client";
import { useState, useEffect } from "react";
import { useRouter } from "next/navigation";
import { Card } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { maskPhone } from "@/lib/format";
import { getMe, logout, type User } from "@/lib/api";

const MENU_SECTIONS = [
  {
    title: "Account",
    items: [
      { icon: "👤", label: "Display name",   href: "/profile/name"          },
      { icon: "🔒", label: "App PIN",         href: "/profile/pin"           },
      { icon: "🌍", label: "Language",        href: "/profile/language", value: "English" },
    ],
  },
  {
    title: "Preferences",
    items: [
      { icon: "🔔", label: "Notifications",   href: "/profile/notifications" },
      { icon: "💱", label: "Display currency", href: "/profile/currency", value: "KES"  },
    ],
  },
  {
    title: "Support",
    items: [
      { icon: "💬", label: "Chat with support", href: "/support" },
      { icon: "❓", label: "Help & FAQ",         href: "/faq"     },
      { icon: "🚩", label: "Report a transaction", href: "/report" },
    ],
  },
  {
    title: "Referrals",
    items: [
      { icon: "🎁", label: "Invite a friend", href: "/profile/referral", badge: "Earn KES 50" },
      { icon: "📊", label: "Your referrals",  href: "/profile/referrals" },
    ],
  },
];

export default function ProfilePage() {
  const router = useRouter();
  const [user, setUser] = useState<User | null>(null);
  const [signingOut, setSigningOut] = useState(false);

  useEffect(() => {
    getMe().then(setUser).catch(() => {});
  }, []);

  async function handleSignOut() {
    setSigningOut(true);
    try { await logout(); } catch {}
    router.push("/login");
  }

  const displayName = user?.full_name || user?.phone || "…";
  const joined = user?.created_at
    ? new Date(user.created_at).toLocaleDateString("en-KE", { month: "long", year: "numeric" })
    : "";

  return (
    <div className="flex flex-col pt-6 pb-4 min-h-full px-4 gap-5">

      {/* Profile header */}
      <div className="flex items-center gap-4">
        <div className="w-16 h-16 rounded-full bg-green-100 flex items-center justify-center text-green-700 font-bold text-2xl flex-shrink-0">
          {displayName[0]?.toUpperCase()}
        </div>
        <div>
          <h1 className="text-xl font-bold text-ink">{displayName}</h1>
          <p className="text-sm text-ink-muted">{user?.phone ? maskPhone(user.phone) : ""}</p>
          {joined && <p className="text-xs text-ink-muted mt-0.5">Member since {joined}</p>}
        </div>
      </div>

      {/* Menu sections */}
      {MENU_SECTIONS.map((section) => (
        <div key={section.title}>
          <p className="text-xs font-semibold text-ink-muted uppercase tracking-wide mb-2 px-1">
            {section.title}
          </p>
          <Card padding="none">
            <ul className="divide-y divide-border">
              {section.items.map((item) => (
                <li key={item.label}>
                  <button
                    onClick={() => router.push(item.href)}
                    className="w-full flex items-center gap-3 px-4 py-4 active:bg-green-50 text-left"
                  >
                    <span className="text-lg w-7 text-center flex-shrink-0">{item.icon}</span>
                    <span className="flex-1 text-sm font-medium text-ink">{item.label}</span>
                    {"badge" in item && item.badge && (
                      <span className="text-xs bg-gold-100 text-gold-500 px-2 py-0.5 rounded-full font-medium mr-1">
                        {item.badge}
                      </span>
                    )}
                    {"value" in item && item.value && (
                      <span className="text-sm text-ink-muted mr-1">{item.value}</span>
                    )}
                    <span className="text-ink-muted">›</span>
                  </button>
                </li>
              ))}
            </ul>
          </Card>
        </div>
      ))}

      {/* Sign out */}
      <Button variant="danger" fullWidth loading={signingOut} onClick={handleSignOut}>
        Sign out
      </Button>

      <p className="text-xs text-ink-muted text-center pb-2">
        Altradits v1.0 · Bitcoin community bank
      </p>

    </div>
  );
}
