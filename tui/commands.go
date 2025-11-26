package tui

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"time"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"

	"seedr/internal" // Added for internal.Log.Debug
	"seedr/pkg/seedr"
)

// COMMANDS
func fetchContents(client *seedr.Client, folderID string) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if client == nil {
			internal.Log.Debug("Seedr client is nil in fetchContents")
			return errMsg{err: errors.New("Seedr client is not initialized")}
		}
		internal.Log.Debug("Seedr client is not nil in fetchContents. Fetching folderID: %s", folderID)

		contents, err := client.ListContents(ctx, folderID) // Use ListContents for the given folderID
		if err != nil {
			internal.Log.Debug("client.ListContents error for folderID %s: %v", folderID, err)
			return errMsg{err: fmt.Errorf("failed to fetch contents for folder %s: %w", folderID, err)}
		}
		internal.Log.Debug("client.ListContents returned %d folders, %d files, %d torrents for folderID %s",
			len(contents.Folders), len(contents.Files), len(contents.Torrents), folderID)

		var allItems []list.Item

		// Add folders
		for _, f := range contents.Folders {
			allItems = append(allItems, item{
				id:       fmt.Sprintf("%d", f.ID),
				itemType: TypeFolder,
				title:    f.Name,
				desc:     fmt.Sprintf("Folder | Size: %.2fGB | Last Update: %s", float64(f.Size)/(1024*1024*1024), func() string {
					if f.LastUpdate != nil {
						return f.LastUpdate.Format("2006-01-02 15:04:05")
					}
					return "N/A"
				}()),
			})
		}

		// Add files
		for _, f := range contents.Files {
			allItems = append(allItems, item{
				id:       fmt.Sprintf("%d", f.FolderFileID), // Corrected: Use f.FolderFileID for file IDs
				itemType: TypeFile,
				title:    f.Name,
				desc:     fmt.Sprintf("File | Size: %.2fGB | Last Update: %s", float64(f.Size)/(1024*1024*1024), func() string {
					if f.LastUpdate != nil {
						return f.LastUpdate.Format("2006-01-02 15:04:05")
					}
					return "N/A"
				}()),
			})
		}
		
		// Add torrents (if any are directly in this folder)
		for _, t := range contents.Torrents {
			allItems = append(allItems, item{
				id:       fmt.Sprintf("%d", t.ID),
				itemType: TypeTorrent,
				title:    t.Name,
				desc:     fmt.Sprintf("Torrent | Size: %.2fGB | Last Update: %s", float64(t.Size)/(1024*1024*1024), func() string {
					if t.LastUpdate != nil {
						return t.LastUpdate.Format("2006-01-02 15:04:05")
					}
					return "N/A"
				}()),
			})
		}

		if len(allItems) == 0 {
			internal.Log.Debug("No items found (folders, files, or torrents) for folderID %s, returning emptyContentsMsg", folderID)
			return emptyContentsMsg{} // Return new message type
		}

		internal.Log.Debug("Returning contentsMsg with %d items (folders, files, and torrents) for folderID %s", len(allItems), folderID)
		return contentsMsg{items: allItems, currentFolderName: contents.Name}
	}
}

func cmdDownloadFile(client *seedr.Client, fileID string, fileName string) tea.Cmd {
	return func() tea.Msg {
		msgChan := make(chan tea.Msg)

		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			fileResult, err := client.FetchFile(ctx, fileID)
			if err != nil {
				msgChan <- downloadErrorMsg{err: fmt.Errorf("failed to get download URL for %s: %w", fileName, err)}
				return
			}

			resp, err := http.Get(fileResult.URL) // nolint:gosec
			if err != nil {
				msgChan <- downloadErrorMsg{err: fmt.Errorf("failed to start download for %s: %w", fileName, err)}
				return
			}
			defer resp.Body.Close()

			if resp.ContentLength <= 0 { // Check ContentLength for progress
				msgChan <- downloadErrorMsg{err: fmt.Errorf("cannot track progress for %s: content length unknown", fileName)}
				return
			}

			outFile, err := os.Create(fileName)
			if err != nil {
				msgChan <- downloadErrorMsg{err: fmt.Errorf("failed to create local file %s: %w", fileName, err)}
				return
			}
			defer outFile.Close()

			totalSize := resp.ContentLength
			downloadedBytes := int64(0)
			buffer := make([]byte, 32*1024) // 32KB buffer

			for {
				n, readErr := resp.Body.Read(buffer)
				if n > 0 {
					_, writeErr := outFile.Write(buffer[:n])
					if writeErr != nil {
						msgChan <- downloadErrorMsg{err: fmt.Errorf("failed to write to file %s: %w", fileName, writeErr)}
						return
					}
					downloadedBytes += int64(n)
					internal.Log.Debug("Download progress: %.2f%%", float64(downloadedBytes)/float64(totalSize)*100)
					msgChan <- progressMsg(float64(downloadedBytes) / float64(totalSize))
				}
				if readErr == io.EOF {
					msgChan <- downloadCompleteMsg(fileName)
					return
				}
				if readErr != nil {
					msgChan <- downloadErrorMsg{err: fmt.Errorf("failed to read download stream for %s: %w", fileName, readErr)}
					return
				}
			}
		}()

		return func() tea.Msg {
			return <-msgChan
		}
	}
}

