## ADDED Requirements

### Requirement: The Go SDK MUST define typed models for all API-defined library schemas
The Go SDK MUST provide Go structs (or, where a field is a closed set of API-defined
strings, typed string constants) for each library resource schema defined by the API
contract, including ordered products, files, filters, history entries, attributes,
product lists, and pagination structures.

#### Scenario: Deserializing an ordered product response
- **WHEN** the Go SDK receives an ordered product response from the API
- **THEN** it deserializes it via `encoding/json` into the documented Go model types
  derived from the API-defined schemas

#### Scenario: Deserializing a null field as its zero value
- **WHEN** the API returns `null` for a field the Rust SDK treats as null-as-default (e.g.
  `OrderProductFile.Checksums`)
- **THEN** the Go type's `UnmarshalJSON` produces the Go zero value for that field instead
  of a nil-pointer/decode error

### Requirement: Go library types MUST derive their structure from API contract schemas
Go library model types MUST follow the field names, optionality, and nesting defined by
the API repository rather than introducing SDK-local interpretations. JSON field-name
mapping MUST use `json` struct tags, not renamed Go field identifiers.

#### Scenario: Adding a new field to a library resource
- **WHEN** the API contract adds a new field to an ordered product or product list schema
- **THEN** the Go SDK type is updated to reflect the API-defined change rather than a
  local guess

### Requirement: JSON:API envelope responses are unwrapped into flat Go structs
Where the API wraps a resource in a JSON:API `{"data": {"attributes": {...}}}` envelope for
a single-purpose response (e.g. `ProductListItemCreateResponse`), the Go SDK MUST provide
a custom `UnmarshalJSON` that unwraps the envelope so callers receive a flat struct.

#### Scenario: Creating a product list item
- **WHEN** `client.AddProductListItem` receives a JSON:API-enveloped creation response
- **THEN** the returned `ProductListItemCreateResponse` exposes `ProductID`,
  `ProductListID`, and `ProductListItemID` as flat fields, with the envelope already
  unwrapped
