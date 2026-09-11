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
    ck "github.com/mtgban/go-cardkingdom"
)

func main() {
    ctx := context.Background()

    singles, err := ck.SinglesPricelist(ctx, nil)
    if err != nil {
        panic(err)
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

Passing `nil` creates a fresh client from `go-cleanhttp` with a 30-second
timeout. To reuse connections across calls, supply a shared `*http.Client`
with your chosen timeout.

## Reading from a local file

For testing or offline runs, you can point `Pricelist` at a local JSON file path:

```go
prods, metadata, err := cardkingdom.Pricelist(ctx, nil, "pricelist.json")
```

If the `link` argument doesn’t start with `http://` or `https://`, the function opens it as a file.

It's possible to parse the metadata CreatedAt field as a time.Time with the
`CreatedAtTime()` method.

## Data model

```go
type Product struct {
    ID          int
    SKU         string
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
    NMPrice float64
    NMQty   int
    EXPrice float64
    EXQty   int
    VGPrice float64
    VGQty   int
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


## Data semantics

`Product.URL` is relative to `Metadata.BaseURL` (for example,
`mtg/4th-edition/abomination`). Condition prices and quantities describe retail
stock; the root `PriceBuy` and `QtyBuying` fields describe the buylist.

`CreatedAt` contains no timezone. `CreatedAtTime()` interprets it as UTC for
compatibility; this does not establish the feed's source timezone. If you know
the source location, use `time.ParseInLocation` on `CreatedAt` instead.

Prices use `float64` to mirror the existing API. Binary floating-point values
are approximate; callers needing exact monetary arithmetic should convert at
their application boundary with an explicit rounding policy.
