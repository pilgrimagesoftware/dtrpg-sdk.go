// Package dtrpg is the root of the DriveThruRPG Go SDK. It wires together configuration
// (aliased from the library package), the authentication session lifecycle (package auth),
// and the library API client (package library) behind a single Sdk entry point.
package dtrpg

import "errors"

// ErrUnconfigured indicates the SDK has not been configured with a Config yet.
//
// Call Sdk.Configure or construct the SDK with NewSdkWithConfig before making API calls.
var ErrUnconfigured = errors.New("dtrpg: sdk is not configured")

// ErrUnauthenticated indicates the SDK has no active authentication session.
//
// Obtain a session by calling Sdk.ApplyAuthResponse with a successful token response from
// the API.
var ErrUnauthenticated = errors.New("dtrpg: sdk does not have an authenticated session")
