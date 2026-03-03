// Package handlers implements HTTP request handlers for the Pi-Star
// dashboard API and page endpoints.
package handlers

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
)

// LoginPost is a placeholder login handler. It accepts any non-empty
// username+password, sets a session cookie, and redirects to /admin.
//
// TODO: Wire up real authentication via auth.VerifyPassword.
func LoginPost(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	username := r.FormValue("username")
	password := r.FormValue("password")

	if username == "" || password == "" {
		http.Redirect(w, r, "/login?error=1", http.StatusSeeOther)
		return
	}

	// TODO: Validate credentials against shadow/chkpwd.
	// For now, accept any non-empty username+password.

	token := generateSessionToken()
	http.SetCookie(w, &http.Cookie{
		Name:     "session",
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
	})

	http.Redirect(w, r, "/admin", http.StatusSeeOther)
}

// Logout clears the session cookie and redirects to /.
func Logout(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:     "session",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
	})
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func generateSessionToken() string {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		// Fallback — should never happen.
		return "fallback-session-token"
	}
	return hex.EncodeToString(b)
}
