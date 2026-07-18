package library

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// Client is an authenticated HTTP client for DriveThruRPG library endpoints.
//
// Client combines SDK configuration and an active bearer token to authenticate all
// outgoing requests. Every method maps to a specific API endpoint and returns a fully
// decoded Go type.
//
// Prefer constructing a Client via the root package's Sdk.LibraryClient, which validates
// both configuration and session state before constructing the client, over calling
// NewClient directly.
type Client struct {
	http   *http.Client
	config Config
	token  string
}

// NewClient creates a new Client from the given configuration and bearer token.
func NewClient(config Config, token string) *Client {
	return &Client{
		http:   &http.Client{},
		config: config,
		token:  token,
	}
}

// endpoint builds the full URL for a versioned API path segment: {base_url}/{api_version}/{path}.
func (c *Client) endpoint(path string) string {
	return fmt.Sprintf("%s/%s/%s", c.config.BaseURL(), c.config.APIVersion(), path)
}

func (c *Client) newRequest(ctx context.Context, method, url string, body io.Reader) (*http.Request, error) {
	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return nil, err
	}
	// The DTRPG API expects the raw JWT token without a "Bearer" prefix. Library requests
	// never attach an applicationKey query parameter — see go-library-client spec.
	req.Header.Set("Authorization", c.token)
	return req, nil
}

// decodeResponse reads a response body and decodes it as T.
//
// A non-success status is treated as a request failure rather than a decode attempt: the
// body is never decoded as T in that case (T describes the success schema, so trying to
// parse an error body against it would produce a confusing "missing field" error instead of
// the API's actual message). Instead a human-readable message is extracted from the body —
// a top-level "message" field, or field-keyed validation errors — and returned via
// *APIError.
//
// On a success status whose body still fails to decode as T, the raw payload is preserved
// (truncated to logPayloadLimit bytes) in a *DecodeError so callers have both the decode
// cause and the offending payload for diagnosis.
func decodeResponse[T any](url string, resp *http.Response) (T, error) {
	var zero T

	retryAfter := parseRetryAfter(resp.Header.Get("Retry-After"))

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return zero, err
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return zero, &APIError{
			URL:        url,
			Status:     resp.StatusCode,
			Message:    extractErrorMessage(body),
			Payload:    truncatedPayload(body),
			RetryAfter: retryAfter,
		}
	}

	var value T
	if err := json.Unmarshal(body, &value); err != nil {
		return zero, &DecodeError{
			URL:     url,
			Status:  resp.StatusCode,
			Cause:   err,
			Payload: truncatedPayload(body),
		}
	}
	return value, nil
}

// parseRetryAfter parses a Retry-After header value as a non-negative integer number of
// delay-seconds (RFC 9110 §10.2.3). HTTP-date values and any other unparseable input yield
// nil rather than an error.
func parseRetryAfter(header string) *time.Duration {
	header = strings.TrimSpace(header)
	if header == "" {
		return nil
	}
	seconds, err := strconv.ParseUint(header, 10, 64)
	if err != nil {
		return nil
	}
	const maxRetryAfterSeconds = uint64((1<<63 - 1) / int64(time.Second))
	if seconds > maxRetryAfterSeconds {
		return nil
	}
	d := time.Duration(seconds) * time.Second
	return &d
}

// ── Ordered Products ────────────────────────────────────────────────────────────────────

