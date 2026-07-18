// Package openapi exposes build-time metadata generated from the dtrpg-api submodule's
// openapi.yaml: the default server URL and the list of operations (method + path) it
// defines. This is metadata-only — no request/response types are generated — used for
// contract-freshness checking against the API/openapi.yaml submodule.
//
// Run `go generate ./...` after API/openapi.yaml changes to regenerate generated.go.
package openapi

//go:generate go run ../internal/opgen -in ../API/openapi.yaml -out generated.go

// Operation is a single HTTP method + path pair declared in the OpenAPI spec.
type Operation struct {
	// Method is the HTTP method in uppercase (e.g. "GET", "POST").
	Method string
	// Path is the OpenAPI path template (e.g. "/{DTRPG_API_VERSION}/order_products").
	Path string
}
