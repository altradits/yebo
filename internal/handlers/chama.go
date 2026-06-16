package handlers

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/yebobank/yebobank/internal/db"
	"github.com/yebobank/yebobank/internal/middleware"
)

func Chama(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserID(r)
	rows, _ := db.DB.Query(`
		SELECT c.id, c.name, c.description, c.status, w.balance_sats,
		       (SELECT COUNT(*) FROM chama_members WHERE chama_id=c.id) as member_count
		FROM chamas c
		JOIN wallets w ON w.id = c.wallet_id
		JOIN chama_members cm ON cm.chama_id = c.id AND cm.user_id = $1
	`, userID)
	defer rows.Close()
	type chama struct {
		ID          int64
		Name, Desc, Status string
		BalanceSats int64
		Members     int
	}
	var chamas []chama
	for rows.Next() {
		var c chama
		rows.Scan(&c.ID, &c.Name, &c.Desc, &c.Status, &c.BalanceSats, &c.Members) //nolint:errcheck
		chamas = append(chamas, c)
	}
	renderTemplate(w, r, "customer/chama.html", map[string]interface{}{"Chamas": chamas})
}

func ChamaCreate(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		renderTemplate(w, r, "customer/chama_create.html", nil)
		return
	}
	name := r.FormValue("name")
	description := r.FormValue("description")
	maxMembers, _ := strconv.Atoi(r.FormValue("max_members"))
	if maxMembers <= 0 {
		maxMembers = 20
	}
	userID := middleware.UserID(r)

	tx, _ := db.DB.Begin()
	defer tx.Rollback() //nolint:errcheck

	// Chama wallet has no individual owner — user_id is NULL
	var walletID int64
	if err := tx.QueryRow(`INSERT INTO wallets (user_id) VALUES (NULL) RETURNING id`).Scan(&walletID); err != nil {
		renderTemplate(w, r, "customer/chama_create.html", map[string]interface{}{"Error": "Could not create chama wallet"})
		return
	}

	var chamaID int64
	if err := tx.QueryRow(`
		INSERT INTO chamas (name, description, wallet_id, created_by, max_members)
		VALUES ($1, $2, $3, $4, $5) RETURNING id
	`, name, description, walletID, userID, maxMembers).Scan(&chamaID); err != nil {
		renderTemplate(w, r, "customer/chama_create.html", map[string]interface{}{"Error": "Could not create chama"})
		return
	}

	tx.Exec(`INSERT INTO chama_members (chama_id, user_id, role) VALUES ($1, $2, 'admin')`,
		chamaID, userID) //nolint:errcheck
	tx.Commit()          //nolint:errcheck

	http.Redirect(w, r, "/chama", http.StatusSeeOther)
}

func ChamaContribute(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/chama", http.StatusSeeOther)
		return
	}
	chamaID, _ := strconv.ParseInt(r.FormValue("chama_id"), 10, 64)
	amountSats, _ := strconv.ParseInt(r.FormValue("amount_sats"), 10, 64)
	memberWalletID := middleware.WalletID(r)
	actorID := middleware.UserID(r)

	var chamaWalletID int64
	db.DB.QueryRow(`SELECT wallet_id FROM chamas WHERE id=$1`, chamaID).Scan(&chamaWalletID) //nolint:errcheck

	tx, _ := db.DB.Begin()
	defer tx.Rollback() //nolint:errcheck

	if err := db.DebitSats(tx, memberWalletID, amountSats, "chama_contribution",
		fmt.Sprintf("chama_%d", chamaID), "Chama contribution", &actorID); err != nil {
		renderTemplate(w, r, "customer/chama.html", map[string]interface{}{"Error": err.Error()})
		return
	}
	if err := db.CreditSats(tx, chamaWalletID, amountSats, "chama_contribution",
		fmt.Sprintf("chama_%d", chamaID), "Member contribution", &actorID); err != nil {
		renderTemplate(w, r, "customer/chama.html", map[string]interface{}{"Error": err.Error()})
		return
	}
	tx.Commit() //nolint:errcheck
	http.Redirect(w, r, "/chama", http.StatusSeeOther)
}
