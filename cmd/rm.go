package cmd

import (
	"context"
	"fmt"
	"seedr/internal"
	"github.com/spf13/cobra"
)

// rmCmd represents the rm command
var rmCmd = &cobra.Command{
	Use:     "rm [item-name]",
	Aliases: []string{"r"},
	Short:   "Delete a file or folder by name",
	Long:    `This command deletes a specified file or folder from your Seedr.cc account using its name.`,

	Run: func(cmd *cobra.Command, args []string) {
		internal.Log.Debug("Running rm command...\n")

		if len(args) == 0 {
			fmt.Println("Please specify the name of the file or folder you want to remove.")
			cmd.Help()
			return
		}
		if len(args) > 1 {
			fmt.Println("Please specify only one item name at a time.")
			return
		}

		itemName := args[0]
		internal.Log.Debug("Trying to Fetch ID for %s to remove", itemName)
		
		// Ensure cache is populated
		_, err := FetchObjectDetails()
		if err != nil {
			fmt.Printf("Error fetching Seedr objects for lookup: %v\n", err)
			return
		}

		obj, ok := allSeedrObjects[itemName]
		if !ok {
			fmt.Printf("Error: Item '%s' not found in your Seedr account. Please check the name and try again.\n", itemName)
			return
		}

		ctx := context.Background()
		if obj.isDir {
			_, err = internal.Account.DeleteFolder(ctx, obj.id)
			if err != nil {
				fmt.Printf("Error deleting folder %s: %v\n", itemName, err)
				return
			}
			fmt.Printf("Successfully deleted folder '%s'.\n", itemName)
		} else {
			_, err = internal.Account.DeleteFile(ctx, obj.id)
			if err != nil {
				fmt.Printf("Error deleting file %s: %v\n", itemName, err)
				return
			}
			fmt.Printf("Successfully deleted file '%s'.\n", itemName)
		}
	},
	ValidArgsFunction: completermPrompt,
}

func init() {
	RootCmd.AddCommand(rmCmd)
}

func completermPrompt(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	return CompleteSeedrObjectPrompt(cmd, args, toComplete)
}
