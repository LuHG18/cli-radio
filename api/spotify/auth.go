package spotify

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

const (
	redirectURI = "http://localhost:8888/callback"
	authURL     = "https://accounts.spotify.com/authorize"
)

var (
	clientID     string
	clientSecret string
)

func init() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	clientID = os.Getenv("CLIENT_ID")
	clientSecret = os.Getenv("CLIENT_SECRET")

	if clientID == "" || clientSecret == "" {
		log.Fatal("Missing CLIENT_ID or CLIENT_SECRET in environment")
	}
}

func Authenticate() (string, error) {
	_, err := GetToken()
	if err == nil {
		return "Spotify already authenticated.", nil
	}

	authPage := fmt.Sprintf("%s?client_id=%s&response_type=code&redirect_uri=%s&scope=playlist-modify-public",
		authURL, clientID, url.QueryEscape(redirectURI))

	// Send this back first to tell the user what to do
	authMessage := fmt.Sprintf("Please authenticate Spotify:\n\n%s\n\nWaiting for confirmation...", authPage)

	// Open browser message will be shown first, and then we'll wait for code
	code, err := StartAuthServer()
	if err != nil {
		return "", fmt.Errorf("failed to receive auth code: %w", err)
	}

	data := url.Values{}
	data.Set("grant_type", "authorization_code")
	data.Set("code", code)
	data.Set("redirect_uri", redirectURI)
	data.Set("client_id", clientID)
	data.Set("client_secret", clientSecret)

	resp, err := http.Post(tokenURL, "application/x-www-form-urlencoded", strings.NewReader(data.Encode()))
	if err != nil {
		return "", fmt.Errorf("failed to exchange code for token: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("failed to authenticate: %s", string(body))
	}

	var token Token
	if err := json.NewDecoder(resp.Body).Decode(&token); err != nil {
		return "", fmt.Errorf("failed to parse token response: %v", err)
	}

	token.ExpiresAt = time.Now().Unix() + 3600
	if err := saveToken(&token); err != nil {
		return "", fmt.Errorf("failed to save token: %v", err)
	}

	// Create playlist if needed
	if _, err := GetPlaylist(); err != nil {
		_, _ = CreatePlaylist(&token)
		return authMessage + "\n\nAuthentication successful. Playlist created.", nil
	}

	return authMessage + "\n\nAuthentication successful.", nil
}

func StartAuthServer() (string, error) {
	authCodeChan := make(chan string)
	errChan := make(chan error)

	server := &http.Server{Addr: ":8888"}

	http.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		code := r.URL.Query().Get("code")
		if code == "" {
			errChan <- fmt.Errorf("authorization code not found")
			return
		}
		fmt.Fprintln(w, "Authentication successful! You can close this page.")
		authCodeChan <- code
	})

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errChan <- err
		}
	}()

	select {
	case code := <-authCodeChan:
		_ = server.Close()
		return code, nil
	case err := <-errChan:
		_ = server.Close()
		return "", err
	}
}
