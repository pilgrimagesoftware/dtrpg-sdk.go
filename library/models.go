package library

import "encoding/json"

// nullAsDefault decodes a JSON null (or missing field, when combined with the omitempty
// struct tag on the caller's side) as the type's zero value, instead of leaving the field
// untouched or erroring.
func nullAsDefault[T any](data []byte) (T, error) {
	var value T
	if string(data) == "null" || len(data) == 0 {
		return value, nil
	}
	err := json.Unmarshal(data, &value)
	return value, err
}

// ── Pagination ──────────────────────────────────────────────────────────────────────────

// PaginationLinks holds pagination links included in all paginated API responses.
type PaginationLinks struct {
	Self  string  `json:"self"`
	First *string `json:"first"`
	Last  *string `json:"last"`
	Prev  *string `json:"prev"`
	Next  *string `json:"next"`
}

// PaginationMeta holds pagination metadata included in all paginated API responses.
type PaginationMeta struct {
	ItemsPerPage uint32 `json:"itemsPerPage"`
	CurrentPage  uint32 `json:"currentPage"`
}

// ── File / Checksum ─────────────────────────────────────────────────────────────────────

// FileChecksum holds checksum information for a single downloadable product file.
type FileChecksum struct {
	Checksum     string `json:"checksum"`
	ChecksumDate string `json:"checksumDate"`
}

// OrderProductFile is a downloadable file associated with an ordered product.
type OrderProductFile struct {
	// Index is the position of this file within the ordered product's file list.
	Index                  uint32 `json:"index"`
	OrderProductDownloadID uint64 `json:"orderProductDownloadId"`
	Title                  string `json:"title"`
	Filename               string `json:"filename"`
	Size                   uint64 `json:"size"`
	SizeMB                 string `json:"sizeMB"`
	// Checksums may be null in the API response; a null decodes to an empty slice.
	Checksums []FileChecksum `json:"checksums"`
}

