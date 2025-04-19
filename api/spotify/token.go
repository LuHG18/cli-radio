package spotify

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

const (
	tokenURL = "https://accounts.spotify.com/api/token"
)

var tokenFile = "api/spotify/token.json"

type Token struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresAt    int64  `json:"expires_at"` // Unix timestamp
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
