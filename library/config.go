// Package library provides the DriveThruRPG library API client, its request/response
// types, and the SDK's Config type (kept here rather than the root package to avoid an
// import cycle between the root, auth, and library packages).
package library

// DefaultBaseURL is the production DriveThruRPG API base URL used when no custom URL is
// provided.
const DefaultBaseURL = "https://api.drivethrurpg.com/api"

// DefaultAPIVersion is the API version path segment used when no custom version is
// provided.
const DefaultAPIVersion = "vBeta"

// Config holds the application-level settings required to make requests to the
// DriveThruRPG API: the application key, the API base URL, and the API version segment
// used in request URLs.
//
// A Config must be provided before any authenticated API calls can be made. It binds an
// application key to an API endpoint, defaulting to the production DriveThruRPG API and
// the current "vBeta" API version.
type Config struct {
	applicationKey string
	baseURL        string
	apiVersion     string
}

// NewConfig creates a Config targeting the production API with the given applicationKey.
// The base URL defaults to DefaultBaseURL and the API version defaults to
// DefaultAPIVersion.
func NewConfig(applicationKey string) Config {
	return Config{
		applicationKey: applicationKey,
		baseURL:        DefaultBaseURL,
		apiVersion:     DefaultAPIVersion,
	}
}

// NewConfigWithBaseURL creates a Config with a custom baseURL, overriding the production
// endpoint. The API version defaults to DefaultAPIVersion. Useful for pointing the SDK at
// a staging server or a local mock during development.
func NewConfigWithBaseURL(applicationKey, baseURL string) Config {
	return Config{
		applicationKey: applicationKey,
		baseURL:        baseURL,
		apiVersion:     DefaultAPIVersion,
	}
}

// ApplicationKey returns the application key used to identify this client to the API.
func (c Config) ApplicationKey() string {
	return c.applicationKey
}

// BaseURL returns the base URL that SDK requests will be sent to.
func (c Config) BaseURL() string {
	return c.baseURL
}

// APIVersion returns the API version path segment used when constructing endpoint URLs.
//
// This value is inserted between the base URL and the resource path:
// {base_url}/{api_version}/{resource}. Defaults to "vBeta".
func (c Config) APIVersion() string {
	return c.apiVersion
}
