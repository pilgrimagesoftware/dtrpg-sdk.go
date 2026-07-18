package auth

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/pilgrimagesoftware/dtrpg-sdk.go/library"
)

// logPayloadLimit is the maximum number of bytes logged from a failing auth response body.
const logPayloadLimit = 2000

// Authenticate exchanges a DriveThruRPG application key for a session token.
//
// Posts to POST /{api_version}/auth_key with applicationKey as a query parameter. On
// success, returns the JWT access token, refresh token, and refresh token TTL.
//
// Authenticate is the only SDK operation that does not require a pre-existing AuthSession.
func Authenticate(ctx context.Context, apiKey string, config library.Config) (AuthTokenResponse, error) {
	url := fmt.Sprintf("%s/%s/auth_key?applicationKey=%s", config.BaseURL(), config.APIVersion(), apiKey)

	// The application key is passed as a query parameter, not in the body, but
	// openapi.yaml still declares an application/json requestBody for this endpoint; send
	// an empty object so the Content-Type matches the spec.
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader([]byte("{}")))
	if err != nil {
		var zero AuthTokenResponse
		return zero, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		var zero AuthTokenResponse
		return zero, err
	}
	defer resp.Body.Close()

	status := resp.StatusCode
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		var zero AuthTokenResponse
		return zero, err
	}

	var response AuthTokenResponse
	if err := json.Unmarshal(body, &response); err != nil {
		payload := body
		if len(payload) > logPayloadLimit {
			payload = payload[:logPayloadLimit]
		}
		var zero AuthTokenResponse
		return zero, &library.DecodeError{
			URL:     url,
			Status:  status,
			Cause:   err,
			Payload: string(payload),
		}
	}
	return response, nil
}
