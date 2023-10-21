// utils/utils.go

package utils

import (
	"encoding/binary"
	"fmt"
	"os"
	"strings"
)

// ClearScreen clears the terminal screen.
func ClearScreen() {
	fmt.Print("\033[H\033[2J")
}

var Debug bool

const (
	volumeBarStart  = "["
	volumeIndicator = "#"
	volumeBarEnd    = "..............................................................]"
	maxBarCount     = len(volumeBarEnd) - 1 // adjusted to account for the starting bracket
)

// GetVolumeBar creates a visual representation of the volume level.
func GetVolumeBar(volume int) string {
	if volume > maxBarCount {
		volume = maxBarCount
	} else if volume < 0 {
		volume = 0
	}

	// This part fills up the volume bar to the current volume with the volumeIndicator character.
	filledPart := strings.Repeat(volumeIndicator, volume)

	// This part creates the empty remainder of the volume bar.
	emptyPartLength := maxBarCount - volume
	emptyPart := ""
	if emptyPartLength > 0 {
		emptyPart = volumeBarEnd[:emptyPartLength]
	}

	// Create the volume bar.
	volumeBar := volumeBarStart + filledPart + emptyPart

	return volumeBar
}

// wavHeaderData is a struct that represents the necessary fields in a WAV file header.
type wavHeaderData struct {
	// Here we define the fields required for a WAV header.
	// Each field must be written in little endian format, and the total header size should be 44 bytes.
	ChunkID       [4]byte // "RIFF"
	ChunkSize     uint32  // 4 + (8 + SubChunk1Size) + (8 + SubChunk2Size)
	Format        [4]byte // "WAVE"
	SubChunk1ID   [4]byte // "fmt "
	SubChunk1Size uint32  // 16 for PCM
	AudioFormat   uint16  // PCM = 1
	NumChannels   uint16  // Mono = 1, Stereo = 2, etc.
	SampleRate    uint32  // 8000, 44100, etc.
	ByteRate      uint32  // SampleRate * NumChannels * BitsPerSample/8
	BlockAlign    uint16  // NumChannels * BitsPerSample/8
	BitsPerSample uint16  // 8 bits = 8, 16 bits = 16, etc.
	SubChunk2ID   [4]byte // "data"
	SubChunk2Size uint32  // data size in bytes
}

// newWAVHeader creates a new WAV header based on the input parameters.
func newWAVHeader(sampleRate, numChannels, bitsPerSample, dataLength int) *wavHeaderData {
	byteRate := sampleRate * numChannels * bitsPerSample / 8
	blockAlign := numChannels * bitsPerSample / 8

	return &wavHeaderData{
		ChunkID:       [4]byte{'R', 'I', 'F', 'F'},
		ChunkSize:     uint32(4 + (8 + 16) + (8 + dataLength)), // Calculate based on the formula given above
		Format:        [4]byte{'W', 'A', 'V', 'E'},
		SubChunk1ID:   [4]byte{'f', 'm', 't', ' '},
		SubChunk1Size: 16, // For PCM
		AudioFormat:   1,  // PCM = 1 (i.e., Linear quantization)
		NumChannels:   uint16(numChannels),
		SampleRate:    uint32(sampleRate),
		ByteRate:      uint32(byteRate),
		BlockAlign:    uint16(blockAlign),
		BitsPerSample: uint16(bitsPerSample),
		SubChunk2ID:   [4]byte{'d', 'a', 't', 'a'},
		SubChunk2Size: uint32(dataLength),
	}
}

// SaveAudioDataToFile saves the audio data to a file as a WAV file.
func SaveAudioDataToFile(filename string, data []byte) error {
	// Create the file
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	// Create the WAV header
	header := newWAVHeader(44100, 1, 16, len(data))
	// Write the WAV header
	err = binary.Write(file, binary.LittleEndian, header)
	if err != nil {
		return err
	}

	// Write the audio data
	err = binary.Write(file, binary.LittleEndian, data)
	if err != nil {
		return err
	}

	return nil
}

// SaveMnemonicToFile saves the mnemonic to a file.
func SaveMnemonicToFile(filename string, mnemonic string) error {
	// Create the file
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	// Write the mnemonic
	_, err = file.WriteString(mnemonic)
	if err != nil {
		return err
	}

	return nil
}

// Float32ToByteSlice converts a float32 slice to a byte slice.
func Float32ToByteSlice(floats []float32) []byte {
	bytes := make([]byte, 4*len(floats))
	for i, f := range floats {
		// Convert the float to a scaled int16
		val := int16(f * 32767)
		// Write the int16 to bytes
		binary.LittleEndian.PutUint16(bytes[i*2:], uint16(val))
	}
	return bytes
}
