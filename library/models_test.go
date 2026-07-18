package library

import (
	"encoding/json"
	"testing"
)

func TestOrderProductFileDecodesNullChecksumsAsEmptySlice(t *testing.T) {
	data := []byte(`{
		"index": 0,
		"orderProductDownloadId": 1,
		"title": "Book",
		"filename": "book.pdf",
		"size": 100,
		"sizeMB": "0.1",
		"checksums": null
	}`)

	var file OrderProductFile
	if err := json.Unmarshal(data, &file); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}
	if len(file.Checksums) != 0 {
		t.Errorf("Checksums = %+v, want empty", file.Checksums)
	}
}

func TestOrderProductFileDecodesPresentChecksums(t *testing.T) {
	data := []byte(`{
		"index": 0,
		"orderProductDownloadId": 1,
		"title": "Book",
		"filename": "book.pdf",
		"size": 100,
		"sizeMB": "0.1",
		"checksums": [{"checksum": "abc123", "checksumDate": "2026-01-01"}]
	}`)

	var file OrderProductFile
	if err := json.Unmarshal(data, &file); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}
	if len(file.Checksums) != 1 || file.Checksums[0].Checksum != "abc123" {
		t.Errorf("Checksums = %+v", file.Checksums)
	}
}

func TestProductListItemCreateResponseUnwrapsEnvelope(t *testing.T) {
	data := []byte(`{
		"data": {
			"id": "/api/vBeta/product_list_items/2634593",
			"type": "ProductListItem",
			"attributes": {
				"productId": 144239,
				"productListId": 86267,
				"productListItemId": 2634593
			}
		}
	}`)

	var response ProductListItemCreateResponse
	if err := json.Unmarshal(data, &response); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}
	if response.ProductID != 144239 || response.ProductListID != 86267 || response.ProductListItemID != 2634593 {
		t.Errorf("response = %+v", response)
	}
}

func TestPaginationLinksUsesSelfKeyword(t *testing.T) {
	data := []byte(`{"self": "url-a", "first": "url-b", "last": null, "prev": null, "next": null}`)

	var links PaginationLinks
	if err := json.Unmarshal(data, &links); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}
	if links.Self != "url-a" {
		t.Errorf("Self = %q, want %q", links.Self, "url-a")
	}
	if links.First == nil || *links.First != "url-b" {
		t.Errorf("First = %v, want url-b", links.First)
	}
	if links.Last != nil {
		t.Errorf("Last = %v, want nil", links.Last)
	}
}

func TestIncludedItemAsPublisherAndAsProduct(t *testing.T) {
	publisherJSON := []byte(`{
		"id": "/api/vBeta/publishers/117",
		"type": "Publisher",
		"attributes": {"name": "The Forge Studios", "publisherId": 117, "slug": "the-forge-studios"}
	}`)
	var publisherItem IncludedItem
	if err := json.Unmarshal(publisherJSON, &publisherItem); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}
	publisher := publisherItem.AsPublisher()
	if publisher == nil || publisher.Name != "The Forge Studios" {
		t.Errorf("AsPublisher() = %v", publisher)
	}
	if publisherItem.AsProduct() != nil {
		t.Error("AsProduct() on a Publisher item should be nil")
	}
}
