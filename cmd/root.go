package cmd

import (
	"fmt"
	"os"
	
	"seedr/internal"

	"github.com/spf13/cobra"
)

var (
	debugMode bool
)

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "seedr",
	Short: "A CLI tool for Seedr.cc",
	Long: `seedr is a command line interface for interacting with Seedr.cc,
a cloud-based torrent downloader.

It allows you to add torrents, list your files, get download links, and more.`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// This will run before any subcommand's Run or PreRun
		// It ensures `internal.Account` is initialized before any command execution.
		// debugMode is already set by PersistentFlags().BoolVar
		if err := internal.FetchSeedrAccessToken(); err != nil {
			return err // Return error to Cobra to stop execution
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		// If no subcommands are provided, launch the TUI.
		if len(args) == 0 && StartTUI != nil { // Only launch TUI if no specific command, not in debug mode, and TUI function is set.
			StartTUI()
		} else {
			cmd.Help() // Show help if arguments are present but no command matched, or if debug mode is on.
		}
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the RootCmd.
func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err.Error())
		os.Exit(1)
	}
}

func init() {
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be available to all subcommands in the application.
	RootCmd.PersistentFlags().BoolVarP(&debugMode, "debug", "d", false, "Enable debug logging")
}

// DebugLog prints messages if debugMode is enabled
func DebugLog(format string, a ...interface{}) {
	if debugMode {
		fmt.Fprintf(os.Stderr, "DEBUG: "+format+"\n", a...)
	}
}

// Function to start TUI. This function will be defined in cli.go and passed to cmd.
var StartTUI func()
