package api

import (
	"context"
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/nielwyn/murmur/internal/database"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/idtoken"
)

const (
	googleStateCookie       = "murmur_oauth_state"
	googleStateTTL          = 10 * time.Minute
	createGoogleUserAttempt = 5
)

var errUnverifiedEmailCollision = errors.New("google: email exists but is not verified")

func (s *Server) googleOAuthConfig() *oauth2.Config {
	return &oauth2.Config{
		ClientID:     s.cfg.GoogleClientID,
		ClientSecret: s.cfg.GoogleClientSecret,
		RedirectURL:  s.cfg.GoogleRedirectURL,
		Scopes:       []string{"openid", "email", "profile"},
		Endpoint:     google.Endpoint,
	}
}

func (s *Server) frontendURL() string {
	if s.cfg.FrontendURL == "" {
		return "/"
	}
	return s.cfg.FrontendURL
}

// popupClose renders a page whose only job is to tell the window that
// opened it (the main app tab) how the Google sign-in attempt went, then
// close itself — this lets the popup flow finish without ever navigating
// the user's original tab.
func (s *Server) popupClose(w http.ResponseWriter, msgType, code string) {
	payload, _ := json.Marshal(map[string]string{"type": msgType, "code": code})
	targetOrigin, _ := json.Marshal(s.frontendURL())

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprintf(w, `<!doctype html>
<html><body><script>
if (window.opener) {
    window.opener.postMessage(%s, %s);
}
window.close();
</script></body></html>`, payload, targetOrigin)
}

func (s *Server) popupError(w http.ResponseWriter, code string) {
	s.popupClose(w, "google-auth-error", code)
}

func (s *Server) popupSuccess(w http.ResponseWriter) {
	s.popupClose(w, "google-auth-success", "")
}

func randomState() (string, error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(buf), nil
}

func (s *Server) handleGoogleLogin(w http.ResponseWriter, r *http.Request) {
	if s.cfg.GoogleClientID == "" || s.cfg.GoogleClientSecret == "" || s.cfg.GoogleRedirectURL == "" {
		s.popupError(w, "google_not_configured")
		return
	}

	state, err := randomState()
	if err != nil {
		s.popupError(w, "google_start_failed")
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     googleStateCookie,
		Value:    state,
		Path:     "/",
		HttpOnly: true,
		Secure:   s.cfg.Secure,
		SameSite: http.SameSiteLaxMode,
		Expires:  time.Now().Add(googleStateTTL),
	})

	http.Redirect(w, r, s.googleOAuthConfig().AuthCodeURL(state, oauth2.AccessTypeOnline), http.StatusFound)
}

func (s *Server) handleGoogleCallback(w http.ResponseWriter, r *http.Request) {
	defer http.SetCookie(w, &http.Cookie{
		Name:     googleStateCookie,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   s.cfg.Secure,
		SameSite: http.SameSiteLaxMode,
		Expires:  time.Unix(0, 0),
		MaxAge:   -1,
	})

	cookie, err := r.Cookie(googleStateCookie)
	if err != nil || subtle.ConstantTimeCompare([]byte(cookie.Value), []byte(r.URL.Query().Get("state"))) != 1 {
		s.popupError(w, "google_state_mismatch")
		return
	}

	code := r.URL.Query().Get("code")
	if code == "" {
		s.popupError(w, "google_denied")
		return
	}

	ctx := r.Context()

	token, err := s.googleOAuthConfig().Exchange(ctx, code)
	if err != nil {
		s.popupError(w, "google_exchange_failed")
		return
	}

	rawIDToken, ok := token.Extra("id_token").(string)
	if !ok || rawIDToken == "" {
		s.popupError(w, "google_no_id_token")
		return
	}

	payload, err := idtoken.Validate(ctx, rawIDToken, s.cfg.GoogleClientID)
	if err != nil {
		s.popupError(w, "google_invalid_token")
		return
	}

	googleID := payload.Subject
	email, _ := payload.Claims["email"].(string)
	emailVerified, _ := payload.Claims["email_verified"].(bool)
	name, _ := payload.Claims["name"].(string)
	if email == "" {
		s.popupError(w, "google_no_email")
		return
	}

	user, err := s.resolveGoogleUser(ctx, googleID, email, emailVerified, name)
	if err != nil {
		if errors.Is(err, errUnverifiedEmailCollision) {
			s.popupError(w, "google_email_taken")
		} else {
			s.popupError(w, "google_login_failed")
		}
		return
	}

	if err := s.issueSession(w, user); err != nil {
		s.popupError(w, "google_login_failed")
		return
	}

	s.popupSuccess(w)
}

func (s *Server) resolveGoogleUser(ctx context.Context, googleID, email string, emailVerified bool, name string) (database.User, error) {
	if u, err := s.db.GetUserByGoogleID(ctx, &googleID); err == nil {
		return u, nil
	}
	if u, err := s.db.GetUserByEmail(ctx, email); err == nil {
		if !emailVerified {
			return database.User{}, errUnverifiedEmailCollision
		}
		return s.db.LinkGoogleAccount(ctx, database.LinkGoogleAccountParams{
			GoogleID: &googleID,
			ID:       u.ID,
		})
	}
	return s.createGoogleUser(ctx, googleID, email, name)
}

func deriveUsername(email, name string) string {
	local, _, _ := strings.Cut(email, "@")
	base := sanitizeUsername(local)
	if base == "" {
		base = sanitizeUsername(name)
	}
	if base == "" {
		base = "user"
	}
	return base
}

func sanitizeUsername(s string) string {
	s = strings.ToLower(s)
	var b strings.Builder
	for _, r := range s {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '_' || r == '.' || r == '-' {
			b.WriteRune(r)
		}
	}
	out := strings.Trim(b.String(), "-_.")
	if len(out) > 30 {
		out = out[:30]
	}
	return out
}

func randomSuffix(n int) (string, error) {
	buf := make([]byte, n)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	const alphabet = "abcdefghijklmnopqrstuvwxyz0123456789"
	for i, b := range buf {
		buf[i] = alphabet[int(b)%len(alphabet)]
	}
	return string(buf), nil
}

func (s *Server) createGoogleUser(ctx context.Context, googleID, email, name string) (database.User, error) {
	base := deriveUsername(email, name)
	for i := range createGoogleUserAttempt {
		candidate := base
		if i > 0 {
			suffix, err := randomSuffix(4)
			if err != nil {
				return database.User{}, err
			}
			candidate = base + "-" + suffix
		}
		u, err := s.db.CreateGoogleUser(ctx, database.CreateGoogleUserParams{
			Username: candidate,
			Email:    email,
			GoogleID: &googleID,
		})
		if err == nil {
			return u, nil
		}
		if !isUniqueViolation(err) {
			return database.User{}, err
		}
	}
	return database.User{}, errors.New("could not allocate a unique username")
}
