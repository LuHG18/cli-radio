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
		m.err = fmt.Errorf("error fetching station: %w", err)
		return nil
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
			m.err = fmt.Errorf("error fetching next station: %w", err)
			return nil
		}
		m.prevStation = m.currentStation
		m.currentStation = station
	}
	playback.PlayStation(m.currentStation.URL, m.currentStation.Name)
	return nil
}

func PlayPreviousStation(m *model) tea.Cmd {
	if m.prevStation == nil {
		m.err = fmt.Errorf("no previous station")
		return nil
	} else if m.prevFlag {
		m.err = fmt.Errorf("can't go back more")
		return nil
	}
	m.prevFlag = true
	playback.PlayStation(m.prevStation.URL, m.prevStation.Name)
	return nil
}

func AddCurrentSong(m *model) tea.Cmd {
	currentSong := playback.GetCurrentSong()
	if strings.ToLower(currentSong) == "song unavailable" || strings.TrimSpace(currentSong) == "" {
		m.err = fmt.Errorf("no song to add")
		return nil
	}

	track, err := spotify.GetSongURI(currentSong)
	if err != nil {
		m.err = fmt.Errorf("get song uri error: %w", err)
		return nil
	}

	if spotify.CompareSongs(currentSong, track) > (len(strings.TrimSpace(currentSong)) / 2) {
		// TODO: present confirmation bubble to user
		return nil
	}

	msg, err := spotify.AddToPlaylist(track.URI)
	if err != nil {
		m.err = fmt.Errorf("add to playlist error: %w", err)
		return nil
	}

	m.err = fmt.Errorf(msg) // show as status
	return nil
}

func DetectAndAddSong(m *model) tea.Cmd {
	songURI, songTitle, err := shazam.DetectSong()
	if err != nil || songTitle == "" {
		m.err = fmt.Errorf("shazam detection error: %w", err)
		return nil
	}

	msg, err := spotify.AddToPlaylist(songURI)
	if err != nil {
		m.err = fmt.Errorf("playlist add error: %w", err)
		return nil
	} else {
		fmt.Println(msg)
	}

	m.err = fmt.Errorf("Added %s", songTitle)
	return nil
}