// UnmarshalJSON treats a null Checksums field as an empty slice instead of a decode error.
func (f *OrderProductFile) UnmarshalJSON(data []byte) error {
	type alias OrderProductFile
	var raw struct {
		alias
		Checksums json.RawMessage `json:"checksums"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	checksums, err := nullAsDefault[[]FileChecksum](raw.Checksums)
	if err != nil {
		return err
	}
	*f = OrderProductFile(raw.alias)
	f.Checksums = checksums
	return nil
}

// ── Filters / History / Attributes ──────────────────────────────────────────────────────

// OrderProductFilter is a filter category associated with an ordered product. Populated
// when getFilters=1 is included in the request.
type OrderProductFilter struct {
	FilterID       uint64 `json:"filterId"`
	ParentFilterID uint64 `json:"parentFilterId"`
	Name           string `json:"name"`
	ParentName     string `json:"parentName"`
}

// OrderProductHistoryEntry records a single change made to an ordered product.
type OrderProductHistoryEntry struct {
	Changed string `json:"changed"`
	Changes string `json:"changes"`
}

// OrderProductAttribute is an individual attribute option associated with an ordered
// product, describing purchase options such as format or edition.
type OrderProductAttribute struct {
	OrderID         uint64 `json:"orderId"`
	OptionName      string `json:"optionName"`
	OptionValueName string `json:"optionValueName"`
	Price           string `json:"price"`
	PricePrefix     string `json:"pricePrefix"`
	OptionValueID   uint64 `json:"optionValueId"`
	OptionType      string `json:"optionType"`
}

// ── OrderProduct ─────────────────────────────────────────────────────────────────────────

// OrderProductAttributes is the full attribute set for an ordered product. It includes
// required fields present on every ordered product as well as optional collections
// (filters, history, attributes) that are populated only when specifically requested.
type OrderProductAttributes struct {
	OrderID            uint64  `json:"orderId"`
	ProductID          uint64  `json:"productId"`
	RoyaltyPublisherID uint64  `json:"royaltyPublisherId"`
	ISBN               *string `json:"isbn"`
	Name               string  `json:"name"`
	DatePurchased      *string `json:"datePurchased"`
	Filesize           *uint64 `json:"filesize"`
	FinalPrice         float64 `json:"finalPrice"`
	Quantity           uint32  `json:"quantity"`
	BundleID           uint64  `json:"bundleId"`
	// Archived indicates whether the product has been archived (1) or not (0).
	Archived           uint8                      `json:"archived"`
	AddOnInfo          *string                    `json:"addOnInfo"`
	OrderProductID     uint64                     `json:"orderProductId"`
	CustomerID         uint64                     `json:"customerId"`
	FileLastModified   *string                    `json:"fileLastModified"`
	FileLastDownloaded *string                    `json:"fileLastDownloaded"`
	Files              []OrderProductFile         `json:"files"`
	Filters            []OrderProductFilter       `json:"filters"`
	History            []OrderProductHistoryEntry `json:"history"`
	Attributes         []OrderProductAttribute    `json:"attributes"`
	// Publisher is embedded publisher metadata, present when the API includes it inline
	// (in addition to, or instead of, sideloaded `included` publisher resources).
	Publisher *OrderProductPublisher `json:"publisher,omitempty"`
	// Product is embedded product catalog metadata (cover images, description).
	Product *OrderProductInfo `json:"product,omitempty"`
	// Order is embedded order summary metadata.
	Order *OrderProductOrder `json:"order,omitempty"`
}

// OrderProductPublisher is publisher metadata embedded directly on an ordered product's
// attributes.
type OrderProductPublisher struct {
	Name        string `json:"name"`
	PublisherID uint64 `json:"publisherId"`
	Slug        string `json:"slug"`
}

// OrderProductDescription is descriptive text for a product, embedded within
// OrderProductInfo.
type OrderProductDescription struct {
	Name             string  `json:"name"`
	PurchaseNote     *string `json:"purchaseNote,omitempty"`
	Slug             string  `json:"slug"`
	ShortDescription *string `json:"shortDescription,omitempty"`
}

// OrderProductInfo is product catalog metadata embedded directly on an ordered product's
// attributes, including relative paths to cover images.
//
// Image paths (Image, WebImage, Thumbnail, Thumbnail100) are relative to the DriveThruRPG
// images base URL (https://api.drivethrurpg.com/images/).
type OrderProductInfo struct {
	Image        *string                  `json:"image,omitempty"`
	WebImage     *string                  `json:"webImage,omitempty"`
	Thumbnail    *string                  `json:"thumbnail,omitempty"`
	Thumbnail100 *string                  `json:"thumbnail100,omitempty"`
	BundleID     uint64                   `json:"bundleId"`
	DateCreated  *string                  `json:"dateCreated,omitempty"`
	ProductID    uint64                   `json:"productId"`
	Description  *OrderProductDescription `json:"description,omitempty"`
	Filesize     *float64                 `json:"filesize,omitempty"`
}

// OrderProductOrder is order summary metadata embedded on an ordered product's attributes.
type OrderProductOrder struct {
	DateCreated *string `json:"dateCreated,omitempty"`
	OrderID     uint64  `json:"orderId"`
}

// OrderProductItem is a single item in an ordered products collection response, following
// the JSON:API resource object structure with ID, ResourceType, and Attributes.
//
// The live API does not embed publisher/product/order metadata directly on Attributes for
// this endpoint — it references them via Relationships, resolved against the response's
// top-level Included array. See OrderProductRelationships and IncludedItem.
type OrderProductItem struct {
	ID            string                     `json:"id"`
	ResourceType  string                     `json:"type"`
	Attributes    OrderProductAttributes     `json:"attributes"`
	Relationships *OrderProductRelationships `json:"relationships,omitempty"`
}

// OrderProductRelationships carries JSON:API relationship references on an
// OrderProductItem.
type OrderProductRelationships struct {
	Publisher *RelationshipRef `json:"publisher,omitempty"`
	Product   *RelationshipRef `json:"product,omitempty"`
	Order     *RelationshipRef `json:"order,omitempty"`
}

// RelationshipRef is a single JSON:API relationship reference, wrapping the referenced
// resource identifier.
type RelationshipRef struct {
	Data *RelationshipData `json:"data"`
}

// RelationshipData is the type/id pair identifying a JSON:API resource referenced by a
// relationship.
type RelationshipData struct {
	ResourceType string `json:"type"`
	ID           string `json:"id"`
}

// ── Sideloaded resources (included) ─────────────────────────────────────────────────────

// PublisherAttributes holds attributes for a publisher resource included alongside ordered
// product responses.
type PublisherAttributes struct {
	Name        string `json:"name"`
	PublisherID uint64 `json:"publisherId"`
	Slug        string `json:"slug"`
}

// PublisherItem is a publisher resource item included in ordered product responses when
// requested, following the JSON:API resource object structure.
type PublisherItem struct {
	ID           string              `json:"id"`
	ResourceType string              `json:"type"`
	Attributes   PublisherAttributes `json:"attributes"`
}

// IncludedItem is a single sideloaded resource entity from the included array of an
// ordered-products list response.
//
// The included array mixes multiple JSON:API resource types (Publisher, Product, Order)
// in a single flat list; ResourceType disambiguates which, and Attributes is kept as raw
// JSON since its shape depends on ResourceType. Decode it via AsPublisher or AsProduct.
type IncludedItem struct {
	ID           string          `json:"id"`
	ResourceType string          `json:"type"`
	Attributes   json.RawMessage `json:"attributes"`
}

// AsPublisher decodes Attributes as PublisherAttributes if ResourceType == "Publisher".
// Returns nil for any other resource type or if decoding fails.
func (i IncludedItem) AsPublisher() *PublisherAttributes {
	if i.ResourceType != "Publisher" {
		return nil
	}
	var attrs PublisherAttributes
	if err := json.Unmarshal(i.Attributes, &attrs); err != nil {
		return nil
	}
	return &attrs
}

// AsProduct decodes Attributes as OrderProductInfo if ResourceType == "Product". Returns
// nil for any other resource type or if decoding fails.
func (i IncludedItem) AsProduct() *OrderProductInfo {
	if i.ResourceType != "Product" {
		return nil
	}
	var info OrderProductInfo
	if err := json.Unmarshal(i.Attributes, &info); err != nil {
		return nil
	}
	return &info
}

// ── Response wrappers ───────────────────────────────────────────────────────────────────

// OrderProductListResponse is a paginated collection of ordered products, returned by
// GET /{api_version}/order_products.
type OrderProductListResponse struct {
	Links    PaginationLinks    `json:"links"`
	Meta     PaginationMeta     `json:"meta"`
	Data     []OrderProductItem `json:"data"`
	Included []IncludedItem     `json:"included,omitempty"`
}

// OrderProductItemResponse is a single ordered product resource response, returned by
// GET /{api_version}/order_products/{id}.
type OrderProductItemResponse struct {
	Data OrderProductItem `json:"data"`
	// Included holds Publisher/Product/Order resources sideloaded alongside the ordered
	// product, resolved by matching relationships.*.data.id against each entry's id
	// (mirrors OrderProductListResponse.Included).
	Included []IncludedItem `json:"included,omitempty"`
}

// ── Product Lists ────────────────────────────────────────────────────────────────────────

// ProductListAttributes holds attributes for a product list resource.
type ProductListAttributes struct {
	CustomerID    uint64 `json:"customerId"`
	Name          string `json:"name"`
	DateCreated   string `json:"dateCreated"`
	ProductListID uint64 `json:"productListId"`
	Slug          string `json:"slug"`
	ItemCount     uint64 `json:"itemCount"`
}

// ProductListItem is a single product list resource item, following the JSON:API resource
// object structure.
type ProductListItem struct {
	ID           string                `json:"id"`
	ResourceType string                `json:"type"`
	Attributes   ProductListAttributes `json:"attributes"`
}

// productListEnvelope unwraps the JSON:API {"data": {...}} envelope returned by
// POST /product_lists.
type productListEnvelope struct {
	Data ProductListItem `json:"data"`
}

// ProductListCollectionResponse is a paginated collection of product lists belonging to
// the authenticated customer, returned by GET /{api_version}/product_lists.
type ProductListCollectionResponse struct {
	Links PaginationLinks   `json:"links"`
	Meta  PaginationMeta    `json:"meta"`
	Data  []ProductListItem `json:"data"`
}

// ProductListItemsResponse is a paginated collection of items within a specific product
// list, returned by GET /{api_version}/product_list_items. Individual item schemas are not
// yet formally defined by the API contract, so items are represented as raw JSON until the
// schema matures.
type ProductListItemsResponse struct {
	Links PaginationLinks   `json:"links"`
	Meta  PaginationMeta    `json:"meta"`
	Data  []json.RawMessage `json:"data"`
}

// ProductListItemCreateRequest is the request body for adding a product to a product list,
// sent by POST /{api_version}/product_list_items.
type ProductListItemCreateRequest struct {
	ProductID     uint64 `json:"productId"`
	ProductListID uint64 `json:"productListId"`
}

// ProductListItemCreateResponse is the created product list item, returned by
// POST /{api_version}/product_list_items. The API wraps this resource in a JSON:API-style
// envelope on the wire ({"data": {"id": ..., "type": ..., "attributes": {"productId": ...,
// "productListId": ..., "productListItemId": ...}}}); UnmarshalJSON unwraps that envelope
// so callers work with a flat struct.
type ProductListItemCreateResponse struct {
	ProductID     uint64 `json:"productId"`
	ProductListID uint64 `json:"productListId"`
	// ProductListItemID is required to remove the item later via
	// DELETE /{api_version}/product_list_items/{id}.
	ProductListItemID uint64 `json:"productListItemId"`
}

// UnmarshalJSON unwraps the JSON:API {"data": {"attributes": {...}}} envelope into a flat
// struct.
func (r *ProductListItemCreateResponse) UnmarshalJSON(data []byte) error {
	var envelope struct {
		Data struct {
			Attributes struct {
				ProductID         uint64 `json:"productId"`
				ProductListID     uint64 `json:"productListId"`
				ProductListItemID uint64 `json:"productListItemId"`
			} `json:"attributes"`
		} `json:"data"`
	}
	if err := json.Unmarshal(data, &envelope); err != nil {
		return err
	}
	r.ProductID = envelope.Data.Attributes.ProductID
	r.ProductListID = envelope.Data.Attributes.ProductListID
	r.ProductListItemID = envelope.Data.Attributes.ProductListItemID
	return nil
}

// ── Query parameter structs ─────────────────────────────────────────────────────────────

// LibraryItemsParams holds query parameters for the GET /order_products (library items)
// endpoint. All fields are optional (nil pointers/zero values are omitted from the
// request). Use the zero value to start with no filters applied.
type LibraryItemsParams struct {
	Page     *uint32
	PageSize *uint32
	// GetChecksum, when true, includes checksum data for each product file
	// (getChecksum=1).
	GetChecksum *bool
	// GetFilters, when true, includes filter category data for each product
	// (getFilters=1).
	GetFilters *bool
	// Library, when true, restricts results to library (non-archived) products
	// (library=true).
	Library *bool
	// Archived, when true, includes archived products; when false, excludes them
	// (archived=1/0).
	Archived *bool
	// UpdatedDateAfter is an ISO 8601 date string. When set, returns only products
	// updated after this date (updatedDate[after]=...).
	UpdatedDateAfter *string
}

// PageParams holds query parameters for paginated collection endpoints such as
// /product_lists. All fields are optional. Use the zero value to retrieve the first page
// with the server's default page size.
type PageParams struct {
	Page     *uint32
	PageSize *uint32
}
