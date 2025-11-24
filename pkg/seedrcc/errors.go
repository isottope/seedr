package seedrcc

import (
	"encoding/json"
	"fmt"
)

// SeedrError is the base interface for all custom Seedr errors.
type SeedrError interface {
	error
	isSeedrError()
}

// APIError represents an error returned by the Seedr API.
type APIError struct {
	Message    string
	StatusCode int
	Code       int    `json:"code,omitempty"`
	ErrorType  string `json:"result,omitempty"` // Corresponds to 'result' in some error responses
	Response   []byte // Raw response body for further inspection
}

func (e *APIError) Error() string {
	if e.Code != 0 || e.ErrorType != "" {
		return fmt.Sprintf("API error: %s (Status: %d, Code: %d, Type: %s)", e.Message, e.StatusCode, e.Code, e.ErrorType)
	}
	return fmt.Sprintf("API error: %s (Status: %d)", e.Message, e.StatusCode)
}
func (e *APIError) isSeedrError() {}

// ServerError represents a 5xx server-side error.
type ServerError struct {
	Message    string
	StatusCode int
	Response   []byte // Raw response body for further inspection
}

func (e *ServerError) Error() string {
	return fmt.Sprintf("Server error: %s (Status: %d)", e.Message, e.StatusCode)
}
func (e *ServerError) isSeedrError() {}

// AuthenticationError represents an error during authentication or token refresh.
type AuthenticationError struct {
	Message    string
	StatusCode int
	ErrorType  string `json:"error,omitempty"` // Corresponds to 'error' in OAuth responses
	Response   []byte // Raw response body for further inspection
}

func (e *AuthenticationError) Error() string {
	if e.ErrorType != "" {
		return fmt.Sprintf("Authentication error: %s (Status: %d, Type: %s)", e.Message, e.StatusCode, e.ErrorType)
	}
	return fmt.Sprintf("Authentication error: %s (Status: %d)", e.Message, e.StatusCode)
}
func (e *AuthenticationError) isSeedrError() {}

// NetworkError represents a network-level error, such as timeouts or connection problems.
type NetworkError struct {
	Message string
	Err     error
}

func (e *NetworkError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("Network error: %s: %v", e.Message, e.Err)
	}
	return fmt.Sprintf("Network error: %s", e.Message)
}
func (e *NetworkError) isSeedrError() {}

// TokenError represents errors related to token serialization or deserialization.
type TokenError struct {
	Message string
	Err     error
}

func (e *TokenError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("Token error: %s: %v", e.Message, e.Err)
	}
	return fmt.Sprintf("Token error: %s", e.Message)
}
func (e *TokenError) isSeedrError() {}

// NewAPIError creates an APIError instance and attempts to parse additional details from the response body.
func NewAPIError(message string, statusCode int, responseBody []byte) *APIError {
	apiErr := &APIError{
		Message:    message,
		StatusCode: statusCode,
		Response:   responseBody,
	}

	if len(responseBody) > 0 {
		var data map[string]interface{}
		if err := json.Unmarshal(responseBody, &data); err == nil {
			if code, ok := data["code"].(float64); ok {
				apiErr.Code = int(code)
			}
			if result, ok := data["result"].(string); ok {
				apiErr.ErrorType = result
			}
		}
	}
	return apiErr
}

// NewAuthenticationError creates an AuthenticationError instance and attempts to parse additional details.
func NewAuthenticationError(message string, statusCode int, responseBody []byte) *AuthenticationError {
	authErr := &AuthenticationError{
		Message:    message,
		StatusCode: statusCode,
		Response:   responseBody,
	}

	if len(responseBody) > 0 {
		var data map[string]interface{}
		if err := json.Unmarshal(responseBody, &data); err == nil {
			if errorDesc, ok := data["error_description"].(string); ok {
				authErr.Message = errorDesc // Use more specific message if available
			}
			if errorType, ok := data["error"].(string); ok {
				authErr.ErrorType = errorType
			}
		}
	}
	return authErr
}
