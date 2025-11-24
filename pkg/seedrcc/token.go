package seedrcc

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
)

// Token represents the authentication tokens for a Seedr session.
type Token struct {
	AccessToken  string  `json:"access_token"`
	RefreshToken *string `json:"refresh_token,omitempty"`
	DeviceCode   *string `json:"device_code,omitempty"`
	mu           sync.RWMutex // Mutex to protect token fields during refresh
}

// NewToken creates and returns a new Token instance.
func NewToken(accessToken string, refreshToken, deviceCode *string) *Token {
	return &Token{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		DeviceCode:   deviceCode,
	}
}

// ToMap returns the token data as a map, excluding any fields that are nil.
func (t *Token) ToMap() (map[string]interface{}, error) {
	t.mu.RLock()
	defer t.mu.RUnlock()

	data := make(map[string]interface{})
	data["access_token"] = t.AccessToken
	if t.RefreshToken != nil {
		data["refresh_token"] = *t.RefreshToken
	}
	if t.DeviceCode != nil {
		data["device_code"] = *t.DeviceCode
	}
	return data, nil
}

// ToJSON returns the token data as a JSON string.
func (t *Token) ToJSON() (string, error) {
	t.mu.RLock()
	defer t.mu.RUnlock()

	data, err := json.Marshal(t)
	if err != nil {
		return "", &TokenError{Message: "Failed to marshal token to JSON", Err: err}
	}
	return string(data), nil
}

// ToBase64 returns the token data as a Base64-encoded JSON string.
func (t *Token) ToBase64() (string, error) {
	jsonStr, err := t.ToJSON()
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString([]byte(jsonStr)), nil
}

// String provides a safe, masked representation of the Token that avoids leaking secrets.
func (t *Token) String() string {
	t.mu.RLock()
	defer t.mu.RUnlock()

	mask := func(value string) string {
		if value == "" {
			return "None"
		}
		if len(value) > 5 {
			return value[:5] + "****"
		}
		return "****"
	}

	parts := []string{
		fmt.Sprintf("access_token=%s", mask(t.AccessToken)),
	}
	if t.RefreshToken != nil {
		parts = append(parts, fmt.Sprintf("refresh_token=%s", mask(*t.RefreshToken)))
	}
	if t.DeviceCode != nil {
		parts = append(parts, fmt.Sprintf("device_code=%s", mask(*t.DeviceCode)))
	}
	return fmt.Sprintf("Token(%s)", strings.Join(parts, ", "))
}

// FromJSON creates a Token object from a JSON string.
func TokenFromJSON(jsonStr string) (*Token, error) {
	var t Token
	if err := json.Unmarshal([]byte(jsonStr), &t); err != nil {
		return nil, &TokenError{Message: "Failed to unmarshal JSON to Token", Err: err}
	}
	return &t, nil
}

// FromBase64 creates a Token object from a Base64-encoded JSON string.
func TokenFromBase64(b64Str string) (*Token, error) {
	decoded, err := base64.StdEncoding.DecodeString(b64Str)
	if err != nil {
		return nil, &TokenError{Message: "Failed to decode Base64 string", Err: err}
	}
	return TokenFromJSON(string(decoded))
}

// Update updates the token's access and refresh tokens.
// This method is thread-safe.
func (t *Token) Update(accessToken string, refreshToken *string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.AccessToken = accessToken
	t.RefreshToken = refreshToken
}

// GetAccessToken returns the current access token. This method is thread-safe.
func (t *Token) GetAccessToken() string {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.AccessToken
}

// GetRefreshToken returns the current refresh token. This method is thread-safe.
func (t *Token) GetRefreshToken() *string {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.RefreshToken
}

// GetDeviceCode returns the current device code. This method is thread-safe.
func (t *Token) GetDeviceCode() *string {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.DeviceCode
}
