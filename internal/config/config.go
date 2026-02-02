package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type HotkeyConfig struct {
	Modifiers []string `json:"modifiers"`
	Keys      []string `json:"keys"`
}

type Config struct {
	WhisperModel         string       `json:"whisperModel"`
	UseOllama            bool         `json:"useOllama"`
	OllamaModel          string       `json:"ollamaModel"`
	Hotkey               HotkeyConfig `json:"hotkey"`
	LLMPrompt            string       `json:"llmPrompt"`
	FilterHallucinations bool         `json:"filterHallucinations"`
	PauseMediaOnRecord   bool         `json:"pauseMediaOnRecord"`
	Microphone           string       `json:"microphone"`
}

type TranscriptionEntry struct {
	Timestamp     int64   `json:"timestamp"`
	WhisperTime   float64 `json:"whisperTime"`
	WhisperOutput string  `json:"whisperOutput"`
	LLMTime       float64 `json:"llmTime"`
	LLMOutput     string  `json:"llmOutput"`
}

const DefaultLLMPrompt = `Clean this voice transcript. Output ONLY the cleaned text, nothing else.
Rules: remove filler words (um, uh, like), fix punctuation and casing, keep original wording.
Transcript:
{{transcript}}`

func DefaultConfig() *Config {
	homeDir, _ := os.UserHomeDir()
	return &Config{
		WhisperModel:         filepath.Join(homeDir, ".local", "share", "jtt", "ggml-small.en.bin"),
		UseOllama:            true,
		OllamaModel:          "llama3.2:3b",
		Hotkey: HotkeyConfig{
			Modifiers: []string{"cmd", "shift"},
			Keys:      []string{"r"},
		},
		LLMPrompt:            DefaultLLMPrompt,
		FilterHallucinations: true,
		PauseMediaOnRecord:   true,
	}
}

func ConfigPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(homeDir, ".config", "jtt", "config.json"), nil
}

func Load() (*Config, error) {
	path, err := ConfigPath()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return DefaultConfig(), nil
		}
		return nil, err
	}

	cfg := &Config{}
	if err := json.Unmarshal(data, cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

// IsFirstRun returns true if config file doesn't exist yet
func IsFirstRun() bool {
	path, err := ConfigPath()
	if err != nil {
		return true
	}
	_, err = os.Stat(path)
	return os.IsNotExist(err)
}

func (c *Config) Save() error {
	path, err := ConfigPath()
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}
