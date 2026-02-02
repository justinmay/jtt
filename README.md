# JTT - Justin's Transcription Tool

A macOS menu bar app for voice-to-text transcription using whisper-cpp, with optional LLM cleaning via Ollama.

## Features

- **Menu bar app** - Lives in your menu bar, always ready
- **Voice transcription** - Uses whisper-cpp for fast local transcription
- **LLM cleaning** (optional) - Cleans up transcripts using Ollama (removes filler words, fixes punctuation)
- **Clipboard integration** - Transcribed text is automatically copied to clipboard
- **Settings UI** - Configure whisper model, Ollama settings, and more

## Installation

### Prerequisites

```bash
# Install sox (audio recording)
brew install sox

# Install whisper-cpp (transcription)
brew install whisper-cpp

# Optional: Install Ollama (LLM text cleaning)
brew install ollama
brew services start ollama
ollama pull llama3.2:3b
```

### Download Whisper Model

On first launch, open Settings and download a whisper model:

| Model | Size | Speed | Quality |
|-------|------|-------|---------|
| tiny.en | 75MB | Fastest | Basic |
| base.en | 142MB | Fast | Good |
| small.en | 466MB | Medium | Better |
| medium.en | 1.5GB | Slow | Great |
| large | 3GB | Slowest | Best |

### Build from Source

```bash
# Install Wails v3 CLI
go install github.com/wailsapp/wails/v3/cmd/wails3@latest

# Development (hot reload)
wails3 dev

# Build for production
wails3 build

# Run
./bin/jtt
```

## Usage

1. Click the microphone icon in the menu bar
2. Select "Start Recording" and speak
3. Select "Stop Recording" when done
4. Transcription is copied to your clipboard - paste anywhere!

### Settings

Click the menu bar icon → Settings to configure:
- **Whisper Model** - Download and select transcription model
- **Ollama** - Enable/disable LLM text cleaning, select model

## Architecture

- **Backend**: Go with Wails v3 (alpha)
- **Frontend**: React + TypeScript + Vite
- **Audio**: sox (`rec` command)
- **Transcription**: whisper-cpp (`whisper-cli` command)
- **LLM**: Ollama HTTP API (localhost:11434)

## Config

Settings stored in `~/.config/jtt/config.json`

## Known Issues

- Global hotkey not yet implemented (use menu bar for now)
- Wails v3 is in alpha - some features may be unstable

## License

MIT
