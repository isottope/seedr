package tui

import "github.com/charmbracelet/lipgloss"

var (
	AppStyle = lipgloss.NewStyle().Padding(1, 2)

	TitleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFDF5")).
			Background(lipgloss.Color("#25A065")).
			Padding(0, 1)

	StatusMessageStyle = lipgloss.NewStyle().
				Foreground(lipgloss.AdaptiveColor{Light: "#04B575", Dark: "#04B575"}).
				Render
)

// MyItemStyles defines styling for a default list item.
type MyItemStyles struct {
	NormalTitle lipgloss.Style
	NormalDesc  lipgloss.Style
	SelectedTitle lipgloss.Style
	SelectedDesc  lipgloss.Style
	DimmedTitle lipgloss.Style
	DimmedDesc  lipgloss.Style
	FilterMatch lipgloss.Style

	FolderTitle lipgloss.Style
	FileTitle   lipgloss.Style
	TorrentTitle lipgloss.Style
}

// NewMyItemStyles returns style definitions for a default item.
func NewMyItemStyles() (s MyItemStyles) {
	// Initialize default styles, similar to list.NewDefaultItemStyles
	s.NormalTitle = lipgloss.NewStyle().
		Foreground(lipgloss.AdaptiveColor{Light: "#1a1a1a", Dark: "#dddddd"}).
		Padding(0, 0, 0, 2)

	s.NormalDesc = s.NormalTitle.
		Foreground(lipgloss.AdaptiveColor{Light: "#A49FA5", Dark: "#777777"})

	s.SelectedTitle = lipgloss.NewStyle().
		Border(lipgloss.NormalBorder(), false, false, false, true).
		BorderForeground(lipgloss.AdaptiveColor{Light: "#F793FF", Dark: "#AD58B4"}).
		Foreground(lipgloss.AdaptiveColor{Light: "#EE6FF8", Dark: "#EE6FF8"}).
		Padding(0, 0, 0, 1)

	s.SelectedDesc = s.SelectedTitle.
		Foreground(lipgloss.AdaptiveColor{Light: "#F793FF", Dark: "#AD58B4"})

	s.DimmedTitle = lipgloss.NewStyle().
		Foreground(lipgloss.AdaptiveColor{Light: "#A49FA5", Dark: "#777777"}).
		Padding(0, 0, 0, 2)

	s.DimmedDesc = s.DimmedTitle.
		Foreground(lipgloss.AdaptiveColor{Light: "#C2B8C2", Dark: "#4D4D4D"})

	s.FilterMatch = lipgloss.NewStyle().Underline(true)

	// Custom type colors
	s.FolderTitle = lipgloss.NewStyle().Foreground(lipgloss.Color("#00BFFF")) // DeepSkyBlue
	s.FileTitle = lipgloss.NewStyle().Foreground(lipgloss.Color("#32CD32"))   // LimeGreen
	s.TorrentTitle = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF4500")) // OrangeRed

	return s
}
