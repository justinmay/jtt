package main

import (
	"embed"
	"fmt"
	"jtt/internal/accessibility"
	"jtt/internal/cleaner"
	"jtt/internal/config"
	"jtt/internal/logger"
	"jtt/internal/media"
	"jtt/internal/recorder"
	"jtt/internal/transcriber"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/wailsapp/wails/v3/pkg/application"
	"golang.design/x/hotkey"
)

//go:embed all:frontend/dist
var assets embed.FS

//go:embed assets/icon-idle.png
var iconIdle []byte

//go:embed assets/icon-recording.png
var iconRecording []byte

//go:embed assets/icon-processing.png
var iconProcessing []byte

type AppState string

const (
	StateIdle       AppState = "idle"
	StateRecording  AppState = "recording"
	StateProcessing AppState = "processing"
)

type JTTApp struct {
	app           *application.App
	systray       *application.SystemTray
	window        *application.WebviewWindow
	cfg           *config.Config
	recorder      *recorder.Recorder
	state         AppState
	history       []config.TranscriptionEntry
	mediaWasPlaying bool
}

func main() {
	if err := logger.Init(); err != nil {
		log.Printf("Failed to init logger: %v", err)
	}
	defer logger.Close()

	cfg, err := config.Load()
	if err != nil {
		logger.Error("Failed to load config: %v", err)
		cfg = config.DefaultConfig()
	}

	jtt := &JTTApp{
		cfg:      cfg,
		recorder: recorder.New(cfg.Microphone),
		state:    StateIdle,
		history:  make([]config.TranscriptionEntry, 0, 5),
	}

	app := application.New(application.Options{
		Name:        "JTT",
		Description: "Justin's Transcription Tool",
		Services: []application.Service{
			application.NewService(NewJTTService(jtt)),
		},
		Assets: application.AssetOptions{
			Handler: application.AssetFileServerFS(assets),
		},
		Mac: application.MacOptions{
			ApplicationShouldTerminateAfterLastWindowClosed: false,
		},
	})

	jtt.app = app
	jtt.setupSystray()

	// Check accessibility permissions on startup (only prompt if not already granted)
	go func() {
		// First check without prompting
		if !accessibility.CheckAccessibility(false) {
			// Not granted, prompt user
			if !accessibility.CheckAccessibility(true) {
				logger.Info("Accessibility permissions not granted - paste feature will not work")
			}
		}
	}()

	// Open settings window on launch
	go func() {
		time.Sleep(500 * time.Millisecond)
		jtt.showSettings()
	}()

	// Start hotkey listener in a goroutine - it will work after app starts
	go func() {
		// Wait a bit for app to initialize
		time.Sleep(2 * time.Second)
		jtt.setupHotkey()
	}()

	err = app.Run()
	if err != nil {
		log.Fatal(err)
	}
}

func (j *JTTApp) setupSystray() {
	j.systray = j.app.SystemTray.New()
	j.systray.SetIcon(iconIdle)

	j.updateMenu()
}

func (j *JTTApp) createWindow() *application.WebviewWindow {
	return j.app.Window.NewWithOptions(application.WebviewWindowOptions{
		Title:    "JTT Settings",
		Width:    640,
		Height:   700,
		MinWidth: 640,
		Mac: application.MacWindow{
			Backdrop:                application.MacBackdropTranslucent,
			TitleBar:                application.MacTitleBarHiddenInset,
			InvisibleTitleBarHeight: 50,
		},
		BackgroundColour: application.NewRGB(28, 28, 30),
		URL:              "/",
	})
}

func (j *JTTApp) showSettings() {
	// Close existing window if any
	if j.window != nil {
		j.window.Close()
	}
	// Create fresh window to avoid Wails v3 hide/show crash bug
	j.window = j.createWindow()
	j.window.Focus()
}

