package cmd

import (
	"context"
	"fmt"
	"seedr/internal" // Assuming internal is where Seedr client and models are
	"github.com/spf13/cobra"
)

// settingsCmd represents the settings command
var settingsCmd = &cobra.Command{
	Use:     "settings",
	Aliases: []string{"s"},
	Short:   "Display Seedr account settings",
	Long:    `This command fetches and displays your Seedr.cc account settings, including username, space usage, and bandwidth.`, 
	Run: func(cmd *cobra.Command, args []string) {
		internal.Log.Debug("Running settings command...")
		ctx := context.Background()
		settings, err := internal.Account.GetSettings(ctx)
		if err != nil {
			fmt.Printf("Error getting settings: %v\n", err)
			return
		}
		printSeedrSettings(settings)
	},
}

func init() {
	RootCmd.AddCommand(settingsCmd)
}

// printSeedrSettings prints formatted account settings.
func printSeedrSettings(data *internal.SeedrUserSettings) {
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
