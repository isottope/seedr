package cmd

import (
	"context"
	"fmt"
	"os" // Add os package for ReadFile
	"strings"

	"seedr/internal"


	"github.com/spf13/cobra"
)

// addCmd represents the add command
var addCmd = &cobra.Command{
	Use:     "add [magnet-link | --file <path> | --url <url>]",
	Aliases: []string{"a"},
	Short:   "Add a torrent to Seedr (magnet, file, or scanned URL)",
	Long: `This command allows you to add a torrent to your Seedr.cc account.
You can provide a magnet link directly, a path to a local .torrent file,
or a URL to a webpage to scan for torrents.

Examples:
  seedr add "magnet:?xt=urn:btih:..."
  seedr add --file /path/to/my.torrent
  seedr add --url "https://example.com/page-with-torrents" --folder "Movies"`,
	Run: func(cmd *cobra.Command, args []string) {
		DebugLog("Running add command...")
		ctx := context.Background()

		// --- Input validation ---
		inputCount := 0
		if len(args) == 1 { // Positional magnet link argument
			inputCount++
		}
		if torrentFilePath != "" { // --file flag
			inputCount++
		}
		if pageURL != "" { // --url flag
			inputCount++
		}

		if inputCount == 0 {
			fmt.Println("Please provide a magnet link, a .torrent file path (--file), or a URL to scan (--url).")
			cmd.Help()
			return
		}
		if inputCount > 1 {
			fmt.Println("Error: Only one of a magnet link, --file, or --url can be provided at a time.")
			return
		}

		var magnetLink *string
		var torrentFileContent []byte

		// Determine input type
		if torrentFilePath != "" {
			// Handle .torrent file upload
			fileBytes, err := os.ReadFile(torrentFilePath)
			if err != nil {
				fmt.Printf("Error reading torrent file '%s': %v\n", torrentFilePath, err)
				return
			}
			torrentFileContent = fileBytes
			DebugLog("Adding torrent from file: %s", torrentFilePath)
		} else if pageURL != "" {
			// Handle URL scan
			DebugLog("Scanning URL for torrents: %s", pageURL)
			scanResult, err := internal.Account.ScanPage(ctx, pageURL)
			if err != nil {
				fmt.Printf("Error scanning URL '%s': %v\n", pageURL, err)
				return
			}

			if len(scanResult.Torrents) == 0 {
				fmt.Printf("No torrents found on page '%s'.\n", pageURL)
				return
			}

			// Simple TUI for selection (for now, just pick the first one or ask user)
			fmt.Println("Torrents found on page:")
			for i, t := range scanResult.Torrents {
				fmt.Printf("[%d] %s (Size: %s, Magnet: %s)\n", i+1, t.Title, internal.HumanReadableBytes(t.Size), t.Magnet)
			}
			fmt.Print("Enter the number of the torrent to add (or 0 to cancel): ")
			var selection int
			_, err = fmt.Scanln(&selection)
			if err != nil || selection < 0 || selection > len(scanResult.Torrents) {
				fmt.Println("Invalid selection. Cancelling add operation.")
				return
			}
			if selection == 0 {
				fmt.Println("Add operation cancelled.")
				return
			}
			selectedTorrent := scanResult.Torrents[selection-1]
			magnetLink = &selectedTorrent.Magnet
			DebugLog("Selected torrent from scan: %s", selectedTorrent.Title)

		} else if len(args) == 1 {
			// Handle magnet link directly from args
			magnetLink = &args[0]
			DebugLog("Adding torrent from magnet link: %s", *magnetLink)
		} else {
			// This else should ideally not be reached due to inputCount validation above
			fmt.Println("Internal error: Unhandled input combination.")
			return
		}

		// Determine target folder ID
		folderID := "-1" // Default to root
		if targetFolderName != "" {
			_, err := FetchObjectDetails() // Ensure cache is populated
			if err != nil {
				fmt.Printf("Error fetching Seedr objects for folder lookup: %v\n", err)
				return
			}
			obj, ok := allSeedrObjects[targetFolderName]
			if !ok || !obj.isDir {
				fmt.Printf("Error: Folder '%s' not found or is not a directory.\n", targetFolderName)
				return
			}
			folderID = obj.id
			DebugLog("Adding to folder: %s (ID: %s)", targetFolderName, folderID)
		}


		addResult, err := internal.Account.AddTorrent(ctx, magnetLink, torrentFileContent, nil, folderID)
		if err != nil {
			if apiErr, ok := err.(*internal.SeedrAPIError); ok {
				if strings.Contains(apiErr.Message, "not_enough_space_added_to_wishlist") {
					fmt.Println("Not enough space to add the torrent.")
					return
				}
			}
			fmt.Printf("Error adding torrent: %v\n", err)
			return
		}
		if !addResult.Result && addResult.Code != nil && *addResult.Code == 409 { // Assuming 409 for already added
			fmt.Println("Torrent already added.")
		} else if addResult.Result {
			fmt.Printf("Added '%s' successfully.\n", addResult.Title)
		} else {
			fmt.Printf("Failed to add torrent: %v\n", err) // Generic error if not caught above
		}
	},
}

var (
	torrentFilePath string
	pageURL         string
	targetFolderName string
)

func init() {
	RootCmd.AddCommand(addCmd)

	addCmd.Flags().StringVarP(&torrentFilePath, "file", "f", "", "Path to a local .torrent file")
	addCmd.Flags().StringVarP(&pageURL, "url", "u", "", "URL to a webpage to scan for torrents")
	addCmd.Flags().StringVarP(&targetFolderName, "folder", "D", "", "Name of the target folder in Seedr")

	// Mutually exclusive flags
	addCmd.MarkFlagsMutuallyExclusive("file", "url")


	// Add completion for --folder flag
	addCmd.RegisterFlagCompletionFunc("folder", completeFolderPrompt)
}

func completeFolderPrompt(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	// Only complete folders
	if len(args) != 0 {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	names, err := FetchObjectDetails()
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}

	var folderNames []string
	for _, name := range names {
		obj, ok := allSeedrObjects[name]
		if ok && obj.isDir {
			folderNames = append(folderNames, obj.name)
		}
	}
	return folderNames, cobra.ShellCompDirectiveNoFileComp
}
