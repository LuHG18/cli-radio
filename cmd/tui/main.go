package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	p := tea.NewProgram(model{
		currentStation: "Loading...",
		currentSong:    "None",
		menu:           []string{"Next Station", "Previous Station", "Add Song to Playlist", "Detect Song"},
	})
	if _, err := p.Run(); err != nil {
		fmt.Println("Error starting UI:", err)
		os.Exit(1)
	}
}
