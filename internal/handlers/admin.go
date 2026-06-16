package handlers

import (
	"net/http"

	"github.com/yebobank/yebobank/internal/db"
	"github.com/yebobank/yebobank/internal/middleware"
)

func AdminDashboard(w http.ResponseWriter, r *http.Request) {
	var userCount, agentCount, activeLocksCount int
	var totalBalanceSats int64
	db.DB.QueryRow(`SELECT COUNT(*) FROM users WHERE role='customer'`).Scan(&userCount)               //nolint:errcheck
	db.DB.QueryRow(`SELECT COUNT(*) FROM agents WHERE status='active'`).Scan(&agentCount)              //nolint:errcheck
	db.DB.QueryRow(`SELECT COUNT(*) FROM savings_locks WHERE status='active'`).Scan(&activeLocksCount) //nolint:errcheck
	db.DB.QueryRow(`SELECT COALESCE(SUM(balance_sats),0) FROM wallets`).Scan(&totalBalanceSats)        //nolint:errcheck

	renderTemplate(w, r, "admin/dashboard.html", map[string]interface{}{
		"UserCount":        userCount,
		"AgentCount":       agentCount,
		"ActiveLocks":      activeLocksCount,
		"TotalBalanceSats": totalBalanceSats,
	})
}

func AdminCustomers(w http.ResponseWriter, r *http.Request) {
	rows, _ := db.DB.Query(`
		SELECT u.id, u.phone, u.full_name, u.role, u.kyc_status, u.is_active, w.balance_sats
		FROM users u JOIN wallets w ON w.user_id=u.id
		ORDER BY u.created_at DESC LIMIT 200
	`)
	defer rows.Close()
	type user struct {
		ID                            int64
		Phone, FullName, Role, KYC   string
		IsActive                      bool
		BalanceSats                   int64
	}
	var users []user
	for rows.Next() {
		var u user
		rows.Scan(&u.ID, &u.Phone, &u.FullName, &u.Role, &u.KYC, &u.IsActive, &u.BalanceSats) //nolint:errcheck
		users = append(users, u)
	}
	renderTemplate(w, r, "admin/customers.html", map[string]interface{}{"Users": users})
}

func AdminToggleUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/admin/customers", http.StatusSeeOther)
		return
	}
	userID := r.FormValue("user_id")
	action := r.FormValue("action")
	actorID := middleware.UserID(r)
	isActive := action == "activate"
	db.DB.Exec(`UPDATE users SET is_active=$1, updated_at=NOW() WHERE id=$2`, isActive, userID) //nolint:errcheck
	db.DB.Exec(`INSERT INTO audit_log (actor_id, action, target_type, target_id, detail) VALUES ($1,$2,'user',$3,$4)`,
		actorID, action, userID, "Admin action") //nolint:errcheck
	http.Redirect(w, r, "/admin/customers", http.StatusSeeOther)
}

func AdminAgents(w http.ResponseWriter, r *http.Request) {
	rows, _ := db.DB.Query(`
		SELECT a.id, u.phone, a.business_name, a.location, a.status, a.float_sats
		FROM agents a JOIN users u ON u.id=a.user_id
		ORDER BY a.created_at DESC
	`)
	defer rows.Close()
	type agent struct {
		ID                              int64
		Phone, BizName, Location, Status string
		FloatSats                       int64
	}
	var agents []agent
	for rows.Next() {
		var a agent
		rows.Scan(&a.ID, &a.Phone, &a.BizName, &a.Location, &a.Status, &a.FloatSats) //nolint:errcheck
		agents = append(agents, a)
	}
	renderTemplate(w, r, "admin/agents.html", map[string]interface{}{"Agents": agents})
}

func AdminApproveAgent(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/admin/agents", http.StatusSeeOther)
		return
	}
	agentID := r.FormValue("agent_id")
	actorID := middleware.UserID(r)
	db.DB.Exec(`UPDATE agents SET status='active', approved_by=$1, approved_at=NOW() WHERE id=$2`, actorID, agentID) //nolint:errcheck
	db.DB.Exec(`INSERT INTO audit_log (actor_id, action, target_type, target_id) VALUES ($1,'approve_agent','agent',$2)`, actorID, agentID) //nolint:errcheck
	http.Redirect(w, r, "/admin/agents", http.StatusSeeOther)
}

func AdminSettings(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		var rateBPS, lockDays int
		var minSats, maxSats int64
		db.DB.QueryRow(`SELECT interest_rate_bps, lock_days, min_savings_sats, max_savings_sats FROM pool_settings WHERE id=1`).
			Scan(&rateBPS, &lockDays, &minSats, &maxSats) //nolint:errcheck
		renderTemplate(w, r, "admin/settings.html", map[string]interface{}{
			"RateBPS": rateBPS, "LockDays": lockDays, "MinSats": minSats, "MaxSats": maxSats,
		})
		return
	}
	// TODO: update pool settings
	http.Redirect(w, r, "/admin/settings", http.StatusSeeOther)
}
