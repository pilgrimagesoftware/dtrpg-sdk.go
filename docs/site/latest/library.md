---
title: Library client
version: latest
slug: library
---

# Library client

`library.Client` is an authenticated HTTP client for DriveThruRPG library endpoints.
Construct it via `sdk.LibraryClient()` (see [Authentication]({{ '/latest/auth/' | relative_url }})),
which validates configuration and session state before returning the client.

## Listing ordered products

```go
client, err := sdk.LibraryClient()
if err != nil {
	log.Fatal(err)
}

products, err := client.ListOrderProducts(ctx, dtrpg.LibraryItemsParams{})
if err != nil {
	log.Fatal(err)
}
for _, item := range products.Data {
	fmt.Println(item.ID)
}
```

`LibraryItemsParams` controls pagination (`Page`, `PageSize`) and filtering
(`GetChecksum`, `GetFilters`, `Library`, `Archived`, `UpdatedDateAfter`). All fields are
pointers; nil fields omit that query parameter. Maps to
`GET /{api_version}/order_products`.

## Fetching a single ordered product

```go
item, err := client.GetOrderProduct(ctx, orderProductID)
```

Maps to `GET /{api_version}/order_products/{order_product_id}`.

## Preparing a download

```go
result, err := client.PrepareDownload(ctx, orderProductID, fileIndex)
```

`fileIndex` identifies which file within the ordered product to prepare — it matches
`OrderProductFile.Index` and is required. The response is returned as a
`map[string]any` since the API has not yet published a formal schema for this endpoint.
Maps to `GET /{api_version}/order_products/{order_product_id}/prepare?index={index}`.

## Listing product lists

```go
lists, err := client.ListProductLists(ctx, dtrpg.PageParams{})
```

`PageParams` controls `Page` and `PageSize`. Maps to `GET /{api_version}/product_lists`.
