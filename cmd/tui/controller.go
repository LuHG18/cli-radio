package main

import (
	"cli-radio/api"
	"cli-radio/api/shazam"
	"cli-radio/api/spotify"
	"cli-radio/playback"
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

func PlayRandomStation(m *model) tea.Cmd {
	station, err := api.FetchStation()
	if err != nil {
		return status(fmt.Sprintf("error fetching station: %v", err))
	}
	m.currentStation = station
	playback.PlayStation(station.URL, station.Name)
	return nil
}

func PlayNextStation(m *model) tea.Cmd {
	if m.prevFlag {
		m.currentStation, m.prevStation = m.prevStation, m.currentStation
		m.prevFlag = false
	} else {
		station, err := api.FetchStation()
		if err != nil {
			return status(fmt.Sprintf("error fetching next station: %v", err))
		}
		m.prevStation = m.currentStation
		m.currentStation = station
	}
	return playStationCmd(m.currentStation.URL, m.currentStation.Name)
}

func PlayPreviousStation(m *model) tea.Cmd {
	if m.prevStation == nil {
		return status(fmt.Sprintf("error fetching previous station"))

	} else if m.prevFlag {
		return status(fmt.Sprintf("can't go back more"))
	}
	m.prevFlag = true
	playback.PlayStation(m.prevStation.URL, m.prevStation.Name)
	return nil
}

func AddCurrentSong(m *model) tea.Cmd {
	currentSong := playback.GetCurrentSong()
	if strings.ToLower(currentSong) == "song unavailable" || strings.TrimSpace(currentSong) == "" {
		return status(fmt.Sprintf("no song to add"))

	}

	track, err := spotify.GetSongURI(currentSong)
	if err != nil {
		return status(fmt.Sprintf("get song uri error: %v", err))

	}

	if spotify.CompareSongs(currentSong, track) > (len(strings.TrimSpace(currentSong)) / 2) {
		// TODO: present confirmation bubble to user
		return nil
	}

	errr := spotify.AddToPlaylist(track.URI)
	if errr != nil {
		return status(fmt.Sprintf("add to playlist error: %v", errr))

	}

	return status(fmt.Sprintf("Added %s", currentSong))
}

func DetectAndAddSong(m *model) tea.Cmd {
	songURI, songTitle, err := shazam.DetectSong()
	if err != nil || songTitle == "" {
		return status(fmt.Sprintf("shazam detection error: %v", err))
	}

	errr := spotify.AddToPlaylist(songURI)
	if err != nil {
		return status(fmt.Sprintf("playlist add error: %v", errr))
	}

	return status(fmt.Sprintf("Added %s", songTitle))
}

func playStationCmd(url, name string) tea.Cmd {
	return func() tea.Msg {
		playback.PlayStation(url, name)
		return statusUpdateMsg(fmt.Sprintf("Playing: %s", name))
	}
}
