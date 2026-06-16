package middleware

import (
	"context"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/yebobank/yebobank/internal/db"
)

type ctxKey string

const (
	CtxUserID   ctxKey = "user_id"
	CtxUserRole ctxKey = "user_role"
	CtxWalletID ctxKey = "wallet_id"
)

// RequireAuth validates the session cookie and injects user context.
// Redirects to /login on failure.
func RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("session")
		if err != nil || cookie.Value == "" {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}
		userID, role, walletID, ok := validateSession(cookie.Value)
		if !ok {
			clearSession(w)
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}
		ctx := context.WithValue(r.Context(), CtxUserID, userID)
		ctx = context.WithValue(ctx, CtxUserRole, role)
		ctx = context.WithValue(ctx, CtxWalletID, walletID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// RequireRole returns a middleware that enforces a specific role.
func RequireRole(role string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Context().Value(CtxUserRole) != role {
				http.Error(w, "Forbidden", http.StatusForbidden)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func validateSession(token string) (userID int64, role string, walletID int64, ok bool) {
	idleHours, _ := strconv.Atoi(os.Getenv("SESSION_IDLE_TIMEOUT_HOURS"))
	if idleHours <= 0 {
		idleHours = 2
	}
	cutoff := time.Now().Add(-time.Duration(idleHours) * time.Hour)

	row := db.DB.QueryRow(`
		SELECT s.user_id, u.role, w.id
		FROM sessions s
		JOIN users u ON u.id = s.user_id
		JOIN wallets w ON w.user_id = s.user_id
		WHERE s.token = $1
		  AND s.expires_at > NOW()
		  AND s.last_seen  > $2
		  AND u.is_active  = TRUE
	`, token, cutoff)

	if err := row.Scan(&userID, &role, &walletID); err != nil {
		return 0, "", 0, false
	}
	db.DB.Exec(`UPDATE sessions SET last_seen = NOW() WHERE token = $1`, token) //nolint:errcheck
	return userID, role, walletID, true
}

func clearSession(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{Name: "session", MaxAge: -1, Path: "/"})
}

// UserID extracts the user ID from context (panics if auth middleware not applied).
func UserID(r *http.Request) int64 {
	return r.Context().Value(CtxUserID).(int64)
}

// WalletID extracts the wallet ID from context.
func WalletID(r *http.Request) int64 {
	return r.Context().Value(CtxWalletID).(int64)
}

// UserRole extracts the role from context.
func UserRole(r *http.Request) string {
	return r.Context().Value(CtxUserRole).(string)
}
