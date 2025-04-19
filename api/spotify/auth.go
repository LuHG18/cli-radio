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

func Authenticate() error {
	token, err := GetToken()
	if err == nil {
		// Token exists and is valid
		fmt.Println("User already authenticated.")
		return nil
	}

	fmt.Println("No valid token found. Starting authentication process.")

	// Open the Spotify authorization page
	authPage := fmt.Sprintf("%s?client_id=%s&response_type=code&redirect_uri=%s&scope=playlist-modify-public",
		authURL, clientID, url.QueryEscape(redirectURI))
	fmt.Printf("Open the following URL in your browser to authenticate:\n%s\n", authPage)

	// Start the local server to capture the auth code
	code := StartAuthServer()

	// Exchange the code for an access token
	data := url.Values{}
	data.Set("grant_type", "authorization_code")
	data.Set("code", code)
	data.Set("redirect_uri", redirectURI)
	data.Set("client_id", clientID)
	data.Set("client_secret", clientSecret)

	resp, err := http.Post(tokenURL, "application/x-www-form-urlencoded", strings.NewReader(data.Encode()))
	if err != nil {
		return fmt.Errorf("failed to exchange code for token: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to authenticate: %s", string(body))
	}

	// Parse the token response and save it
	if err := json.NewDecoder(resp.Body).Decode(&token); err != nil {
		return fmt.Errorf("failed to parse token response: %v", err)
	}

	// Tokens are valid for 1 hour
	token.ExpiresAt = time.Now().Unix() + 3600
	if err := saveToken(token); err != nil {
		return fmt.Errorf("failed to save token: %v", err)
	}
	// check if playlist was created yet
	if _, err := GetPlaylist(); err != nil {
		CreatePlaylist(token)
		fmt.Println("Authentication successful and TEMPLE playlist created")
		return nil
	}

	fmt.Println("Authentication successful ")
	return nil
}

func StartAuthServer() string {
	var authCode string

	http.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		authCode = r.URL.Query().Get("code")
		if authCode == "" {
			http.Error(w, "Authorization code not found", http.StatusBadRequest)
			return
		}

		fmt.Fprintln(w, "Authentication successful! You can close this page.")
		fmt.Printf("Authorization code received: %s\n", authCode)
	})

	// Start the server in a Goroutine
	go func() {
		err := http.ListenAndServe(":8888", nil)
		if err != nil {
			fmt.Printf("Error starting server: %v\n", err)
		}
	}()

	fmt.Println("Waiting for authorization code...")
	for authCode == "" {
		// Wait until the user logs in and the server receives the callback
	}
	return authCode
}
