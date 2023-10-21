// audio/audio.go

package audio

import (
	"encoding/binary"
	"fmt"
	"log"
	"math"
	"strings"
	"sync"
	"time"

	"github.com/gianlucamazza/audio-entropy-bip39/internal/utils"
	"github.com/gordonklaus/portaudio"
)

const (
	sampleRate  = 44100 // 44.1 kHz
	numSeconds  = 15    // Number of seconds to record audio
	maxBarCount = 50    // Maximum size of the volume bar
)

// RecordAudio performs audio recording and returns the recorded data.
func RecordAudio() ([]byte, error) {
	// Initialize portaudio.
	if err := portaudio.Initialize(); err != nil {
		return nil, fmt.Errorf("portaudio init error: %w", err)
	}
	defer portaudio.Terminate()

	// Clean up the terminal screen.
	utils.ClearScreen()

	bufferSize := sampleRate / 10
	shortTermBuffer := make([]float32, bufferSize)

	// Open the default audio stream.
	stream, err := portaudio.OpenDefaultStream(1, 0, float64(sampleRate), bufferSize, shortTermBuffer)
	if err != nil {
		return nil, fmt.Errorf("error opening audio stream: %w", err)
	}
	defer stream.Close()

	fmt.Println("Recording. Speak into the microphone.")

	if err := stream.Start(); err != nil {
		return nil, fmt.Errorf("error starting audio stream: %w", err)
	}
	defer stream.Stop()

	var wg sync.WaitGroup
	done := make(chan struct{})

	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			select {
			case <-done:
				return
			default:
				if err := stream.Read(); err != nil {
					log.Fatalf("Error during reading: %s", err) // Consider handling this error without terminating the program
				}

				volume := calculateVolume(shortTermBuffer)
				barLength := int(volume * maxBarCount)
				bar := fmt.Sprintf("\r%v", getVolumeBar(barLength))
				fmt.Print(bar)
			}
		}
	}()

	time.Sleep(numSeconds * time.Second)
	close(done)
	wg.Wait()

	fmt.Println("\nRecording complete. Processing...")

	// Convert the float32 buffer to a byte slice.
	audioData := float32ToInt16Bytes(shortTermBuffer)

	return audioData, nil
}

// calculateVolume calculates the volume of the audio samples.
func calculateVolume(samples []float32) float32 {
	sum := float32(0)
	for _, sample := range samples {
		sum += sample * sample
	}
	rms := math.Sqrt(float64(sum / float32(len(samples))))
	volume := float32(rms)

	if volume > 1.0 {
		volume = 1.0
	}

	return volume
}

// getVolumeBar creates a visual representation of the volume level.
func getVolumeBar(length int) string {
	if length > maxBarCount {
		length = maxBarCount
	} else if length < 0 {
		length = 0
	}

	return fmt.Sprintf("[%s%s]",
		strings.Repeat("#", length),
		strings.Repeat(" ", maxBarCount-length))
}

// clearScreen clears the terminal screen.
func clearScreen() {
	fmt.Print("\033[H\033[2J")
}

// float32ToInt16Bytes converts a slice of float32s to a slice of bytes.
func float32ToInt16Bytes(floats []float32) []byte {
	bytes := make([]byte, 2*len(floats))
	for i, f := range floats {
		binary.LittleEndian.PutUint16(bytes[i*2:], uint16(f*32767))
	}
	return bytes
}
