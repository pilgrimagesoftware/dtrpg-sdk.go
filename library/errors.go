package library

import (
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"
)

// logPayloadLimit is the maximum number of bytes logged from a failing response body.
const logPayloadLimit = 2000

// ErrInvalidCredentials is returned by LoginWithCredentials when
// validate_login_credentials.php indicates the supplied email/password pair is invalid.
var ErrInvalidCredentials = errors.New("library: invalid credentials")

// ApplicationKeyRequestError is returned by LoginWithCredentials when credentials pass
// validation but create_account_app.php returns a non-success status.
type ApplicationKeyRequestError struct {
	// Status is the status string returned by create_account_app.php.
	Status string
}

func (e *ApplicationKeyRequestError) Error() string {
	return fmt.Sprintf("application key request failed (status: %s)", e.Status)
}

// DecodeError indicates the HTTP response had a successful status but the body could not
// be decoded into the expected type. The raw response body (truncated to logPayloadLimit
// bytes) is preserved so callers can log the offending payload for diagnosis.
type DecodeError struct {
	// URL is the URL that was requested.
	URL string
	// Status is the HTTP status code of the response.
	Status int
	// Cause is the underlying decode error.
	Cause error
	// Payload is the raw response body, truncated to logPayloadLimit bytes.
	Payload string
}

func (e *DecodeError) Error() string {
	return fmt.Sprintf("response decode failed [%s] (HTTP %d): %v", e.URL, e.Status, e.Cause)
}

func (e *DecodeError) Unwrap() error {
	return e.Cause
}

// APIError indicates the API returned a non-success status. Message, when present, is a
// human-readable explanation extracted from the response body (either a top-level message
// field or field-keyed validation errors, e.g.
// {"productId": "Requires a valid Product ID. Invalid value 22654728."}).
//
// The raw response body (truncated to logPayloadLimit bytes) is preserved so callers can
// log the offending payload when no Message could be extracted.
type APIError struct {
	// URL is the URL that was requested.
	URL string
	// Status is the HTTP status code of the response.
	Status int
	// Message is a human-readable message extracted from the response body, if any.
	Message *string
	// Payload is the raw response body, truncated to logPayloadLimit bytes.
	Payload string
	// RetryAfter is the delay specified by the response's Retry-After header, if present
	// and parseable as a delay-seconds value (RFC 9110 §10.2.3).
	RetryAfter *time.Duration
}

func (e *APIError) Error() string {
	detail := e.Payload
	if e.Message != nil {
		detail = *e.Message
	}
	return fmt.Sprintf("API request failed [%s] (HTTP %d): %s", e.URL, e.Status, detail)
}

// extractErrorMessage extracts a human-readable error message from a non-success JSON
// response body.
//
// Recognizes three shapes seen across DriveThruRPG API error responses: a top-level
// message string; a nested {"error": {"message": "..."}} object (e.g. product_list_items
// failures); or a flat object keyed by field name whose values are a validation message
// string or an array of message strings (e.g. {"productId": "Requires a valid Product ID.
// Invalid value 22654728."}). Returns nil if the body isn't JSON or matches none of these
// shapes, so the caller falls back to the raw payload.
func extractErrorMessage(body []byte) *string {
	var value map[string]any
	if err := json.Unmarshal(body, &value); err != nil {
		return nil
	}

	if message, ok := value["message"].(string); ok {
		return &message
	}

	if errObj, ok := value["error"].(map[string]any); ok {
		if message, ok := errObj["message"].(string); ok {
			return &message
		}
	}

	fields := make([]string, 0, len(value))
	for field := range value {
		fields = append(fields, field)
	}
	sort.Strings(fields)

	var parts []string
	for _, field := range fields {
		switch v := value[field].(type) {
		case string:
			parts = append(parts, fmt.Sprintf("%s: %s", field, v))
		case []any:
			for _, item := range v {
				if message, ok := item.(string); ok {
					parts = append(parts, fmt.Sprintf("%s: %s", field, message))
				}
			}
		}
	}

	if len(parts) == 0 {
		return nil
	}
	joined := strings.Join(parts, "; ")
	return &joined
}

func truncatedPayload(body []byte) string {
	if len(body) > logPayloadLimit {
		return string(body[:logPayloadLimit]) + "… (truncated)"
	}
	return string(body)
}
