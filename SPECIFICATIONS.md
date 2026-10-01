# SPECIFICATIONS.md

Technical and domain reference for `go-cardkingdom`. This documents the
actual wire format of Card Kingdom's feeds (verified directly against the
live endpoints, not inferred from the struct that decodes them) and the
semantics of every exported field. README.md is the user-facing quick start;
this is the source of truth for what the vendor actually sends.

## Endpoints

| Constant | URL | Content | Records observed (2026-10-01) |
|---|---|---|---|
| `PricelistURL` | `https://api.cardkingdom.com/api/v2/pricelist` | Singles (individual cards) | 152,175 |
| `SealedListURL` | `https://api.cardkingdom.com/api/sealed_pricelist` | Sealed product | 2,128 |

Both return the same envelope shape (`meta` + `data`), decoded
by the same `Product` struct for every record — but the two endpoints send
**structurally disjoint** sets of fields per record. This was confirmed by
fetching every record from both live endpoints on 2026-10-01 and computing
the exhaustive union of JSON keys present, at every level, not a sample.
`Product` decodes every key either feed sends.

### Field presence by endpoint

| JSON key | Singles | Sealed | Decoded by `Product` |
|---|---|---|---|
| `id` | always | always | yes (`ID`) |
| `sku` | always | **never** | yes (`SKU`) |
| `scryfall_id` | always (`null` on some rows) | **never** | yes (`ScryfallID`) |
| `url` | always | always | yes (`URL`) |
| `name` | always | always | yes (`Name`) |
| `variation` | always (often `""`) | **never** | yes (`Variation`) |
| `edition` | always | always | yes (`Edition`) |
| `is_foil` | always | **never** | yes (`IsFoil`) |
| `price_retail` | always | always | yes (`PriceRetail`) |
| `qty_retail` | always | always | yes (`QtyRetail`) |
| `price_buy` | always | always | yes (`PriceBuy`) |
| `qty_buying` | always | always | yes (`QtyBuying`) |
| `condition_values` | always | **never** | yes (`ConditionValues`) |
| `ships_internationally` | **never** | always | yes (`ShipsInternationally`) |

Practical consequence: decoding a sealed record into `Product` silently
zero-values `SKU`, `ScryfallID`, `Variation`, `IsFoil`, and
`ConditionValues` on every sealed record — this is not an error.
`ShipsInternationally` (`ships_internationally`) is a native JSON boolean
(`true`/`false`), not a `json:",string"` field like `IsFoil` — the vendor is
not consistent about which representation it uses per field.

There is no field in either feed that names which endpoint a record came
from — callers distinguish singles from sealed only by which URL/function
they used to fetch it (`SinglesPricelist` vs `SealedPricelist`), not by
anything in the `Product` value itself.

## Types

### `PricelistFile`

The envelope `{"meta": Metadata, "data": []Product}`, decoded as sent and
returned by `SinglesPricelistFile`, `SealedPricelistFile` and
`LoadPricelistFile`.

### `Metadata`

- `CreatedAt` (`created_at`): a timestamp with **no timezone indicator**,
  formatted `"2006-01-02 15:04:05"`, kept as the string sent.
  `CreatedAtTime()` parses it as UTC for a deterministic, comparable value —
  not a claim about the feed's actual source timezone, which is unknown. If
  the true source timezone is ever confirmed, parse `CreatedAt` with
  `time.ParseInLocation` instead.
- `BaseURL` (`base_url`): observed as `"https://www.cardkingdom.com/"` on
  both endpoints. `Product.URL` is a path relative to this, not an absolute
  URL (see below).

### `Product`

One purchasable item — a card (singles feed) or a sealed product (sealed
feed); see the field-presence table above for which fields are meaningful
for which.

- `ID` (`id`, int): Card Kingdom's internal product identifier. Stable
  across both feeds' numbering (a sealed product's `id` and a card's `id`
  share one number space, per the `sealed.go` consumer pattern below).
- `SKU` (`sku`, string, singles only): the stock-keeping unit. Encodes the
  set code as a prefix (e.g. `"4ED-117"`), with `F`/`E`/`FE` prefixes for
  some foil/etched variants — see the downstream-consumer section, this
  encoding is load-bearing for at least one real consumer.
- `ScryfallID` (`scryfall_id`, string, singles only): a Scryfall UUID for
  cross-referencing. The feed sends JSON `null` rather than `""` when there
  is none, which decodes to `""` — 281 singles on 2026-10-01, mostly the
  `"Promo Pack"` edition (222) and Card Kingdom's own tokens (36, `CKT-`
  SKUs).
- `URL` (`url`, string): a **path**, not an absolute URL — e.g.
  `"mtg/4th-edition/abomination"` or `"mtg-sealed/mercadian-masques-booster-box"`.
  Join it with `Metadata.BaseURL` to get a usable link; do not use it
  standalone as a URL.
- `Name`, `Edition` (`name`, `edition`, string): always present, both feeds.
- `Variation` (`variation`, string, singles only): printing detail (e.g.
  `"Borderless"`, `"Extended Art"`), often `""`.
- `IsFoil` (`is_foil`, bool via `json:",string"`, singles only): the vendor
  sends this as the literal JSON strings `"true"`/`"false"`, not a native
  boolean — unlike `ships_internationally` on the sealed feed, which is a
  native boolean. The two feeds are not consistent with each other on this.
- `PriceRetail` / `QtyRetail` (`price_retail` json-string float64 /
  `qty_retail` int): current sale price and in-stock quantity, both feeds.
