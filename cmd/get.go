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
		
		// Ensure cache is populated
		_, err := fetchObjectDetails()
		if err != nil {
			fmt.Printf("Error fetching Seedr objects for lookup: %v\n", err)
			return
		}

		obj, ok := allSeedrObjects[itemName]
		if !ok {
			fmt.Printf("Error: Item '%s' not found in your Seedr account. Please check the name and try again.\n", itemName)
			return
		}
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
	if len(args) != 0 {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	names, err := fetchObjectDetails()
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}

	return names, cobra.ShellCompDirectiveNoFileComp
}

type SeedrObject struct {
	isDir bool
	name  string
	id    string
}

var allSeedrObjects map[string]SeedrObject // Global map to store all objects for quick lookup
var objectNames []string                 // Global slice for auto-completion names

// getFolderContents recursively traverses Seedr folders and collects all files and subfolders.
func getFolderContents(ctx context.Context, currentFolder *internal.SeedrListContentsResult, collectedObjects *[]SeedrObject) {
	for _, subfolder := range currentFolder.Folders {
		// Add subfolder itself
		*collectedObjects = append(*collectedObjects, SeedrObject{isDir: true, name: subfolder.Name, id: fmt.Sprintf("%d", subfolder.ID)})

		// Recursively get contents of subfolder
		subfolderData, err := internal.Account.ListContents(ctx, fmt.Sprintf("%d", subfolder.ID))
		if err != nil {
			internal.DebugLog("Error listing contents of folder %d: %v", subfolder.ID, err)
			continue
		}
		getFolderContents(ctx, subfolderData, collectedObjects)
	}

	for _, file := range currentFolder.Files {
		// Use file.FileID for files
		*collectedObjects = append(*collectedObjects, SeedrObject{isDir: false, name: file.Name, id: fmt.Sprintf("%d", file.FileID)})
	}
}

// fetchObjectDetails retrieves all Seedr files and folders, populating global maps for lookup and auto-completion.
func fetchObjectDetails() ([]string, error) {
	// If already populated, return cached names
	if allSeedrObjects != nil && len(objectNames) > 0 {
		return objectNames, nil
	}

	ctx := context.Background()
	rootData, err := internal.Account.ListContents(ctx, "0") // Root folder has ID "0"
	if err != nil {
		return nil, fmt.Errorf("error listing root contents: %w", err)
	}

	var collectedObjects []SeedrObject
	// Process immediate subfolders of the root
	for _, subfolder := range rootData.Folders {
		collectedObjects = append(collectedObjects, SeedrObject{isDir: true, name: subfolder.Name, id: fmt.Sprintf("%d", subfolder.ID)})
		subfolderData, err := internal.Account.ListContents(ctx, fmt.Sprintf("%d", subfolder.ID))
		if err != nil {
			internal.DebugLog("Error listing contents of folder %d: %v", subfolder.ID, err)
			continue
		}
		getFolderContents(ctx, subfolderData, &collectedObjects) // Recursively add sub-contents
	}

	// Process immediate files in the root
	for _, file := range rootData.Files {
		collectedObjects = append(collectedObjects, SeedrObject{isDir: false, name: file.Name, id: fmt.Sprintf("%d", file.FileID)})
	}

	allSeedrObjects = make(map[string]SeedrObject)
	objectNames = make([]string, 0, len(collectedObjects))
	for _, obj := range collectedObjects {
		// If names are not unique, this map will only store the last encountered object for a given name.
		// For a more robust solution, names could be disambiguated (e.g., by appending parent folder name).
		allSeedrObjects[obj.name] = obj
		objectNames = append(objectNames, obj.name)
	}

	return objectNames, nil
}
