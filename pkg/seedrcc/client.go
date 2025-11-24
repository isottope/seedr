package seedrcc

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	//"strconv" // Removed unused import
	"strings"
	"sync"
	"time"
)

// OnTokenRefreshCallback defines the signature for the token refresh callback function.
type OnTokenRefreshCallback func(newToken *Token)

// Client represents a Seedr API client.
type Client struct {
	httpClient *http.Client
	token      *Token
	onTokenRefresh OnTokenRefreshCallback
	mu         sync.Mutex // Mutex for protecting client-wide state, especially during token refresh

	// Stores whether the client manages its own http.Client lifecycle.
	// If true, httpClient.CloseIdleConnections() will be called on Client.Close().
	managesClientLifecycle bool
}

// ClientOption is a function type for configuring the Client.
type ClientOption func(*Client)

// WithHTTPClient provides a custom http.Client for the Seedr client.
func WithHTTPClient(httpClient *http.Client) ClientOption {
	return func(c *Client) {
		c.httpClient = httpClient
		c.managesClientLifecycle = false // User provided client, so we don't manage its lifecycle
	}
}

// WithTimeout sets the timeout for the HTTP client.
func WithTimeout(timeout time.Duration) ClientOption {
	return func(c *Client) {
		if c.managesClientLifecycle && c.httpClient != nil {
			c.httpClient.Timeout = timeout
		} else if c.httpClient == nil {
			// If no client is provided yet, set it up with a new one
			c.httpClient = &http.Client{Timeout: timeout}
			c.managesClientLifecycle = true
		}
	}
}

// WithProxy sets the proxy for the HTTP client.
func WithProxy(proxyURL *url.URL) ClientOption {
	return func(c *Client) {
		if c.managesClientLifecycle && c.httpClient != nil {
			if c.httpClient.Transport == nil {
				c.httpClient.Transport = &http.Transport{}
			}
			if tr, ok := c.httpClient.Transport.(*http.Transport); ok {
				tr.Proxy = http.ProxyURL(proxyURL)
			}
		} else if c.httpClient == nil {
			// If no client is provided yet, set it up with a new one
			tr := &http.Transport{Proxy: http.ProxyURL(proxyURL)}
			c.httpClient = &http.Client{Transport: tr}
			c.managesClientLifecycle = true
		}
	}
}

// WithTokenRefreshCallback sets the callback function for token refreshes.
func WithTokenRefreshCallback(callback OnTokenRefreshCallback) ClientOption {
	return func(c *Client) {
		c.onTokenRefresh = callback
	}
}

// NewClient creates a new Seedr API client with the given token and options.
func NewClient(token *Token, opts ...ClientOption) *Client {
	c := &Client{
		token: token,
	}

	// Apply options
	for _, opt := range opts {
		opt(c)
	}

	// If no http client was provided, create a default one
	if c.httpClient == nil {
		c.httpClient = &http.Client{Timeout: 30 * time.Second} // Default timeout
		c.managesClientLifecycle = true
	}

	return c
}

// Close closes the underlying HTTP client if its lifecycle is managed by this Client instance.
func (c *Client) Close() {
	if c.managesClientLifecycle && c.httpClient != nil {
		c.httpClient.CloseIdleConnections()
	}
}

// Token returns the current authentication token used by the client.
func (c *Client) Token() *Token {
	c.mu.Lock() // Use Lock as Token can be updated concurrently
	defer c.mu.Unlock()
	return c.token
}

