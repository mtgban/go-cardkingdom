# SPECIFICATIONS.md

Technical and domain reference for `go-cardkingdom`. This documents the
actual wire format of Card Kingdom's feeds (verified directly against the
live endpoints, not inferred from the struct that decodes them) and the
semantics of every exported field. README.md is the user-facing quick start;
this is the source of truth for what the vendor actually sends.

## Endpoints

| Constant | URL | Content | Records observed (2026-09-14) |
|---|---|---|---|
| `PricelistURL` | `https://api.cardkingdom.com/api/v2/pricelist` | Singles (individual cards) | 151,082 |
| `SealedListURL` | `https://api.cardkingdom.com/api/sealed_pricelist` | Sealed product | 2,129 |

Both return the same envelope shape (`Response`: `meta` + `data`), decoded
by the same `Product` struct for every record — but the two endpoints send
**structurally disjoint** sets of fields per record. This was confirmed by
fetching every record from both live endpoints on 2026-09-14 and computing
the exhaustive union of JSON keys present, not a sample.

### Field presence by endpoint

| JSON key | Singles | Sealed | Decoded by `Product` today |
|---|---|---|---|
| `id` | always | always | yes (`ID`) |
| `sku` | always | **never** | yes (`SKU`) |
| `scryfall_id` | always (empty string on some rows) | **never** | yes (`ScryfallID`) |
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
`ConditionValues` — this is not an error, and `ScryfallID` being empty for
sealed product was already documented, but the other four fields going
empty for every sealed record was not, until this was written up.
`ShipsInternationally` (`ships_internationally`) is a native JSON boolean
(`true`/`false`), not a `json:",string"` field like `IsFoil` — the vendor is
not consistent about which representation it uses per field.

There is no field in either feed that names which endpoint a record came
from — callers distinguish singles from sealed only by which URL/function
they used to fetch it (`SinglesPricelist` vs `SealedPricelist`), not by
anything in the `Product` value itself.

## Types

### `Response`

The top-level envelope: `{"meta": Metadata, "data": []Product}`. Most
callers use `SinglesPricelist`, `SealedPricelist`, or `Pricelist` instead of
decoding this directly.

### `Metadata`

- `CreatedAt` (`created_at`): a timestamp with **no timezone indicator**,
  formatted `"2006-01-02 15:04:05"`. `CreatedAtTime()` parses it as UTC for
  a deterministic, comparable value — this is a compatibility choice, not a
  claim about the feed's actual source timezone, which is unknown. If the
  true source timezone is ever confirmed, parse `CreatedAt` directly with
  `time.ParseInLocation` instead of using `CreatedAtTime()`.
- `BaseURL` (`base_url`): observed as `"https://www.cardkingdom.com/"` on
  both endpoints. `Product.URL` is a path relative to this, not an absolute
  URL (see below) — despite the field's name.

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
- `ScryfallID` (`scryfall_id`, string, singles only, may be `""`): a
  Scryfall UUID for cross-referencing. Empty even on some singles rows.
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

Per-condition **retail** breakdown for a singles row — despite the field
names originally documented as buylist data, this was corrected in
[PR #4](https://github.com/mtgban/go-cardkingdom/pull/4) (merged) after
checking a real consumer: `go-mtgban` explicitly names its use of these
fields `retailPrices`/`qtys` and feeds them into inventory (for-sale)
records, never buylist records. There is no vendor-provided per-condition
*buylist* breakdown in either feed — a consumer wanting buylist prices by
condition has to derive them (see below).

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

## Versioning history

| Tag | Notes |
|---|---|
| `v0.0.1` | **Retracted** — contains a compilation error. |
| `v0.0.2` | — |
| `v0.0.3` | Hardening pass: default HTTP client timeout, explicit `http://`/`https://` scheme check (was a bare `"http"` prefix), test suite added, CI added, `%q`-quoted error strings, `v0.0.1` retraction added. |
| `v0.1.0` | **Breaking**: exported field initialisms cased per the Go style guide (`Sku`→`SKU`, `Nm`/`Ex`/`Vg`→`NM`/`EX`/`VG`). |
| (unreleased, on `master`) | `PR #4`: corrected `ConditionValue`/`URL` doc comments from buylist to retail semantics, documented the `CreatedAt` timezone assumption. `PR #3`: added explicit `PricelistFromURL`/`PricelistFromFile`/`DecodePricelist` alongside the prefix-sniffing `Pricelist` (see API surface, below). `PR #5`: added `Product.ShipsInternationally`; `Pricelist` now checks `ctx.Err()` before a local-file read. |

## API surface

- `SinglesPricelist(ctx, client)` / `SealedPricelist(ctx, client)`: the two
  main entry points, thin wrappers around `PricelistFromURL` against the
  fixed `PricelistURL`/`SealedListURL` constants. A `nil` client gets a
  fresh `go-cleanhttp` client with `DefaultTimeout` (30s).
- `Pricelist(ctx, client, link)`: dispatches on `link`'s prefix — a
  `http://`/`https://` URL goes to `PricelistFromURL`; anything else goes to
  `PricelistFromFile`, after a `ctx.Err()` check (added in `PR #5`) so an
  already-cancelled/expired context is honored before the read starts. Kept
  for compatibility; prefer the explicit functions below in new code.
- `PricelistFromURL(ctx, client, url)`: HTTP(S) only — rejects any other
  scheme. `ctx` governs the request and body read throughout. Non-200
  responses return an error with the status and up to 4KB of body; decode
  errors are wrapped with the source URL via `%w`.
- `PricelistFromFile(path)`: opens and reads a local file, closes it before
  returning. Takes no `ctx` and does not support cancellation once the read
  has started — this is a deliberate, documented design choice, not an
  oversight (contrast with `Pricelist`'s upfront check, above).
- `DecodePricelist(reader)`: the lowest-level primitive — decodes one JSON
  price list from an already-open `io.Reader` without closing it or
  observing any cancellation; both are entirely the caller's
  responsibility.
- `Metadata.CreatedAtTime()`: parses `CreatedAt` assuming UTC (see Known
  limitations, below).

## Downstream consumer: `go-mtgban`

`github.com/mtgban/go-mtgban`'s `cardkingdom/` package is a real, live
consumer (pinned at `v0.1.0` as of this writing) and the best available
ground truth for how these fields are actually used:

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
- **`PricelistFromFile` has no cancellation once reading starts**: `Pricelist`
  checks `ctx` before delegating to it, but the read itself can't be
  interrupted mid-flight — this is a deliberate scope decision (local file
  reads are fast and finite), not a gap.

## Testing

`testdata/pricelist.json` is a two-record fixture used by
`cardkingdom_test.go` for both the singles and sealed code paths (via a
fake `http.RoundTripper` that serves the same body regardless of which URL
is requested). Its second record ("Booster Box") now carries
`ships_internationally: true` and both feed-shape assertions
(`ShipsInternationally` false-on-singles / true-on-sealed) are checked, but
it still also carries singles-only fields (`sku`, `scryfall_id`,
`variation`, `is_foil`, `condition_values`, all as zero/empty values) that a
real sealed record never sends — see `todo/001` for making the fixture's
sealed record match the real disjoint shape exactly, rather than only
approximating it.
