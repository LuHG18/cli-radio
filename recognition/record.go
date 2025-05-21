package recognition

import (
	"fmt"
	"os/exec"
)

const OutputFile = "recognition/clip.raw"

// const ConvertedFile = "recognition/converted.dat"

func RecordClip() error {
	cmd := exec.Command(
		"ffmpeg",
		"-y",
		"-f", "avfoundation",
		"-i", ":1",
		"-t", "7",
		"-filter:a", "volume=7.0",
		"-ch_layout", "mono",
		"-ar", "44100",
		"-acodec", "pcm_s16le",
		"-f", "s16le",
		OutputFile,
	)

	fmt.Println("Recording audio...")
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("recording failed: %w", err)
	}
	fmt.Println("Recording complete.")
	return nil
}
