package playback

import (
	"fmt"
	"os/exec"
	"strings"
)

const MultiOutputDevice = "Blackhole+Bose"

var originalDevice string

// check that SwitchAudioSource is installed and in PATH
func checkSAS() bool {
	_, err := exec.LookPath("SwitchAudioSource")
	return err == nil
}

func getCurrentAudioDevice() (string, error) {
	currDevice, err := exec.Command("SwitchAudioSource", "-c").Output()
	if err != nil {
		return "", fmt.Errorf("could not get current device: %w", err)
	}
	return strings.TrimSpace(string(currDevice)), nil
}

func switchAudioDevice(device string) error {
	cmd := exec.Command("SwitchAudioSource", "-s", device)
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("issue when switching audio source: %w", err)
	}
	fmt.Printf("Switched to audio device: %s\n", device)
	return nil
}

// sets up the audio output to our Blackhole device
func SetupAudio() error {
	if !checkSAS() {
		return fmt.Errorf("package SwitchAudioSource not found. Please install it using 'brew install switchaudio-osx'")
	}

	device, err := getCurrentAudioDevice()
	if err != nil {
		return err
	}

	originalDevice = device
	fmt.Printf("Current audio device: %s\n", originalDevice)

	// now we want to check if the Multi-Output Device has been set up
	allDevices, err := exec.Command("SwitchAudioSource", "-a").Output() // grab all the devices
	if err != nil {
		return fmt.Errorf("audio devices could not be found: %w", err)
	}

	// check that the multi-output device exists in the list of available devices
	if !strings.Contains(string(allDevices), MultiOutputDevice) {
		return fmt.Errorf("multi output device not configured... see documentation to set up Blackhole with Bluetooth")
	}

	return switchAudioDevice(MultiOutputDevice)
}

// FIXME: might need to introduce some error handling here
func RestoreAudio() {
	if originalDevice != "" {
		switchAudioDevice(originalDevice)
		fmt.Printf("Restored to original audio device: %s\n", originalDevice)
	}
}
