package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

// OAuth2 / OIDC (Authorization Code Flow) — cara "Login with Google/GitHub":
//
//  1. User diarahkan ke provider -> login -> provider balik dengan CODE.
//  2. Backend menukar CODE + client_secret -> ACCESS TOKEN (langkah rahasia).
//  3. Backend memakai token untuk mengambil profil user (/userinfo).
//
// Di sini langkah 2 & 3 (sisi backend) diimplementasikan; provider di-mock.

// --- Sisi CLIENT (aplikasimu) ---

type tokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
}

// ExchangeCode menukar authorization code dengan access token (server-to-server).
func ExchangeCode(tokenURL, clientID, clientSecret, code, redirectURI string) (string, error) {
	form := url.Values{
		"grant_type":    {"authorization_code"},
		"code":          {code},
		"client_id":     {clientID},
		"client_secret": {clientSecret},
		"redirect_uri":  {redirectURI},
	}
	resp, err := http.PostForm(tokenURL, form)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("token exchange gagal (%d): %s", resp.StatusCode, b)
	}
	var tr tokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tr); err != nil {
		return "", err
	}
	return tr.AccessToken, nil
}

// FetchUserInfo mengambil profil user memakai access token.
func FetchUserInfo(userinfoURL, accessToken string) (map[string]any, error) {
	req, _ := http.NewRequest(http.MethodGet, userinfoURL, nil)
	req.Header.Set("Authorization", "Bearer "+accessToken)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("userinfo gagal: %d", resp.StatusCode)
	}
	var info map[string]any
	return info, json.NewDecoder(resp.Body).Decode(&info)
}

// --- Sisi PROVIDER (di-mock untuk demo/test) ---

type MockProvider struct {
	clientID, clientSecret string
	validCode              string
	accessToken            string
	userInfo               map[string]any
}

func NewMockProvider() *MockProvider {
	return &MockProvider{
		clientID:     "my-app",
		clientSecret: "s3cret",
		validCode:    "auth-code-xyz",
		accessToken:  "access-token-abc",
		userInfo:     map[string]any{"sub": "user-123", "email": "ana@mail.id", "name": "Ana"},
	}
}

func (p *MockProvider) Handler() http.Handler {
	mux := http.NewServeMux()

	// Tukar code -> token (cek client secret & code).
	mux.HandleFunc("POST /token", func(w http.ResponseWriter, r *http.Request) {
		_ = r.ParseForm()
		if r.FormValue("code") != p.validCode ||
			r.FormValue("client_id") != p.clientID ||
			r.FormValue("client_secret") != p.clientSecret {
			http.Error(w, "invalid_grant", http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(tokenResponse{AccessToken: p.accessToken, TokenType: "Bearer", ExpiresIn: 3600})
	})

	// Profil user (cek Bearer token).
	mux.HandleFunc("GET /userinfo", func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer "+p.accessToken {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(p.userInfo)
	})

	return mux
}
