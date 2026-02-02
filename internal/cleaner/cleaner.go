package cleaner

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type CleanResult struct {
	Text    string
	Seconds float64
}

const ollamaURL = "http://localhost:11434/api/generate"

type Cleaner struct {
	model   string
	enabled bool
	prompt  string
}

func New(model string, enabled bool, prompt string) *Cleaner {
	return &Cleaner{model: model, enabled: enabled, prompt: prompt}
}

func (c *Cleaner) Clean(text string) (*CleanResult, error) {
	if !c.enabled {
		return &CleanResult{Text: text, Seconds: 0}, nil
	}

	prompt := strings.ReplaceAll(c.prompt, "{{transcript}}", text)

	reqBody := map[string]interface{}{
		"model":  c.model,
		"prompt": prompt,
		"stream": false,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return &CleanResult{Text: text, Seconds: 0}, err
	}

	start := time.Now()
	resp, err := http.Post(ollamaURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return &CleanResult{Text: text, Seconds: 0}, fmt.Errorf("ollama not running: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	elapsed := time.Since(start).Seconds()
	if err != nil {
		return &CleanResult{Text: text, Seconds: elapsed}, err
	}

	var result struct {
		Response string `json:"response"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return &CleanResult{Text: text, Seconds: elapsed}, err
	}

	return &CleanResult{
		Text:    strings.TrimSpace(result.Response),
		Seconds: elapsed,
	}, nil
}

func ListModels() ([]string, error) {
	resp, err := http.Get("http://localhost:11434/api/tags")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result struct {
		Models []struct {
			Name string `json:"name"`
		} `json:"models"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	names := make([]string, len(result.Models))
	for i, m := range result.Models {
		names[i] = m.Name
	}
	return names, nil
}

func IsOllamaRunning() bool {
	resp, err := http.Get("http://localhost:11434/api/tags")
	if err != nil {
		return false
	}
	resp.Body.Close()
	return true
}
