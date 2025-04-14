package main

import (
	"cli-radio/api"
	"cli-radio/api/spotify"
	"cli-radio/playback"
	"fmt"
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
		case "p":
			station, err := api.FetchStation()
			if err != nil {
				fmt.Printf("Error fetching station: %v\n", err)
				continue
			}
			currentStation = station
			fmt.Printf("Playing: %s\n", station.Name)
			playback.PlayStation(station.URL, station.Name)
		case "n":
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
		case "prev":
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
		case "end":
			playback.StopPlayback()
			fmt.Println("Playback stopped")
		case "q":
			playback.StopPlayback()
			fmt.Println("Exiting...")
			prevStation, currentStation = nil, nil
			return
		default:
			fmt.Println("Invalid Command...")
		}
	}
}
