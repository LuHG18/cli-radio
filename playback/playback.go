package playback

import (
	"bufio"
	"fmt"
	"os/exec"
	"strings"
	"sync"
	"syscall"
)

var (
	currentProcess   *exec.Cmd
	CurrentSong      string
	playbackMutex    sync.Mutex
	SongUpdateChan   chan string
	StatusUpdateChan chan string
)

func PlayStation(url string, stationName string) {
	playbackMutex.Lock()
	StopPlayback()
	playbackMutex.Unlock()

	updateCurrentSong("")

	if StatusUpdateChan != nil {
		StatusUpdateChan <- fmt.Sprintf("Starting playback: %s", stationName)
	}

	// brings every station to the same volume with loudnorm filter
	// single pass, set consistent sample rate
	audioFix := "lavfi=[loudnorm=I=-16:TP=-1.5:LRA=11," + "aresample=44100]"
	currentProcess = exec.Command("mpv", "--no-video", "--af="+audioFix, url)

	stdoutPipe, err := currentProcess.StdoutPipe()
	if err != nil {
		if StatusUpdateChan != nil {
			StatusUpdateChan <- fmt.Sprintf("failed to create stdout pipe: %v", err)
		}
	}
	// Detach the process
	currentProcess.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true, // Detach from the parent process group
	}
	if err := currentProcess.Start(); err != nil {
		if StatusUpdateChan != nil {
			StatusUpdateChan <- fmt.Sprintf("Failed to play station: %v", err)
		}
		return
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
							if StatusUpdateChan != nil {
								StatusUpdateChan <- "Song unavailable"
							}
							return
						} else {
							updateCurrentSong(songInfo)
							if StatusUpdateChan != nil {
								StatusUpdateChan <- fmt.Sprintf("Now playing: %s", songInfo)
							}
						}
					}
				}
			}
		}

		if err := scanner.Err(); err != nil {
			if StatusUpdateChan != nil {
				StatusUpdateChan <- fmt.Sprintf("Error reading metadata: %v", err)
			}
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
			if StatusUpdateChan != nil {
				StatusUpdateChan <- fmt.Sprintf("Playback finished with error: %v", err)
			}
		} else {
			if StatusUpdateChan != nil {
				StatusUpdateChan <- "Playback finished."
			}
		}
	}()
}

func StopPlayback() {
	if currentProcess != nil {
		// Send a SIGKILL to the current process group
		err := syscall.Kill(-currentProcess.Process.Pid, syscall.SIGKILL)
		if err != nil {
			if StatusUpdateChan != nil {
				StatusUpdateChan <- fmt.Sprintf("Failed to stop playback: %v", err)
			}
		} else {
			if StatusUpdateChan != nil {
				StatusUpdateChan <- "Stopped current playback."
			}
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

	if SongUpdateChan != nil {
		select {
		case SongUpdateChan <- song:
		default: // don’t block if no one’s listening
		}
	}
}

func init() {
	SongUpdateChan = make(chan string)
	StatusUpdateChan = make(chan string)
}

func sendStatus(msg string) {
	if StatusUpdateChan != nil {
		StatusUpdateChan <- msg
	}
}
