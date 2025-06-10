package main

import (
	"cli-radio/playback"
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	p := tea.NewProgram(model{
		currentStation: nil,
		currentSong:    "None",
		menu:           []string{"Play a Station", "Next Station", "Previous Station", "Add Song to Playlist", "Detect Song"},
	})

	go func(p *tea.Program) {
		for song := range playback.SongUpdateChan {
			p.Send(songUpdateMsg(song))
		}
	}(p)

	go func(p *tea.Program) {
		for status := range playback.StatusUpdateChan {
			p.Send(statusUpdateMsg(status))
		}
	}(p)

	if _, err := p.Run(); err != nil {
		fmt.Println("Error starting UI:", err)
		os.Exit(1)
	}
}
