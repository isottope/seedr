package tui

import (
	"fmt"
	"strings"
	"time" // Import time package

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea" // Corrected: Added alias 'tea'
	"seedr/internal" // Import internal package
	"seedr/pkg/seedr"
	"github.com/charmbracelet/lipgloss"
)

// appState describes the current state of the application.
type appState int

const (
	stateLoading appState = iota
	stateDownloading
	stateReady
	stateError
	stateEmpty
)

type (
	itemChosenMsg string
	clearChosenMessageMsg struct{}
)

func clearChosenMessageAfter(d time.Duration) tea.Cmd {
	return tea.Tick(d, func(t time.Time) tea.Msg {
		return clearChosenMessageMsg{}
	})
}

type model struct {
	list            list.Model
	spinner         spinner.Model
	progress        progress.Model
	quitting        bool
	state           appState
	err             error
	client          *seedr.Client
	folderHistory   []string
	currentFolderID string
	contentCache    map[string]contentsMsg
	markedFiles     map[string]item // Map to store marked files by their ID
	currentFolderPath string // Stores the current folder's path in a Linux-like format
	chosenMessage   string // New field to display messages below the title
	keys            KeyMap
}

func newModel(client *seedr.Client) model {
	s := spinner.New()
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("69"))
	s.Spinner = spinner.Dot
	myStyles := NewMyItemStyles() // From styles.go
	itemDel := itemDelegate{styles: myStyles, keys: newDelegateKeyMap()} // Initialize delegate with keys

	// Initialize the list component
	l := list.New([]list.Item{}, itemDel, 0, 0)
	l.Title = "Loading..." // Initial title
	l.SetShowStatusBar(false) // Disable default status bar
	l.SetFilteringEnabled(true)
	l.KeyMap = DefaultKeyMap.KeyMap // Assign the embedded list.KeyMap from keys.go
	l.Styles.Title = TitleStyle // Use TitleStyle from styles.go
	l.AdditionalFullHelpKeys = func() []key.Binding {
		return []key.Binding{
			DefaultKeyMap.Download,
			DefaultKeyMap.CopyURL,
			DefaultKeyMap.OpenMPV,
			DefaultKeyMap.Mark,
			DefaultKeyMap.Retry,
			DefaultKeyMap.ToggleSpinner,
			DefaultKeyMap.ToggleTitleBar,
			DefaultKeyMap.ToggleStatusBar,
			DefaultKeyMap.TogglePagination,
			DefaultKeyMap.ToggleHelpMenu,
		}
	}

	return model{
		list:            l,
		spinner:         s,
		progress:        progress.New(progress.WithDefaultGradient()), // Initialize progress model
		state:           stateLoading,
		client:          client,
		folderHistory:   []string{"0"}, // Start at root folder "0"
		currentFolderID: "0",
		contentCache:    make(map[string]contentsMsg),
		markedFiles:     make(map[string]item), // Initialize the map
		currentFolderPath: "/",
		chosenMessage:   "", // Initialize chosenMessage
		keys:            DefaultKeyMap, // Assign the DefaultKeyMap from keys.go
	}
}

