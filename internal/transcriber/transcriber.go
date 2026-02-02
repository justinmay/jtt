package transcriber

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

type TranscribeResult struct {
	Text    string
	Seconds float64
}

type Transcriber struct {
	modelPath            string
	filterHallucinations bool
}

// findWhisperBinary locates the whisper-cli binary, checking common Homebrew paths
// since bundled macOS apps don't inherit shell PATH
func findWhisperBinary() string {
	homebrewPaths := []string{
		"/opt/homebrew/bin/whisper-cli", // Apple Silicon
		"/usr/local/bin/whisper-cli",    // Intel Mac
	}
	for _, p := range homebrewPaths {
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}
	return "whisper-cli"
}

func New(modelPath string, filterHallucinations bool) *Transcriber {
	return &Transcriber{modelPath: modelPath, filterHallucinations: filterHallucinations}
}

func (t *Transcriber) Transcribe(audioPath string) (*TranscribeResult, error) {
	if _, err := os.Stat(t.modelPath); os.IsNotExist(err) {
		return nil, err
	}

	outputBase := strings.TrimSuffix(audioPath, filepath.Ext(audioPath))

	cmd := exec.Command(findWhisperBinary(),
		"-m", t.modelPath,
		"-f", audioPath,
		"--no-timestamps",
		"--language", "en",
		"--output-txt",
		"--output-file", outputBase,
		"--no-fallback",
		"-et", "2.4",
		"-lpt", "-1.0",
	)

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	start := time.Now()
	if err := cmd.Run(); err != nil {
		errMsg := strings.TrimSpace(stderr.String())
		if errMsg == "" {
			errMsg = "unknown error"
		}
		return nil, fmt.Errorf("%w: %s", err, errMsg)
	}
	elapsed := time.Since(start).Seconds()

	txtPath := outputBase + ".txt"
	data, err := os.ReadFile(txtPath)
	if err != nil {
		return nil, err
	}

	text := strings.TrimSpace(string(data))

	// Filter out known whisper hallucinations on silence/noise
	if t.filterHallucinations {
		normalized := strings.ToLower(strings.TrimRight(text, ".,!?"))
		if normalized == "you" {
			log.Println("transcriber: filtered hallucination on silence/noise")
			text = ""
		}
	}

	if text == "" {
		log.Println("transcriber: no audio detected (empty transcription)")
	}

	return &TranscribeResult{
		Text:    text,
		Seconds: elapsed,
	}, nil
}
