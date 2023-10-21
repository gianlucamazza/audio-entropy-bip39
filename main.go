package main

import (
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"log"
	"math"
	"time"

	"github.com/gordonklaus/portaudio"
	"github.com/tyler-smith/go-bip39"
	"golang.org/x/crypto/hkdf"
)

const (
	sampleRate = 44100
	numSeconds = 5
	keySize    = 32
)

func main() {
	// Initialize portaudio.
	err := portaudio.Initialize()
	if err != nil {
		log.Fatal(err)
	}
	defer portaudio.Terminate()

	// Generate 256 bits of entropy.
	entropy, err := bip39.NewEntropy(256)
	if err != nil {
		log.Fatal(err)
	}

	// Create a new HKDF reader.
	hkdfReader := hkdf.New(sha256.New, entropy, nil, nil)
	key := make([]byte, keySize)
	_, err = hkdfReader.Read(key)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Entropy: %x\nDerived key: %x\n", entropy, key)

	in := make([]float32, numSeconds*sampleRate)
	stream, err := portaudio.OpenDefaultStream(1, 0, float64(sampleRate), len(in), in)
	if err != nil {
		log.Fatal(err)
	}
	defer stream.Close()

	fmt.Println("Recording. Speak into the microphone.")

	err = stream.Start()
	if err != nil {
		log.Fatal(err)
	}
	defer stream.Stop() // Ensure the stream is stopped.

	done := make(chan bool)
	go func() {
		time.Sleep(numSeconds * time.Second)
		done <- true
	}()

	<-done // Wait for recording to finish.

	fmt.Println("Recording complete. Processing...")

	audioData := make([]byte, 4*len(in))
	for i, sample := range in {
		binary.LittleEndian.PutUint32(audioData[i*4:], math.Float32bits(sample))
	}

	audioHash := sha256.Sum256(audioData)
	combinedData := append(entropy, audioHash[:]...)
	combinedHash := sha256.Sum256(combinedData)

	fmt.Printf("Audio hash: %x\nCombined hash: %x\n", audioHash, combinedHash)

	// Generate mnemonic from the combined hash.
	mnemonic, err := bip39.NewMnemonic(combinedHash[:])
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Mnemonic: %s\n", mnemonic)
}
