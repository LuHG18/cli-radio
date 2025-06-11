package main

import (
	"cli-radio/api"
	"cli-radio/api/spotify"
	"cli-radio/playback"
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

	initialized bool
	initStatus  string

	err    error
	cursor int
	menu   []string
}

type statusUpdateMsg string
type songUpdateMsg string
type initCompleteMsg struct{}
type initStatusMsg string

func setInitStatus(msg string) tea.Cmd {
	return func() tea.Msg {
		return initStatusMsg(msg)
	}
}

func initStartup() tea.Cmd {
	return func() tea.Msg {
		// Try to get a token or trigger authentication
		if err := spotify.Authenticate(); err != nil {
			return initStatusMsg(fmt.Sprintf("Auth failed: %v", err))
		}

		// Setup audio
		if err := playback.SetupAudio(); err != nil {
			return initStatusMsg(fmt.Sprintf("Audio setup failed: %v", err))
		}

		return initCompleteMsg{}
	}
}

// can be used to return a command for some initial IO
// might use this to do PlayStation -> different UI setup when app first spins up
// notice the (m model) syntax "attaches" Init() function to model structs
func (m model) Init() tea.Cmd {
	return tea.Batch(
		setInitStatus("Initializing..."),
		initStartup(),
	)
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
	case initStatusMsg:
		m.initStatus = string(msg)
		return m, nil
	case initCompleteMsg:
		m.initialized = true
		m.initStatus = "All set! Use the menu below."
		return m, nil
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
	if !m.initialized {
		var b strings.Builder
		b.WriteString("Starting up...\n\n")
		b.WriteString(fmt.Sprintf("%s\n", m.initStatus))
		return b.String()
	}

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
