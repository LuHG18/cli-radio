package playback

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
)

func HandleSignals(StopPlayback func()) {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, os.Interrupt, syscall.SIGTERM)
	go func() {
		for sig := range sigs {
			fmt.Printf("\nReceived signal: %s. Cleaning up...\n", sig)
			StopPlayback()
			os.Exit(0)
		}
	}()
}
