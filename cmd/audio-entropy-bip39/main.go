package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/gianlucamazza/audio-entropy-bip39/internal/audio"
	"github.com/gianlucamazza/audio-entropy-bip39/internal/crypto"
	"github.com/gianlucamazza/audio-entropy-bip39/internal/utils"
)

const (
	savedAudioDataFilename = "audio-data.wav"
	savedMnemonicFilename  = "mnemonic.txt"
	debug                  = false
	buffersize             = 512
)

func main() {
	// Set the debug flag.
	var debugMode bool
	flag.BoolVar(&debugMode, "debug", debug, "Enable debug mode")
	flag.Parse()
	fmt.Printf("Debug mode is: %t\n", debugMode) // Aggiungi questa riga

	// Set the debug print function.
	debugPrint := func(format string, args ...interface{}) {
		if !debugMode {
			fmt.Printf(format, args...)
		}
	}

	// Initialize the audio stream.
	stream, cleanup, err := audio.NewConcreteAudioStream(buffersize)
	if err != nil {
		log.Fatalf("Error creating audio stream: %v", err)
	}

	defer cleanup()

	// Clear the screen before starting the audio recording if debug mode is enabled.
	if !debugMode {
		utils.ClearScreen()
	}

	fmt.Println("Starting audio recording...")
	audioData, err := audio.RecordAudio(stream, audio.CalculateVolume)
	if err != nil {
		log.Fatalf("Error recording audio: %v", err)
	}

	// Clear the screen after stopping the audio recording if debug mode is enabled.
	if !debugMode {
		utils.ClearScreen()
	}

	debugPrint("Generating cryptographic entropy...\n")
	entropy, err := crypto.GenerateEntropy(256) // Assuming 256 bits for strong security.
	if err != nil {
		log.Fatalf("Error generating entropy: %v", err)
	}

	debugPrint("Deriving cryptographic key...\n")
	key, err := crypto.DeriveKey(entropy)
	if err != nil {
		log.Fatalf("Error deriving key: %v", err)
	}

	// Print the generated entropy and derived key in hexadecimal.
	debugPrint("Entropy: %x\n", entropy)
	debugPrint("Key: %x\n", key)

	debugPrint("Hashing recorded audio data...\n")
	audioHash := crypto.HashAudioData(audioData)

	debugPrint("Combining entropy with audio data hash and re-hashing...\n")
	combinedDataHash := crypto.CombineAndHashData(entropy, audioHash[:])

	debugPrint("Generating BIP-39 mnemonic from combined data hash...\n")
	mnemonic, err := crypto.GenerateMnemonic(combinedDataHash[:])
	if err != nil {
		log.Fatalf("Error generating mnemonic: %v", err)
	}

	// Display the generated mnemonic.
	fmt.Printf("Mnemonic: %s\n", mnemonic)

	// Save audio data to file
	fmt.Println("Saving audio data to file...")
	if err := utils.SaveAudioDataToFile(savedAudioDataFilename, audioData); err != nil {
		log.Fatalf("Error saving audio data to file: %v", err)
	}

	// Save mnemonic to file.
	fmt.Println("Saving mnemonic to file...")
	if err := utils.SaveMnemonicToFile(savedMnemonicFilename, mnemonic); err != nil {
		log.Fatalf("Error saving mnemonic to file: %v", err)
	}

}
