package playback

import (
	"bytes"
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

func getAvailableAudioDevices() ([]string, error) {
	cmd := exec.Command("SwitchAudioSource", "-a", "-t", "output")

	var out bytes.Buffer
	cmd.Stdout = &out

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("failed to list audio devices: %w", err)
	}

	lines := strings.Split(out.String(), "\n")
	var devices []string
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed != "" {
			devices = append(devices, trimmed)
		}
	}

	return devices, nil
}

func switchAudioDevice(device string) error {
	cmd := exec.Command("SwitchAudioSource", "-s", device)
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("issue when switching audio source: %w", err)
	}
	// fmt.Printf("Switched to audio device: %s\n", device)
	return nil
}

func IsBluetoothDevice(device string) bool {
	keywords := []string{"Headphones", "Bose", "Sony", "AirPods", "Bluetooth", "BT"}
	for _, keyword := range keywords {
		if strings.Contains(strings.ToLower(device), strings.ToLower(keyword)) {
			return true
		}
	}
	return false
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

	if IsBluetoothDevice(originalDevice) {
		// now we want to check if the Multi-Output Device has been set up
		allDevices, err := exec.Command("SwitchAudioSource", "-a").Output() // grab all the devices
		if err != nil {
			return fmt.Errorf("audio devices could not be found: %w", err)
		}

		// check that the multi-output device exists in the list of available devices
		if !strings.Contains(string(allDevices), MultiOutputDevice) {
			return fmt.Errorf("multi output device not configured... see documentation to set up Blackhole with Bluetooth")
		}
		switchAudioDevice(MultiOutputDevice)
	}

	return nil
}

// contains checks whether the target string exists in the list of strings.
func contains(list []string, target string) bool {
	for _, item := range list {
		if item == target {
			return true
		}
	}
	return false
}

func RestoreAudio() error {
	current, _ := getCurrentAudioDevice()

	if strings.Contains(current, "Blackhole") || strings.Contains(current, "Multi") {
		available, _ := getAvailableAudioDevices()

		// if the headphones are still availble (meaning we didn't turn them off during playback)
		if contains(available, originalDevice) {
			if err := switchAudioDevice(originalDevice); err != nil {
				return err
			}
			// we want to remove the Blackhole output from the stack for when we turn off headphones at some point
			// FIXME: need to make this more robust for other users (not hardcoded)
			switchAudioDevice("MacBook Pro Speakers")
			switchAudioDevice(originalDevice)
		} else {
			// 3. Fallback to MacBook speakers
			switchAudioDevice("MacBook Pro Speakers")
		}
	}

	return nil
}
