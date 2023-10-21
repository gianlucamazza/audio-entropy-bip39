package main

import (
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"log"
	"math"
	"strings"
	"sync"
	"time"

	"github.com/gordonklaus/portaudio"
	"github.com/tyler-smith/go-bip39"
	"golang.org/x/crypto/hkdf"
)

const (
	sampleRate  = 44100
	numSeconds  = 15
	keySize     = 32
	maxBarCount = 50 // Define a constant for the maximum size of the volume bar
)

func main() {
	// Initialize portaudio.
	if err := portaudio.Initialize(); err != nil {
		log.Fatalf("Portaudio init error: %s", err)
	}
	defer portaudio.Terminate()

	// Generate 256 bits of entropy.
	entropy, err := bip39.NewEntropy(256)
	if err != nil {
		log.Fatalf("Entropy generation error: %s", err)
	}

	// Create a new HKDF reader.
	hkdfReader := hkdf.New(sha256.New, entropy, nil, nil)
	key := make([]byte, keySize)
	if _, err := hkdfReader.Read(key); err != nil {
		log.Fatalf("HKDF read error: %s", err)
	}

	fmt.Printf("Entropy: %x\nDerived key: %x\n", entropy, key)

	bufferSize := sampleRate / 10
	shortTermBuffer := make([]float32, bufferSize)

	stream, err := portaudio.OpenDefaultStream(1, 0, float64(sampleRate), bufferSize, shortTermBuffer)
	if err != nil {
		log.Fatalf("Error opening audio stream: %s", err)
	}
	defer stream.Close()

	fmt.Println("Recording. Speak into the microphone.")

	if err := stream.Start(); err != nil {
		log.Fatalf("Error starting audio stream: %s", err)
	}
	defer stream.Stop()

	var wg sync.WaitGroup
	done := make(chan struct{}) // Use a struct{} channel which doesn't occupy memory

	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			select {
			case <-done:
				return
			default:
				if err := stream.Read(); err != nil {
					log.Fatalf("Error during reading: %s", err)
				}

				volume := calculateVolume(shortTermBuffer)

				barLength := int(volume * maxBarCount)
				bar := fmt.Sprintf("\r%v", getVolumeBar(barLength))
				fmt.Print(bar)
			}
		}
	}()

	time.Sleep(numSeconds * time.Second)
	close(done) // Closing channel to signal goroutine
	wg.Wait()   // Wait for the goroutine to finish

	fmt.Println("\nRecording complete. Processing...")

	audioData := float32ToInt16Bytes(shortTermBuffer)

	audioHash := sha256.Sum256(audioData)
	combinedData := append(entropy, audioHash[:]...)
	combinedHash := sha256.Sum256(combinedData)

	fmt.Printf("Audio hash: %x\nCombined hash: %x\n", audioHash, combinedHash)

	mnemonic, err := bip39.NewMnemonic(combinedHash[:])
	if err != nil {
		log.Fatalf("Error generating mnemonic: %s", err)
	}

	fmt.Printf("Mnemonic: %s\n", mnemonic)
}

// calculateVolume calculates the volume of the audio samples.
func calculateVolume(samples []float32) float32 {
	sum := float32(0)
	for _, sample := range samples {
		sum += sample * sample
	}
	rms := math.Sqrt(float64(sum / float32(len(samples))))
	volume := float32(rms)

	// Limit the volume to a maximum of 1.0.
	if volume > 1.0 {
		volume = 1.0
	}

	return volume
}

// getVolumeBar returns a string representing the volume bar.
func getVolumeBar(length int) string {
	// Ensure the bar length is within the allowable range.
	if length > maxBarCount {
		length = maxBarCount
	} else if length < 0 { // This condition is for safety, 'length' should not be negative after our volume adjustment.
		length = 0
	}

	return fmt.Sprintf("[%s%s]",
		strings.Repeat("#", length),
		strings.Repeat(" ", maxBarCount-length)) // Filling rest of the bar with spaces for a steady bar size
}

func float32ToInt16Bytes(floats []float32) []byte {
	bytes := make([]byte, 2*len(floats))
	for i, f := range floats {
		binary.LittleEndian.PutUint16(bytes[i*2:], uint16(f*32767))
	}
	return bytes
}
