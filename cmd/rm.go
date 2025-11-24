package cmd

import (
	"context"
	"fmt"

	"seedrcc/internal"

	"github.com/spf13/cobra"
)

// rmCmd represents the rm command
var rmCmd = &cobra.Command{
	Use:     "rm [folder-id]",
	Aliases: []string{"r"},
	Short:   "Delete a torrent folder by ID",
	Long:    `This command deletes a specified torrent folder from your Seedr.cc account using its ID.`,
	Args:    cobra.ExactArgs(1), // Requires exactly one argument (the folder ID)
	Run: func(cmd *cobra.Command, args []string) {
		DebugLog("Running rm command...")
		folderID := args[0]
		ctx := context.Background()
		_, err := internal.Account.DeleteFolder(ctx, folderID)
		if err != nil {
			fmt.Printf("Error deleting folder %s: %v\n", folderID, err)
			return
		}
		fmt.Printf("Successfully deleted folder %s.\n", folderID)
	},
}

func init() {
	RootCmd.AddCommand(rmCmd)
}
