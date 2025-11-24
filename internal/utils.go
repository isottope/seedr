package internal

import (
	"github.com/dustin/go-humanize"
)

// humanReadableBytes converts a byte count into a human-readable string.
func HumanReadableBytes(byteCount int) string {
	return humanize.Bytes(uint64(byteCount))
}