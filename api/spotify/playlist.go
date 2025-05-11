package spotify

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/lithammer/fuzzysearch/fuzzy"
)

const (
	apiBaseURL       = "https://api.spotify.com/v1"
	playlistUrl      = "https://api.spotify.com/v1/me/playlists"
	addToPlaylistUrl = "https://api.spotify.com/v1/playlists/%s/tracks"
	SearchUrl        = "https://api.spotify.com/v1/search"
	playlistName     = "TEMPLE"
)

var playlistFile = "api/spotify/playlist.json"

type Playlist struct {
	ID   string `json:"playlist_id"`
	Name string `json:"name"`
}

func AddToPlaylist(songUri string) (string, error) {
	token, err := GetToken()
	if err != nil {
		fmt.Println("Erorr from GetToken() :", err)
	}

	playlist, err := GetPlaylist()
	if err != nil {
		return "", err
	}
	fullUrl := fmt.Sprintf(addToPlaylistUrl, playlist.ID)

	payload := map[string][]string{
		"uris": {songUri},
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest("POST", fullUrl, bytes.NewBuffer(body))
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
	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("failed to add track: %s — %s", resp.Status, string(respBody))
	}

	return "Song Added", nil
}

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

type Track struct {
	URI     string `json:"uri"`
	Name    string `json:"name"`
	Artists []struct {
		Name string `json:"name"`
	} `json:"artists"`
}

type searchResponse struct {
	Tracks struct {
		Items []Track `json:"items"`
	} `json:"tracks"`
}

func GetSongURI(song string) (*Track, error) {
	token, err := GetToken()
	if err != nil {
		fmt.Println("Erorr from GetToken() :", err)
	}
	query := song
	if query == "" {
		return nil, fmt.Errorf("invalid song string: %q", song)
	}

	fullUrl := fmt.Sprintf("%s?q=%s&type=track&limit=1", SearchUrl, url.QueryEscape(query))
	// build the body of the request
	req, err := http.NewRequest("GET", fullUrl, nil)
	if err != nil {
		return nil, fmt.Errorf("couldn't build song request: %w", err)
	}

	// add a header to the request with our access token
	req.Header.Set("Authorization", "Bearer "+token.AccessToken)

	// make the actual request
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("search request failed: %w", err)
	}

	defer resp.Body.Close() // close the TCP connection
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("spotify search failed: %q", resp.Status)
	}

	var data searchResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil { // err stays within the scope of the conditional
		return nil, fmt.Errorf("could not decode response: %w", err)
	}

	if len(data.Tracks.Items) == 0 {
		return nil, fmt.Errorf("no tracks found for this search: %q", query)
	}

	return &data.Tracks.Items[0], nil
}

func CompareSongs(currentSong string, track *Track) int {
	rawInput := strings.ToLower(strings.TrimSpace(currentSong))

	trackArtist := strings.ToLower(strings.TrimSpace(track.Artists[0].Name))
	trackName := strings.ToLower(strings.TrimSpace(track.Name))

	firstOrder := trackArtist + " - " + trackName
	secondOrder := trackName + " - " + trackArtist

	score1 := fuzzy.LevenshteinDistance(firstOrder, rawInput)
	score2 := fuzzy.LevenshteinDistance(secondOrder, rawInput)

	if score1 < score2 {
		return score1
	}

	return score2
}
