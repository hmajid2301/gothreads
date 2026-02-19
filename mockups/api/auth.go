package main

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type User struct {
	ID            int64     `json:"id"`
	Email         string    `json:"email"`
	Name          string    `json:"name"`
	AvatarURL     *string   `json:"avatar_url,omitempty"`
	OAuthProvider string    `json:"oauth_provider"`
	OAuthID       string    `json:"oauth_id"`
	IsAdmin       bool      `json:"is_admin"`
	IsApproved    bool      `json:"is_approved"`
	CreatedAt     time.Time `json:"created_at"`
}

type OAuthConfig struct {
	Enabled      bool
	Provider     string
	ClientID     string
	ClientSecret string
	RedirectURL  string
	AuthURL      string
	TokenURL     string
	UserInfoURL  string
	Scopes       []string
}

type AuthConfig struct {
	SkipAuth      bool
	JWTSecret     string
	SessionExpiry time.Duration
	OAuth         map[string]OAuthConfig
}

var (
	authConfig     AuthConfig
	oauthStates    = make(map[string]time.Time)
	oauthStatesMux sync.Mutex
)

func initAuth() {
	authConfig = AuthConfig{
		SkipAuth:      getEnv("SKIP_AUTH", "false") == "true",
		JWTSecret:     getEnv("JWT_SECRET", "gothreads-dev-secret-change-in-production"),
		SessionExpiry: 24 * 7 * time.Hour,
		OAuth:         make(map[string]OAuthConfig),
	}

	if authConfig.SkipAuth {
		log.Println("⚠️  SKIP_AUTH enabled - authentication disabled")
		return
	}

	oauthProviders := []string{"authentik", "authelia", "google", "github"}
	for _, provider := range oauthProviders {
		clientID := getEnv(fmt.Sprintf("OAUTH_%s_CLIENT_ID", strings.ToUpper(provider)), "")
		if clientID == "" {
			continue
		}

		cfg := OAuthConfig{
			Enabled:      true,
			Provider:     provider,
			ClientID:     clientID,
			ClientSecret: getEnv(fmt.Sprintf("OAUTH_%s_CLIENT_SECRET", strings.ToUpper(provider)), ""),
			RedirectURL:  getEnv(fmt.Sprintf("OAUTH_%s_REDIRECT_URL", strings.ToUpper(provider)), ""),
			AuthURL:      getEnv(fmt.Sprintf("OAUTH_%s_AUTH_URL", strings.ToUpper(provider)), getDefaultAuthURL(provider)),
			TokenURL:     getEnv(fmt.Sprintf("OAUTH_%s_TOKEN_URL", strings.ToUpper(provider)), getDefaultTokenURL(provider)),
			UserInfoURL:  getEnv(fmt.Sprintf("OAUTH_%s_USERINFO_URL", strings.ToUpper(provider)), getDefaultUserInfoURL(provider)),
			Scopes:       strings.Split(getEnv(fmt.Sprintf("OAUTH_%s_SCOPES", strings.ToUpper(provider)), getDefaultScopes(provider)), ","),
		}

		authConfig.OAuth[provider] = cfg
		log.Printf("✅ OAuth provider configured: %s", provider)
	}
}

func getDefaultAuthURL(provider string) string {
	switch provider {
	case "authentik":
		return getEnv("AUTHENTIK_URL", "http://localhost:9000") + "/application/o/authorize/"
	case "authelia":
		return getEnv("AUTHELIA_URL", "http://localhost:9091") + "/api/oidc/authorization"
	case "google":
		return "https://accounts.google.com/o/oauth2/v2/auth"
	case "github":
		return "https://github.com/login/oauth/authorize"
	}
	return ""
}

func getDefaultTokenURL(provider string) string {
	switch provider {
	case "authentik":
		return getEnv("AUTHENTIK_URL", "http://localhost:9000") + "/application/o/token/"
	case "authelia":
		return getEnv("AUTHELIA_URL", "http://localhost:9091") + "/api/oidc/token"
	case "google":
		return "https://oauth2.googleapis.com/token"
	case "github":
		return "https://github.com/login/oauth/access_token"
	}
	return ""
}

func getDefaultUserInfoURL(provider string) string {
	switch provider {
	case "authentik":
		return getEnv("AUTHENTIK_URL", "http://localhost:9000") + "/application/o/userinfo/"
	case "authelia":
		return getEnv("AUTHELIA_URL", "http://localhost:9091") + "/api/oidc/userinfo"
	case "google":
		return "https://www.googleapis.com/oauth2/v3/userinfo"
	case "github":
		return "https://api.github.com/user"
	}
	return ""
}