func (j *JTTApp) setupHotkey() {
	if len(j.cfg.Hotkey.Keys) == 0 {
		log.Printf("No hotkey configured")
		return
	}

	key := j.cfg.Hotkey.Keys[0]
	mods := parseModifiers(j.cfg.Hotkey.Modifiers)
	k := parseKey(key)

	j.registerHotkey(mods, k)
}

func (j *JTTApp) registerHotkey(mods []hotkey.Modifier, key hotkey.Key) {
	hk := hotkey.New(mods, key)
	err := hk.Register()
	if err != nil {
		log.Printf("Failed to register hotkey: %v", err)
		return
	}

	log.Printf("Hotkey registered: %v + %v", mods, key)

	// Listen for keydown (start recording)
	<-hk.Keydown()
	log.Printf("Hotkey pressed - starting recording")
	j.StartRecording()

	// Listen for keyup (stop recording)
	<-hk.Keyup()
	log.Printf("Hotkey released - stopping recording")
	j.StopRecording()

	hk.Unregister()

	// Re-register to listen again
	j.registerHotkey(mods, key)
}

func parseModifiers(mods []string) []hotkey.Modifier {
	var result []hotkey.Modifier
	for _, m := range mods {
		switch m {
		case "cmd", "command":
			result = append(result, hotkey.ModCmd)
		case "ctrl", "control":
			result = append(result, hotkey.ModCtrl)
		case "alt", "option":
			result = append(result, hotkey.ModOption)
		case "shift":
			result = append(result, hotkey.ModShift)
		}
	}
	return result
}

func parseKey(key string) hotkey.Key {
	switch key {
	case "a", "A":
		return hotkey.KeyA
	case "b", "B":
		return hotkey.KeyB
	case "c", "C":
		return hotkey.KeyC
	case "d", "D":
		return hotkey.KeyD
	case "e", "E":
		return hotkey.KeyE
	case "f", "F":
		return hotkey.KeyF
	case "g", "G":
		return hotkey.KeyG
	case "h", "H":
		return hotkey.KeyH
	case "i", "I":
		return hotkey.KeyI
	case "j", "J":
		return hotkey.KeyJ
	case "k", "K":
		return hotkey.KeyK
	case "l", "L":
		return hotkey.KeyL
	case "m", "M":
		return hotkey.KeyM
	case "n", "N":
		return hotkey.KeyN
	case "o", "O":
		return hotkey.KeyO
	case "p", "P":
		return hotkey.KeyP
	case "q", "Q":
		return hotkey.KeyQ
	case "r", "R":
		return hotkey.KeyR
	case "s", "S":
		return hotkey.KeyS
	case "t", "T":
		return hotkey.KeyT
	case "u", "U":
		return hotkey.KeyU
	case "v", "V":
		return hotkey.KeyV
	case "w", "W":
		return hotkey.KeyW
	case "x", "X":
		return hotkey.KeyX
	case "y", "Y":
		return hotkey.KeyY
	case "z", "Z":
		return hotkey.KeyZ
	case "0":
		return hotkey.Key0
	case "1":
		return hotkey.Key1
	case "2":
		return hotkey.Key2
	case "3":
		return hotkey.Key3
	case "4":
		return hotkey.Key4
	case "5":
		return hotkey.Key5
	case "6":
		return hotkey.Key6
	case "7":
		return hotkey.Key7
	case "8":
		return hotkey.Key8
	case "9":
		return hotkey.Key9
	case "space", " ":
		return hotkey.KeySpace
	case "return", "enter":
		return hotkey.KeyReturn
	case "escape", "esc":
		return hotkey.KeyEscape
	case "f1":
		return hotkey.KeyF1
	case "f2":
		return hotkey.KeyF2
	case "f3":
		return hotkey.KeyF3
	case "f4":
		return hotkey.KeyF4
	case "f5":
		return hotkey.KeyF5
	case "f6":
		return hotkey.KeyF6
	case "f7":
		return hotkey.KeyF7
	case "f8":
		return hotkey.KeyF8
	case "f9":
		return hotkey.KeyF9
	case "f10":
		return hotkey.KeyF10
	case "f11":
		return hotkey.KeyF11
	case "f12":
		return hotkey.KeyF12
	default:
		return hotkey.KeySpace
	}
}

