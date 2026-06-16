package handlers

import (
	"net/http"

	"github.com/yebobank/yebobank/internal/db"
	"github.com/yebobank/yebobank/internal/middleware"
	"github.com/yebobank/yebobank/internal/services/interest"
)

func TraderDashboard(w http.ResponseWriter, r *http.Request) {
	var totalLocked int64
	var activeDistributions int
	db.DB.QueryRow(`SELECT COALESCE(SUM(amount_sats),0) FROM savings_locks WHERE status='active'`).Scan(&totalLocked)           //nolint:errcheck
	db.DB.QueryRow(`SELECT COUNT(*) FROM interest_distributions WHERE status='complete'`).Scan(&activeDistributions)            //nolint:errcheck
	renderTemplate(w, r, "trader/dashboard.html", map[string]interface{}{
		"TotalLocked":      totalLocked,
		"DistributionCount": activeDistributions,
	})
}

func TraderAssets(w http.ResponseWriter, r *http.Request) {
	rows, _ := db.DB.Query(`SELECT id, name, asset_type, balance_sats, apy_bps FROM treasury_assets ORDER BY id`)
	defer rows.Close()
	type asset struct {
		ID            int64
		Name, Type    string
		BalanceSats   int64
		APYBPS        int
	}
	var assets []asset
	for rows.Next() {
		var a asset
		rows.Scan(&a.ID, &a.Name, &a.Type, &a.BalanceSats, &a.APYBPS) //nolint:errcheck
		assets = append(assets, a)
	}
	renderTemplate(w, r, "trader/assets.html", map[string]interface{}{"Assets": assets})
}

func TraderRunDistribution(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/trader", http.StatusSeeOther)
		return
	}
	actorID := middleware.UserID(r)
	if err := interest.Distribute(&actorID); err != nil {
		renderTemplate(w, r, "trader/dashboard.html", map[string]interface{}{"Error": err.Error()})
		return
	}
	db.DB.Exec(`INSERT INTO audit_log (actor_id, action) VALUES ($1, 'interest_distribution')`, actorID) //nolint:errcheck
	http.Redirect(w, r, "/trader?distributed=1", http.StatusSeeOther)
}

func TraderProfit(w http.ResponseWriter, r *http.Request) {
	rows, _ := db.DB.Query(`
		SELECT run_at, total_locked_sats, total_interest_sats, accounts_credited, rate_bps
		FROM interest_distributions ORDER BY run_at DESC LIMIT 24
	`)
	defer rows.Close()
	type dist struct {
		RunAt           string
		TotalLocked     int64
		TotalInterest   int64
		AccountsCredited int
		RateBPS         int
	}
	var dists []dist
	for rows.Next() {
		var d dist
		rows.Scan(&d.RunAt, &d.TotalLocked, &d.TotalInterest, &d.AccountsCredited, &d.RateBPS) //nolint:errcheck
		dists = append(dists, d)
	}
	renderTemplate(w, r, "trader/profit.html", map[string]interface{}{"Distributions": dists})
}
