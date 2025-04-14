package api

import (
	"testing"
)

func TestGetServer(t *testing.T) {
	server, err := GetServer()
	if err != nil {
		t.Fatalf("GetServer failed: %v", err)
	}

	t.Logf("Valid server: %s", server)
}

func TestFetchStation(t *testing.T) {
	station, err := FetchStation()
	if err != nil {
		t.Fatalf("FetchStation failed: %v", err)
	}

	t.Logf("Fetched Station: Name=%s, URL=%s, Tags=%s", station.Name, station.URL, station.Tags)
}
