package cmd

import (
	"context"
	"fmt"
	"os"
	"regexp" // Added for magnet link detection
	"strings"

	"seedr/internal"


	"github.com/spf13/cobra"
)

// addCmd represents the add command
var addCmd = &cobra.Command{
	Use:     "add <torrent-source>",
	Aliases: []string{"a"},
	Short:   "Add a torrent to Seedr (magnet link, .torrent file, or URL to scan)",
	Long: `This command allows you to add a torrent to your Seedr.cc account.
You can provide a magnet link, a path to a local .torrent file,
or a URL to a webpage to scan for torrents. The command will automatically
detect the type of input.

The target directory can optionally be specified using the --td flag.

Examples:
  seedr add "magnet:?xt=urn:btih:..."
  seedr add /path/to/my.torrent --td Movies
  seedr add "https://example.com/page-with-torrents"`,
	Run: func(cmd *cobra.Command, args []string) {
		internal.Log.Debug("Running add command...")
		ctx := context.Background()

		if len(args) != 1 {
			fmt.Println("Please provide a magnet link, a .torrent file path, or a URL to scan.")
			cmd.Help()
			return
		}

		input := args[0]
		var magnetLink *string
		var torrentFileContent []byte
		var err error

		// Regex to detect magnet links
		isMagnet, err := regexp.MatchString("^magnet:.*", input)
		if err != nil {
			fmt.Printf("Error checking magnet link regex: %v\n", err)
			return
		}

		if isMagnet {
			magnetLink = &input
			internal.Log.Debug("Detected input as magnet link: %s", *magnetLink)
		} else if strings.HasSuffix(strings.ToLower(input), ".torrent") {
			// Handle .torrent file upload
			fileBytes, err := os.ReadFile(input)
			if err != nil {
				fmt.Printf("Error reading torrent file '%s': %v\n", input, err)
				return
			}
			torrentFileContent = fileBytes
			internal.Log.Debug("Detected input as local torrent file: %s", input)
		} else {
			// Assume it's a URL to scan
			internal.Log.Debug("Detected input as URL to scan for torrents: %s", input)
			scanResult, err := internal.Account.ScanPage(ctx, input)
			if err != nil {
				fmt.Printf("Error scanning URL '%s': %v\n", input, err)
				return
			}

			if len(scanResult.Torrents) == 0 {
				fmt.Printf("No torrents found on page '%s'.\n", input)
				return
			}

			// Simple TUI for selection
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
			internal.Log.Debug("Selected torrent from scan: %s", selectedTorrent.Title)
		}

		// Determine target folder ID
		folderID := "-1" // Default to root
		if targetDirectoryName != "" {
			_, err := FetchObjectDetails() // Ensure cache is populated
			if err != nil {
				fmt.Printf("Error fetching Seedr objects for folder lookup: %v\n", err)
				return
			}
			obj, ok := allSeedrObjects[targetDirectoryName]
			if !ok || !obj.isDir {
				fmt.Printf("Error: Directory '%s' not found or is not a directory.\n", targetDirectoryName)
				return
			}
			folderID = obj.id
			internal.Log.Debug("Adding to directory: %s (ID: %s)", targetDirectoryName, folderID)
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
	targetDirectoryName string
)

func init() {
	RootCmd.AddCommand(addCmd)
	addCmd.Flags().StringVarP(&targetDirectoryName, "target-directory", "t", "", "Name of the target directory in Seedr (optional)")

	// Add completion for --td flag
	addCmd.RegisterFlagCompletionFunc("target-directory", completeFolderPrompt)
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
