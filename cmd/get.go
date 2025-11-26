package cmd

import (
	"context"
	"fmt"
	"seedr/internal"
	"github.com/spf13/cobra"
)

// getCmd represents the get command
var getCmd = &cobra.Command{
	Use:     "get",
	Aliases: []string{"g"},
	Short:   "Get download URL of files/folders",
	Long:    `This command fetches and prints the download URL for a specified file or folder from your Seedr.cc account.`,
	Run: func(cmd *cobra.Command, args []string) {
		internal.Log.Debug("Running get command...\n")

		if len(args) == 0 {
			fmt.Println("Please specify the name of the file or folder you want to get the download URL for.")
			cmd.Help()
			return
		}
		if len(args) > 1 {
			fmt.Println("Please specify only one item name at a time.")
			return
		}

		itemName := args[0]
		internal.Log.Debug("Trying to Fetch ID for %s", itemName)
		
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
		internal.Log.Debug("Trying to Fetch ID for %s - ID : %s", itemName, obj.id)
		getDownloadURL(obj.isDir, obj.id)
	},
	ValidArgsFunction: completegetPrompt,
}

func init() {
	RootCmd.AddCommand(getCmd)
}

// getDownloadURL fetches and prints the download URL for a file or folder.
func getDownloadURL(isDirectory bool, id string) {
	ctx := context.Background()
	var downloadURL string

	if isDirectory {
		dirArchive, err := internal.Account.CreateArchive(ctx, id)
		if err != nil {
			fmt.Printf("Error creating archive for folder %s: %v\n", id, err)
			return
		}
		downloadURL = dirArchive.ArchiveURL
		fmt.Printf("Archive URL: %s\n", downloadURL)
	} else {
		fileResult, err := internal.Account.FetchFile(ctx, id)
		if err != nil {
			fmt.Printf("Error fetching file %s: %v\n", id, err)
			return
		}
		downloadURL = fileResult.URL
		fmt.Printf("File Name: %s\n", fileResult.Name)
		fmt.Printf("Download URL: %s\n", downloadURL)
	}
}

func completegetPrompt(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	return CompleteSeedrObjectPrompt(cmd, args, toComplete)
}