func (j *JTTApp) updateHotkey(modifiers []string, key string) error {
	j.cfg.Hotkey.Modifiers = modifiers
	j.cfg.Hotkey.Keys = []string{key}
	return j.cfg.Save()
	// Note: Hotkey will be re-registered on next app restart
}

func (j *JTTApp) updateMenu() {
	menu := j.app.NewMenu()

	statusLabel := "Ready"
	if j.state == StateRecording {
		statusLabel = "Recording..."
	} else if j.state == StateProcessing {
		statusLabel = "Processing..."
	}

	status := menu.Add(statusLabel)
	status.SetEnabled(false)

	menu.AddSeparator()

	if j.state == StateIdle {
		menu.Add("Start Recording").OnClick(func(ctx *application.Context) {
			j.StartRecording()
		})
	} else if j.state == StateRecording {
		menu.Add("Stop Recording").OnClick(func(ctx *application.Context) {
			j.StopRecording()
		})
	}

	menu.AddSeparator()

	menu.Add("Settings...").OnClick(func(ctx *application.Context) {
		j.showSettings()
	})

	menu.AddSeparator()

	menu.Add("Quit").OnClick(func(ctx *application.Context) {
		j.app.Quit()
	})

	j.systray.SetMenu(menu)
}

func (j *JTTApp) updateState(state AppState) {
	j.state = state

	switch state {
	case StateIdle:
		j.systray.SetTemplateIcon(iconIdle)
	case StateRecording:
		j.systray.SetTemplateIcon(iconRecording)
	case StateProcessing:
		j.systray.SetTemplateIcon(iconProcessing)
	}

	j.updateMenu()
	// Emit event to frontend
	j.app.Event.Emit("state-change", string(state))
}

func (j *JTTApp) StartRecording() error {
	if j.state != StateIdle {
		return nil
	}

	// Pause media if enabled and playing
	j.mediaWasPlaying = false
	if j.cfg.PauseMediaOnRecord && media.IsPlaying() {
		media.Pause()
		j.mediaWasPlaying = true
		logger.Info("Paused media playback")
	}

	if err := j.recorder.Start(); err != nil {
		logger.Error("Failed to start recording: %v", err)
		// Resume media if we paused it
		if j.mediaWasPlaying {
			media.Play()
		}
		return err
	}

	logger.Info("Recording started")
	j.updateState(StateRecording)
	return nil
}

