// audio/audio.go

package audio

import (
	"fmt"
	"log"
	"math"
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

// AudioStream is an interface that represents an audio stream.
type AudioStream interface {
	Read() (int, error)
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
	// Ensure PortAudio is initialized before creating the stream.
	if err := portaudio.Initialize(); err != nil {
		return nil, nil, fmt.Errorf("error initializing PortAudio: %w", err)
	}

	// Updated stream creation to accommodate input processing.
	input := make([]float32, bufferSize) // Buffer for incoming audio.
	stream, err := portaudio.OpenDefaultStream(1, 0, sampleRate, bufferSize, input)
	if err != nil {
		portaudio.Terminate() // Properly terminate PortAudio if there's an error.
		return nil, nil, fmt.Errorf("error opening default stream: %w", err)
	}

	// Create a cleanup function that will terminate PortAudio and close the stream.
	cleanup := func() {
		if err := stream.Close(); err != nil {
			log.Printf("Error closing the stream: %v", err)
		}
		if err := portaudio.Terminate(); err != nil {
			log.Printf("Error terminating PortAudio: %v", err)
		}
	}

	return &ConcreteAudioStream{stream: stream, buffer: input}, cleanup, nil
}

// Read reads from the audio stream.
func (cas *ConcreteAudioStream) Read() (int, error) {
	// Read from the stream into the buffer.
	err := cas.stream.Read()
	if err != nil {
		return 0, fmt.Errorf("error reading from audio stream: %w", err)
	}

	// Number of frames read.
	return len(cas.buffer), nil
}

// Close closes the audio stream.
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
	// Calculate the number of bars to display.
	vb.BarCount = int(math.Round(float64(volume * maxBarCount)))
}

// RecordAudio performs audio recording and returns the recorded data.
func RecordAudio(stream AudioStream, calculateVolumeFunc func(buffer []float32) (float32, error)) ([]byte, error) {
	bufferSize := sampleRate * numSeconds
	fullBuffer := make([]float32, 0, bufferSize) // Buffer to store complete recording.

	fmt.Println("Recording. Speak into the microphone...")

	if err := stream.Start(); err != nil {
		return nil, fmt.Errorf("error starting audio stream: %w", err)
	}
	defer stream.Stop()

	var wg sync.WaitGroup
	done := make(chan struct{})
	errChan := make(chan error, 1)

	wg.Add(1)
	go func() {
		defer wg.Done()

		for {
			select {
			case <-done:
				return
			default:
				// Read from the stream.
				n, err := stream.Read()
				if err != nil {
					errChan <- err
					return
				}

				// Append the read data to the full buffer.
				fullBuffer = append(fullBuffer, stream.(*ConcreteAudioStream).buffer[:n]...)

				// Calculate the volume and update the display bar.
				volume, err := calculateVolumeFunc(stream.(*ConcreteAudioStream).buffer[:n])
				if err != nil {
					log.Printf("Error calculating volume: %s", err)
					continue
				}

				volumeBar := NewVolumeBar()
				volumeBar.Update(volume)
				fmt.Printf("\rVolume: %s", utils.GetVolumeBar(volumeBar.BarCount))
			}
		}
	}()

	// Record for the specified duration.
	timer := time.NewTimer(time.Duration(numSeconds) * time.Second)
	<-timer.C
	close(done)
	wg.Wait()

	select {
	case err := <-errChan:
		return nil, err // Handle potential error from the recording goroutine.
	default:
		// No error, continue.
	}

	fmt.Println("\nRecording complete. Processing...")

	// Convert the recorded float32 data to bytes.
	audioData := utils.Float32ToByteSlice(fullBuffer)

	return audioData, nil
}

// CalculateVolume calculates the volume of the audio data.
func CalculateVolume(buffer []float32) (float32, error) {
	var sum float32
	for _, f := range buffer {
		sum += f * f
	}
	rms := float32(sum / float32(len(buffer)))
	volume := float32(0)
	if rms > 0 {
		volume = float32(20 * (1 + math.Log10(float64(rms))))
	}
	return volume, nil
}