// ListOrderProducts fetches a paginated list of ordered products from the authenticated
// user's library.
//
// Maps to GET /{api_version}/order_products. All non-nil fields of params are included as
// query parameters.
func (c *Client) ListOrderProducts(ctx context.Context, params LibraryItemsParams) (OrderProductListResponse, error) {
	endpoint := c.endpoint("order_products")

	query := url.Values{}
	if params.Page != nil {
		query.Set("page", strconv.FormatUint(uint64(*params.Page), 10))
	}
	if params.PageSize != nil {
		query.Set("pageSize", strconv.FormatUint(uint64(*params.PageSize), 10))
	}
	if params.GetChecksum != nil && *params.GetChecksum {
		query.Set("getChecksum", "1")
	}
	if params.GetFilters != nil && *params.GetFilters {
		query.Set("getFilters", "1")
	}
	if params.Library != nil && *params.Library {
		query.Set("library", "true")
	}
	if params.Archived != nil {
		if *params.Archived {
			query.Set("archived", "1")
		} else {
			query.Set("archived", "0")
		}
	}
	if params.UpdatedDateAfter != nil {
		query.Set("updatedDate[after]", *params.UpdatedDateAfter)
	}

	reqURL := endpoint
	if encoded := query.Encode(); encoded != "" {
		reqURL = endpoint + "?" + encoded
	}

	req, err := c.newRequest(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		var zero OrderProductListResponse
		return zero, err
	}
	resp, err := c.http.Do(req)
	if err != nil {
		var zero OrderProductListResponse
		return zero, err
	}
	defer resp.Body.Close()

	return decodeResponse[OrderProductListResponse](endpoint, resp)
}

// GetOrderProduct fetches the details of a single ordered product by its identifier.
//
// Maps to GET /{api_version}/order_products/{order_product_id}.
func (c *Client) GetOrderProduct(ctx context.Context, orderProductID uint64) (OrderProductItemResponse, error) {
	endpoint := c.endpoint(fmt.Sprintf("order_products/%d", orderProductID))

	req, err := c.newRequest(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		var zero OrderProductItemResponse
		return zero, err
	}
	resp, err := c.http.Do(req)
	if err != nil {
		var zero OrderProductItemResponse
		return zero, err
	}
	defer resp.Body.Close()

	return decodeResponse[OrderProductItemResponse](endpoint, resp)
}

