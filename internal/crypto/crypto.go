// crypto/crypto.go

package crypto

import (
	"crypto/sha256"
	"fmt"

	"github.com/tyler-smith/go-bip39"
	"golang.org/x/crypto/hkdf"
)

const (
	keySize = 32 // 256 bits
)

// GenerateEntropy uses the bip39 package to generate cryptographic entropy of a specified size.
func GenerateEntropy(bitSize int) ([]byte, error) {
	entropy, err := bip39.NewEntropy(bitSize)
	if err != nil {
		return nil, fmt.Errorf("entropy generation error: %w", err)
	}
	return entropy, nil
}

// DeriveKey uses the HKDF to derive a key from the entropy.
func DeriveKey(entropy []byte) ([]byte, error) {
	// Create a new HKDF reader.
	hkdfReader := hkdf.New(sha256.New, entropy, nil, nil)

	key := make([]byte, keySize)
	if _, err := hkdfReader.Read(key); err != nil {
		return nil, fmt.Errorf("HKDF read error: %w", err)
	}
	return key, nil
}

// GenerateMnemonic creates a mnemonic based on the input data (usually a hash).
func GenerateMnemonic(inputData []byte) (string, error) {
	mnemonic, err := bip39.NewMnemonic(inputData)
	if err != nil {
		return "", fmt.Errorf("error generating mnemonic: %w", err)
	}
	return mnemonic, nil
}

// HashAudioData creates a SHA-256 hash of the input data.
func HashAudioData(data []byte) [sha256.Size]byte {
	return sha256.Sum256(data)
}

// CombineAndHashData combines two byte slices and hashes the resulting data.
func CombineAndHashData(data1, data2 []byte) [sha256.Size]byte {
	combinedData := append(data1, data2...)
	return sha256.Sum256(combinedData)
}
