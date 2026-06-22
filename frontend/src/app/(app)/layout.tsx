import { BottomNav } from "@/components/bottom-nav";

export default function AppLayout({ children }: { children: React.ReactNode }) {
  return (
    <div className="flex flex-col min-h-full bg-surface">
      {/* Page content — bottom padding clears the nav bar */}
      <main className="flex-1 pb-20 safe-top">{children}</main>
      <BottomNav />
    </div>
  );
}
