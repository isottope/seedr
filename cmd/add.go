package cmd

import (
	"context"
	"fmt"
	"strings"

	"seedrcc/internal"

	"github.com/spf13/cobra"
)

// addCmd represents the add command
var addCmd = &cobra.Command{
	Use:     "add [magnet-link]",
	Aliases: []string{"a"},
	Short:   "Add a torrent magnet link to Seedr",
	Long:    `This command allows you to add a torrent to your Seedr.cc account using a magnet link.`,
	Args:    cobra.ExactArgs(1), // Requires exactly one argument (the magnet link)
	Run: func(cmd *cobra.Command, args []string) {
		DebugLog("Running add command...")
		magnetLink := args[0]
		ctx := context.Background()
		addResult, err := internal.Account.AddTorrent(ctx, &magnetLink, nil, nil, "-1") // Default folder_id
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
			fmt.Printf("Added %s.\n", addResult.Title)
		} else {
			fmt.Printf("Failed to add torrent: %v\n", err) // Generic error if not caught above
		}
	},
}

func init() {
	RootCmd.AddCommand(addCmd)
}