func cmdCopyURL(client *seedr.Client, fileID string) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		fileResult, err := client.FetchFile(ctx, fileID)
		if err != nil {
			return clipboardErrorMsg{err: fmt.Errorf("failed to get download URL for clipboard: %w", err)}
		}

		// Attempt to copy to clipboard using wl-copy (Wayland) or xclip (X11)
		copyCmd := exec.Command("wl-copy", fileResult.URL)
		if err := copyCmd.Run(); err != nil {
			copyCmd = exec.Command("xclip", "-selection", "clipboard")
			stdin, _ := copyCmd.StdinPipe()
			go func() {
				defer stdin.Close()
				io.WriteString(stdin, fileResult.URL)
			}()
			if err := copyCmd.Run(); err != nil {
				return clipboardErrorMsg{err: fmt.Errorf("failed to copy URL to clipboard: %w", err)}
			}
		}
		return clipboardCompleteMsg("URL copied to clipboard!")
	}
}

func cmdOpenMPV(client *seedr.Client, fileID string) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		fileResult, err := client.FetchFile(ctx, fileID)
		if err != nil {
			return openMPVErrorMsg{err: fmt.Errorf("failed to get download URL for MPV: %w", err)}
		}

		// Run mpv as a detached process
		cmd := exec.Command("mpv", fileResult.URL)
		err = cmd.Start() // Use Start to run in background
		if err != nil {
			return openMPVErrorMsg{err: fmt.Errorf("failed to start mpv: %w", err)}
		}
		// We don't wait for it to finish, just that it started
		return openMPVCompleteMsg(fmt.Sprintf("Opening %s with mpv...", fileResult.URL))
	}
}

func cmdBatchDownloadFiles(client *seedr.Client, files []item) tea.Cmd {
	return func() tea.Msg {
		msgChan := make(chan tea.Msg)

		go func() {
			var batchErrors []error
			for _, file := range files {
				ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second) // Longer timeout per file
				fileResult, err := client.FetchFile(ctx, file.id)
				if err != nil {
					batchErrors = append(batchErrors, downloadErrorMsg{err: fmt.Errorf("failed to get download URL for %s: %w", file.title, err)}.err)
					cancel()
					continue // Move to next file
				}
				
				resp, err := http.Get(fileResult.URL)
				if err != nil {
					batchErrors = append(batchErrors, downloadErrorMsg{err: fmt.Errorf("failed to start download for %s: %w", file.title, err)}.err)
					cancel()
					continue // Move to next file
				}
				defer resp.Body.Close() // Defer inside the loop so it's called for each response

				if resp.ContentLength <= 0 {
					batchErrors = append(batchErrors, downloadErrorMsg{err: fmt.Errorf("cannot track progress for %s: content length unknown", file.title)}.err)
					cancel()
					continue // Move to next file
				}

				outFile, err := os.Create(file.title)
				if err != nil {
					batchErrors = append(batchErrors, downloadErrorMsg{err: fmt.Errorf("failed to create local file %s: %w", file.title, err)}.err)
					cancel()
					continue // Move to next file
				}
				defer outFile.Close()

				totalSize := resp.ContentLength
				downloadedBytes := int64(0)
				buffer := make([]byte, 32*1024)

				for {
					n, readErr := resp.Body.Read(buffer)
					if n > 0 {
						_, writeErr := outFile.Write(buffer[:n])
						if writeErr != nil {
							batchErrors = append(batchErrors, downloadErrorMsg{err: fmt.Errorf("failed to write to file %s: %w", file.title, writeErr)}.err)
							break // Break inner loop, move to next file
						}
						downloadedBytes += int64(n)
						internal.Log.Debug("Batch download progress for %s: %.2f%%", file.title, float64(downloadedBytes)/float64(totalSize)*100)
						msgChan <- progressMsg(float64(downloadedBytes) / float64(totalSize))
					}
					if readErr == io.EOF {
						msgChan <- downloadCompleteMsg(file.title)
						break // Current file download complete
					}
					if readErr != nil {
						batchErrors = append(batchErrors, downloadErrorMsg{err: fmt.Errorf("failed to read download stream for %s: %w", file.title, readErr)}.err)
						break // Error reading current file, move to next
					}
				}
				cancel() // Ensure context is cancelled for each file after its download loop
			}

			if len(batchErrors) > 0 {
				msgChan <- batchDownloadErrorMsg{err: errors.Join(batchErrors...)}
			} else {
				msgChan <- batchDownloadCompleteMsg(fmt.Sprintf("Successfully downloaded %d files.", len(files)))
			}
			close(msgChan)
		}()

		return func() tea.Msg {
			return <-msgChan
		}
	}
}
