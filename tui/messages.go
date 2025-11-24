package tui

import (
	"github.com/charmbracelet/bubbles/list"
)

// MESSAGES
type contentsMsg struct{ items []list.Item; currentFolderName string } // Add currentFolderName
type errMsg struct{ err error }
type emptyContentsMsg struct{}
type downloadCompleteMsg string
type downloadErrorMsg struct{ err error }
type clipboardCompleteMsg string
type clipboardErrorMsg struct{ err error }
type openMPVCompleteMsg string
type openMPVErrorMsg struct{ err error }
type batchDownloadCompleteMsg string
type batchDownloadErrorMsg struct{ err error }

type progressMsg float64        // New: for progress updates
type progressErrMsg struct{ err error } // New: for progress errors

func (e errMsg) Error() string { return e.err.Error() }
func (e downloadErrorMsg) Error() string { return e.err.Error() }
func (e clipboardErrorMsg) Error() string { return e.err.Error() }
func (e openMPVErrorMsg) Error() string { return e.err.Error() }
func (e batchDownloadErrorMsg) Error() string { return e.err.Error() }
