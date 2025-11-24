package cmd

import (
	"context"
	"fmt"

	"seedrcc/internal"

	"github.com/spf13/cobra"
)

// SeedrObject represents a file or folder from Seedr.cc
type SeedrObject struct {
	isDir bool
	name  string
	id    string
}

var allSeedrObjects map[string]SeedrObject // Global map to store all objects for quick lookup
var objectNames []string                 // Global slice for auto-completion names

// GetFolderContents recursively traverses Seedr folders and collects all files and subfolders.
func GetFolderContents(ctx context.Context, currentFolder *internal.SeedrListContentsResult, collectedObjects *[]SeedrObject) {
	// Process immediate subfolders of the current folder
	for _, subfolder := range currentFolder.Folders {
		// Add subfolder itself
		*collectedObjects = append(*collectedObjects, SeedrObject{isDir: true, name: subfolder.Name, id: fmt.Sprintf("%d", subfolder.ID)})

		// Recursively get contents of subfolder
		subfolderData, err := internal.Account.ListContents(ctx, fmt.Sprintf("%d", subfolder.ID))
		if err != nil {
			DebugLog("Error listing contents of folder %d (%s): %v", subfolder.ID, subfolder.Name, err)
			continue
		}
		GetFolderContents(ctx, subfolderData, collectedObjects)
	}

	// Process immediate files in the current folder
	for _, file := range currentFolder.Files {
		// Use file.FolderFileID for files
		*collectedObjects = append(*collectedObjects, SeedrObject{isDir: false, name: file.Name, id: fmt.Sprintf("%d", file.FolderFileID)})
	}
}

// FetchObjectDetails retrieves all Seedr files and folders, populating global maps for lookup and auto-completion.
func FetchObjectDetails() ([]string, error) {
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
		// Add subfolder itself
		collectedObjects = append(collectedObjects, SeedrObject{isDir: true, name: subfolder.Name, id: fmt.Sprintf("%d", subfolder.ID)})
		
		// Recursively get contents of subfolder
		subfolderData, err := internal.Account.ListContents(ctx, fmt.Sprintf("%d", subfolder.ID))
		if err != nil {
			DebugLog("Error listing contents of folder %d (%s): %v", subfolder.ID, subfolder.Name, err)
			continue
		}
		GetFolderContents(ctx, subfolderData, &collectedObjects) // Recursively add sub-contents
	}

	// Process immediate files in the root
	for _, file := range rootData.Files {
		collectedObjects = append(collectedObjects, SeedrObject{isDir: false, name: file.Name, id: fmt.Sprintf("%d", file.FolderFileID)})
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

// CompleteSeedrObjectPrompt provides tab completion for Seedr objects (files and folders).
func CompleteSeedrObjectPrompt(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	if len(args) != 0 {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	names, err := FetchObjectDetails()
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}

	return names, cobra.ShellCompDirectiveNoFileComp
}
