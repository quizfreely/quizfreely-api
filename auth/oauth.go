package auth

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/url"
	"os"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"github.com/rs/zerolog/log"
	"github.com/georgysavva/scany/v2/pgxscan"
)

var googleOauthConfig *oauth2.Config

func InitOAuthGoogle() {
	/* this gets called after env vars are loaded by main() in server.go */

	googleOauthConfig = &oauth2.Config{
		ClientID:     os.Getenv("OAUTH_GOOGLE_CLIENT_ID"),
		ClientSecret: os.Getenv("OAUTH_GOOGLE_CLIENT_SECRET"),
		RedirectURL:  os.Getenv("OAUTH_GOOGLE_CALLBACK_URL"),
		Scopes: []string{
			"openid",
			"profile",
			"email",
		},
		Endpoint: google.Endpoint,
	}
}

func generateStateParam(length int) (string, error) {
	// length is number of bytes before base64 encoding
	b := make([]byte, length)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	// URL-safe base64 encoding
	return base64.URLEncoding.WithPadding(base64.NoPadding).EncodeToString(b), nil
}

func (ah *AuthHandler) OAuthGoogleRedirect(w http.ResponseWriter, r *http.Request) {
	// Generate random state
	state, err := generateStateParam(16) // 16 bytes → ~22 chars after base64
	if err != nil {
		log.Error().Err(err).Msg("Failed to generate state before Google OAuth redirect")
		redirUrl := os.Getenv("OAUTH_FINAL_REDIRECT_URL")
		redirUrl += "?error="+url.QueryEscape("Failed to generate state")
		http.Redirect(w, r, redirUrl, http.StatusTemporaryRedirect)
		return
	}

	// Store it in a secure cookie or server-side session store
	http.SetCookie(w, &http.Cookie{
		Name:     "qzfr_oauth_g_state",
		Value:    state,
		HttpOnly: true,
		Secure:   true, // only over HTTPS
		Path:     "/",
		SameSite: http.SameSiteLaxMode,
		MaxAge: 300, /* 5 mins * 60s/min = 300 sec = 5 min */
	})

	// Redirect to Google
	redirUrl := googleOauthConfig.AuthCodeURL(state, oauth2.AccessTypeOffline)
	http.Redirect(w, r, redirUrl, http.StatusTemporaryRedirect)
}

func (ah *AuthHandler) OAuthGoogleCallback(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("qzfr_oauth_g_state")
	if err != nil {
		log.Warn().Msg("Google OAuth callback missing state cookie")
		redirUrl := os.Getenv("OAUTH_FINAL_REDIRECT_URL")
		redirUrl += "?error="+url.QueryEscape("State cookie missing")
		http.Redirect(w, r, redirUrl, http.StatusTemporaryRedirect)
		return
	}

	if r.FormValue("state") != cookie.Value {
		log.Warn().Msg("Google OAuth callback invalid state")
		redirUrl := os.Getenv("OAUTH_FINAL_REDIRECT_URL")
		redirUrl += "?error="+url.QueryEscape("Invalid state")
		http.Redirect(w, r, redirUrl, http.StatusTemporaryRedirect)
		return
	}

	token, err := googleOauthConfig.Exchange(r.Context(), r.FormValue("code"))
	if err != nil {
		log.Warn().Err(err).Msg("Google OAuth code exchange failed")
		redirUrl := os.Getenv("OAUTH_FINAL_REDIRECT_URL")
		redirUrl += "?error="+url.QueryEscape("Code exchange failed")
		http.Redirect(w, r, redirUrl, http.StatusTemporaryRedirect)
		return
	}

	resp, err := http.Get("https://www.googleapis.com/oauth2/v3/userinfo?access_token=" + token.AccessToken)
	if err != nil {
		log.Warn().Err(err).Msg("Failed to get user info from googleapis.com/oauth2/v3/userinfo")
		redirUrl := os.Getenv("OAUTH_FINAL_REDIRECT_URL")
		redirUrl += "?error="+url.QueryEscape("Failed to get user info from Google")
		http.Redirect(w, r, redirUrl, http.StatusTemporaryRedirect)
		return
	}
	defer resp.Body.Close()

	var userInfo struct {
		Sub     string `json:"sub"`
		Email   string `json:"email"`
		Picture string `json:"picture"`
		Name    string `json:"name"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		log.Warn().Err(err).Msg("Failed to decode user info")
		redirUrl := os.Getenv("OAUTH_FINAL_REDIRECT_URL")
		redirUrl += "?error="+url.QueryEscape("Failed to decode user info")
		http.Redirect(w, r, redirUrl, http.StatusTemporaryRedirect)
		return
	}
	
	var qzfrUserID string
	err = pgxscan.Get(
		r.Context(),
		ah.DB,
		&qzfrUserID,
		`INSERT INTO auth.users (oauth_google_sub, auth_type, oauth_google_email, display_name)
VALUES ($1, 'OAUTH_GOOGLE', $2, $3) ON CONFLICT (oauth_google_sub) DO UPDATE
SET oauth_google_email = $3 RETURNING id`,
		userInfo.Sub,
		userInfo.Email,
		userInfo.Name,
	)
	if err != nil {
		log.Error().Err(err).Msg("Database error while adding google oauth user")
		redirUrl := os.Getenv("OAUTH_FINAL_REDIRECT_URL")
		redirUrl += "?error="+url.QueryEscape("Database error while adding user")
		http.Redirect(w, r, redirUrl, http.StatusTemporaryRedirect)
		return
	}

	var qzfrToken string
	err = pgxscan.Get(
		r.Context(),
		ah.DB,
		&qzfrToken,
		`INSERT INTO auth.sessions (user_id)
VALUES ($1) RETURNING token`,
		qzfrUserID,
	)
	if err != nil {
		log.Error().Err(err).Msg("Database error while adding session for google oauth")
		redirUrl := os.Getenv("OAUTH_FINAL_REDIRECT_URL")
		redirUrl += "?error="+url.QueryEscape("Database error while adding session")
		http.Redirect(w, r, redirUrl, http.StatusTemporaryRedirect)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:  "auth",
		Value: qzfrToken,
		Path:  "/",
		/* 10 days * 24 hours per day * 60 mins per hour * 60s per min
		= 864000 seconds = 10 days */
		MaxAge:   864000,
		Secure:   true,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
	http.Redirect(w, r, os.Getenv("OAUTH_FINAL_REDIRECT_URL"), http.StatusTemporaryRedirect)
}
