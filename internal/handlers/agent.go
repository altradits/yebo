package handlers

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/yebobank/yebobank/internal/db"
	"github.com/yebobank/yebobank/internal/middleware"
	"github.com/yebobank/yebobank/internal/services/rates"
	"github.com/yebobank/yebobank/internal/utils"
)

func AgentDashboard(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserID(r)
	var agentID int64
	var bizName, location, status string
	var commissionBPS int
	var floatSats int64
	db.DB.QueryRow(`
		SELECT id, business_name, location, status, commission_rate_bps, float_sats
		FROM agents WHERE user_id=$1
	`, userID).Scan(&agentID, &bizName, &location, &status, &commissionBPS, &floatSats) //nolint:errcheck

	btcKES := rates.Global.GetKES()
	renderTemplate(w, r, "agent/dashboard.html", map[string]interface{}{
		"AgentID":       agentID,
		"BusinessName":  bizName,
		"Location":      location,
		"Status":        status,
		"CommissionBPS": commissionBPS,
		"FloatSats":     floatSats,
		"FloatKES":      utils.SatsToKES(floatSats, btcKES),
	})
}

// AgentCashIn credits a customer's wallet from cash received by the agent.
func AgentCashIn(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		renderTemplate(w, r, "agent/cash_in.html", nil)
		return
	}
	agentUserID := middleware.UserID(r)
	customerPhone := utils.NormalisePhone(r.FormValue("customer_phone"))
	amountSats, err := strconv.ParseInt(r.FormValue("amount_sats"), 10, 64)
	if err != nil || amountSats <= 0 {
		renderTemplate(w, r, "agent/cash_in.html", map[string]interface{}{"Error": "Invalid amount"})
		return
	}

	// Find customer wallet
	var customerWalletID int64
	err = db.DB.QueryRow(`SELECT w.id FROM wallets w JOIN users u ON u.id=w.user_id WHERE u.phone=$1`,
		customerPhone).Scan(&customerWalletID)
	if err != nil {
		renderTemplate(w, r, "agent/cash_in.html", map[string]interface{}{"Error": "Customer not found"})
		return
	}

	var agentID int64
	var commissionBPS int
	db.DB.QueryRow(`SELECT id, commission_rate_bps FROM agents WHERE user_id=$1`, agentUserID).
		Scan(&agentID, &commissionBPS) //nolint:errcheck

	commissionSats := amountSats * int64(commissionBPS) / 10000

	tx, _ := db.DB.Begin()
	defer tx.Rollback() //nolint:errcheck

	if err := db.CreditSats(tx, customerWalletID, amountSats, "agent_cashin",
		fmt.Sprintf("agent_%d", agentID), "Agent cash-in", &agentUserID); err != nil {
		renderTemplate(w, r, "agent/cash_in.html", map[string]interface{}{"Error": err.Error()})
		return
	}

	// Credit agent commission
	var agentWalletID int64
	tx.QueryRow(`SELECT id FROM wallets WHERE user_id=$1`, agentUserID).Scan(&agentWalletID) //nolint:errcheck
	if commissionSats > 0 {
		db.CreditSats(tx, agentWalletID, commissionSats, "agent_commission", //nolint:errcheck
			"", "Cash-in commission", &agentUserID)
	}

	tx.Exec(`INSERT INTO agent_transactions (agent_id, customer_wallet, type, amount_sats, commission_sats) VALUES ($1,$2,'cash_in',$3,$4)`,
		agentID, customerWalletID, amountSats, commissionSats) //nolint:errcheck
	tx.Commit()                                                 //nolint:errcheck

	renderTemplate(w, r, "agent/cash_in.html", map[string]interface{}{
		"Success": fmt.Sprintf("Credited %s to %s", utils.FormatSats(amountSats), customerPhone),
	})
}

// AgentCashOut debits a customer's wallet when they withdraw cash from the agent.
func AgentCashOut(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		renderTemplate(w, r, "agent/cash_out.html", nil)
		return
	}
	agentUserID := middleware.UserID(r)
	customerPhone := utils.NormalisePhone(r.FormValue("customer_phone"))
	amountSats, err := strconv.ParseInt(r.FormValue("amount_sats"), 10, 64)
	if err != nil || amountSats <= 0 {
		renderTemplate(w, r, "agent/cash_out.html", map[string]interface{}{"Error": "Invalid amount"})
		return
	}

	var customerWalletID int64
	err = db.DB.QueryRow(`SELECT w.id FROM wallets w JOIN users u ON u.id=w.user_id WHERE u.phone=$1`,
		customerPhone).Scan(&customerWalletID)
	if err != nil {
		renderTemplate(w, r, "agent/cash_out.html", map[string]interface{}{"Error": "Customer not found"})
		return
	}

	var agentID int64
	db.DB.QueryRow(`SELECT id FROM agents WHERE user_id=$1`, agentUserID).Scan(&agentID) //nolint:errcheck

	tx, _ := db.DB.Begin()
	defer tx.Rollback() //nolint:errcheck

	if err := db.DebitSats(tx, customerWalletID, amountSats, "agent_cashout",
		fmt.Sprintf("agent_%d", agentID), "Agent cash-out", &agentUserID); err != nil {
		renderTemplate(w, r, "agent/cash_out.html", map[string]interface{}{"Error": err.Error()})
		return
	}
	tx.Exec(`INSERT INTO agent_transactions (agent_id, customer_wallet, type, amount_sats) VALUES ($1,$2,'cash_out',$3)`,
		agentID, customerWalletID, amountSats) //nolint:errcheck
	tx.Commit()                                //nolint:errcheck

	renderTemplate(w, r, "agent/cash_out.html", map[string]interface{}{
		"Success": fmt.Sprintf("Debited %s from %s. Hand over the cash.", utils.FormatSats(amountSats), customerPhone),
	})
}