// makeHTTPRequest performs the raw HTTP request and handles low-level network or HTTP status errors.
func (c *Client) makeHTTPRequest(
	ctx context.Context,
	method, rawURL string,
	params map[string]string,
	data map[string]string,
	files map[string][]byte, // file_field_name -> file_content
) (map[string]interface{}, error) {
	reqURL, err := url.Parse(rawURL)
	if err != nil {
		return nil, &NetworkError{Message: "Failed to parse URL", Err: err}
	}

	query := reqURL.Query()
	for k, v := range params {
		query.Set(k, v)
	}
	reqURL.RawQuery = query.Encode()

	var reqBody io.Reader
	contentType := "application/x-www-form-urlencoded" // Default for form data

	if files != nil && len(files) > 0 {
		// Use multipart/form-data if files are present
		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)

		for fieldName, fileContent := range files {
			part, err := writer.CreateFormFile(fieldName, "filename") // Generic filename
			if err != nil {
				return nil, fmt.Errorf("failed to create form file: %w", err)
			}
			if _, err := part.Write(fileContent); err != nil {
				return nil, fmt.Errorf("failed to write file content: %w", err)
			}
		}

		for k, v := range data {
			_ = writer.WriteField(k, v)
		}

		writer.Close() // Important: close the writer to finalize the body
		reqBody = body
		contentType = writer.FormDataContentType()
	} else if data != nil && len(data) > 0 {
		// Use form-urlencoded for data if no files
		formData := url.Values{}
		for k, v := range data {
			formData.Set(k, v)
		}
		reqBody = strings.NewReader(formData.Encode())
	} else if method == http.MethodPost || method == http.MethodPut {
		// If it's a POST/PUT but no data, send an empty body
		reqBody = strings.NewReader("")
	}


	req, err := http.NewRequestWithContext(ctx, method, reqURL.String(), reqBody)
	if err != nil {
		return nil, &NetworkError{Message: "Failed to create HTTP request", Err: err}
	}

	req.Header.Set("Content-Type", contentType)
	req.Header.Set("User-Agent", "seedrcc-go/1.0") // Custom User-Agent

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, &NetworkError{Message: "HTTP request failed", Err: err}
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, &NetworkError{Message: "Failed to read response body", Err: err}
	}

	if resp.StatusCode >= 500 && resp.StatusCode < 600 {
		return nil, &ServerError{
			Message:    fmt.Sprintf("Server returned status code %d", resp.StatusCode),
			StatusCode: resp.StatusCode,
			Response:   respBody,
		}
	}

	if resp.StatusCode >= 400 && resp.StatusCode < 500 {
		// Attempt to parse as AuthenticationError if 401
		if resp.StatusCode == http.StatusUnauthorized {
			return nil, NewAuthenticationError(
				fmt.Sprintf("Authentication failed with status code %d", resp.StatusCode),
				resp.StatusCode,
				respBody,
			)
		}
		// Otherwise, general APIError
		return nil, NewAPIError(
			fmt.Sprintf("API returned status code %d", resp.StatusCode),
			resp.StatusCode,
			respBody,
		)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, &APIError{
			Message:    "Failed to parse API response as JSON",
			StatusCode: resp.StatusCode,
			Response:   respBody,
			ErrorType:  "parsing_error",
		}
	}

	return result, nil
}

// apiRequest handles the core logic for making authenticated API requests, including token refreshes.
func (c *Client) apiRequest(
	ctx context.Context,
	method, funcName string,
	data map[string]string,
	files map[string][]byte,
	extraParams map[string]string, // For URL params not part of the 'data' payload
	rawURL string, // Optional: override default URL
) (map[string]interface{}, error) {
	c.mu.Lock() // Protect client state during token handling
	defer c.mu.Unlock()

	requestURL := ResourceURL
	if rawURL != "" {
		requestURL = rawURL
	}

	params := make(map[string]string)
	params["access_token"] = c.token.GetAccessToken()
	if funcName != "" {
		params["func"] = funcName
	}
	// Add any extra URL parameters provided by the caller
	for k, v := range extraParams {
		params[k] = v
	}


	// First attempt
	response, err := c.makeHTTPRequest(ctx, method, requestURL, params, data, files)
	if err != nil {
		if apiErr, ok := err.(*APIError); ok {
			if apiErr.ErrorType == "expired_token" {
				// Token expired, attempt refresh
				if refreshErr := c.refreshAccessToken(ctx); refreshErr != nil {
					return nil, refreshErr // Refresh failed
				}
				// Retry with new access token
				params["access_token"] = c.token.GetAccessToken()
				response, err = c.makeHTTPRequest(ctx, method, requestURL, params, data, files)
			}
		}
		if err != nil { // Re-check err after potential retry
			return nil, err
		}
	}

	// Check for API-specific result=false or error fields
	if result, ok := response["result"].(bool); ok && !result {
		if errorMsg, ok := response["error"].(string); ok {
			return nil, NewAPIError(errorMsg, 0, nil) // status code 0 as it's from response body, not HTTP status
		}
		return nil, NewAPIError("Unknown API error from response body (result=false)", 0, nil)
	}
	// Some API calls don't have a "result" field but still return data.
	// If "result" is true, or not present, consider it a success.

	return response, nil
}