func (j *JTTApp) StopRecording() (string, error) {
	if j.state != StateRecording {
		return "", nil
	}

	j.updateState(StateProcessing)

	if err := j.recorder.Stop(); err != nil {
		logger.Error("Failed to stop recording: %v", err)
		j.updateState(StateIdle)
		return "", err
	}

	logger.Info("Recording stopped, starting transcription")
	trans := transcriber.New(j.cfg.WhisperModel, j.cfg.FilterHallucinations)
	whisperResult, err := trans.Transcribe(j.recorder.AudioPath())
	if err != nil {
		logger.Error("Transcription failed: %v", err)
		j.updateState(StateIdle)
		return "", err
	}
	logger.Info("Transcription completed in %.2fs", whisperResult.Seconds)

	// Skip LLM cleaning if there's no text
	var cleanResult *cleaner.CleanResult
	if whisperResult.Text == "" {
		cleanResult = &cleaner.CleanResult{Text: "", Seconds: 0}
	} else {
		prompt := j.cfg.LLMPrompt
		if prompt == "" {
			prompt = config.DefaultLLMPrompt
		}
		clean := cleaner.New(j.cfg.OllamaModel, j.cfg.UseOllama, prompt)
		cleanResult, err = clean.Clean(whisperResult.Text)
		if err != nil {
			logger.Error("LLM cleaning failed: %v", err)
			cleanResult = &cleaner.CleanResult{Text: whisperResult.Text, Seconds: 0}
		}
	}

	// Add to history (keep last 5)
	entry := config.TranscriptionEntry{
		Timestamp:     time.Now().Unix(),
		WhisperTime:   whisperResult.Seconds,
		WhisperOutput: whisperResult.Text,
		LLMTime:       cleanResult.Seconds,
		LLMOutput:     cleanResult.Text,
	}
	j.history = append(j.history, entry)
	if len(j.history) > 5 {
		j.history = j.history[len(j.history)-5:]
	}

	// Copy to clipboard using pbcopy
	clipCmd := exec.Command("pbcopy")
	clipCmd.Stdin = strings.NewReader(cleanResult.Text)
	if err := clipCmd.Run(); err != nil {
		logger.Error("Failed to copy to clipboard: %v", err)
	} else {
		logger.Info("Copied to clipboard: %d chars", len(cleanResult.Text))
	}

	// Small delay to ensure clipboard is ready
	time.Sleep(50 * time.Millisecond)

	// Paste using Cmd+V
	pasteCmd := exec.Command("osascript", "-e", `tell application "System Events" to keystroke "v" using command down`)
	if output, err := pasteCmd.CombinedOutput(); err != nil {
		logger.Error("Failed to paste: %v, output: %s", err, string(output))
	} else {
		logger.Info("Paste command executed")
	}

	// Resume media if it was playing before recording
	if j.mediaWasPlaying {
		media.Play()
		logger.Info("Resumed media playback")
		j.mediaWasPlaying = false
	}

	j.updateState(StateIdle)
	return cleanResult.Text, nil
}

// JTTService exposes methods to the frontend
type JTTService struct {
	jtt *JTTApp
}

func NewJTTService(jtt *JTTApp) *JTTService {
	return &JTTService{jtt: jtt}
}

func (s *JTTService) GetState() string {
	return string(s.jtt.state)
}

func (s *JTTService) GetConfig() *config.Config {
	return s.jtt.cfg
}

func (s *JTTService) SaveConfig(cfg *config.Config) error {
	s.jtt.cfg = cfg
	// Update recorder's microphone setting
	s.jtt.recorder.SetMicrophone(cfg.Microphone)
	return cfg.Save()
}

func (s *JTTService) UpdateHotkey(modifiers []string, key string) error {
	return s.jtt.updateHotkey(modifiers, key)
}

func (s *JTTService) StartRecording() error {
	return s.jtt.StartRecording()
}

func (s *JTTService) StopRecording() (string, error) {
	return s.jtt.StopRecording()
}

func (s *JTTService) GetOllamaModels() []string {
	models, err := cleaner.ListModels()
	if err != nil {
		return []string{}
	}
	return models
}

func (s *JTTService) IsOllamaRunning() bool {
	return cleaner.IsOllamaRunning()
}

type DependencyStatus struct {
	Sox        bool `json:"sox"`
	Whisper    bool `json:"whisper"`
	Ollama     bool `json:"ollama"`
	HasModel   bool `json:"hasModel"`
	NowPlaying bool `json:"nowPlaying"`
}

func (s *JTTService) CheckDependencies() DependencyStatus {
	status := DependencyStatus{}

	// Check Homebrew paths directly since bundled apps don't have shell PATH
	soxPaths := []string{"/opt/homebrew/bin/rec", "/usr/local/bin/rec"}
	for _, p := range soxPaths {
		if _, err := os.Stat(p); err == nil {
			status.Sox = true
			break
		}
	}

	whisperPaths := []string{"/opt/homebrew/bin/whisper-cli", "/usr/local/bin/whisper-cli"}
	for _, p := range whisperPaths {
		if _, err := os.Stat(p); err == nil {
			status.Whisper = true
			break
		}
	}

	status.Ollama = cleaner.IsOllamaRunning()
	status.NowPlaying = media.IsAvailable()

	_, err := os.Stat(s.jtt.cfg.WhisperModel)
	status.HasModel = err == nil

	return status
}

