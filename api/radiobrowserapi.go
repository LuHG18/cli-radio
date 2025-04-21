package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"time"
)

type Station struct {
	Name string `json:"name"`
	URL  string `json:"url"`
	Tags string `json:"tags"`
}

var excludedTags = []string{"news", "news+talk", "military", "sports", "podcast", "podcasts"}
var excludedLanguages = []string{"chinese", "iranian", "mandarin"}

// Get valid servers
func GetServer() (string, error) {
	// Perform DNS lookup
	ips, err := net.LookupIP("all.api.radio-browser.info")
	if err != nil {
		return "", fmt.Errorf("DNS lookup failed: %w", err)
	}

	// Reverse DNS to get hostnames
	for _, ip := range ips {
		names, err := net.LookupAddr(ip.String())
		if err == nil && len(names) > 0 {
			name := strings.TrimSuffix(names[0], ".")
			return "https://" + name, nil
		}
	}
	return "", fmt.Errorf("no valid servers")
}

func FetchStation() (*Station, error) {
	server, err := GetServer()
	if err != nil {
		return nil, fmt.Errorf("failed to get server: %w", err)
	}

	url := buildFilterURL(server)
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("API request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	// Read the response body
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var stations []Station
	// We need to try and put the selected station into a Station struct
	if err := json.Unmarshal(bodyBytes, &stations); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if len(stations) == 0 {
		return nil, fmt.Errorf("no stations found")
	}
	return &stations[0], nil
}

func buildFilterURL(baseURL string) string {
	url := baseURL + "/json/stations/bytag/?"
	for _, tag := range excludedTags {
		url += fmt.Sprintf("tag=!%s&", tag)
	}
	for _, lang := range excludedLanguages {
		url += fmt.Sprintf("language=!%s&", lang)
	}
	// We use a unique time stamp to prevent from getting the same stations
	url += fmt.Sprintf("order=random&limit=1&nocache=%d", time.Now().UnixNano())
	return url
}
