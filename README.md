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

	"github.com/mtgban/go-cardkingdom"
)

func main() {
	singles, err := cardkingdom.SinglesPricelist(context.Background(), nil)
	if err != nil {
		panic(err)
	}
	fmt.Printf("found %d products\n", len(singles))
}
```

## Custom HTTP client (timeouts, proxies)

You can inject any `*http.Client`:

```go
client := &http.Client{Timeout: 10 * time.Second}
items, err := cardkingdom.SinglesPricelist(ctx, client)
```

Passing `nil` creates a fresh client from `go-cleanhttp` with a 30-second
timeout. To reuse connections across calls, supply a shared `*http.Client`
with your chosen timeout.

## Reading from a local file

For testing or offline runs, you can point `Pricelist` at a local JSON file path:

```go
prods, err := cardkingdom.Pricelist(ctx, nil, "pricelist.json")
```

If the `link` argument doesn’t start with `http://` or `https://`, the function opens it as a file.

## When the list was built

`SinglesPricelistFile`, `SealedPricelistFile` and `LoadPricelistFile` return
the whole `PricelistFile` as Card Kingdom sends it: `Meta` (when the list was
built, and the base URL product links are relative to) and `Data` (the
products). Use it to refuse a list that has stopped updating; how old is too
old is up to you.

```go
file, err := cardkingdom.SinglesPricelistFile(ctx, nil)
if err != nil {
	return err
}
created, err := file.Meta.CreatedAtTime()
if err != nil {
	return err
}
if time.Since(created) > 48*time.Hour {
	return fmt.Errorf("card kingdom list is from %s", created)
}
```

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
    ShipsInternationally bool
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

Every function accepts a `context.Context`. Pass deadlines or cancel to abort in-flight HTTP requests; a local file is only checked against the context before it is opened:

```go
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()
prods, err := cardkingdom.SinglesPricelist(ctx, nil)
```

## Error handling

- Non-200 responses include a short body preview to help troubleshooting.
- JSON decoding errors are wrapped with the endpoint/file path that failed.

## Data semantics

`Product.URL` is relative to `Metadata.BaseURL` (for example,
`mtg/4th-edition/abomination`). Condition prices and quantities describe retail
stock; the root `PriceBuy` and `QtyBuying` fields describe the buylist.

Singles and sealed product are structurally disjoint records decoded into
the same `Product` struct: `SKU`, `ScryfallID`, `Variation`, `IsFoil`, and
`ConditionValues` are singles-only (always zero/empty on sealed product),
while `ShipsInternationally` is sealed-only (always `false` on singles).
Nothing in the feed itself flags which shape a record is — that's implied
by which endpoint you fetched it from.

`CreatedAt` contains no timezone. `CreatedAtTime()` interprets it as UTC;
this does not establish the feed's source timezone. If you know the source
location, use `time.ParseInLocation` on `CreatedAt` instead.

Prices use `float64`. Binary floating-point values
are approximate; callers needing exact monetary arithmetic should convert at
their application boundary with an explicit rounding policy.

## License

MIT
