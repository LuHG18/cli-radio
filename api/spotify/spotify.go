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
	redirectURI  = "http://localhost:8888/callback"
	authURL      = "https://accounts.spotify.com/authorize"
	tokenURL     = "https://accounts.spotify.com/api/token"
	apiBaseURL   = "https://api.spotify.com/v1"
	playlistUrl  = "https://api.spotify.com/v1/me/playlists"
	searchUrl    = "https://api.spotify.com/v1/search"
	playlistName = "TEMPLE"
)

var (
	clientID     string
	clientSecret string
)

var tokenFile = "api/spotify/token.json"
var playlistFile = "api/spotify/playlist.json"

type Token struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresAt    int64  `json:"expires_at"` // Unix timestamp
}

type Playlist struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

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

// func AddToPlaylist(Playlist playlist) (string, error) {
// 	return nil
// }

func CreatePlaylist(token *Token) (string, error) {
	req, err := http.NewRequest("POST", playlistUrl, strings.NewReader(fmt.Sprintf(`{"name":"%s"}`, playlistName)))
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+token.AccessToken)
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var result Playlist
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		return "", err
	}

	data := map[string]string{"playlist_id": result.ID}
	fileData, err := json.Marshal(data)
	if err != nil {
		return "", err
	}
	err = os.WriteFile(playlistFile, fileData, 0644)
	if err != nil {
		return "", err
	}

	return result.ID, nil
}

func GetPlaylist() (*Playlist, error) {
	if _, err := os.Stat(playlistFile); os.IsNotExist(err) {
		return nil, fmt.Errorf("playlist file does not exist")
	}

	data, err := os.ReadFile(playlistFile)
	if err != nil {
		return nil, err
	}

	var playlist Playlist
	if err := json.Unmarshal(data, &playlist); err != nil {
		return nil, err
	}
	return &playlist, nil
}

func GetToken() (*Token, error) {
	// Check if the token file exists
	if _, err := os.Stat(tokenFile); os.IsNotExist(err) {
		return nil, fmt.Errorf("token file does not exist")
	}

	token, err := loadToken()
	if err != nil {
		return nil, fmt.Errorf("failed to load token: %v", err)
	}

	if time.Now().Unix() >= token.ExpiresAt {
		// Refresh the token
		token, err = refreshToken(token.RefreshToken)
		if err != nil {
			return nil, fmt.Errorf("failed to refresh token: %v", err)
		}
	}

	return token, nil
}

func loadToken() (*Token, error) {
	data, err := os.ReadFile(tokenFile)
	if err != nil {
		return nil, err
	}

	var token Token
	if err := json.Unmarshal(data, &token); err != nil {
		return nil, err
	}
	return &token, nil
}

func refreshToken(refreshToken string) (*Token, error) {
	data := url.Values{}
	data.Set("grant_type", "refresh_token")
	data.Set("refresh_token", refreshToken)
	data.Set("client_id", clientID)
	data.Set("client_secret", clientSecret)

	resp, err := http.Post(tokenURL, "application/x-www-form-urlencoded", strings.NewReader(data.Encode()))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to refresh token: %s", string(body))
	}

	var token Token
	if err := json.NewDecoder(resp.Body).Decode(&token); err != nil {
		return nil, err
	}

	token.ExpiresAt = time.Now().Unix() + 3600
	if err := saveToken(&token); err != nil {
		return nil, err
	}

	return &token, nil
}

func saveToken(token *Token) error {
	data, err := json.Marshal(token)
	if err != nil {
		return err
	}
	return os.WriteFile(tokenFile, data, 0644)
}
