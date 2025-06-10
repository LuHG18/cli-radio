package main

import (
	"cli-radio/api"
	"strings"

	// "cli-radio/playback"
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
)

type model struct {
	currentStation *api.Station
	currentSong    string
	prevStation    *api.Station
	prevFlag       bool

	statusMsg string

	err    error
	cursor int
	menu   []string
}

type statusUpdateMsg string
type songUpdateMsg string

// can be used to return a command for some initial IO
// might use this to do PlayStation -> different UI setup when app first spins up
// notice the (m model) syntax "attaches" Init() function to model structs
func (m model) Init() tea.Cmd {
	return nil
}

// return the updated model
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	// test for key press first
	case tea.KeyMsg:
		// see what was actually pressed
		switch msg.String() {
		case "up":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down":
			if m.cursor < len(m.menu)-1 {
				m.cursor++
			}
		case "enter":
			return handleUserAction(m)
		case "ctrl+c", "q":
			return m, tea.Quit
		}
	case statusUpdateMsg:
		m.statusMsg = string(msg)
	case songUpdateMsg:
		m.currentSong = string(msg)

	}
	return m, nil
}

func handleUserAction(m model) (tea.Model, tea.Cmd) {
	selected := m.menu[m.cursor]

	switch selected {
	case "Play a Station":
		return m, PlayRandomStation(&m)
	case "Next Station":
		return m, PlayNextStation(&m)
	case "Previous Station":
		return m, PlayPreviousStation(&m)
	case "Add Song to Playlist":
		return m, AddCurrentSong(&m)
	case "Detect Song":
		return m, DetectAndAddSong(&m)
	}
	return m, nil
}

func (m model) View() string {
	var b strings.Builder

	stationName := "None"
	if m.currentStation != nil {
		stationName = m.currentStation.Name
	}
	b.WriteString(fmt.Sprintf("Current Station: %v\n", stationName))

	b.WriteString(fmt.Sprintf("Current Song: %s\n\n", m.currentSong))

	// Render menu
	for i, option := range m.menu {
		cursor := " "
		if m.cursor == i {
			cursor = ">"
		}
		b.WriteString(fmt.Sprintf("%s %s\n", cursor, option))
	}

	// Optional: add spacing
	b.WriteString("\n")

	b.WriteString("\n--- Status ---\n")
	b.WriteString(fmt.Sprintf("%s\n", m.statusMsg))

	return b.String()
}

func status(text string) tea.Cmd {
	return func() tea.Msg {
		return statusUpdateMsg(text)
	}
}
