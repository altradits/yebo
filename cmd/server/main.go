package main

import (
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/yebobank/yebobank/internal/db"
	"github.com/yebobank/yebobank/internal/handlers"
	"github.com/yebobank/yebobank/internal/middleware"
	"github.com/yebobank/yebobank/internal/services/interest"
	"github.com/yebobank/yebobank/internal/services/lightning"
	"github.com/yebobank/yebobank/internal/services/mpesa"
	"github.com/yebobank/yebobank/internal/services/rates"
)

func main() {
	// ── Database ────────────────────────────────────────────────────────────────
	if err := db.Open(); err != nil {
		log.Fatalf("db: %v", err)
	}

	migrationsDir := envOr("MIGRATIONS_DIR", "docs/database/migrations")
	if err := db.Migrate(migrationsDir); err != nil {
		log.Fatalf("migrations: %v", err)
	}
	if err := db.Seed(); err != nil {
		log.Fatalf("seed: %v", err)
	}

	// ── Templates ───────────────────────────────────────────────────────────────
	tmplDir := envOr("TEMPLATES_DIR", filepath.Join("web", "templates"))
	if err := handlers.InitTemplates(tmplDir); err != nil {
		log.Fatalf("templates: %v", err)
	}

	// ── External services ───────────────────────────────────────────────────────
	mpesa.Init()
	if err := lightning.Init(); err != nil {
		log.Printf("lightning: %v (continuing without LND)", err)
	}

	// ── Background jobs ─────────────────────────────────────────────────────────
	rates.StartFeed()
	interest.StartScheduler()
	interest.StartUnlockChecker()

	// ── Routes ──────────────────────────────────────────────────────────────────
	mux := http.NewServeMux()

	// Static assets
	staticDir := envOr("STATIC_DIR", filepath.Join("web", "static"))
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir(staticDir))))

	// Public
	mux.HandleFunc("/", handlers.Home)
	mux.HandleFunc("/login", middleware.IPRateLimit(10)(http.HandlerFunc(handlers.Login)).ServeHTTP)
	mux.HandleFunc("/register", handlers.Register)
	mux.HandleFunc("/logout", handlers.Logout)

	// LNURL-pay (Lightning Address protocol)
	mux.HandleFunc("/.well-known/lnurlp/", handlers.LNURLPay)
	mux.HandleFunc("/lnurl/pay/", handlers.LNURLPayCallback)

	// Webhooks (public — called by Safaricom and LND)
	mux.HandleFunc("/webhook/mpesa", handlers.MpesaSTKCallback)
	mux.HandleFunc("/webhook/lnd", handlers.LNDInvoiceSettled)

	// Customer (authenticated)
	auth := middleware.RequireAuth
	mux.Handle("/dashboard", auth(http.HandlerFunc(handlers.Dashboard)))
	mux.Handle("/deposit", auth(http.HandlerFunc(handlers.Deposit)))
	mux.Handle("/deposit/mpesa", auth(http.HandlerFunc(handlers.DepositMpesa)))
	mux.Handle("/deposit/lightning", auth(http.HandlerFunc(handlers.DepositLightning)))
	mux.Handle("/withdraw", auth(http.HandlerFunc(handlers.Withdraw)))
	mux.Handle("/withdraw/mpesa", auth(http.HandlerFunc(handlers.WithdrawMpesa)))
	mux.Handle("/withdraw/lightning", auth(http.HandlerFunc(handlers.WithdrawLightning)))
	mux.Handle("/history", auth(http.HandlerFunc(handlers.History)))
	mux.Handle("/savings", auth(http.HandlerFunc(handlers.Savings)))
	mux.Handle("/savings/lock", auth(http.HandlerFunc(handlers.SavingsLock)))
	mux.Handle("/chama", auth(http.HandlerFunc(handlers.Chama)))
	mux.Handle("/chama/create", auth(http.HandlerFunc(handlers.ChamaCreate)))
	mux.Handle("/chama/contribute", auth(http.HandlerFunc(handlers.ChamaContribute)))
	mux.Handle("/settings", auth(http.HandlerFunc(handlers.Settings)))

	// Agent (authenticated + role guard)
	agentAuth := func(h http.Handler) http.Handler {
		return auth(middleware.RequireRole("agent")(h))
	}
	mux.Handle("/agent", agentAuth(http.HandlerFunc(handlers.AgentDashboard)))
	mux.Handle("/agent/cashin", agentAuth(http.HandlerFunc(handlers.AgentCashIn)))
	mux.Handle("/agent/cashout", agentAuth(http.HandlerFunc(handlers.AgentCashOut)))

	// Trader
	traderAuth := func(h http.Handler) http.Handler {
		return auth(middleware.RequireRole("trader")(h))
	}
	mux.Handle("/trader", traderAuth(http.HandlerFunc(handlers.TraderDashboard)))
	mux.Handle("/trader/assets", traderAuth(http.HandlerFunc(handlers.TraderAssets)))
	mux.Handle("/trader/distribute", traderAuth(http.HandlerFunc(handlers.TraderRunDistribution)))
	mux.Handle("/trader/profit", traderAuth(http.HandlerFunc(handlers.TraderProfit)))

	// Admin
	adminAuth := func(h http.Handler) http.Handler {
		return auth(middleware.RequireRole("admin")(h))
	}
	mux.Handle("/admin", adminAuth(http.HandlerFunc(handlers.AdminDashboard)))
	mux.Handle("/admin/customers", adminAuth(http.HandlerFunc(handlers.AdminCustomers)))
	mux.Handle("/admin/customers/toggle", adminAuth(http.HandlerFunc(handlers.AdminToggleUser)))
	mux.Handle("/admin/agents", adminAuth(http.HandlerFunc(handlers.AdminAgents)))
	mux.Handle("/admin/agents/approve", adminAuth(http.HandlerFunc(handlers.AdminApproveAgent)))
	mux.Handle("/admin/settings", adminAuth(http.HandlerFunc(handlers.AdminSettings)))

	// ── Listen ──────────────────────────────────────────────────────────────────
	port := envOr("PORT", "8080")
	log.Printf("yebobank: listening on :%s", port)
	if err := http.ListenAndServe(":"+port, mux); err != nil {
		log.Fatalf("server: %v", err)
	}
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
