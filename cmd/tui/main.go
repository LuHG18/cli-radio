package main

import (
	"bufio"
	"fmt"
	"log"
	"os"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	logChan := make(chan string, 100)

	r, w, _ := os.Pipe()
	log.SetOutput(w) // Redirect only logs

	// Goroutine to capture logs from the pipe
	go func() {
		scanner := bufio.NewScanner(r)
		for scanner.Scan() {
			logChan <- scanner.Text()
		}
	}()

	p := tea.NewProgram(model{
		currentStation: nil,
		currentSong:    "None",
		menu:           []string{"Play a Station", "Next Station", "Previous Station", "Add Song to Playlist", "Detect Song"},
		logLines:       []string{},
		maxLogLines:    5,
		logBuffer:      logChan,
	})
	if _, err := p.Run(); err != nil {
		fmt.Println("Error starting UI:", err)
		os.Exit(1)
	}
}
