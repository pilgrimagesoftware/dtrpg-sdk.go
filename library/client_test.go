package library

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func clientFor(server *httptest.Server) *Client {
	config := NewConfigWithBaseURL("test-app-key", server.URL)
	return NewClient(config, "test-token")
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}

func TestListOrderProductsSendsAuthorizationHeaderAndQuery(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("Authorization"); got != "test-token" {
			t.Errorf("Authorization = %q, want raw token %q (no Bearer prefix)", got, "test-token")
		}
		if r.URL.Query().Get("applicationKey") != "" {
			t.Error("request must not include an applicationKey query parameter")
		}
		if r.URL.Query().Get("page") != "2" {
			t.Errorf("page = %q, want %q", r.URL.Query().Get("page"), "2")
		}
		if r.URL.Query().Get("getChecksum") != "1" {
			t.Errorf("getChecksum = %q, want %q", r.URL.Query().Get("getChecksum"), "1")
		}
		writeJSON(w, http.StatusOK, map[string]any{
			"links": map[string]any{"self": "x"},
			"meta":  map[string]any{"itemsPerPage": 25, "currentPage": 2},
			"data":  []any{},
		})
	}))
	defer server.Close()

	page := uint32(2)
	getChecksum := true
	client := clientFor(server)
	_, err := client.ListOrderProducts(context.Background(), LibraryItemsParams{Page: &page, GetChecksum: &getChecksum})
	if err != nil {
		t.Fatalf("ListOrderProducts() error = %v", err)
	}
}

func TestGetOrderProductDecodesSideloadedIncludedArray(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/vBeta/order_products/22654728" {
			t.Errorf("path = %s", r.URL.Path)
		}
		writeJSON(w, http.StatusOK, map[string]any{
			"data": map[string]any{
				"id":   "/api/vBeta/order_products/22654728",
				"type": "order_product",
				"attributes": map[string]any{
					"orderId":            7332333,
					"productId":          144239,
					"royaltyPublisherId": 117,
					"name":               "Common Places - Free Map #1",
					"finalPrice":         0.0,
					"quantity":           1,
					"bundleId":           0,
					"archived":           0,
					"orderProductId":     22654728,
					"customerId":         399144,
					"files":              []any{},
				},
				"relationships": map[string]any{
					"publisher": map[string]any{
						"data": map[string]any{"type": "Publisher", "id": "/api/vBeta/publishers/117"},
					},
				},
			},
			"included": []any{
				map[string]any{
					"id":   "/api/vBeta/publishers/117",
					"type": "Publisher",
					"attributes": map[string]any{
						"name":        "The Forge Studios",
						"publisherId": 117,
						"slug":        "the-forge-studios",
					},
				},
			},
		})
	}))
	defer server.Close()

	client := clientFor(server)
	result, err := client.GetOrderProduct(context.Background(), 22654728)
	if err != nil {
		t.Fatalf("GetOrderProduct() error = %v", err)
	}
	if len(result.Included) != 1 {
		t.Fatalf("Included has %d items, want 1", len(result.Included))
	}
	publisher := result.Included[0].AsPublisher()
	if publisher == nil {
		t.Fatal("AsPublisher() = nil, want decoded publisher")
	}
	if publisher.Name != "The Forge Studios" {
		t.Errorf("publisher.Name = %q, want %q", publisher.Name, "The Forge Studios")
	}
}

func TestPrepareDownloadRequiresIndex(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("index") != "0" {
			t.Errorf("index = %q, want %q", r.URL.Query().Get("index"), "0")
		}
		writeJSON(w, http.StatusOK, map[string]any{"url": "https://example.com/download"})
	}))
	defer server.Close()

	client := clientFor(server)
	result, err := client.PrepareDownload(context.Background(), 515276, 0)
	if err != nil {
		t.Fatalf("PrepareDownload() error = %v", err)
	}
	if result["url"] != "https://example.com/download" {
		t.Errorf("result = %+v", result)
	}
}

func TestCreateProductListDecodesJSONAPIEnvelope(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusCreated, map[string]any{
			"data": map[string]any{
				"id":   "/api/vBeta/product_lists/86267",
				"type": "ProductList",
				"attributes": map[string]any{
					"customerId":    399144,
					"name":          "Testing",
					"dateCreated":   "2026-07-09T00:42:39-05:00",
					"productListId": 86267,
					"slug":          "testing",
					"itemCount":     0,
				},
			},
		})
	}))
	defer server.Close()

	client := clientFor(server)
	result, err := client.CreateProductList(context.Background(), "Testing")
	if err != nil {
		t.Fatalf("CreateProductList() error = %v", err)
	}
	if result.ID != "/api/vBeta/product_lists/86267" {
		t.Errorf("ID = %q", result.ID)
	}
	if result.Attributes.ProductListID != 86267 {
		t.Errorf("ProductListID = %d, want 86267", result.Attributes.ProductListID)
	}
}

func TestAddProductListItemReturnsCreatedItem(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body ProductListItemCreateRequest
		_ = json.NewDecoder(r.Body).Decode(&body)
		if body.ProductID != 515276 || body.ProductListID != 86151 {
			t.Errorf("body = %+v", body)
		}
		writeJSON(w, http.StatusCreated, map[string]any{
			"data": map[string]any{
				"id":   "/api/vBeta/product_list_items/2629321",
				"type": "ProductListItem",
				"attributes": map[string]any{
					"productId":         515276,
					"productListId":     86151,
					"productListItemId": 2629321,
				},
			},
		})
	}))
	defer server.Close()

	client := clientFor(server)
	result, err := client.AddProductListItem(context.Background(), 86151, 515276)
	if err != nil {
		t.Fatalf("AddProductListItem() error = %v", err)
	}
	if result.ProductID != 515276 || result.ProductListID != 86151 || result.ProductListItemID != 2629321 {
		t.Errorf("result = %+v", result)
	}
}

