package tui

import (
	"context"
	"fmt"
)

// RunTUI is the exported function to start the TUI.
func RunTUI(client *seedr.Client) error {
	p := tea.NewProgram(newModel(client), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		return fmt.Errorf("error running program: %w", err)
	}
	return nil
}