// refreshAccessToken refreshes the access token using the refresh token or device code.
func (c *Client) refreshAccessToken(ctx context.Context) error {
	var (
		response map[string]interface{}
		err      error
	)

	refreshToken := c.token.GetRefreshToken()
	deviceCode := c.token.GetDeviceCode()

	if refreshToken != nil && *refreshToken != "" {
		payload := PrepareRefreshTokenPayload(*refreshToken)
		response, err = c.makeHTTPRequest(ctx, http.MethodPost, TokenURL, nil, payload, nil)
	} else if deviceCode != nil && *deviceCode != "" {
		params := PrepareDeviceCodeParams(*deviceCode)
		response, err = c.makeHTTPRequest(ctx, http.MethodGet, DeviceAuthorizeURL, params, nil, nil)
	} else {
		return NewAuthenticationError("Session expired. No refresh token or device code available.", 0, nil)
	}

	if err != nil {
		// Wrap the underlying error message, as AuthenticationError doesn't have an 'Err' field.
		return NewAuthenticationError(fmt.Sprintf("Failed to refresh token: %v", err.Error()), 0, nil)
	}

	accessToken, ok := response["access_token"].(string)
	if !ok || accessToken == "" {
		return NewAuthenticationError("Token refresh failed. The response did not contain a new access token.", 0, nil)
	}

	// Update the token in a thread-safe manner
	c.token.Update(accessToken, refreshToken) // Keep the same refresh token unless a new one is provided.

	if c.onTokenRefresh != nil {
		c.onTokenRefresh(c.token)
	}

	return nil
}

// initializeClient is a factory helper that orchestrates the authentication process and constructs the client.
func initializeClient(
	ctx context.Context,
	authCallable func(*http.Client) (map[string]interface{}, error),
	tokenExtrasCallable func(map[string]interface{}) map[string]string,
	onTokenRefresh OnTokenRefreshCallback,
	opts ...ClientOption,
) (*Client, error) {
	tempClient := NewClient(nil, opts...) // Create a temporary client to perform initial auth
	defer func() {
		// Only close if it was internally created and auth failed
		if tempClient.managesClientLifecycle {
			tempClient.Close()
		}
	}()

	response_data, err := authCallable(tempClient.httpClient)
	if err != nil {
		return nil, err
	}

	tokenExtras := tokenExtrasCallable(response_data)
	
	var refreshToken *string
	if rt, ok := response_data["refresh_token"].(string); ok {
		refreshToken = &rt
	}

	token := NewToken(
		response_data["access_token"].(string),
		refreshToken,
		nil, // Device code is handled by tokenExtrasCallable if applicable
	)

	if deviceCode, ok := tokenExtras["device_code"]; ok {
		token.DeviceCode = &deviceCode
	}

	// Create the actual client with the obtained token
	client := NewClient(token, opts...)
	client.onTokenRefresh = onTokenRefresh // Ensure the callback is set on the final client
	return client, nil
}