func TestAddProductListItemReturnsAPIErrorOnFailureStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	client := clientFor(server)
	_, err := client.AddProductListItem(context.Background(), 86151, 515276)

	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("error = %v (%T), want *APIError", err, err)
	}
	if apiErr.Status != 404 {
		t.Errorf("Status = %d, want 404", apiErr.Status)
	}
	if apiErr.RetryAfter != nil {
		t.Errorf("RetryAfter = %v, want nil", apiErr.RetryAfter)
	}
}

func TestAddProductListItemReturnsRetryAfterOn429(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Retry-After", "30")
		w.WriteHeader(http.StatusTooManyRequests)
	}))
	defer server.Close()

	client := clientFor(server)
	_, err := client.AddProductListItem(context.Background(), 86151, 515276)

	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("error = %v (%T), want *APIError", err, err)
	}
	if apiErr.Status != 429 {
		t.Errorf("Status = %d, want 429", apiErr.Status)
	}
	if apiErr.RetryAfter == nil || *apiErr.RetryAfter != 30*time.Second {
		t.Errorf("RetryAfter = %v, want 30s", apiErr.RetryAfter)
	}
}

func TestAddProductListItemIgnoresUnparseableRetryAfter(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Retry-After", "Wed, 21 Oct 2026 07:28:00 GMT")
		w.WriteHeader(http.StatusTooManyRequests)
	}))
	defer server.Close()

	client := clientFor(server)
	_, err := client.AddProductListItem(context.Background(), 86151, 515276)

	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("error = %v (%T), want *APIError", err, err)
	}
	if apiErr.RetryAfter != nil {
		t.Errorf("RetryAfter = %v, want nil for an HTTP-date value", apiErr.RetryAfter)
	}
}

func TestAddProductListItemSurfacesValidationMessageOnConflict(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusConflict, map[string]any{
			"error": map[string]any{
				"id":      "6a4ee5880bfda",
				"message": "productId: Requires a valid Product ID. Invalid value 22654728.",
				"code":    409,
				"status":  409,
			},
		})
	}))
	defer server.Close()

	client := clientFor(server)
	_, err := client.AddProductListItem(context.Background(), 86151, 22654728)

	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("error = %v (%T), want *APIError", err, err)
	}
	want := "productId: Requires a valid Product ID. Invalid value 22654728."
	if apiErr.Message == nil || *apiErr.Message != want {
		t.Errorf("Message = %v, want %q", apiErr.Message, want)
	}
}

func TestDeleteProductListItemSucceedsOnNoContent(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("method = %s, want DELETE", r.Method)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client := clientFor(server)
	if err := client.DeleteProductListItem(context.Background(), 2629321); err != nil {
		t.Fatalf("DeleteProductListItem() error = %v", err)
	}
}

func TestDeleteProductListItemReturnsAPIErrorOnFailureStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	client := clientFor(server)
	err := client.DeleteProductListItem(context.Background(), 2629321)

	// Unlike the Rust SDK (which surfaces a bare transport error here via
	// error_for_status()), the Go client routes deletes through the same decode path as
	// every other endpoint, per go-library-client and rate-limit-retry-after.
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("error = %v (%T), want *APIError", err, err)
	}
	if apiErr.Status != 404 {
		t.Errorf("Status = %d, want 404", apiErr.Status)
	}
}

func TestDeleteProductListReturnsRetryAfterOnFailure(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Retry-After", "5")
		w.WriteHeader(http.StatusTooManyRequests)
	}))
	defer server.Close()

	client := clientFor(server)
	err := client.DeleteProductList(context.Background(), 86151)

	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("error = %v (%T), want *APIError", err, err)
	}
	if apiErr.RetryAfter == nil || *apiErr.RetryAfter != 5*time.Second {
		t.Errorf("RetryAfter = %v, want 5s", apiErr.RetryAfter)
	}
}

func TestListProductListsAndItems(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/vBeta/product_lists":
			writeJSON(w, http.StatusOK, map[string]any{
				"links": map[string]any{"self": "x"},
				"meta":  map[string]any{"itemsPerPage": 25, "currentPage": 1},
				"data":  []any{},
			})
		case "/vBeta/product_list_items":
			if r.URL.Query().Get("productListId") != "86151" {
				t.Errorf("productListId = %q", r.URL.Query().Get("productListId"))
			}
			writeJSON(w, http.StatusOK, map[string]any{
				"links": map[string]any{"self": "x"},
				"meta":  map[string]any{"itemsPerPage": 25, "currentPage": 1},
				"data":  []any{},
			})
		default:
			t.Errorf("unexpected path %s", r.URL.Path)
		}
	}))
	defer server.Close()

	client := clientFor(server)
	if _, err := client.ListProductLists(context.Background(), PageParams{}); err != nil {
		t.Fatalf("ListProductLists() error = %v", err)
	}
	if _, err := client.ListProductListItems(context.Background(), 86151, PageParams{}); err != nil {
		t.Fatalf("ListProductListItems() error = %v", err)
	}
}
