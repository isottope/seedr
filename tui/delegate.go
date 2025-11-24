package tui

import (
	"fmt"
	"io"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
)

const (
	bullet   = "•"
	ellipsis = "…"
)

// itemType describes the type of a list item.
type itemType int

const (
	TypeFolder itemType = iota
	TypeFile
	TypeTorrent
)

// item implements the list.Item interface.
type item struct {
	id       string
	itemType itemType
	title    string
	desc     string
	marked   bool // Add marked field
}

func (i item) FilterValue() string { return i.title }
func (i item) Title() string {
	if i.marked {
		return "✅ " + i.title // Visual indicator for marked items
	}
	return i.title
}
func (i item) Description() string { return i.desc }

// itemDelegate implements list.ItemDelegate to customize rendering
type itemDelegate struct {
	styles MyItemStyles
	keys   *delegateKeyMap // Add delegateKeyMap to the delegate
}

func (d itemDelegate) Height() int                             { return 2 }
func (d itemDelegate) Spacing() int                            { return 1 }
func (d itemDelegate) Update(msg tea.Msg, m *list.Model) tea.Cmd {
	var title string

	if i, ok := m.SelectedItem().(item); ok {
		title = i.Title()
	} else {
		return nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, d.keys.choose): // Use d.keys for choose
			// Handle the choose action. In the main model, this means navigating into a folder or performing an action.
			// Send a custom message to the main model to display below the title.
			return func() tea.Msg { return itemChosenMsg("You chose " + title) }

		case key.Matches(msg, d.keys.remove): // Use d.keys for remove
			// Handle the remove action. This is for deleting an item from the list.
			index := m.Index()
			m.RemoveItem(index)
			if len(m.Items()) == 0 {
				d.keys.remove.SetEnabled(false) // Disable remove key if list is empty
			}
			return m.NewStatusMessage(StatusMessageStyle("Deleted " + title))
		}
	}
	return nil
}

func (d itemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	var (
		title, desc  string
		matchedRunes []int
	)

	i, ok := listItem.(item);
	if !ok {
		return
	}

	title = i.Title()
	desc = i.Description()

	if m.Width() <= 0 {
		return
	}

	// Conditions
	var (
		isSelected  = index == m.Index()
		emptyFilter = m.FilterState() == list.Filtering && m.FilterValue() == ""
		isFiltered  = m.FilterState() == list.Filtering || m.FilterState() == list.FilterApplied
	)

	// Determine base styles for title and description
	var currentTitleStyle, currentDescStyle lipgloss.Style

	if emptyFilter {
		currentTitleStyle = d.styles.DimmedTitle
		currentDescStyle = d.styles.DimmedDesc
	} else if isSelected && m.FilterState() != list.Filtering {
		currentTitleStyle = d.styles.SelectedTitle
		currentDescStyle = d.styles.SelectedDesc
	} else {
		currentTitleStyle = d.styles.NormalTitle
		currentDescStyle = d.styles.NormalDesc
	}

	// Apply type-specific colors to the title, only if not selected and not dimmed
	if !isSelected && !emptyFilter {
		switch i.itemType {
		case TypeFolder:
			currentTitleStyle = currentTitleStyle.Inherit(d.styles.FolderTitle)
		case TypeFile:
			currentTitleStyle = currentTitleStyle.Inherit(d.styles.FileTitle)
		case TypeTorrent:
			currentTitleStyle = currentTitleStyle.Inherit(d.styles.TorrentTitle)
		}
	}

	// Handle filter matches
	if isFiltered && index < len(m.VisibleItems()) {
		matchedRunes = m.MatchesForItem(index)
		unmatched := currentTitleStyle.Inline(true)
		matched := unmatched.Inherit(d.styles.FilterMatch)
		title = lipgloss.StyleRunes(title, matchedRunes, matched, unmatched)
	}

	// Truncate title and description
	// Use lipgloss.Width to properly calculate widths of styled strings
	textWidth := m.Width() - currentTitleStyle.GetPaddingLeft() - currentTitleStyle.GetPaddingRight()
	if textWidth < 0 { textWidth = 0 } // Ensure non-negative width

	title = ansi.Truncate(title, textWidth, ellipsis)

	desc = ansi.Truncate(desc, textWidth, ellipsis)

	// Print title and description
	fmt.Fprintf(w, "%s\n%s", currentTitleStyle.Render(title), currentDescStyle.Render(desc)) //nolint: errcheck
}

type delegateKeyMap struct {
	choose key.Binding
	remove key.Binding
}

func newDelegateKeyMap() *delegateKeyMap {
	return &delegateKeyMap{
		choose: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "choose"),
		),
		remove: key.NewBinding(
			key.WithKeys("x", "backspace"),
			key.WithHelp("x", "delete"),
		),
	}
}