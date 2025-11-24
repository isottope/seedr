package seedrcc

import (
	"fmt"
	"strconv"
	"time"
)

// ParseDateTime parses a datetime string or timestamp from the API.
// It returns nil if the input is invalid or nil.
func ParseDateTime(dt interface{}) *time.Time {
	if dt == nil {
		return nil
	}

	switch v := dt.(type) {
	case float64: // JSON numbers are often float64
		t := time.Unix(int64(v), 0)
		return &t
	case int:
		t := time.Unix(int64(v), 0)
		return &t
	case string:
		// Attempt to parse "YYYY-MM-DD HH:MM:SS" format
		parsedTime, err := time.Parse("2006-01-02 15:04:05", v)
		if err == nil {
			return &parsedTime
		}
		// If string is a timestamp
		timestamp, err := strconv.ParseInt(v, 10, 64)
		if err == nil {
			t := time.Unix(timestamp, 0)
			return &t
		}
	}
	return nil
}

// PreparePasswordPayload prepares the data payload for password-based authentication.
func PreparePasswordPayload(username, password string) map[string]string {
	return map[string]string{
		"grant_type": "password",
		"client_id":  PswrdClientID,
		"type":       "login",
		"username":   username,
		"password":   password,
	}
}

// PrepareRefreshTokenPayload prepares the data payload for refreshing an access token.
func PrepareRefreshTokenPayload(refreshToken string) map[string]string {
	return map[string]string{
		"grant_type":    "refresh_token",
		"refresh_token": refreshToken,
		"client_id":     PswrdClientID,
	}
}

// PrepareDeviceCodeParams prepares the URL parameters for device code authorization.
func PrepareDeviceCodeParams(deviceCode string) map[string]string {
	return map[string]string{
		"client_id":  DeviceClientID,
		"device_code": deviceCode,
	}
}

// PrepareAddTorrentPayload prepares the data payload for adding a new torrent.
func PrepareAddTorrentPayload(magnetLink, wishlistID, folderID string) map[string]string {
	payload := make(map[string]string)
	if magnetLink != "" {
		payload["torrent_magnet"] = magnetLink
	}
	if wishlistID != "" {
		payload["wishlist_id"] = wishlistID
	}
	payload["folder_id"] = folderID
	return payload
}

// PrepareScanPagePayload prepares the data payload for scanning a page.
func PrepareScanPagePayload(url string) map[string]string {
	return map[string]string{"url": url}
}

// PrepareCreateArchivePayload prepares the data payload for creating an archive.
func PrepareCreateArchivePayload(folderID string) map[string]string {
	// The Python version uses a JSON string, which is unusual for form data.
	// We'll mimic this for now, but it might need adjustment if the API expects
	// a different format for multipart/form-data.
	return map[string]string{"archive_arr": fmt.Sprintf(`[{"type":"folder","id":%s}]`, folderID)}
}

// PrepareFetchFilePayload prepares the data payload for fetching a file.
func PrepareFetchFilePayload(fileID string) map[string]string {
	return map[string]string{"folder_file_id": fileID}
}

// PrepareListContentsPayload prepares the data payload for listing contents.
func PrepareListContentsPayload(folderID string) map[string]string {
	return map[string]string{"content_type": "folder", "content_id": folderID}
}

// PrepareRenamePayload prepares the data payload for renaming a file or folder.
func PrepareRenamePayload(renameTo string, fileID, folderID string) map[string]string {
	payload := map[string]string{"rename_to": renameTo}
	if fileID != "" {
		payload["file_id"] = fileID
	}
	if folderID != "" {
		payload["folder_id"] = folderID
	}
	return payload
}

// PrepareDeleteItemPayload prepares the data payload for deleting an item.
func PrepareDeleteItemPayload(itemType, itemID string) map[string]string {
	// The Python version uses a JSON string.
	return map[string]string{"delete_arr": fmt.Sprintf(`[{"type":"%s","id":%s}]`, itemType, itemID)}
}

// PrepareRemoveWishlistPayload prepares the data payload for removing a wishlist item.
func PrepareRemoveWishlistPayload(wishlistID string) map[string]string {
	return map[string]string{"id": wishlistID}
}

// PrepareAddFolderPayload prepares the data payload for adding a folder.
func PrepareAddFolderPayload(name string) map[string]string {
	return map[string]string{"name": name}
}

// PrepareSearchFilesPayload prepares the data payload for searching files.
func PrepareSearchFilesPayload(query string) map[string]string {
	return map[string]string{"search_query": query}
}

// PrepareChangeNamePayload prepares the data payload for changing the account name.
func PrepareChangeNamePayload(name, password string) map[string]string {
	return map[string]string{"setting": "fullname", "password": password, "fullname": name}
}

// PrepareChangePasswordPayload prepares the data payload for changing the account password.
func PrepareChangePasswordPayload(oldPassword, newPassword string) map[string]string {
	return map[string]string{
		"setting":             "password",
		"password":            oldPassword,
		"new_password":        newPassword,
		"new_password_repeat": newPassword,
	}
}