func findBrewBinary() string {
	paths := []string{
		"/opt/homebrew/bin/brew", // Apple Silicon
		"/usr/local/bin/brew",    // Intel Mac
	}
	for _, p := range paths {
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}
	return "brew"
}

func (s *JTTService) InstallDependency(dep string) error {
	brew := findBrewBinary()
	var cmd *exec.Cmd
	switch dep {
	case "sox":
		cmd = exec.Command(brew, "install", "sox")
	case "whisper":
		cmd = exec.Command(brew, "install", "whisper-cpp")
	case "nowplaying":
		cmd = exec.Command(brew, "install", "nowplaying-cli")
	default:
		return fmt.Errorf("unknown dependency: %s", dep)
	}
	if err := cmd.Run(); err != nil {
		logger.Error("Failed to install %s: %v", dep, err)
		return err
	}
	logger.Info("Successfully installed %s", dep)
	return nil
}

type WhisperModel struct {
	Name    string `json:"name"`
	Size    string `json:"size"`
	Speed   string `json:"speed"`
	Quality string `json:"quality"`
	URL     string `json:"url"`
}

func (s *JTTService) GetAvailableWhisperModels() []WhisperModel {
	return []WhisperModel{
		{Name: "tiny.en", Size: "75MB", Speed: "Fastest", Quality: "Basic", URL: "https://huggingface.co/ggerganov/whisper.cpp/resolve/main/ggml-tiny.en.bin"},
		{Name: "base.en", Size: "142MB", Speed: "Fast", Quality: "Good", URL: "https://huggingface.co/ggerganov/whisper.cpp/resolve/main/ggml-base.en.bin"},
		{Name: "small.en", Size: "466MB", Speed: "Medium", Quality: "Better", URL: "https://huggingface.co/ggerganov/whisper.cpp/resolve/main/ggml-small.en.bin"},
		{Name: "medium.en", Size: "1.5GB", Speed: "Slow", Quality: "Great", URL: "https://huggingface.co/ggerganov/whisper.cpp/resolve/main/ggml-medium.en.bin"},
		{Name: "large", Size: "3GB", Speed: "Slowest", Quality: "Best", URL: "https://huggingface.co/ggerganov/whisper.cpp/resolve/main/ggml-large.bin"},
	}
}

func (s *JTTService) DownloadWhisperModel(name, url string) error {
	homeDir, _ := os.UserHomeDir()
	modelDir := filepath.Join(homeDir, ".local", "share", "jtt")
	os.MkdirAll(modelDir, 0755)

	modelPath := filepath.Join(modelDir, "ggml-"+name+".bin")

	cmd := exec.Command("curl", "-L", "-o", modelPath, url)
	if err := cmd.Run(); err != nil {
		return err
	}

	s.jtt.cfg.WhisperModel = modelPath
	return s.jtt.cfg.Save()
}

func (s *JTTService) GetDownloadedModels() []string {
	homeDir, _ := os.UserHomeDir()
	modelDir := filepath.Join(homeDir, ".local", "share", "jtt")

	entries, err := os.ReadDir(modelDir)
	if err != nil {
		return []string{}
	}

	var models []string
	for _, e := range entries {
		if filepath.Ext(e.Name()) == ".bin" {
			models = append(models, e.Name())
		}
	}
	return models
}

func (s *JTTService) GetHistory() []config.TranscriptionEntry {
	return s.jtt.history
}

func (s *JTTService) GetDefaultPrompt() string {
	return config.DefaultLLMPrompt
}

func (s *JTTService) GetLogPath() string {
	return logger.LogPath()
}

func (s *JTTService) GetRecentLogs() string {
	return logger.GetRecentLogs(100)
}

func (s *JTTService) GetMicrophones() []recorder.Microphone {
	mics, err := recorder.ListMicrophones()
	if err != nil {
		return []recorder.Microphone{{ID: "", Name: "System Default"}}
	}
	return mics
}
