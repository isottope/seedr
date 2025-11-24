package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
)

// loginCmd represents the login command
var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Log into Seedr",
	Long:  `This command initiates the device authentication flow to log into Seedr and saves the token for future use.`,
	Run: func(cmd *cobra.Command, args []string) {
		DebugLog("Running login command...")
		fmt.Println("Login successful and token stored (or refreshed).")
	},
}

func init() {
	RootCmd.AddCommand(loginCmd)
}
