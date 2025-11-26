package internal

import (
	"bytes"
	"fmt"
	"os/exec"
	"runtime"

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
