package internal

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sync"
	
	"github.com/dustin/go-humanize"
)

// humanReadableBytes converts a byte count into a human-readable string.
func HumanReadableBytes(byteCount int) string {
	return humanize.Bytes(uint64(byteCount))
}

// CopyToClipboard copies the given text to the system clipboard.
// It uses "xclip" for Linux and "pbcopy" for macOS.
func CopyToClipboard(text string) error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "linux":
		cmd = exec.Command("xclip", "-selection", "clipboard")
	case "darwin":
		cmd = exec.Command("pbcopy")
	default:
		return fmt.Errorf("unsupported operating system for clipboard operations: %s", runtime.GOOS)
	}

	cmd.Stdin = bytes.NewReader([]byte(text))
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to copy to clipboard: %w", err)
	}
	return nil
}

// Logger provides a flexible logging mechanism.
type Logger struct {
	fileLogger    *log.Logger
	consoleLogger *log.Logger
	debugMode     bool
	logToFile     bool
	logToConsole  bool
	mode          string // "TUI" or "CLI"
	mu            sync.Mutex
}

// Log is the global logger instance.
var Log *Logger

// NewLogger initializes and returns a new Logger instance.
func NewLogger(debugMode, isTUI bool) *Logger {
	l := &Logger{
		debugMode:    debugMode,
		logToConsole: true, // Always log to console in debug mode for CLI
	}

	if isTUI {
		l.mode = "TUI"
		l.logToConsole = false // Don't log to console for TUI by default
		if debugMode {
			l.logToFile = true // Log to file if TUI is in debug mode
		}
	} else {
		l.mode = "CLI"
		l.logToConsole = true // Always log to console for CLI
		if debugMode {
			l.logToFile = true // Log to file if CLI is in debug mode
		}
	}

	if l.logToConsole {
		l.consoleLogger = log.New(os.Stderr, fmt.Sprintf("[%s] ", l.mode), log.Ldate|log.Ltime|log.Lshortfile)
	}

	if l.logToFile {
		logDirPath := filepath.Join(os.Getenv("HOME"), ".local", "share", "logs")
		if err := os.MkdirAll(logDirPath, 0755); err != nil {
			log.Printf("Failed to create log directory: %v", err)
			return l // Return logger without file logging
		}
		logFilePath := filepath.Join(logDirPath, "seedr.log")

		file, err := os.OpenFile(logFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			log.Printf("Failed to open log file %s: %v", logFilePath, err)
			return l // Return logger without file logging
		}
		l.fileLogger = log.New(file, fmt.Sprintf("[%s] ", l.mode), log.Ldate|log.Ltime|log.Lshortfile)
	}

	// If no loggers are active, set a dummy logger
	if !l.logToConsole && !l.logToFile {
		l.consoleLogger = log.New(io.Discard, "", 0) // Discard all output
		l.fileLogger = log.New(io.Discard, "", 0)
	} else if !l.logToConsole {
		l.consoleLogger = log.New(io.Discard, "", 0)
	} else if !l.logToFile {
		l.fileLogger = log.New(io.Discard, "", 0)
	}


	return l
}

// Debug logs a debug message if debug mode is enabled.
func (l *Logger) Debug(format string, v ...interface{}) {
	if l.debugMode {
		l.mu.Lock()
		defer l.mu.Unlock()

		msg := fmt.Sprintf(format, v...)

		if l.logToConsole && l.consoleLogger != nil {
			l.consoleLogger.Output(2, "DEBUG: "+msg)
		}
		if l.logToFile && l.fileLogger != nil {
			l.fileLogger.Output(2, "DEBUG: "+msg)
		}
	}
}

// Close closes any open file handles.
func (l *Logger) Close() {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.fileLogger != nil {
		if f, ok := l.fileLogger.Writer().(*os.File); ok {
			f.Close()
		}
	}
}

