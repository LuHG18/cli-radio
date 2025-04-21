package playback

import (
	"bufio"
	"cli-radio/api/spotify"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"

	"os/exec"

	"strings"
	"sync"
	"syscall"
)

var (
	currentProcess *exec.Cmd
	CurrentSong    string
	playbackMutex  sync.Mutex
)

func PlayStation(url string, stationName string) {
	playbackMutex.Lock()
	StopPlayback()
	playbackMutex.Unlock()

	updateCurrentSong("")

	fmt.Printf("Starting playback: %s\n", stationName)
	currentProcess = exec.Command("mpv", "--no-video", url)

	stdoutPipe, err := currentProcess.StdoutPipe()
	if err != nil {
		log.Fatalf("Failed to create stdout pipe: %v", err)
	}
	// Detach the process
	currentProcess.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true, // Detach from the parent process group
	}
	if err := currentProcess.Start(); err != nil {
		log.Fatalf("Failed to play station: %v", err)
	}

	go func() {
		scanner := bufio.NewScanner(stdoutPipe)
		var inFileTagsSection bool // Tracks if we are in the "File tags" section
		for scanner.Scan() {
			line := scanner.Text()
			// fmt.Printf("\rmpv output: %s\n> ", line) // Debugging raw mpv output

			// Detect the start of the "File tags" section
			if strings.HasPrefix(line, "File tags:") {
				inFileTagsSection = true
				continue
			}

			// Parse metadata inside the "File tags" section
			if inFileTagsSection {
				if strings.TrimSpace(line) == "" {
					// End of "File tags" section
					inFileTagsSection = false
					continue
				}

				// Check if the line contains "icy-title"
				if strings.Contains(line, "icy-title") {
					parts := strings.SplitN(line, ": ", 2)
					if len(parts) == 2 {
						songInfo := strings.TrimSpace(parts[1])
						if songInfo == "" || songInfo == "-" {
							songInfo = "Song unavailable"
							updateCurrentSong(songInfo)
							fmt.Printf("\rNow playing: %s\n> ", songInfo)
							return
						}
						updateCurrentSong(songInfo) // Update the current song
						fmt.Printf("\rNow playing: %s\n> ", songInfo)
					}
				}
			}
		}

		if err := scanner.Err(); err != nil {
			fmt.Printf("Error reading metadata: %v\n", err)
		}
	}()

	go func() {
		err := currentProcess.Wait()
		if err != nil {
			// Suppress the "signal: killed" message when the process is intentionally stopped
			if exitError, ok := err.(*exec.ExitError); ok && exitError.ProcessState != nil && exitError.ProcessState.ExitCode() == -1 {
				// -1 indicates the process was killed (signal SIGKILL)
				return
			}
			fmt.Printf("Playback finished with error: %v\n", err)
		} else {
			fmt.Println("Playback finished.")
		}
	}()
}

func StopPlayback() {
	if currentProcess != nil {
		// Send a SIGKILL to the current process group
		err := syscall.Kill(-currentProcess.Process.Pid, syscall.SIGKILL)
		if err != nil {
			fmt.Printf("Failed to stop playback: %v\n", err)
		} else {
			fmt.Println("Stopped current playback.")
		}

		currentProcess = nil
	}
}

func GetCurrentSong() string {
	playbackMutex.Lock()
	defer playbackMutex.Unlock()
	return CurrentSong
}

func updateCurrentSong(song string) {
	playbackMutex.Lock()
	defer playbackMutex.Unlock()
	CurrentSong = song
}

type searchResponse struct {
	Tracks struct {
		Items []struct {
			URI     string `json:"uri"`
			Name    string `json:"name"`
			Artists []struct {
				Name string `json:"name"`
			} `json:"artists"`
		} `json:"items"`
	} `json:"tracks"`
}

func GetSongURI() (string, error) {
	token, err := spotify.GetToken()
	if err != nil {
		fmt.Println("Erorr from GetToken() :", err)
	}
	query := CurrentSong
	if query == "" {
		return "", fmt.Errorf("invalid song string: %q", CurrentSong)
	}

	fullUrl := fmt.Sprintf("%s?q=%s&type=track&limit=1", spotify.SearchUrl, url.QueryEscape(query))
	// build the body of the request
	req, err := http.NewRequest("GET", fullUrl, nil)
	if err != nil {
		return "", fmt.Errorf("couldn't build song request: %w", err)
	}

	// add a header to the request with our access token
	req.Header.Set("Authorization", "Bearer "+token.AccessToken)

	// make the actual request
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("search request failed: %w", err)
	}

	defer resp.Body.Close() // close the TCP connection
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("spotify search failed: %q", resp.Status)
	}

	var data searchResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil { // err stays within the scope of the conditional
		return "", fmt.Errorf("could not decode response: %w", err)
	}

	if len(data.Tracks.Items) == 0 {
		return "", fmt.Errorf("no tracks found for this search: %q", query)
	}

	return data.Tracks.Items[0].URI, nil
}
