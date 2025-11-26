package main

import (
	"fmt"
	"os"

	"seedr/cmd"
	"seedr/internal" // Import the internal package
	"seedr/pkg/seedr"
	"seedr/tui" // Import the new tui package

	"github.com/spf13/cobra"

)

// getSeedrSettings prints formatted account settings.
func getSeedrSettings(data *seedr.UserSettings) {
	accountInfo := data.Account

	fmt.Printf("Username: %s\n", accountInfo.Username)
	fmt.Printf("User ID: %d\n", accountInfo.UserID)

	spaceUsed := internal.HumanReadableBytes(accountInfo.SpaceUsed)
	spaceMax := internal.HumanReadableBytes(accountInfo.SpaceMax)
	bandwidthUsed := internal.HumanReadableBytes(accountInfo.BandwidthUsed)

	fmt.Printf("Space Used: %s\n", spaceUsed)
	fmt.Printf("Space Max: %s\n", spaceMax)
	fmt.Printf("Bandwidth Used: %s\n", bandwidthUsed)
	fmt.Printf("Country: %s\n", data.Country)
}

func main() {
	cmd.Execute()
}

func init() {
	// Add the Cobra root command's PersistentPreRun hook to initialize the Seedr client.
	// This ensures `internal.Account` is available for all commands.
	cmd.RootCmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		if err := internal.FetchSeedrAccessToken(); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to initialize Seedr client: %v\n", err)
			return err // Return error to Cobra to stop execution
		}
		return nil
	}

	// Ensure account is closed after CLI commands, or TUI exits
	cobra.OnFinalize(func() {
		if internal.Account != nil {
			internal.Account.Close()
		}
		if internal.Log != nil { // Close the logger file handle
			internal.Log.Close()
		}
	})

	// Assign the TUI start function to the cmd package variable
	cmd.StartTUI = startTUI
}

// Function to start TUI. This will be called only if no commands or flags are passed.
func startTUI() {
	if err := tui.RunTUI(internal.Account); err != nil {
		internal.Log.Debug("Error running TUI: %v", err)
		os.Exit(1)
	}
}



// Removed runCli function as its logic is being replaced by Cobra commands.