// PrepareDownload prepares a download for the given ordered product's file and returns the
// raw API response.
//
// Maps to GET /{api_version}/order_products/{order_product_id}/prepare?index={index}.
// index identifies which file within the ordered product to prepare — it matches
// OrderProductFile.Index — and is required: the API rejects the request with an error if
// it is omitted, so there is no overload that defaults it.
//
// The response is returned as a map because the response schema for this endpoint has not
// yet been formally defined by the API contract.
func (c *Client) PrepareDownload(ctx context.Context, orderProductID uint64, index uint32) (map[string]any, error) {
	base := c.endpoint(fmt.Sprintf("order_products/%d/prepare", orderProductID))
	reqURL := base + "?index=" + strconv.FormatUint(uint64(index), 10)

	req, err := c.newRequest(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return decodeResponse[map[string]any](base, resp)
}

// ── Product Lists ───────────────────────────────────────────────────────────────────────

// ListProductLists fetches a paginated list of product lists belonging to the
// authenticated user.
//
// Maps to GET /{api_version}/product_lists. Pagination is controlled via params.
func (c *Client) ListProductLists(ctx context.Context, params PageParams) (ProductListCollectionResponse, error) {
	endpoint := c.endpoint("product_lists")

	query := url.Values{}
	if params.Page != nil {
		query.Set("page", strconv.FormatUint(uint64(*params.Page), 10))
	}
	if params.PageSize != nil {
		query.Set("pageSize", strconv.FormatUint(uint64(*params.PageSize), 10))
	}

	reqURL := endpoint
	if encoded := query.Encode(); encoded != "" {
		reqURL = endpoint + "?" + encoded
	}

	req, err := c.newRequest(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		var zero ProductListCollectionResponse
		return zero, err
	}
	resp, err := c.http.Do(req)
	if err != nil {
		var zero ProductListCollectionResponse
		return zero, err
	}
	defer resp.Body.Close()

	return decodeResponse[ProductListCollectionResponse](endpoint, resp)
}

// ListProductListItems fetches a paginated list of items within a specific product list.
//
// Maps to GET /{api_version}/product_list_items?productListId={product_list_id}.
// Pagination is controlled via params.
func (c *Client) ListProductListItems(ctx context.Context, productListID uint64, params PageParams) (ProductListItemsResponse, error) {
	endpoint := c.endpoint("product_list_items")

	query := url.Values{}
	query.Set("productListId", strconv.FormatUint(productListID, 10))
	if params.Page != nil {
		query.Set("page", strconv.FormatUint(uint64(*params.Page), 10))
	}
	if params.PageSize != nil {
		query.Set("pageSize", strconv.FormatUint(uint64(*params.PageSize), 10))
	}

	reqURL := endpoint + "?" + query.Encode()

	req, err := c.newRequest(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		var zero ProductListItemsResponse
		return zero, err
	}
	resp, err := c.http.Do(req)
	if err != nil {
		var zero ProductListItemsResponse
		return zero, err
	}
	defer resp.Body.Close()

	return decodeResponse[ProductListItemsResponse](endpoint, resp)
}

// CreateProductList creates a new product list with the given name.
//
// Maps to POST /{api_version}/product_lists with a JSON body {"name": "<name>"}.
func (c *Client) CreateProductList(ctx context.Context, name string) (ProductListItem, error) {
	endpoint := c.endpoint("product_lists")

	body, err := json.Marshal(map[string]string{"name": name})
	if err != nil {
		var zero ProductListItem
		return zero, err
	}

	req, err := c.newRequest(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		var zero ProductListItem
		return zero, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		var zero ProductListItem
		return zero, err
	}
	defer resp.Body.Close()

	envelope, err := decodeResponse[productListEnvelope](endpoint, resp)
	if err != nil {
		var zero ProductListItem
		return zero, err
	}
	return envelope.Data, nil
}

// DeleteProductList deletes a product list by id.
//
// Maps to DELETE /{api_version}/product_lists/{id}. Unlike the Rust SDK, this method
// routes its response through decodeResponse (via the anonymous struct{} decode target)
// so failures produce a consistent *APIError with an extracted message and RetryAfter,
// rather than a bare transport error — see go-library-client and rate-limit-retry-after.
func (c *Client) DeleteProductList(ctx context.Context, id uint64) error {
	endpoint := c.endpoint(fmt.Sprintf("product_lists/%d", id))

	req, err := c.newRequest(ctx, http.MethodDelete, endpoint, nil)
	if err != nil {
		return err
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return decodeEmptyResponse(endpoint, resp)
}

// AddProductListItem adds a product to a product list as a member.
//
// Maps to POST /{api_version}/product_list_items.
func (c *Client) AddProductListItem(ctx context.Context, productListID, productID uint64) (ProductListItemCreateResponse, error) {
	endpoint := c.endpoint("product_list_items")

	body, err := json.Marshal(ProductListItemCreateRequest{
		ProductID:     productID,
		ProductListID: productListID,
	})
	if err != nil {
		var zero ProductListItemCreateResponse
		return zero, err
	}

	req, err := c.newRequest(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		var zero ProductListItemCreateResponse
		return zero, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		var zero ProductListItemCreateResponse
		return zero, err
	}
	defer resp.Body.Close()

	return decodeResponse[ProductListItemCreateResponse](endpoint, resp)
}

// DeleteProductListItem removes a product list item by its own id (not the product's id).
//
// Maps to DELETE /{api_version}/product_list_items/{product_list_item_id}. Like
// DeleteProductList, this routes through decodeResponse for consistent error decoding.
func (c *Client) DeleteProductListItem(ctx context.Context, productListItemID uint64) error {
	endpoint := c.endpoint(fmt.Sprintf("product_list_items/%d", productListItemID))

	req, err := c.newRequest(ctx, http.MethodDelete, endpoint, nil)
	if err != nil {
		return err
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return decodeEmptyResponse(endpoint, resp)
}

// decodeEmptyResponse decodes a response expected to have no meaningful body (e.g. HTTP
// 204 from a DELETE), still routing non-success statuses through the same *APIError path
// as decodeResponse.
func decodeEmptyResponse(url string, resp *http.Response) error {
	retryAfter := parseRetryAfter(resp.Header.Get("Retry-After"))

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return &APIError{
			URL:        url,
			Status:     resp.StatusCode,
			Message:    extractErrorMessage(body),
			Payload:    truncatedPayload(body),
			RetryAfter: retryAfter,
		}
	}
	return nil
}