func (m model) Init() tea.Cmd {
	return tea.Batch(m.spinner.Tick, fetchContents(m.client, m.currentFolderID))
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.list.SetWidth(msg.Width)
		// Calculate the height considering the appStyle padding
		h, v := AppStyle.GetFrameSize() // Use AppStyle from styles.go
		m.list.SetSize(msg.Width-h, msg.Height-v)
		return m, nil

	case itemChosenMsg:
		m.chosenMessage = string(msg)
		return m, clearChosenMessageAfter(2 * time.Second) // Clear message after 2 seconds

	case clearChosenMessageMsg:
		m.chosenMessage = ""
		return m, nil

	case tea.KeyMsg:
		// Don't match any of the keys below if we're actively filtering.
		if m.list.FilterState() == list.Filtering {
			break
		}

		var cmd tea.Cmd
		switch {
		// General Keys (Seedr-specific)
		case key.Matches(msg, m.keys.Quit):
			m.quitting = true
			return m, tea.Quit

		case key.Matches(msg, m.keys.Retry):
			if m.state == stateError || m.state == stateEmpty {
				m.state = stateLoading
				m.err = nil
				// Always fetch on retry to ensure fresh data and clear cache for this folder
				delete(m.contentCache, m.currentFolderID)
				return m, tea.Batch(m.spinner.Tick, fetchContents(m.client, m.currentFolderID))
			}

		case key.Matches(msg, m.keys.Enter):
			if m.state == stateReady {
				selectedItem := m.list.SelectedItem()
				if selectedItem == nil {
					return m, nil
				}
				item := selectedItem.(item)
				if item.itemType == TypeFolder {
					m.folderHistory = append(m.folderHistory, m.currentFolderID) // Push current folder to history
					m.currentFolderID = item.id
					// Append new folder to path, ensuring it's always rooted
					if m.currentFolderPath == "/" { // If currently at root, special handling for first folder
						m.currentFolderPath = item.title
					} else {
						m.currentFolderPath = m.currentFolderPath + "/" + item.title
					}
				
					if _, ok := m.contentCache[m.currentFolderID]; ok {
						// Found in cache, use it immediately
						m.state = stateReady
						m.list.Title = "SEEDR" + " " + m.currentFolderPath // Update title with current path
						m.list.Select(0) // Reset cursor to top
						return m, nil
					} else {
						// Not in cache, fetch
						m.state = stateLoading
						m.list.Title = "SEEDR" + " " + m.currentFolderPath // Update title with current path
						m.list.Select(0) // Reset cursor to top when entering a new folder
						return m, tea.Batch(m.spinner.Tick, fetchContents(m.client, m.currentFolderID))
					}
				}
			}

		case key.Matches(msg, m.keys.Back):
			if m.state == stateReady && len(m.folderHistory) > 1 {
				internal.Log.Debug("Back key pressed. Current Folder ID: %s, History: %v", m.currentFolderID, m.folderHistory)

				// Pop from history
				m.folderHistory = m.folderHistory[:len(m.folderHistory)-1]
				prevFolderID := m.folderHistory[len(m.folderHistory)-1]

				// Update current folder path
				if m.currentFolderPath != "/" {
					parts := strings.Split(m.currentFolderPath, "/")
					if len(parts) > 1 {
						// Reconstruct path by removing the last segment
						m.currentFolderPath = strings.Join(parts[:len(parts)-1], "/")
						if m.currentFolderPath == "" { // Handle case where path becomes empty string after split
							m.currentFolderPath = "/"
						}
					} else {
						m.currentFolderPath = "/" // Fallback to root if path logic somehow fails
					}
				}
				internal.Log.Debug("Going back to Folder ID: %s, Path: %s", prevFolderID, m.currentFolderPath)

				m.currentFolderID = prevFolderID // Set to new current
				
				if cachedContents, ok := m.contentCache[m.currentFolderID]; ok {
					// Found in cache, use it immediately
					m.state = stateReady
					m.list.SetItems(cachedContents.items)
					m.list.Title = "SEEDR" + " " + m.currentFolderPath // Always update title
					m.list.Select(0) // Reset cursor to top
					return m, nil
				} else {
					// Not in cache, fetch
					m.state = stateLoading
					m.list.Title = "SEEDR" + " " + m.currentFolderPath // Always update title
					m.list.Select(0) // Reset cursor to top when going back
					return m, tea.Batch(m.spinner.Tick, fetchContents(m.client, m.currentFolderID))
				}
			}

		case key.Matches(msg, m.keys.Mark):
			if m.state == stateReady {
				selectedItem := m.list.SelectedItem()
				if selectedItem == nil {
					return m, nil
				}
				selectedListItem := selectedItem.(item)
				
				if selectedListItem.itemType == TypeFile {
					selectedListItem.marked = !selectedListItem.marked // Toggle marked status
					if selectedListItem.marked {
						m.markedFiles[selectedListItem.id] = selectedListItem // Add to marked files
					} else {
						delete(m.markedFiles, selectedListItem.id) // Remove from marked files
					}
					// Update the item in the list
					items := m.list.Items()
					for i, listItem := range items { // Ensure that all marked files are updated
						if it, ok := listItem.(item); ok && it.id == selectedListItem.id {
							items[i] = it // Assign the updated 'it' to the slice
							break
						}
					}
					m.list.SetItems(items)
				}
			}

		case key.Matches(msg, m.keys.Download):
			if m.state == stateReady {
				if len(m.markedFiles) > 0 {
					// Batch download marked files
					filesToDownload := make([]item, 0, len(m.markedFiles))
					for _, markedFile := range m.markedFiles {
						filesToDownload = append(filesToDownload, markedFile)
					}
					m.state = stateDownloading // Show spinner and progress bar while batch downloading
					// Reset progress bar to 0 when starting a new download
					m.progress = progress.New(progress.WithDefaultGradient())
					return m, tea.Batch(m.spinner.Tick, cmdBatchDownloadFiles(m.client, filesToDownload))
				} else {
					// Single file download
					selectedItem := m.list.SelectedItem()
					if selectedItem == nil {
						return m, nil
					}
					item := selectedItem.(item)
					if item.itemType == TypeFile {
						m.state = stateDownloading // Show spinner and progress bar while downloading
						// Reset progress bar to 0 when starting a new download
						m.progress = progress.New(progress.WithDefaultGradient())
						return m, tea.Batch(m.spinner.Tick, cmdDownloadFile(m.client, item.id, item.title))
					}
				}
			}
		case key.Matches(msg, m.keys.CopyURL):
			if m.state == stateReady {
				if len(m.markedFiles) > 0 {
					return m, m.list.NewStatusMessage(StatusMessageStyle("cannot copy download link when files are marked for batch operations"))
				}
				selectedItem := m.list.SelectedItem()
				if selectedItem == nil {
					return m, nil
				}
				item := selectedItem.(item)
				if item.itemType == TypeFile {
					internal.Log.Debug("CopyURL: Attempting to copy URL for file ID: %s", item.id)
					m.state = stateLoading // Show spinner
					return m, tea.Batch(m.spinner.Tick, cmdCopyURL(m.client, item.id))
				}
			}
		case key.Matches(msg, m.keys.OpenMPV):
			if m.state == stateReady {
				if len(m.markedFiles) > 0 {
					return m, m.list.NewStatusMessage(StatusMessageStyle("Cannot open with MPV when files are marked for batch operations"))
				}
				selectedItem := m.list.SelectedItem()
				if selectedItem == nil {
					return m, nil
				}
				item := selectedItem.(item)
				if item.itemType == TypeFile {
					m.state = stateLoading // Show spinner
					return m, tea.Batch(m.spinner.Tick, cmdOpenMPV(m.client, item.id))
				}
			}

		// List-fancy specific keybindings (toggles)
		case key.Matches(msg, m.keys.ToggleSpinner):
			cmd = m.list.ToggleSpinner()
			return m, cmd

		case key.Matches(msg, m.keys.ToggleTitleBar):
			v := !m.list.ShowTitle()
			m.list.SetShowTitle(v)
			m.list.SetShowFilter(v)
			m.list.SetFilteringEnabled(v)
			return m, nil

		case key.Matches(msg, m.keys.ToggleStatusBar):
			m.list.SetShowStatusBar(!m.list.ShowStatusBar())
			return m, nil

		case key.Matches(msg, m.keys.TogglePagination):
			m.list.SetShowPagination(!m.list.ShowPagination())
			return m, nil

		case key.Matches(msg, m.keys.ToggleHelpMenu):
			m.list.SetShowHelp(!m.list.ShowHelp())
			return m, nil

		// The InsertItem key from list-fancy, which added random items, is not integrated
		// as Seedr functionality revolves around existing files.
		}

		// Allow the list to handle its own key presses
		m.list, cmd = m.list.Update(msg)
		return m, cmd

	case spinner.TickMsg:
		var cmd tea.Cmd
		if m.state == stateLoading {
			m.spinner, cmd = m.spinner.Update(msg)
		}
		return m, cmd

	case contentsMsg:
		m.state = stateReady
		m.list.SetItems(msg.items)
		// Set title to current path, which is updated on enter/backspace
		m.list.Title = m.currentFolderPath
		m.contentCache[m.currentFolderID] = msg // Cache the fetched contents
		// If len(msg.items) == 0, emptyContentsMsg would have been returned instead
		return m, nil

	case errMsg:
		m.state = stateError
		m.err = msg.err
		return m, nil

	case progressErrMsg: // New: Handle errors from progress reporting
		m.state = stateError
		m.err = msg.err
		return m, nil

	case emptyContentsMsg:
		m.state = stateEmpty
		return m, nil

	case downloadCompleteMsg:
		m.state = stateReady // Return to ready state after download attempt
		// Optionally display a temporary status message
		m.err = nil // Clear any previous error
		return m, m.list.NewStatusMessage(fmt.Sprintf("Downloaded: %s", string(msg)))
	case downloadErrorMsg:
		m.state = stateError
		m.err = msg.err
		return m, nil

	case batchDownloadCompleteMsg:
		m.state = stateReady
		m.err = nil
		return m, m.list.NewStatusMessage(string(msg))
	case batchDownloadErrorMsg:
		m.state = stateError
		m.err = msg.err
		return m, nil

	case clipboardCompleteMsg:
		m.state = stateReady // Return to ready state
		m.err = nil
		m.chosenMessage = string(msg) // Set the chosen message
		return m, clearChosenMessageAfter(2 * time.Second) // Clear message after 2 seconds
	case clipboardErrorMsg:
		m.state = stateError
		m.err = msg.err
		return m, nil

	case openMPVCompleteMsg:
		m.state = stateReady // Return to ready state
		m.err = nil
		return m, m.list.NewStatusMessage(string(msg))
	case openMPVErrorMsg:
		m.state = stateError
		m.err = msg.err
		return m, nil

	case progressMsg: // New: Handle progress updates
		internal.Log.Debug("Received progressMsg: %.2f", float64(msg)*100)
		var cmd tea.Cmd
		var updatedProgressModel tea.Model
		updatedProgressModel, cmd = m.progress.Update(msg)
		m.progress = updatedProgressModel.(progress.Model) // Type assertion here
		return m, cmd

	default:
		var cmd tea.Cmd
		m.list, cmd = m.list.Update(msg)
		return m, cmd
	}

	return m, nil
}

func (m model) View() string {
	if m.quitting {
		return ""
	}
	
	var viewString string
	switch m.state {
	case stateLoading:
		viewString = fmt.Sprintf("%s Loading Contents...", m.spinner.View())
	case stateDownloading:
		viewString = fmt.Sprintf("%s Downloading... %s", m.spinner.View(), m.progress.View())
	case stateError:
		viewString = fmt.Sprintf("Error: %v\n\nPress 'r' to retry, 'q' to quit.", m.err)
	case stateReady:
		var s strings.Builder
		s.WriteString(m.list.View())
		if m.chosenMessage != "" {
			s.WriteString("\n" + StatusMessageStyle(m.chosenMessage))
		}
		viewString = s.String()
	case stateEmpty:
		viewString = "No contents Found in this Folder.\n\nPress 'r' to retry, 'backspace' to go back, 'q' to quit."
	default:
		viewString = "I Don't Even Know What's Going On :("
	}
	return AppStyle.Render(viewString) // Wrap all views with AppStyle
}

// RunTUI is the exported function to start the TUI.
func RunTUI(client *seedr.Client) error {
	p := tea.NewProgram(newModel(client), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		return fmt.Errorf("error running program: %w", err)
	}
	return nil
}