// FromPassword creates a new client by authenticating with a username and password.
func FromPassword(ctx context.Context, username, password string, opts ...ClientOption) (*Client, error) {
	authCallable := func(httpClient *http.Client) (map[string]interface{}, error) {
		payload := PreparePasswordPayload(username, password)
		resp, err := (&Client{httpClient: httpClient}).makeHTTPRequest(ctx, http.MethodPost, TokenURL, nil, payload, nil)
		if err != nil {
			if apiErr, ok := err.(*APIError); ok {
				return nil, NewAuthenticationError("Authentication failed", apiErr.StatusCode, apiErr.Response)
			}
			return nil, err
		}
		return resp, nil
	}

	tokenExtrasCallable := func(map[string]interface{}) map[string]string {
		return make(map[string]string)
	}

	return initializeClient(ctx, authCallable, tokenExtrasCallable, nil, opts...)
}

// FromDeviceCode creates a new client by authorizing with a device code.
func FromDeviceCode(ctx context.Context, deviceCode string, opts ...ClientOption) (*Client, error) {
	authCallable := func(httpClient *http.Client) (map[string]interface{}, error) {
		params := PrepareDeviceCodeParams(deviceCode)
		resp, err := (&Client{httpClient: httpClient}).makeHTTPRequest(ctx, http.MethodGet, DeviceAuthorizeURL, params, nil, nil)
		if err != nil {
			if apiErr, ok := err.(*APIError); ok {
				return nil, NewAuthenticationError("Failed to authorize device", apiErr.StatusCode, apiErr.Response)
			}
			return nil, err
		}
		return resp, nil
	}

	tokenExtrasCallable := func(map[string]interface{}) map[string]string {
		return map[string]string{"device_code": deviceCode}
	}

	return initializeClient(ctx, authCallable, tokenExtrasCallable, nil, opts...)
}

// FromRefreshToken creates a new client by using an existing refresh token.
func FromRefreshToken(ctx context.Context, refreshToken string, opts ...ClientOption) (*Client, error) {
	authCallable := func(httpClient *http.Client) (map[string]interface{}, error) {
		payload := PrepareRefreshTokenPayload(refreshToken)
		resp, err := (&Client{httpClient: httpClient}).makeHTTPRequest(ctx, http.MethodPost, TokenURL, nil, payload, nil)
		if err != nil {
			if apiErr, ok := err.(*APIError); ok {
				return nil, NewAuthenticationError("Failed to refresh token", apiErr.StatusCode, apiErr.Response)
			}
			return nil, err
		}
		return resp, nil
	}

	tokenExtrasCallable := func(map[string]interface{}) map[string]string {
		return map[string]string{"refresh_token": refreshToken}
	}

	return initializeClient(ctx, authCallable, tokenExtrasCallable, nil, opts...)
}

// GetDeviceCode retrieves the device and user codes required for authorization.
func GetDeviceCode(ctx context.Context) (*DeviceCode, error) {
	params := map[string]string{"client_id": DeviceClientID}
	
	// Use a temporary client for this static method, as it doesn't require prior authentication
	tempClient := NewClient(nil) 
	defer tempClient.Close()

	response_data, err := tempClient.makeHTTPRequest(ctx, http.MethodGet, DeviceCodeURL, params, nil, nil)
	if err != nil {
		return nil, err
	}
	
	deviceCode := NewDeviceCodeFromMap(response_data)
	return &deviceCode, nil
}

// GetSettings retrieves the user settings.
func (c *Client) GetSettings(ctx context.Context) (*UserSettings, error) {
	response_data, err := c.apiRequest(ctx, http.MethodGet, "get_settings", nil, nil, nil, "")
	if err != nil {
		return nil, err
	}
	settings := NewUserSettingsFromMap(response_data)
	return &settings, nil
}

