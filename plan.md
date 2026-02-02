# JTT Refactoring Plan

## Goal
Replace the current Python CLI + Hammerspoon setup with a native macOS menu bar app using **Wails** (Go + webview frontend).

## Current Architecture
- **Python CLI** (`jtt start`/`jtt stop`) - recording (sox), transcription (whisper-cpp), cleaning (ollama)
- **Hammerspoon Spoon** - global hotkey listener (`[` + `]`), triggers CLI, pastes result

## Proposed Architecture

### Why Wails?
- Compiles to a **single native binary** - no Go installation required for users
- Built-in systray/menu bar support
- Lightweight (~10MB binary)
- Go is simpler for CLI/process management than Rust

### Project Structure
```
jtt/
├── main.go                 # Wails app entry
├── app.go                  # Backend methods exposed to frontend
├── internal/
│   ├── recorder/           # sox wrapper
│   ├── transcriber/        # whisper-cpp integration
│   ├── cleaner/            # ollama integration (optional)
│   ├── config/             # settings persistence
│   └── hotkey/             # global hotkey (golang.design/x/hotkey)
├── frontend/               # Svelte/React settings UI
│   ├── src/
│   └── ...
├── build/                  # macOS app bundle output
└── wails.json
```

### Menu Bar States
| State | Icon | Description |
|-------|------|-------------|
| Idle | 🎤 (gray) | Ready to record |
| Recording | 🔴 (red) | Actively capturing audio |
| Processing | ⏳ (spinner) | Transcribing / cleaning |

### Settings Page (React)
- [ ] **Whisper Model** selector with download buttons:
  | Model | Size | Speed | Quality |
  |-------|------|-------|---------|
  | tiny.en | 75MB | Fastest | Basic |
  | base.en | 142MB | Fast | Good |
  | small.en | 466MB | Medium | Better |
  | medium.en | 1.5GB | Slow | Great |
  | large | 3GB | Slowest | Best |
- [ ] **Use Ollama** toggle (on/off)
- [ ] **Ollama Model** dropdown (fetched from `ollama list`)
- [ ] **Hotkey** configuration (e.g., `[+]`, `Cmd+Shift+R`, etc.)

### Config File
Location: `~/.config/jtt/config.json`
```json
{
  "whisperModel": "~/.local/share/jtt/ggml-small.en.bin",
  "useOllama": true,
  "ollamaModel": "llama3.2:3b",
  "hotkey": {
    "modifiers": [],
    "keys": ["[", "]"]
  }
}
```

## Dependency Strategy

### Shell-out vs Native Go

| Component | Shell-out | Native Go | Recommendation |
|-----------|-----------|-----------|----------------|
| **Audio Recording** | `sox` (rec) | `gen2brain/malgo` (miniaudio) | Native - miniaudio is header-only C, easy to bundle |
| **Transcription** | `whisper-cli` | `ggerganov/whisper.cpp/bindings/go` | Native - official Go bindings |
| **LLM Cleaning** | `ollama run` | HTTP API (`localhost:11434`) | HTTP - cleaner, no process spawning |
| **Clipboard** | `pbcopy` | `golang.design/x/clipboard` | Native - pure Go |
| **Hotkey** | (hammerspoon) | `golang.design/x/hotkey` | Native - pure Go |

### Shell out to
- `rec` (sox) - audio capture
- `whisper-cli` - transcription

### Native (in binary)
- Hotkey listener (`golang.design/x/hotkey`)
- Clipboard (`golang.design/x/clipboard`)
- Ollama client (HTTP to `localhost:11434`)
- Menu bar / systray
- Settings UI (React)

### External Dependencies
- `sox` - audio recording
- `whisper-cpp` - transcription
- `ollama` (optional) - LLM cleaning
- Whisper model weights (~150MB for small.en)

## Install Script

On first launch (or via `jtt install`), check for and install dependencies:

```bash
#!/bin/bash
set -e

# Check for Homebrew
if ! command -v brew &> /dev/null; then
    echo "Homebrew not found. Install from https://brew.sh"
    exit 1
fi

# Install sox if missing
if ! command -v rec &> /dev/null; then
    echo "Installing sox..."
    brew install sox
fi

# Install whisper-cpp if missing
if ! command -v whisper-cli &> /dev/null; then
    echo "Installing whisper-cpp..."
    brew install whisper-cpp
fi

# Note: Whisper model downloaded via Settings UI (user chooses size/quality)

# Optional: Install ollama
if ! command -v ollama &> /dev/null; then
    read -p "Install Ollama for LLM text cleaning? (y/n) " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        brew install ollama
        brew services start ollama
        ollama pull llama3.2:3b
    fi
fi

echo "Done! JTT is ready to use."
```

### Install Options
1. **Standalone script**: `scripts/install.sh` - user runs manually
2. **In-app check**: App detects missing deps on launch, prompts to run install
3. **Embedded**: App runs the checks itself via Go `exec.Command`

Recommendation: **Option 3** - App checks on launch, shows friendly UI if deps missing with "Install" button that runs the commands

## macOS Permissions

### Info.plist
Required for microphone access prompt:
```xml
<key>NSMicrophoneUsageDescription</key>
<string>JTT needs microphone access to record your voice for transcription</string>
```

### Entitlements (for distribution/notarization)
```xml
<!-- Sandbox entitlement -->
<key>com.apple.security.device.microphone</key>
<true/>

<!-- Hardened runtime entitlement -->
<key>com.apple.security.device.audio-input</key>
<true/>
```

### Accessibility (for global hotkey)
The app will need Accessibility permissions to capture global hotkeys. macOS will prompt automatically when the app tries to register a global hotkey. Users grant this in:
**System Preferences → Privacy & Security → Accessibility**

## Migration Steps

### Phase 1: Scaffold
- [ ] Initialize Wails project
- [ ] Set up Go module structure
- [ ] Basic systray with idle/recording states

### Phase 2: Core Logic
- [ ] Port recorder (sox) to Go
- [ ] Port transcriber (whisper-cpp CLI wrapper) to Go
- [ ] Port cleaner (ollama CLI wrapper) to Go
- [ ] Implement config loading/saving

### Phase 3: Hotkey
- [ ] Integrate `golang.design/x/hotkey` for global hotkey
- [ ] Support configurable hotkey combos (modifiers + keys)
- [ ] Remove Hammerspoon dependency

### Phase 4: Settings UI (React)
- [ ] Build settings page with React
- [ ] Wire up config changes to backend
- [ ] Fetch available Ollama models
- [ ] Hotkey recorder component (press keys to set new hotkey)

### Phase 5: Polish
- [ ] App icon design
- [ ] Build `.app` bundle
- [ ] Update README with new install instructions
- [ ] Clean up old Python/Hammerspoon code

## Open Questions
- [ ] Should we use go-whisper bindings instead of shelling out to whisper-cli?
- [ ] Auto-update mechanism?
