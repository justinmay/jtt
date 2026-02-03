package recorder

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"
)

type Recorder struct {
	cmd        *exec.Cmd
	audioPath  string
	pidPath    string
	microphone string
}

type Microphone struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// ListMicrophones returns available audio input devices on macOS
func ListMicrophones() ([]Microphone, error) {
	return listMicrophonesFromSystemProfiler()
}

func listMicrophonesFromSystemProfiler() ([]Microphone, error) {
	cmd := exec.Command("system_profiler", "SPAudioDataType")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	var mics []Microphone
	scanner := bufio.NewScanner(strings.NewReader(string(output)))

	var currentDevice string
	var hasInput bool

	for scanner.Scan() {
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)

		// Detect device names: lines with 8 spaces indent ending with ":"
		// (not property lines which have 10+ spaces)
		if strings.HasPrefix(line, "        ") && !strings.HasPrefix(line, "          ") &&
			strings.HasSuffix(trimmed, ":") &&
			!strings.Contains(trimmed, "Devices:") {
			// Save previous device if it had input
			if hasInput && currentDevice != "" {
				found := false
				for _, m := range mics {
					if m.Name == currentDevice {
						found = true
						break
					}
				}
				if !found {
					mics = append(mics, Microphone{
						ID:   currentDevice,
						Name: currentDevice,
					})
				}
			}
			currentDevice = strings.TrimSuffix(trimmed, ":")
			hasInput = false
		}

		// Check if device has input capabilities
		if strings.Contains(trimmed, "Input Channels:") || strings.Contains(trimmed, "Input Source:") {
			hasInput = true
		}
	}

	// Don't forget the last device
	if hasInput && currentDevice != "" {
		found := false
		for _, m := range mics {
			if m.Name == currentDevice {
				found = true
				break
			}
		}
		if !found {
			mics = append(mics, Microphone{
				ID:   currentDevice,
				Name: currentDevice,
			})
		}
	}

	// Always include "default" option
	defaultMic := Microphone{ID: "", Name: "System Default"}
	mics = append([]Microphone{defaultMic}, mics...)

	return mics, nil
}

// findRecBinary locates the rec binary, checking common Homebrew paths
// since bundled macOS apps don't inherit shell PATH
func findRecBinary() string {
	// Check common Homebrew locations first (for bundled app)
	homebrewPaths := []string{
		"/opt/homebrew/bin/rec", // Apple Silicon
		"/usr/local/bin/rec",    // Intel Mac
	}
	for _, p := range homebrewPaths {
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}
	// Fall back to PATH lookup (works in dev mode)
	return "rec"
}

func New(microphone string) *Recorder {
	homeDir, _ := os.UserHomeDir()
	cacheDir := filepath.Join(homeDir, ".cache", "jtt")
	return &Recorder{
		audioPath:  filepath.Join(cacheDir, "recording.wav"),
		pidPath:    filepath.Join(cacheDir, "rec.pid"),
		microphone: microphone,
	}
}

func (r *Recorder) SetMicrophone(mic string) {
	r.microphone = mic
}

func (r *Recorder) AudioPath() string {
	return r.audioPath
}

func (r *Recorder) IsRecording() bool {
	_, err := os.Stat(r.pidPath)
	return err == nil
}

func (r *Recorder) Start() error {
	if r.IsRecording() {
		return nil
	}

	cacheDir := filepath.Dir(r.audioPath)
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		return err
	}

	os.Remove(r.audioPath)

	args := []string{"-q", "-c", "1", "-r", "16000", r.audioPath, "trim", "0", "600"}
	
	r.cmd = exec.Command(findRecBinary(), args...)
	
	// NOTE: On macOS, sox's coreaudio driver does not support device selection via AUDIODEV
	// with device names like "MacBook Pro Microphone". It only uses the system default.
	// Setting an invalid AUDIODEV causes rec to fail silently.
	// For now, we always use the system default audio input.
	r.cmd.Stdout = nil
	r.cmd.Stderr = nil

	if err := r.cmd.Start(); err != nil {
		return err
	}

	return os.WriteFile(r.pidPath, []byte(fmt.Sprintf("%d", r.cmd.Process.Pid)), 0644)
}

func (r *Recorder) Stop() error {
	if !r.IsRecording() {
		return nil
	}

	defer os.Remove(r.pidPath)

	pidData, err := os.ReadFile(r.pidPath)
	if err != nil {
		return err
	}

	pid, err := strconv.Atoi(string(pidData))
	if err != nil {
		return err
	}

	proc, err := os.FindProcess(pid)
	if err != nil {
		return err
	}

	// Send SIGINT (Ctrl+C) instead of SIGTERM - rec handles this better
	// and properly finalizes the WAV file header
	proc.Signal(syscall.SIGINT)

	// Wait for the process to actually exit (up to 2 seconds)
	done := make(chan error, 1)
	go func() {
		_, err := proc.Wait()
		done <- err
	}()

	select {
	case <-done:
		// Process exited cleanly
	case <-time.After(2 * time.Second):
		// Timeout - force kill
		proc.Kill()
	}

	// Give filesystem a moment to sync
	time.Sleep(50 * time.Millisecond)

	return nil
}