func getDefaultScopes(provider string) string {
	switch provider {
	case "authentik", "authelia":
		return "openid,profile,email"
	case "google":
		return "openid,profile,email"
	case "github":
		return "read:user,user:email"
	}
	return "openid,profile,email"
}

func generateState() string {
	b := make([]byte, 32)
	rand.Read(b)
	state := base64.URLEncoding.EncodeToString(b)

	oauthStatesMux.Lock()
	defer oauthStatesMux.Unlock()
	oauthStates[state] = time.Now().Add(15 * time.Minute)

	return state
}

func validateState(state string) bool {
	oauthStatesMux.Lock()
	defer oauthStatesMux.Unlock()

	expiry, exists := oauthStates[state]
	if !exists {
		return false
	}
	delete(oauthStates, state)
	return time.Now().Before(expiry)
}

func generateJWT(user *User) (string, error) {
	claims := jwt.MapClaims{
		"user_id":     user.ID,
		"email":       user.Email,
		"name":        user.Name,
		"is_admin":    user.IsAdmin,
		"is_approved": user.IsApproved,
		"exp":         time.Now().Add(authConfig.SessionExpiry).Unix(),
		"iat":         time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(authConfig.JWTSecret))
}

func validateJWT(tokenString string) (*User, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(authConfig.JWTSecret), nil
	})

	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	user := &User{
		ID:         int64(claims["user_id"].(float64)),
		Email:      claims["email"].(string),
		Name:       claims["name"].(string),
		IsAdmin:    claims["is_admin"].(bool),
		IsApproved: claims["is_approved"].(bool),
	}

	return user, nil
}

func handleOAuthLogin(w http.ResponseWriter, r *http.Request) {
	provider := r.PathValue("provider")
	cfg, exists := authConfig.OAuth[provider]
	if !exists || !cfg.Enabled {
		writeError(w, http.StatusBadRequest, "OAuth provider not configured")
		return
	}

	state := generateState()
	params := url.Values{
		"client_id":     {cfg.ClientID},
		"redirect_uri":  {cfg.RedirectURL},
		"response_type": {"code"},
		"scope":         {strings.Join(cfg.Scopes, " ")},
		"state":         {state},
	}

	http.Redirect(w, r, cfg.AuthURL+"?"+params.Encode(), http.StatusTemporaryRedirect)
}

func handleOAuthCallback(w http.ResponseWriter, r *http.Request) {
	provider := r.PathValue("provider")
	cfg, exists := authConfig.OAuth[provider]
	if !exists || !cfg.Enabled {
		writeError(w, http.StatusBadRequest, "OAuth provider not configured")
		return
	}

	state := r.URL.Query().Get("state")
	if !validateState(state) {
		writeError(w, http.StatusBadRequest, "Invalid or expired state")
		return
	}

	code := r.URL.Query().Get("code")
	if code == "" {
		writeError(w, http.StatusBadRequest, "No authorization code")
		return
	}

	token, err := exchangeCodeForToken(cfg, code)
	if err != nil {
		log.Printf("Token exchange error: %v", err)
		writeError(w, http.StatusInternalServerError, "Failed to exchange token")
		return
	}

	userInfo, err := fetchUserInfo(cfg, token)
	if err != nil {
		log.Printf("UserInfo fetch error: %v", err)
		writeError(w, http.StatusInternalServerError, "Failed to fetch user info")
		return
	}

	user, err := getOrCreateUser(provider, userInfo)
	if err != nil {
		log.Printf("User creation error: %v", err)
		writeError(w, http.StatusInternalServerError, "Failed to create user")
		return
	}

	jwtToken, err := generateJWT(user)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to generate token")
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "auth_token",
		Value:    jwtToken,
		Path:     "/",
		HttpOnly: true,
		Secure:   getEnv("SECURE_COOKIES", "false") == "true",
		SameSite: http.SameSiteLaxMode,
		Expires:  time.Now().Add(authConfig.SessionExpiry),
	})

	if !user.IsApproved {
		http.Redirect(w, r, "/pending.html", http.StatusSeeOther)
		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func exchangeCodeForToken(cfg OAuthConfig, code string) (string, error) {
	data := url.Values{
		"client_id":     {cfg.ClientID},
		"client_secret": {cfg.ClientSecret},
		"code":          {code},
		"redirect_uri":  {cfg.RedirectURL},
		"grant_type":    {"authorization_code"},
	}

	resp, err := http.PostForm(cfg.TokenURL, data)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return "", err
	}

	if accessToken, ok := result["access_token"].(string); ok {
		return accessToken, nil
	}

	return "", fmt.Errorf("no access token in response")
}

