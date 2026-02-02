# JTT - Agent Instructions

## Project Overview
JTT (Justin's Transcription Tool) is a macOS menu bar app for voice-to-text transcription using whisper-cpp and optional Ollama cleaning.

## Technology Stack
- **Backend**: Go with Wails v3 (ALPHA - v3.0.0-alpha.65)
- **Frontend**: React + TypeScript + Vite
- **Transcription**: whisper-cpp (shell out to `whisper-cli`)
- **Audio Recording**: sox (shell out to `rec`)
- **LLM Cleaning**: Ollama HTTP API (localhost:11434)

## IMPORTANT: Wails v3 Alpha

This project uses **Wails v3 alpha**, NOT v2. The APIs are different.

### Wails v3 Documentation
- Main docs: https://v3alpha.wails.io/
- Systray: https://v3alpha.wails.io/features/menus/systray/
- Services: https://v3alpha.wails.io/services/
- Events: https://v3alpha.wails.io/guides/events/

### Key Wails v3 Differences from v2
- Use `application.New()` instead of `wails.Run()`
- Services replace bound structs: `application.NewService(&MyService{})`
- Systray is built-in: `app.SystemTray.New()`
- Window creation: `app.Window.New()` or `app.Window.NewWithOptions()`
- Events: `app.Event.Emit()` and `app.Event.On()`
- Mac options: `application.MacOptions{}` not `mac.Options{}`

### CLI Commands
```bash
# Development
wails3 dev

# Build
wails3 build

# Generate bindings
wails3 generate bindings
```

## Project Structure
```
jtt/
├── main.go              # Wails app entry with systray
├── services.go          # Go services exposed to frontend
├── internal/
│   ├── config/          # Config loading/saving
│   ├── recorder/        # sox wrapper
│   ├── transcriber/     # whisper-cpp wrapper
│   └── cleaner/         # Ollama HTTP client
├── assets/              # Systray icons
├── frontend/            # React UI
└── build/               # Build config
```

## Systray States
- **Idle**: Gray microphone icon, ready to record
- **Recording**: Red icon, actively capturing audio
- **Processing**: Spinner/yellow icon, transcribing/cleaning

## Config Location
`~/.config/jtt/config.json`
