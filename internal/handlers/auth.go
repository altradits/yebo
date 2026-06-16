package handlers

import (
	"database/sql"
	"net/http"
	"time"

	"github.com/yebobank/yebobank/internal/db"
	"github.com/yebobank/yebobank/internal/utils"
)

func Register(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		renderTemplate(w, r, "register.html", nil)
		return
	}
	phone := utils.NormalisePhone(r.FormValue("phone"))
	password := r.FormValue("password")
	fullName := r.FormValue("full_name")

	if err := utils.ValidatePhone(phone); err != nil {
		renderTemplate(w, r, "register.html", map[string]interface{}{"Error": err.Error()})
		return
	}
	hash, err := utils.HashPassword(password)
	if err != nil {
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}
	tx, _ := db.DB.Begin()
	defer tx.Rollback() //nolint:errcheck

	var userID int64
	err = tx.QueryRow(`
		INSERT INTO users (phone, password_hash, full_name)
		VALUES ($1, $2, $3) RETURNING id
	`, phone, hash, fullName).Scan(&userID)
	if err != nil {
		renderTemplate(w, r, "register.html", map[string]interface{}{"Error": "Phone number already registered"})
		return
	}
	if _, err := tx.Exec(`INSERT INTO wallets (user_id) VALUES ($1)`, userID); err != nil {
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}
	if err := tx.Commit(); err != nil {
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/login?registered=1", http.StatusSeeOther)
}

func Login(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		renderTemplate(w, r, "login.html", map[string]interface{}{
			"Registered": r.URL.Query().Get("registered") == "1",
		})
		return
	}
	phone := utils.NormalisePhone(r.FormValue("phone"))
	password := r.FormValue("password")

	var userID int64
	var hash, role string
	var isActive bool
	err := db.DB.QueryRow(`
		SELECT id, password_hash, role, is_active FROM users WHERE phone=$1
	`, phone).Scan(&userID, &hash, &role, &isActive)
	if err == sql.ErrNoRows || !utils.CheckPassword(password, hash) {
		renderTemplate(w, r, "login.html", map[string]interface{}{"Error": "Invalid phone or password"})
		return
	}
	if err != nil {
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}
	if !isActive {
		renderTemplate(w, r, "login.html", map[string]interface{}{"Error": "Account suspended"})
		return
	}
	token, err := utils.GenerateToken()
	if err != nil {
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}
	db.DB.Exec(`INSERT INTO sessions (token, user_id) VALUES ($1, $2)`, token, userID) //nolint:errcheck
	http.SetCookie(w, &http.Cookie{
		Name:     "session",
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		Secure:   r.TLS != nil,
		SameSite: http.SameSiteLaxMode,
		Expires:  time.Now().Add(72 * time.Hour),
	})
	switch role {
	case "admin":
		http.Redirect(w, r, "/admin", http.StatusSeeOther)
	case "agent":
		http.Redirect(w, r, "/agent", http.StatusSeeOther)
	case "trader":
		http.Redirect(w, r, "/trader", http.StatusSeeOther)
	default:
		http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
	}
}

func Logout(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("session")
	if err == nil {
		db.DB.Exec(`DELETE FROM sessions WHERE token=$1`, cookie.Value) //nolint:errcheck
	}
	http.SetCookie(w, &http.Cookie{Name: "session", MaxAge: -1, Path: "/"})
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}
