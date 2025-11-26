package cmd

import (
	"context"
	"fmt"
	"strings"

	"seedr/internal"

	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
)

// Define styles
var (
	rootStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("12")).Bold(true) // Blue
	folderStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("10"))           // Green
	fileStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("15"))           // White/Light Gray
	idStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))            // Dark Gray
	sizeStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("6"))            // Cyan
	errorStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("9")).Bold(true) // Red
	branchStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("7"))            // Light Gray for tree lines
)

// listCmd represents the list command
var listCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"l"},
	Short:   "List folders and files on Seedr",
	Long:    `This command lists your torrents, folders, and files on Seedr.cc in a tree-like structure.`,
	Run: func(cmd *cobra.Command, args []string) {
		internal.Log.Debug("Running list command...")
		ctx := context.Background()
		
		settings, err := internal.Account.GetSettings(ctx)
		if err != nil {
			internal.Log.Debug("Error getting username in listTorrentFolders: %v", err)
			fmt.Printf("%s\n", errorStyle.Render("Error getting username: "+err.Error()))
			return
		}
		username := settings.Account.Username

		rootData, err := internal.Account.ListContents(ctx, "0") // Root folder
		if err != nil {
			internal.Log.Debug("Error listing root contents in listTorrentFolders: %v", err)
			fmt.Printf("%s\n", errorStyle.Render("Error listing root contents: "+err.Error()))
			return
		}
		internal.Log.Debug("listTorrentFolders found %d torrents (via rootData.Torrents)", len(rootData.Torrents))

		// Print root entry
		fmt.Printf("%s %s\n", 
			rootStyle.Render("/"+username), 
			idStyle.Render(fmt.Sprintf("(ID: %d)", rootData.ID)))
		
		// Print contents starting at level 0 (children of root)
		printFolderContents(ctx, rootData, 0)
	},
}

func init() {
	RootCmd.AddCommand(listCmd)
}

// printFolderContents recursively prints the contents of a folder.
// The level parameter indicates the current depth in the tree, starting at 0 for direct children of the root.
func printFolderContents(ctx context.Context, folder *internal.SeedrListContentsResult, level int) {
	// Calculate the base indentation (spaces) for the current level
	baseIndent := strings.Repeat("  ", level)
	
	// The tree branch visual part (e.g., "|- ") rendered with its style
	// This will be prepended to the actual item name.
	treeBranch := baseIndent + branchStyle.Render("|-") + " "

	// Print subfolders
	for _, subfolder := range folder.Folders {
		subfolderData, err := internal.Account.ListContents(ctx, fmt.Sprintf("%d", subfolder.ID))
		if err != nil {
			fmt.Printf("%s%s %s\n", 
				treeBranch, 
				folderStyle.Render(subfolder.Name), 
				errorStyle.Render("(Error: " + err.Error() + ")"))
			continue
		}
		fmt.Printf("%s%s %s\n", 
			treeBranch, 
			folderStyle.Render(subfolderData.Name), 
			idStyle.Render(fmt.Sprintf("(ID: %d)", subfolderData.ID)))
		
		printFolderContents(ctx, subfolderData, level+1) // Recurse with incremented level
	}

	// Print files
	for _, file := range folder.Files {
		fileSize := internal.HumanReadableBytes(file.Size)
		fmt.Printf("%s%s %s %s \n", 
			treeBranch, 
			fileStyle.Render(file.Name), 
			idStyle.Render(fmt.Sprintf("(ID: %d)", file.FolderFileID)), 
			sizeStyle.Render(fmt.Sprintf("(Size: %s)", fileSize)))
	}
}