func fetchUserInfo(cfg OAuthConfig, token string) (map[string]interface{}, error) {
	req, err := http.NewRequest("GET", cfg.UserInfoURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var userInfo map[string]interface{}
	if err := json.Unmarshal(body, &userInfo); err != nil {
		return nil, err
	}

	return userInfo, nil
}

func getOrCreateUser(provider string, userInfo map[string]interface{}) (*User, error) {
	var email, name, oauthID string
	var avatarURL *string

	switch provider {
	case "github":
		if id, ok := userInfo["id"].(float64); ok {
			oauthID = fmt.Sprintf("%.0f", id)
		}
		email, _ = userInfo["email"].(string)
		name, _ = userInfo["name"].(string)
		if name == "" {
			name, _ = userInfo["login"].(string)
		}
		if avatar, ok := userInfo["avatar_url"].(string); ok {
			avatarURL = &avatar
		}
	default:
		oauthID, _ = userInfo["sub"].(string)
		email, _ = userInfo["email"].(string)
		name, _ = userInfo["name"].(string)
		if name == "" {
			name = strings.Split(email, "@")[0]
		}
		if avatar, ok := userInfo["picture"].(string); ok {
			avatarURL = &avatar
		}
	}

	if email == "" {
		email = oauthID + "@" + provider + ".oauth"
	}

	var user User
	var isFirstUser bool

	err := dbPool.QueryRow(context.Background(),
		"SELECT id, email, name, avatar_url, oauth_provider, oauth_id, is_admin, is_approved, created_at FROM users WHERE oauth_provider = $1 AND oauth_id = $2",
		provider, oauthID).Scan(&user.ID, &user.Email, &user.Name, &user.AvatarURL, &user.OAuthProvider, &user.OAuthID, &user.IsAdmin, &user.IsApproved, &user.CreatedAt)

	if err != nil {
		err = dbPool.QueryRow(context.Background(),
			"SELECT COUNT(*) = 0 FROM users").Scan(&isFirstUser)
		if err != nil {
			isFirstUser = false
		}

		err = dbPool.QueryRow(context.Background(),
			`INSERT INTO users (email, name, avatar_url, oauth_provider, oauth_id, is_admin, is_approved) 
			 VALUES ($1, $2, $3, $4, $5, $6, $7) 
			 RETURNING id, email, name, avatar_url, oauth_provider, oauth_id, is_admin, is_approved, created_at`,
			email, name, avatarURL, provider, oauthID, isFirstUser, isFirstUser).Scan(
			&user.ID, &user.Email, &user.Name, &user.AvatarURL, &user.OAuthProvider, &user.OAuthID, &user.IsAdmin, &user.IsApproved, &user.CreatedAt)

		if err != nil {
			return nil, fmt.Errorf("failed to create user: %w", err)
		}

		log.Printf("✅ New user created: %s (admin: %v, approved: %v)", email, user.IsAdmin, user.IsApproved)
	}

	return &user, nil
}

func handleLogout(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:     "auth_token",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		MaxAge:   -1,
	})
	http.Redirect(w, r, "/login.html", http.StatusSeeOther)
}

func handleGetCurrentUser(w http.ResponseWriter, r *http.Request) {
	user := getUserFromRequest(r)
	if user == nil {
		writeError(w, http.StatusUnauthorized, "Not authenticated")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

func getUserFromRequest(r *http.Request) *User {
	if authConfig.SkipAuth {
		return &User{
			ID:         1,
			Email:      "demo@gothreads.dev",
			Name:       "Demo User",
			IsAdmin:    true,
			IsApproved: true,
		}
	}

	cookie, err := r.Cookie("auth_token")
	if err != nil {
		return nil
	}

	user, err := validateJWT(cookie.Value)
	if err != nil {
		return nil
	}

	return user
}

func authMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if authConfig.SkipAuth {
			next(w, r)
			return
		}

		user := getUserFromRequest(r)
		if user == nil {
			if strings.HasPrefix(r.URL.Path, "/api/") {
				writeError(w, http.StatusUnauthorized, "Authentication required")
				return
			}
			http.Redirect(w, r, "/login.html", http.StatusSeeOther)
			return
		}

		if !user.IsApproved {
			if strings.HasPrefix(r.URL.Path, "/api/") {
				writeError(w, http.StatusForbidden, "Account pending approval")
				return
			}
			http.Redirect(w, r, "/pending.html", http.StatusSeeOther)
			return
		}

		ctx := context.WithValue(r.Context(), "user", user)
		next(w, r.WithContext(ctx))
	}
}

func adminMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return authMiddleware(func(w http.ResponseWriter, r *http.Request) {
		user := getUserFromRequest(r)
		if user == nil || !user.IsAdmin {
			writeError(w, http.StatusForbidden, "Admin access required")
			return
		}
		next(w, r)
	})
}

func getUserID(r *http.Request) int64 {
	user := getUserFromRequest(r)
	if user != nil {
		return user.ID
	}
	return 1
}
