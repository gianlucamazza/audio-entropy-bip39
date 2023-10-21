# Enhanced Security Mnemonic Generation Using Ambient Sound Entropy

This project showcases the process of generating a secure mnemonic phrase by incorporating standard entropy with additional entropy captured from microphone audio input.

## Overview

Security is paramount when generating mnemonic phrases for use in cryptocurrency wallets or other applications demanding robust encryption. By introducing entropy from a microphone's audio input, the mnemonic generation process becomes less predictable and more resistant to attacks compared to using standard entropy alone.

In this example, we record a short audio sample, compute its hash, and merge it with securely generated entropy before creating the mnemonic phrase. This approach enhances security while adding a unique element to each generated mnemonic phrase.

## How It Works

1. **Entropy Generation**: Secure entropy is generated using a trusted cryptographic library.
2. **Audio Recording**: Audio input is captured from the user's microphone.
3. **Audio Hash Calculation**: An audio sample is hashed after recording.
4. **Entropy Combination**: Both the generated entropy and audio hash are merged into a single entity.
5. **Mnemonic Generation**: The combined entropy is used to produce a mnemonic phrase following the BIP39 standard.

## Prerequisites

- Go (Recommended version: 1.x or higher)
- PortAudio
- Go packages:
    - `github.com/gordonklaus/portaudio`
    - `github.com/tyler-smith/go-bip39`
    - `golang.org/x/crypto/hkdf`

## Setup

To execute this example, follow these steps:

1. Install the necessary dependencies.
2. Clone the repository or download the source code.
3. Navigate to the project directory via the command line.
4. Run `go run main.go` to start the application.

During execution, the application will prompt you to speak into the microphone and briefly record audio. After recording, it processes the audio, generates combined entropy, and ultimately prints out the mnemonic phrase.

## Example Output

```bash
Entropy: 7bdcfaebbafb0365553c2f3fba9b9bccdd8be86deba806ae91d1a8a7249cc2ba
Derived key: 1c98b1fea2a15f95e0faf7c0295a397b63ed3a1a91f81643da7dd193a60e704f
Recording. Speak into the microphone.
[###                                               ]
Recording complete. Processing...
Audio hash: 2b7403769a83516d2aa1d8354eed51eaa6d2c3d21262239d9549db7cbd595212
Combined hash: 7c04f865fa40fac3310a14c1d91b30fd455f158babff73e097ac3569f5def878
Mnemonic: lab chief bonus virus autumn ghost service dream scrub silver slow whisper field member concert lend initial again twelve hello palm urge tiger barely
```

Security Considerations
While adding entropy from audio provides an additional security layer, it's vital to note that the quality of entropy will depend on environmental conditions and the microphone hardware's quality. This method should be used as an extra security layer in conjunction with other reliable entropy generation methods.

Contributing
Contributions, enhancements, and bug reports are always welcome. Please refer to the contribution guidelines for more details.

