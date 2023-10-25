// audio/audio.go

package audio

import (
	"errors"
	"fmt"
	"github.com/gianlucamazza/audio-entropy-bip39/internal/utils"
	"github.com/gordonklaus/portaudio"
	"log"
	"math"
	"strings"
	"sync"
	"time"
)

const (
	sampleRate  = 44100 // 44.1 kHz
	numSeconds  = 15    // Number of seconds to record audio
	maxBarCount = 50    // Maximum size of the volume bar
)

// AudioStream is an interface that represents an audio stream.
type AudioStream interface {
	Read() error
	Start() error
	Stop() error
	Close() error
}

// ConcreteAudioStream is a concrete implementation of the AudioStream interface.
type ConcreteAudioStream struct {
	stream *portaudio.Stream
	buffer []float32
}

// NewConcreteAudioStream creates a new ConcreteAudioStream.
func NewConcreteAudioStream(bufferSize int) (*ConcreteAudioStream, func(), error) {
	// Initialize PortAudio once during the program lifecycle.
	err := portaudio.Initialize()
	if err != nil {
		return nil, nil, fmt.Errorf("error initializing PortAudio: %w", err)
	}

	// Buffer for incoming audio.
	input := make([]float32, bufferSize)

	// Updated stream creation to accommodate input processing.
	stream, err := portaudio.OpenDefaultStream(1, 0, sampleRate, bufferSize, &input)
	if err != nil {
		portaudio.Terminate() // It's important to terminate after a failed initialization.
		return nil, nil, fmt.Errorf("error opening default stream: %w", err)
	}

	// Create a cleanup function.
	cleanup := func() {
		err := stream.Close()
		if err != nil {
			log.Printf("Error closing the stream: %v", err)
		}
		err = portaudio.Terminate()
		if err != nil {
			log.Printf("Error terminating PortAudio: %v", err)
		}
	}

	return &ConcreteAudioStream{stream: stream, buffer: input}, cleanup, nil
}

// Read from the audio stream into the buffer.
// Read fills the buffer with audio data.
func (cas *ConcreteAudioStream) Read() error {
	err := cas.stream.Read()
	if err != nil {
		if err != portaudio.InputOverflowed {
			return fmt.Errorf("error reading from audio stream: %w", err)
		}
		log.Printf("Input overflow occurred: %v", err)
	}
	return nil
}

// Close the audio stream.
func (cas *ConcreteAudioStream) Close() error {
	if cas.stream != nil {
		err := cas.stream.Close()
		if err != nil {
			return fmt.Errorf("failed to close audio stream: %w", err)
		}
	}
	return nil
}

// Start starts the audio stream.
func (cas *ConcreteAudioStream) Start() error {
	return cas.stream.Start()
}

// Stop stops the audio stream.
func (cas *ConcreteAudioStream) Stop() error {
	return cas.stream.Stop()
}

// VolumeBar represents a volume bar.
type VolumeBar struct {
	BarCount int
}

// NewVolumeBar creates a new VolumeBar.
func NewVolumeBar() *VolumeBar {
	return &VolumeBar{BarCount: maxBarCount}
}

// Update updates the volume bar.
func (vb *VolumeBar) Update(volume float32) {
	const maxVolume = 1.0
	volume = volume / maxVolume

	vb.BarCount = int(volume * float32(maxBarCount))

	if vb.BarCount < 0 {
		vb.BarCount = 0
	} else if vb.BarCount > maxBarCount {
		vb.BarCount = maxBarCount
	}
}

// Draw draws the volume bar.
func (vb *VolumeBar) Draw() string {
	bar := strings.Repeat("#", vb.BarCount)
	return fmt.Sprintf("[%s%s]", bar, strings.Repeat(" ", maxBarCount-vb.BarCount))
}

// RecordAudio performs audio recording and returns the recorded data.
func RecordAudio(stream AudioStream, calculateVolumeFunc func(buffer []float32) (float32, error)) ([]byte, error) {
	bufferSize := sampleRate * numSeconds
	fullBuffer := make([]float32, 0, bufferSize)

	fmt.Println("Recording. Speak into the microphone...")

	// Start the audio stream.
	if err := stream.Start(); err != nil {
		return nil, fmt.Errorf("error starting audio stream: %w", err)
	}
	defer func() {
		err := stream.Stop() // Ensure the stream is stopped.
		if err != nil {
			log.Printf("Error stopping audio stream: %v", err)
		}
	}()

	var wg sync.WaitGroup
	done := make(chan bool)
	errChan := make(chan error)

	// Recording routine.
	wg.Add(1)
	go func() {
		defer wg.Done()

		fmt.Println("Press Ctrl-C to stop recording...")
		for {
			select {
			case <-done:
				return
			default:
				// Read from the audio stream.
				err := stream.Read()
				if err != nil {
					errChan <- fmt.Errorf("error reading from audio stream: %w", err)
					return
				}

				// Calculate the volume.
				volume, err := calculateVolumeFunc(stream.(*ConcreteAudioStream).buffer)
				if err != nil {
					errChan <- fmt.Errorf("error calculating volume: %w", err)
					return
				}

				fmt.Printf("\rVolume: %f", volume)

				// Update the volume bar.
				volumeBar := NewVolumeBar()
				volumeBar.Update(volume)

				// Draw the volume bar.
				fmt.Printf("\r%s", volumeBar.Draw())
			}
		}
	}()

	// Wait for the recording to complete.
	timer := time.NewTimer(time.Duration(numSeconds) * time.Second)
	<-timer.C
	close(done)
	wg.Wait()

	// Check for any errors that occurred during recording.
	select {
	case err := <-errChan:
		return nil, err
	default:
		// No errors.
	}

	fmt.Println("\nRecording complete. Processing...")

	// Convert the audio buffer to bytes.
	audioData := utils.Float32ToByteSlice(fullBuffer)

	return audioData, nil
}

// ErrInvalidBuffer indicates an operation on an invalid buffer.
var ErrInvalidBuffer = errors.New("invalid buffer")

// CalculateVolume calculates the volume of the audio data in decibels.
func CalculateVolume(buffer []float32) (float32, error) {
	if len(buffer) == 0 {
		return 0, ErrInvalidBuffer
	}

	var sumSquares float64
	for _, sample := range buffer {
		sumSquares += float64(sample) * float64(sample) // Squaring each sample.
	}

	// Calculate the mean of the squares.
	meanSquare := sumSquares / float64(len(buffer))

	// Calculate the root of the mean square, i.e., RMS.
	rms := math.Sqrt(meanSquare)

	// Normalizing the volume so that it fits in a 0-1 range for visualization.
	// This approach avoids using dB and keeps the volume in a linear scale.
	normalizedVolume := float32(rms)

	return normalizedVolume, nil
}
