package main

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
)

type model struct {
	currentStation string
	currentSong    string
	cursor         int
	menu           []string
}

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
			return nil, nil
		case "ctrl+c", "q":
			return m, tea.Quit
		}
	}
	return m, nil
}

func handleUserAction(m model) (tea.Model, tea.Cmd) {
	selected := m.menu[m.cursor]

	switch selected {
	case "Next Station":
	case "Previous Station":
	case "Add Song to Playlist":
	case "Detect Song":
	}
	return m, nil
}

func (m model) View() string {
	s := fmt.Sprintf("Current Station: %s\n", m.currentStation)
	s += fmt.Sprintf("Current Song: %s\n", m.currentSong)

	for i, option := range m.menu {
		cursor := " "
		if m.cursor == i {
			cursor = ">"
		}
		s += fmt.Sprintf("%s %s\n", cursor, option)
	}
	return s
}
