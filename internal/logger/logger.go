package logger

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"
)

var logFile *os.File

func Init() error {
	homeDir, _ := os.UserHomeDir()
	logDir := filepath.Join(homeDir, ".cache", "jtt")
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return err
	}

	logPath := filepath.Join(logDir, "jtt.log")
	var err error
	logFile, err = os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return err
	}

	log.SetOutput(logFile)
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
	
	log.Printf("=== JTT started at %s ===", time.Now().Format(time.RFC3339))
	return nil
}

func Close() {
	if logFile != nil {
		logFile.Close()
	}
}

func Error(format string, v ...interface{}) {
	log.Printf("[ERROR] "+format, v...)
}

func Info(format string, v ...interface{}) {
	log.Printf("[INFO] "+format, v...)
}

func Debug(format string, v ...interface{}) {
	log.Printf("[DEBUG] "+format, v...)
}

func LogPath() string {
	homeDir, _ := os.UserHomeDir()
	return filepath.Join(homeDir, ".cache", "jtt", "jtt.log")
}

func GetRecentLogs(lines int) string {
	data, err := os.ReadFile(LogPath())
	if err != nil {
		return fmt.Sprintf("Error reading logs: %v", err)
	}
	
	content := string(data)
	// Return last N lines
	allLines := splitLines(content)
	if len(allLines) <= lines {
		return content
	}
	start := len(allLines) - lines
	result := ""
	for i := start; i < len(allLines); i++ {
		result += allLines[i] + "\n"
	}
	return result
}

func splitLines(s string) []string {
	var lines []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == '\n' {
			lines = append(lines, s[start:i])
			start = i + 1
		}
	}
	if start < len(s) {
		lines = append(lines, s[start:])
	}
	return lines
}
