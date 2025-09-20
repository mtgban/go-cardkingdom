# go-cardkingdom

A tiny Go client for fetching Card Kingdom’s public price lists.

- **Singles:** `https://api.cardkingdom.com/api/v2/pricelist`
- **Sealed:** `https://api.cardkingdom.com/api/sealed_pricelist`

This package provides simple, ergonomic helpers to download and decode those lists into Go structs.

## Install

```bash
go get github.com/mtgban/go-cardkingdom
```

## Quick start

```go
package main

import (
    "context"
    "fmt"
    "net/http"
    ck "github.com/your/module/path/cardkingdom"
)

func main() {
    ctx := context.Background()

    singles, err := ck.SinglesPricelist(ctx, http.DefaultClient)
    if err != nil {
        // Handle err
    }
    fmt.Printf("found %d products\n", len(singles))
}
```

## Custom HTTP client (timeouts, proxies)

You can inject any `*http.Client`:

```go
client := &http.Client{ Timeout: 10 * time.Second }
items, err := cardkingdom.SinglesPricelist(ctx, client)
```

Passing `nil` will default to a clean, connection-reusing client from `go-cleanhttp`.

## Reading from a local file

For testing or offline runs, you can point `Pricelist` at a local JSON file path:

```go
prods, metadata, err := cardkingdom.Pricelist(ctx, nil, "pricelist.json")
```

If the `link` argument doesn’t start with `http`, the function opens it as a file.

## Data model

```go
type Product struct {
    ID          int
    Sku         string
    ScryfallID  string
    URL         string
    Name        string
    Variation   string
    Edition     string
    IsFoil      bool
    PriceRetail float64
    QtyRetail   int
    PriceBuy    float64
    QtyBuying   int
    ConditionValues ConditionValue
}

type ConditionValue struct {
    NmPrice float64
    NmQty   int
    ExPrice float64
    ExQty   int
    VgPrice float64
    VgQty   int
    GPrice  float64
    GQty    int
}
```

**Note on types:** the Card Kingdom API returns many numeric fields as **strings**.  
This package uses the `json:",string"` tag to parse those into `float64`/`bool` automatically.

## Context & cancellation

All functions accept a `context.Context`. Pass deadlines or cancel to abort in-flight HTTP requests:

```go
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()
prods, err := cardkingdom.SinglesPricelist(ctx, nil)
```

## Error handling

- Non-200 responses include a short body preview to help troubleshooting.
- JSON decoding errors are wrapped with the endpoint/file path that failed.

## License

MIT