// GetMemoryBandwidth retrieves the memory and bandwidth usage.
func (c *Client) GetMemoryBandwidth(ctx context.Context) (*MemoryBandwidth, error) {
	response_data, err := c.apiRequest(ctx, http.MethodGet, "get_memory_bandwidth", nil, nil, nil, "")
	if err != nil {
		return nil, err
	}
	mb := NewMemoryBandwidthFromMap(response_data)
	return &mb, nil
}

// ListContents lists the contents of a folder.
func (c *Client) ListContents(ctx context.Context, folderID string) (*ListContentsResult, error) {
	data := PrepareListContentsPayload(folderID)
	response_data, err := c.apiRequest(ctx, http.MethodPost, "list_contents", data, nil, nil, "")
	if err != nil {
		return nil, err
	}
	lcr := NewListContentsResultFromMap(response_data)
	return &lcr, nil
}

// AddTorrent adds a torrent to the Seedr account for downloading.
func (c *Client) AddTorrent(
	ctx context.Context,
	magnetLink *string,
	torrentFileContent []byte, // Use []byte for file content
	wishlistID *string,
	folderID string,
) (*AddTorrentResult, error) {
	data := make(map[string]string)
	if magnetLink != nil {
		data["torrent_magnet"] = *magnetLink
	}
	if wishlistID != nil {
		data["wishlist_id"] = *wishlistID
	}
	data["folder_id"] = folderID

	files := make(map[string][]byte)
	if torrentFileContent != nil {
		files["torrent_file"] = torrentFileContent
	}

	response_data, err := c.apiRequest(ctx, http.MethodPost, "add_torrent", data, files, nil, "")
	if err != nil {
		return nil, err
	}
	atr := NewAddTorrentResultFromMap(response_data)
	return &atr, nil
}

// ScanPage scans a page for torrents and magnet links.
func (c *Client) ScanPage(ctx context.Context, pageURL string) (*ScanPageResult, error) {
	data := PrepareScanPagePayload(pageURL)
	response_data, err := c.apiRequest(ctx, http.MethodPost, "scan_page", data, nil, nil, "")
	if err != nil {
		return nil, err
	}
	spr := NewScanPageResultFromMap(response_data)
	return &spr, nil
}

// FetchFile creates a link of a file.
func (c *Client) FetchFile(ctx context.Context, fileID string) (*FetchFileResult, error) {
	data := PrepareFetchFilePayload(fileID)
	response_data, err := c.apiRequest(ctx, http.MethodPost, "fetch_file", data, nil, nil, "")
	if err != nil {
		return nil, err
	}
	ffr := NewFetchFileResultFromMap(response_data)
	return &ffr, nil
}

// CreateArchive creates an archive link of a folder.
func (c *Client) CreateArchive(ctx context.Context, folderID string) (*CreateArchiveResult, error) {
	data := PrepareCreateArchivePayload(folderID)
	response_data, err := c.apiRequest(ctx, http.MethodPost, "create_empty_archive", data, nil, nil, "")
	if err != nil {
		return nil, err
	}
	car := NewCreateArchiveResultFromMap(response_data)
	return &car, nil
}

// SearchFiles searches for files.
func (c *Client) SearchFiles(ctx context.Context, query string) (*Folder, error) {
	data := PrepareSearchFilesPayload(query)
	response_data, err := c.apiRequest(ctx, http.MethodPost, "search_files", data, nil, nil, "")
	if err != nil {
		return nil, err
	}
	folder := NewFolderFromMap(response_data)
	return &folder, nil
}

// AddFolder adds a folder.
func (c *Client) AddFolder(ctx context.Context, name string) (*APIResult, error) {
	data := PrepareAddFolderPayload(name)
	response_data, err := c.apiRequest(ctx, http.MethodPost, "add_folder", data, nil, nil, "")
	if err != nil {
		return nil, err
	}
	result := NewAPIResultFromMap(response_data)
	return &result, nil
}

