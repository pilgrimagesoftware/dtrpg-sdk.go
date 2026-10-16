package auth

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"

	"github.com/pilgrimagesoftware/dtrpg-sdk.go/library"
)

// RefreshToken exchanges a refresh token for a new access token.
//
// Posts to POST /{api_version}/auth/refresh with the token in a refreshToken cookie. On
// success, returns a new access token together with the refresh token and expiry the API
// now considers current; callers should store all three, because the API may rotate the
// refresh token.
//
// The refresh endpoint is not described by the bundled OpenAPI document. Its shape is
// inferred from the Set-Cookie header on the auth_key response, which scopes the
// refreshToken cookie to /api/vBeta/auth/refresh: the token is sent back as that cookie,
// and the response is assumed to carry the same payload as auth_key.
//
// A rejected refresh token, such as an expired or revoked one, surfaces as a
// *library.DecodeError whose Status is the HTTP status of the rejection.
func RefreshToken(ctx context.Context, refreshToken string, config library.Config) (AuthTokenResponse, error) {
	url := fmt.Sprintf("%s/%s/auth/refresh", config.BaseURL(), config.APIVersion())

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader([]byte("{}")))
	if err != nil {
		var zero AuthTokenResponse
		return zero, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(&http.Cookie{Name: "refreshToken", Value: refreshToken})

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		var zero AuthTokenResponse
		return zero, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		var zero AuthTokenResponse
		return zero, err
	}
	return decodeTokenResponse(url, resp.StatusCode, body)
}
