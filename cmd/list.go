package cmd

import (
	"context"
	"fmt"
	"strings"

	"seedrcc/internal"

	"github.com/spf13/cobra"
)

// listCmd represents the list command
var listCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"l"},
	Short:   "List folders and files on Seedr",
	Long:    `This command lists your torrents, folders, and files on Seedr.cc in a tree-like structure.`,
	Run: func(cmd *cobra.Command, args []string) {
		DebugLog("Running list command...")
		ctx := context.Background()
		
		settings, err := internal.Account.GetSettings(ctx)
		if err != nil {
			DebugLog("Error getting username in listTorrentFolders: %v", err)
			fmt.Printf("Error getting username: %v\n", err)
			return
		}
		username := settings.Account.Username

		rootData, err := internal.Account.ListContents(ctx, "0") // Root folder
		if err != nil {
			DebugLog("Error listing root contents in listTorrentFolders: %v", err)
			fmt.Printf("Error listing root contents: %v\n", err)
			return
		}
		DebugLog("listTorrentFolders found %d torrents (via rootData.Torrents)", len(rootData.Torrents))

		fmt.Printf("/%s (ID: %d)\n", username, rootData.ID)
		printFolderContents(ctx, rootData, 1)
	},
}

func init() {
	RootCmd.AddCommand(listCmd)
}

func printFolderContents(ctx context.Context, folder *internal.SeedrListContentsResult, indentLevel int) {
	indent := strings.Repeat("  ", indentLevel)

	// Print subfolders
	for _, subfolder := range folder.Folders {
		subfolderData, err := internal.Account.ListContents(ctx, fmt.Sprintf("%d", subfolder.ID))
		if err != nil {
			fmt.Printf("%sError listing subfolder %s contents: %v\n", indent, subfolder.Name, err)
			continue
		}
		fmt.Printf("%s%s (ID: %d)\n", indent, subfolderData.Name, subfolderData.ID)
		printFolderContents(ctx, subfolderData, indentLevel+1)
	}

	// Print files
	for _, file := range folder.Files {
		fileSize := internal.HumanReadableBytes(file.Size)
		fmt.Printf("%s%s (ID: %d) (Size: %s)\n", indent, file.Name, file.FolderFileID, fileSize)
	}
}