// RenameFile renames a file.
func (c *Client) RenameFile(ctx context.Context, fileID, renameTo string) (*APIResult, error) {
	data := PrepareRenamePayload(renameTo, fileID, "") // fileID set, folderID empty
	response_data, err := c.apiRequest(ctx, http.MethodPost, "rename", data, nil, nil, "")
	if err != nil {
		return nil, err
	}
	result := NewAPIResultFromMap(response_data)
	return &result, nil
}

// RenameFolder renames a folder.
func (c *Client) RenameFolder(ctx context.Context, folderID, renameTo string) (*APIResult, error) {
	data := PrepareRenamePayload(renameTo, "", folderID) // folderID set, fileID empty
	response_data, err := c.apiRequest(ctx, http.MethodPost, "rename", data, nil, nil, "")
	if err != nil {
		return nil, err
	}
	result := NewAPIResultFromMap(response_data)
	return &result, nil
}

// deleteAPIItem is a helper for deleting various item types.
func (c *Client) deleteAPIItem(ctx context.Context, itemType, itemID string) (*APIResult, error) {
	data := PrepareDeleteItemPayload(itemType, itemID)
	response_data, err := c.apiRequest(ctx, http.MethodPost, "delete", data, nil, nil, "")
	if err != nil {
		return nil, err
	}
	result := NewAPIResultFromMap(response_data)
	return &result, nil
}

// DeleteFile deletes a file.
func (c *Client) DeleteFile(ctx context.Context, fileID string) (*APIResult, error) {
	return c.deleteAPIItem(ctx, "file", fileID)
}

// DeleteFolder deletes a folder.
func (c *Client) DeleteFolder(ctx context.Context, folderID string) (*APIResult, error) {
	return c.deleteAPIItem(ctx, "folder", folderID)
}

// DeleteTorrent deletes an active downloading torrent.
func (c *Client) DeleteTorrent(ctx context.Context, torrentID string) (*APIResult, error) {
	return c.deleteAPIItem(ctx, "torrent", torrentID)
}

// DeleteWishlist deletes an item from the wishlist.
func (c *Client) DeleteWishlist(ctx context.Context, wishlistID string) (*APIResult, error) {
	data := PrepareRemoveWishlistPayload(wishlistID)
	response_data, err := c.apiRequest(ctx, http.MethodPost, "remove_wishlist", data, nil, nil, "")
	if err != nil {
		return nil, err
	}
	result := NewAPIResultFromMap(response_data)
	return &result, nil
}

// GetDevices retrieves the devices connected to the Seedr account.
func (c *Client) GetDevices(ctx context.Context) ([]Device, error) {
	response_data, err := c.apiRequest(ctx, http.MethodGet, "get_devices", nil, nil, nil, "")
	if err != nil {
		return nil, err
	}
	devicesData, ok := response_data["devices"].([]interface{})
	if !ok {
		return nil, fmt.Errorf("expected 'devices' field to be a list")
	}
	
	var devices []Device
	for _, item := range devicesData {
		if deviceMap, isMap := item.(map[string]interface{}); isMap {
			devices = append(devices, NewDeviceFromMap(deviceMap))
		}
	}
	return devices, nil
}

// ChangeName changes the name of the account.
func (c *Client) ChangeName(ctx context.Context, name, password string) (*APIResult, error) {
	data := PrepareChangeNamePayload(name, password)
	response_data, err := c.apiRequest(ctx, http.MethodPost, "user_account_modify", data, nil, nil, "")
	if err != nil {
		return nil, err
	}
	result := NewAPIResultFromMap(response_data)
	return &result, nil
}

// ChangePassword changes the password of the account.
func (c *Client) ChangePassword(ctx context.Context, oldPassword, newPassword string) (*APIResult, error) {
	data := PrepareChangePasswordPayload(oldPassword, newPassword)
	response_data, err := c.apiRequest(ctx, http.MethodPost, "user_account_modify", data, nil, nil, "")
	if err != nil {
		return nil, err
	}
	result := NewAPIResultFromMap(response_data)
	return &result, nil
}