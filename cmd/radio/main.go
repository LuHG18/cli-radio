package main

import (
	"cli-radio/api"
	"cli-radio/api/shazam"
	"cli-radio/api/spotify"
	"cli-radio/playback"
	recogntion "cli-radio/recognition"
	"fmt"
	"strings"
)

func main() {
	fmt.Println("Welcome")

	playback.HandleSignals(playback.StopPlayback)
	var prevFlag bool = false
	var currentStation, prevStation *api.Station = nil, nil
	var command string

	spotify.Authenticate()

	for {
		fmt.Print("> ")
		_, err := fmt.Scanln(&command)
		if err != nil {
			fmt.Println("Input not valid. Try again.")
			continue
		}

		switch command {
		case "p", "play":
			station, err := api.FetchStation()
			if err != nil {
				fmt.Printf("Error fetching station: %v\n", err)
				continue
			}
			currentStation = station
			fmt.Printf("Playing: %s\n", station.Name)
			playback.PlayStation(station.URL, station.Name)
		case "n", "next":
			var newStation *api.Station
			if prevFlag {
				newStation = currentStation
				prevFlag = false
			} else {
				newStation, err = api.FetchStation()
				if err != nil {
					fmt.Printf("Error fetching station: %v\n", err)
					continue
				}
			}
			prevStation = currentStation
			currentStation = newStation
			fmt.Printf("Playing next: %s\n", newStation.Name)
			playback.PlayStation(newStation.URL, newStation.Name)
		case "pr", "prev":
			if prevStation == nil {
				fmt.Println("No previous stations")
				continue
			} else if prevFlag {
				fmt.Println("Can't go back any more")
				fmt.Printf("Still playing: %s\n", prevStation.Name)
				continue
			}
			prevFlag = true
			fmt.Printf("Playing previous: %s\n", prevStation.Name)
			playback.PlayStation(prevStation.URL, prevStation.Name)
		case "a", "add":
			currentSong := playback.GetCurrentSong()
			if strings.ToLower(currentSong) == "song unavailable" || strings.TrimSpace(currentSong) == "" {
				fmt.Println("Song not currently available. Wait for a track to play to add.")
				continue
			}
			track, err := spotify.GetSongURI(currentSong)
			if err != nil {
				fmt.Printf("Error getting song URI: %s\n", err)
			}

			if spotify.CompareSongs(currentSong, track) > (len(strings.TrimSpace(currentSong)) / 2) {
				fmt.Printf("The song we found seems to be a bit different than we expected.\nFound: %s by %s\nProceed? (y/n): ", track.Name, track.Artists[0].Name)
				var response string
				fmt.Scanln(&response)
				if strings.ToLower(response) != "y" {
					fmt.Println("Not adding song.")
					continue
				}
			}
			msg, err := spotify.AddToPlaylist(track.URI)
			if err != nil {
				fmt.Printf("Error adding to playlist: %s\n", err)
				continue
			}
			fmt.Println(msg)
		case "d":
			err := recogntion.RecordClip()
			if err != nil {
				fmt.Printf("Error in RecordClip: %s\n", err)
				continue
			}

			apiResponse, err := shazam.IdentifySong()
			if err != nil {
				fmt.Printf("Error in IdentifySong: %s\n", err)
				continue
			}

			songURI := shazam.ExtractSpotifyURI(apiResponse)
			fmt.Printf("Detected Song URI: %s\n", songURI)

		case "e", "end":
			playback.StopPlayback()
			fmt.Println("Playback stopped")
		case "q", "quit":
			playback.StopPlayback()
			fmt.Println("Exiting...")
			prevStation, currentStation = nil, nil
			return
		default:
			fmt.Println("Invalid Command...")
		}
	}
}