- `PriceBuy` / `QtyBuying` (`price_buy` json-string float64 / `qty_buying`
  int): current buylist (Card Kingdom buying from you) price and desired
  quantity, both feeds. `0` means not currently buying.
- `ConditionValues` (`condition_values`, singles only): see below.
- `ShipsInternationally` (`ships_internationally`, bool, sealed only): a
  native JSON boolean (unlike `IsFoil`, above) reporting whether the item
  ships outside the United States.

### `ConditionValue`

Per-condition **retail** breakdown for a singles row, not buylist data:
`go-mtgban` names its use of these fields `retailPrices`/`qtys` and feeds
them into inventory (for-sale) records, never buylist records. There is no
vendor-provided per-condition *buylist* breakdown in either feed — a
consumer wanting buylist prices by condition has to derive them (see
below).

- `NMPrice`/`NMQty`, `EXPrice`/`EXQty`, `VGPrice`/`VGQty`, `GPrice`/`GQty`:
  retail price and in-stock quantity for Near Mint, Excellent, Very Good,
  and Good (heavily played) condition respectively. All eight are
  `json:",string"`-decoded except the four `*Qty` ints, which are native
  JSON numbers.
- A newly-listed card can have `NMPrice == 0` before this breakdown is
  populated; fall back to `Product.PriceRetail` in that case (this is what
  the real consumer does — see below).

## Numeric encoding

Most price and boolean fields arrive from the vendor as **JSON strings**
(`"price_retail": "0.35"`, `"is_foil": "false"`), decoded via the
`json:",string"` struct tag into `float64`/`bool`. Quantities
(`qty_retail`, `qty_buying`, the four `*Qty` condition fields) and `id` are
native JSON numbers. `ships_internationally` (`ShipsInternationally`,
sealed-only) is also a native JSON boolean — the vendor is not consistent
about which representation it uses per field.

## API surface

Each entry point comes in two forms, as in `go-cardmarket`: one returns the
products, its `…File` counterpart the whole `PricelistFile`.

- `SinglesPricelist(ctx, client)` / `SealedPricelist(ctx, client)` and
  `SinglesPricelistFile` / `SealedPricelistFile`: fetch the fixed
  `PricelistURL` / `SealedListURL`. A `nil` client gets a fresh
  `go-cleanhttp` client with `DefaultTimeout` (30s).
- `Pricelist(ctx, client, link)` / `LoadPricelistFile(ctx, client, link)`:
  a `http://`/`https://` link is fetched, with `ctx` governing the request
  and body read; anything else is read as a local file, after a `ctx.Err()`
  check. Non-200 responses return an error with the status and up to 4KB of
  body; decode errors are wrapped with the source link via `%w`.

## Downstream consumer: `go-mtgban`

`github.com/mtgban/go-mtgban`'s `cardkingdom/` package is a real, live
consumer and the best available ground truth for how these fields are
actually used:

- Builds retail inventory entries from `ConditionValues`, indexed
  positionally against `mtgban.DefaultGradeTags` — `NMPrice`/`EXPrice`/
  `VGPrice`/`GPrice` and their `*Qty` counterparts are read into two
  parallel slices, zipped by index with that external ordering. This is the
  scenario a `Condition`-keyed accessor (`todo/002`) would remove the
  indexing risk from, if revisited.
- Synthesizes a **per-condition buylist price** that the feed does not
  provide, by scaling `PriceBuy` by the ratio of each condition's retail
  price to NM retail price: `buyPrice = PriceBuy * retailPrices[i] /
  retailPrices[0]`. `QtyBuying` (not condition-specific) is applied
  uniformly to every synthesized grade.
- Parses `SKU` as `"<setCode>-<number>"`, stripping a leading `F`/`E`/`FE`
  for some foil/etched variants before further lookup — `SKU`'s structure
  is load-bearing, not just an opaque identifier, for at least this
  consumer.
- Joins `Product.URL` onto a hardcoded `"https://www.cardkingdom.com/"`
  base rather than `Metadata.BaseURL` — consistent with `URL` being a
  relative path, per the table above.
- Uses `ID` and `SKU` as opaque `string`/`int` identifiers
  (`OriginalID`/`InstanceID`), and separately matches sealed-product `ID`
  values against IDs recovered from a different source
  (`cardkingdom/sealed.go`), confirming `ID` is one shared number space
  across both feeds.

## Known limitations

- **Floating-point currency**: prices are `float64`. This mirrors the
  vendor's own precision (two decimal places as sent) but is not exact
  decimal arithmetic; a caller doing exact monetary math should convert at
  its own boundary with an explicit rounding policy. See `todo/004`.
- **`CreatedAt` has no timezone**: `CreatedAtTime()` assumes UTC for
  determinism, not because the feed says so.
- **One struct, two schemas**: `Product` decodes both feeds by field
  presence/absence rather than by an explicit discriminator (see the
  field-presence table above).
- **Local files have no cancellation once reading starts**: `Pricelist` and
  `LoadPricelistFile` check `ctx` before opening a file, but the read itself
  can't be interrupted mid-flight — local file reads are fast and finite.

## Testing

The fixtures are real records copied unchanged from both live feeds
(2026-10-01), so they carry exactly the keys each endpoint sends:

- `testdata/pricelist.json` (singles): a foil card with stock in every
  condition, a non-foil `Variation`, and a `Promo Pack` card whose
  `scryfall_id` is `null`.
- `testdata/sealed_pricelist.json` (sealed): one product that ships
  internationally and one that does not; none of the singles-only keys.

`cardkingdom_test.go` compares every decoded `Product` field for field. The
tests for the fixed endpoints, plain and `…File`, serve each endpoint its
own fixture through a fake `http.RoundTripper`.
