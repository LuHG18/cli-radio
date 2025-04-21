package spotify

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
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
	// should just call spotify API call for adding to a playlist
	// we need the spotify ID of the playlist, track uri
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
