package media

import (
	"os/exec"
	"strings"
)

// findNowPlayingCLI locates the nowplaying-cli binary
func findNowPlayingCLI() string {
	paths := []string{
		"/opt/homebrew/bin/nowplaying-cli",
		"/usr/local/bin/nowplaying-cli",
	}
	for _, p := range paths {
		if _, err := exec.LookPath(p); err == nil {
			return p
		}
	}
	return ""
}

// IsPlaying checks if media is currently playing
func IsPlaying() bool {
	cli := findNowPlayingCLI()
	if cli == "" {
		return false
	}

	out, err := exec.Command(cli, "get", "playbackRate").Output()
	if err != nil {
		return false
	}

	// playbackRate of 1 means playing, 0 means paused
	return strings.TrimSpace(string(out)) == "1"
}

// Pause pauses the currently playing media
func Pause() {
	cli := findNowPlayingCLI()
	if cli == "" {
		return
	}
	exec.Command(cli, "pause").Run()
}

// Play resumes media playback
func Play() {
	cli := findNowPlayingCLI()
	if cli == "" {
		return
	}
	exec.Command(cli, "play").Run()
}

// IsAvailable checks if nowplaying-cli is installed
func IsAvailable() bool {
	return findNowPlayingCLI() != ""
}
