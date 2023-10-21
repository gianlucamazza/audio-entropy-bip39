package main

import (
	"fmt"
	"log"

	"github.com/gianlucamazza/audio-entropy-bip39/internal/audio"
	"github.com/gianlucamazza/audio-entropy-bip39/internal/crypto"
)

func main() {
	// Initialize portaudio.
	audioData, err := audio.RecordAudio()
	if err != nil {
		log.Fatalf("Error recording audio: %s", err)
	}

	// Generate 256 bits of entropy.
	entropy, err := crypto.GenerateEntropy(256)
	if err != nil {
		log.Fatalf("Error: %s", err)
	}

	// Derive a key from the entropy.
	key, err := crypto.DeriveKey(entropy)
	if err != nil {
		log.Fatalf("Error: %s", err)
	}

	fmt.Printf("Entropy: %x\nDerived key: %x\n", entropy, key)

	// Hash the audio data.
	audioHash := crypto.HashAudioData(audioData)

	// Combine the entropy and the audio hash.
	combinedDataHash := crypto.CombineAndHashData(entropy, audioHash[:])

	// Generate a mnemonic from the combined data hash.
	mnemonic, err := crypto.GenerateMnemonic(combinedDataHash[:])
	if err != nil {
		log.Fatalf("Error: %s", err)
	}

	fmt.Printf("Mnemonic: %s\n", mnemonic)
}
