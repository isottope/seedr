package internal

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"

	"seedr/pkg/seedr"
)

// Account is the global Seedr client variable
var Account *seedr.Client

// SeedrAPIError is an alias for seedr.APIError to avoid direct dependency in cmd package
type SeedrAPIError = seedr.APIError

// SeedrListContentsResult is an alias for seedr.ListContentsResult
type SeedrListContentsResult = seedr.ListContentsResult

// SeedrUserSettings is an alias for seedr.UserSettings
type SeedrUserSettings = seedr.UserSettings

// DebugLog is a package-level variable to hold the debug logging function.
// It is meant to be set by an external package (e.g., cmd) to route debug messages.
// By default, it's a no-op function.
var DebugLog = func(format string, a ...interface{}) {}


// onTokenRefresh is a global callback function for token refreshes.
// It saves the new token to file.
var onTokenRefresh = func(newToken *seedr.Token) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting user home directory for token refresh: %v\n", err)
		return
	}
	seedrFolder := filepath.Join(homeDir, ".cache", "seedr")
	tokenLocation := filepath.Join(seedrFolder, "token.txt")

	jsonStr, err := newToken.ToJSON()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error saving refreshed token to JSON: %v\n", err)
		return
	}
	if err := os.WriteFile(tokenLocation, []byte(jsonStr), 0600); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing refreshed token to file: %v\n", err)
	} else {
		fmt.Println("Token refreshed and saved.")
	}
}

// FetchSeedrAccessToken handles token retrieval and persistence.
func FetchSeedrAccessToken() error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("could not get user home directory: %w", err)
	}
	seedrFolder := filepath.Join(homeDir, ".cache", "seedr")
	if _, err := os.Stat(seedrFolder); os.IsNotExist(err) {
		os.MkdirAll(seedrFolder, 0755)
	}
	tokenLocation := filepath.Join(seedrFolder, "token.txt")

	ctx := context.Background()

	if _, err := os.Stat(tokenLocation); os.IsNotExist(err) {
		// No token file, perform device authentication
		DebugLog("No token found. Initiating device authentication flow...")
	
	
codes, err := seedr.GetDeviceCode(ctx)
		if err != nil {
			return fmt.Errorf("error getting device code: %w", err)
		}

		fmt.Printf("Please go to %s and enter the code: %s\n", codes.VerificationURL, codes.UserCode)
		fmt.Print("Press Enter after authorizing the device.")
		bufio.NewReader(os.Stdin).ReadBytes('\n') // Wait for user to press Enter

		client, err := seedr.FromDeviceCode(ctx, codes.DeviceCode, seedr.WithTokenRefreshCallback(onTokenRefresh))
		if err != nil {
			return fmt.Errorf("error creating client from device code: %w", err)
		}
		Account = client // Set the global client
		tokenJson, err := Account.Token().ToJSON() // Use Token() accessor
		if err != nil {
			return fmt.Errorf("error converting new token to JSON: %w", err)
		}
		fmt.Printf("Authorization Successful. Token: %s\n", Account.Token().String()) // Use Token() accessor

		if err := os.WriteFile(tokenLocation, []byte(tokenJson), 0600); err != nil {
			return fmt.Errorf("error writing token to file: %w", err)
		}
		return nil

	} else {
		// Token file exists, load it
		DebugLog("Token file found. Loading existing token...")
		tokenBytes, err := os.ReadFile(tokenLocation)
		if err != nil {
			return fmt.Errorf("error reading token file: %w", err)
		}
		token, err := seedr.TokenFromJSON(string(tokenBytes))
		if err != nil {
			return fmt.Errorf("error parsing token from JSON: %w", err)
		}
		// Create client from existing token
		client := seedr.NewClient(token, seedr.WithTokenRefreshCallback(onTokenRefresh))
		Account = client // Set the global client
		return nil
	}
}
