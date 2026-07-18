package dtrpg

import (
	"github.com/pilgrimagesoftware/dtrpg-sdk.go/auth"
	"github.com/pilgrimagesoftware/dtrpg-sdk.go/library"
)

// Config holds the application-level settings required to make requests to the
// DriveThruRPG API. See library.Config for field documentation; it lives in the library
// package to avoid an import cycle between this package, auth, and library, and is aliased
// here for the common single-import call site.
type Config = library.Config

// DefaultBaseURL is the production DriveThruRPG API base URL used when no custom URL is
// provided.
const DefaultBaseURL = library.DefaultBaseURL

// DefaultAPIVersion is the API version path segment used when no custom version is
// provided.
const DefaultAPIVersion = library.DefaultAPIVersion

// NewConfig creates a Config targeting the production API with the given applicationKey.
func NewConfig(applicationKey string) Config {
	return library.NewConfig(applicationKey)
}

// NewConfigWithBaseURL creates a Config with a custom baseURL, overriding the production
// endpoint.
func NewConfigWithBaseURL(applicationKey, baseURL string) Config {
	return library.NewConfigWithBaseURL(applicationKey, baseURL)
}

// AuthSession is an active authentication session with the DriveThruRPG API. See
// auth.AuthSession for field documentation.
type AuthSession = auth.AuthSession

// AuthState is the authentication failure state reported by the DriveThruRPG API. See
// auth.AuthState.
type AuthState = auth.AuthState

// AuthTokenResponse is the raw authentication token payload returned by the DriveThruRPG
// API. See auth.AuthTokenResponse.
type AuthTokenResponse = auth.AuthTokenResponse

// AuthSessionError is a structured authentication error returned by the DriveThruRPG API.
// See auth.AuthSessionError.
type AuthSessionError = auth.AuthSessionError

// SessionTransition is the outcome of invalidating an AuthSession. See
// auth.SessionTransition.
type SessionTransition = auth.SessionTransition

// NewAuthTokenResponse constructs a new AuthTokenResponse from its fields.
func NewAuthTokenResponse(token, refreshToken string, refreshTokenTTL uint64) AuthTokenResponse {
	return auth.NewAuthTokenResponse(token, refreshToken, refreshTokenTTL)
}

// LibraryClient is an authenticated HTTP client for DriveThruRPG library endpoints. See
// library.Client.
type LibraryClient = library.Client

// LibraryItemsParams holds query parameters for the order-products (library items)
// endpoint. See library.LibraryItemsParams.
type LibraryItemsParams = library.LibraryItemsParams

// PageParams holds query parameters for paginated collection endpoints. See
// library.PageParams.
type PageParams = library.PageParams
