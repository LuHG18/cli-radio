package shazam

import (
	"bytes"
	recogntion "cli-radio/recognition"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
)

const (
	ShazamAPI    = "https://shazam.p.rapidapi.com/songs/v2/detect"
	rapidAPIHost = "shazam.p.rapidapi.com"
)

var (
	rapidAPIKey string
)

type ShazamResponse struct {
	Track struct {
		Title    string `json:"title"`
		Subtitle string `json:"subtitle"`
		Hub      struct {
			Providers []struct {
				Type    string `json:"type"`
				Actions []struct {
					URI string `json:"uri"`
				} `json:"actions"`
			} `json:"providers"`
		} `json:"hub"`
	} `json:"track"`
}

func init() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	rapidAPIKey = os.Getenv("RAPID_API_KEY")

	if rapidAPIKey == "" {
		log.Fatal("Missing Rapid API Key in environment")
	}
}

func IdentifySong() (*ShazamResponse, error) {
	// Read the converted audio file
	file, err := os.ReadFile(recogntion.OutputFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	// Encode to base64
	encoded := base64.StdEncoding.EncodeToString(file)

	// Prepare the request body
	reqBody := []byte(encoded)

	// Send the POST request
	req, err := http.NewRequest("POST", ShazamAPI, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("content-type", "text/plain")
	req.Header.Set("x-rapidapi-key", rapidAPIKey)
	req.Header.Set("x-rapidapi-host", "shazam.p.rapidapi.com")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("non-200 response: %s\n%s", resp.Status, string(body))
	}

	var result ShazamResponse
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &result, nil
}

func ExtractSpotifyURI(response *ShazamResponse) string {
	for _, provider := range response.Track.Hub.Providers {
		if provider.Type == "SPOTIFY" {
			for _, action := range provider.Actions {
				if action.URI != "" {
					return action.URI
				}
			}
		}
	}
	return ""
}
