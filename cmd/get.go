package cmd

import (
	"context"
	"fmt"

	"seedrcc/internal"

	"github.com/spf13/cobra"
)

// getCmd represents the get command
var getCmd = &cobra.Command{
	Use:     "get",
	Aliases: []string{"g"},
	Short:   "Get download URL of files/folders",
	Long:    `This command fetches and prints the download URL for a specified file or folder from your Seedr.cc account.`,
	Run: func(cmd *cobra.Command, args []string) {
		DebugLog("Running get command...\n")
		fileID, _ := cmd.Flags().GetString("file")
		dirID, _ := cmd.Flags().GetString("dir")

		if fileID != "" {
			getDownloadURL(false, fileID)
		} else if dirID != "" {
			getDownloadURL(true, dirID)
		} else {
			fmt.Println("Please specify either --file or --dir with --get.")
			cmd.Help()
			return
		}
	},
}

func init() {
	RootCmd.AddCommand(getCmd)

	getCmd.Flags().StringP("file", "f", "", "Specify a file ID to get its download URL")
	getCmd.Flags().StringP("dir", "D", "", "Specify a directory ID to get its archive URL (note: uppercase D to avoid conflict with debug flag)")
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
