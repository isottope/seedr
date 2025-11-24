package tui

import (
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
)

type KeyMap struct {
	list.KeyMap

	// Custom commands from original tui
	Download key.Binding
	CopyURL  key.Binding
	OpenMPV  key.Binding
	Mark     key.Binding
	Retry    key.Binding
	Enter    key.Binding
	Back     key.Binding

	// Commands from list-fancy example
	ToggleSpinner    key.Binding
	ToggleTitleBar   key.Binding
	ToggleStatusBar  key.Binding
	TogglePagination key.Binding
	ToggleHelpMenu   key.Binding
	InsertItem       key.Binding // Consider if this is truly needed for Seedr context
}

// ShortHelp returns a slice of key.Binding that provides a brief
// help menu for the current view.
func (k KeyMap) ShortHelp() []key.Binding {
	return []key.Binding{
		k.Quit,
		k.Enter,
		k.Back,
		k.Download,
		k.Mark,
		k.Retry,
		k.CopyURL,
		k.OpenMPV,
		k.Filter,
		// list-fancy keys for short help
		k.ToggleHelpMenu,
	}
}

// FullHelp returns a slice of slices of key.Binding that provides a
// more detailed help menu.
func (k KeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Quit, k.Enter, k.Back, k.Download, k.Mark, k.Retry, k.CopyURL, k.OpenMPV},
		{k.CursorUp, k.CursorDown, k.GoToStart, k.GoToEnd},
		{k.Filter, k.ClearFilter, k.CancelWhileFiltering, k.AcceptWhileFiltering},
		// list-fancy keys for full help
		{k.ToggleSpinner, k.ToggleTitleBar, k.ToggleStatusBar, k.TogglePagination, k.ToggleHelpMenu},
		{k.InsertItem}, // Keep insertItem for completeness if it needs to be there
	}
}

// DefaultKeyMap defines the custom keybindings.
var DefaultKeyMap = KeyMap{
	KeyMap: list.DefaultKeyMap(), // Initialize the embedded default keymap

	// Custom commands from original tui
	Download: key.NewBinding(
		key.WithKeys("d"),
		key.WithHelp("d", "download"),
	),
	CopyURL: key.NewBinding(
		key.WithKeys("c"),
		key.WithHelp("c", "copy URL"),
	),
	OpenMPV: key.NewBinding(
		key.WithKeys("o"),
		key.WithHelp("o", "open MPV"),
	),
	Mark: key.NewBinding(
		key.WithKeys("m"),
		key.WithHelp("m", "mark/unmark"),
	),
	Retry: key.NewBinding(
		key.WithKeys("r"),
		key.WithHelp("r", "retry"),
	),
	Enter: key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("enter", "select/open"),
	),
	Back: key.NewBinding(
		key.WithKeys("backspace", "h", "left"),
		key.WithHelp("backspace", "go back"),
	),

	// Commands from list-fancy example
	ToggleSpinner: key.NewBinding(
		key.WithKeys("s"),
		key.WithHelp("s", "toggle spinner"),
	),
	ToggleTitleBar: key.NewBinding(
		key.WithKeys("T"),
		key.WithHelp("T", "toggle title"),
	),
	ToggleStatusBar: key.NewBinding(
		key.WithKeys("S"),
		key.WithHelp("S", "toggle status"),
	),
	TogglePagination: key.NewBinding(
		key.WithKeys("P"),
		key.WithHelp("P", "toggle pagination"),
	),
	ToggleHelpMenu: key.NewBinding(
		key.WithKeys("H"),
		key.WithHelp("H", "toggle help"),
	),
	InsertItem: key.NewBinding(
		key.WithKeys("a"),
		key.WithHelp("a", "add item"),
	),
}