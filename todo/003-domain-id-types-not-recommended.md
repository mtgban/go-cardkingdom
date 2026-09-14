# `ProductID`/`SKU` domain types — not recommended

**Status:** closed without merging —
[PR #1 "Introduce ProductID and SKU domain types"](https://github.com/mtgban/go-cardkingdom/pull/1).
No review comments are recorded on the PR. This file exists so the idea
isn't re-proposed without accounting for the measured evidence already
gathered against it.

## What it proposed

`type ProductID int` / `type SKU string` in place of the plain `int`/
`string` fields, to prevent accidentally assigning an unrelated `int` or
`string` where a product ID or SKU was expected. Additive to JSON encoding;
literal assignments (`Product{ID: 5}`) still compile unchanged.

## Measured impact (from the PR body)

The PR was validated against `go-mtgban` at commit `815b076f` using a
temporary `replace` directive: restoring `go build ./...` and `go vet ./...`
needed **11 boundary conversions** across three files (`string(card.SKU)` /
`int(card.ID)` at each point the value flows into a plain `string`/`int`
parameter — `strconv.Itoa`, `map[string]string` values, a `unindexedTokenSheet(string)`
helper, `strings.TrimPrefix`/`strings.Split`). No test-literal changes were
needed. This is a real, measured number, not an estimate.

## Why this is not recommended without new evidence

The 11 conversions all sit at points where the value is *consumed*
(formatted into a string, passed to a stdlib function, used as a map key) —
not at the point where a real mixup could occur. The actual risk this kind
of typing usually guards against — assigning a `SKU` where an `ID` was
meant, or vice versa — happens in `go-mtgban` when both are funneled into
the *same* `string` field type on an unrelated struct:

```go
out := &mtgban.InventoryEntry{
    OriginalID: strconv.Itoa(card.ID), // int -> string
    InstanceID: card.SKU,              // string -> string
}
```

By the time `card.ID` reaches `OriginalID`, it has already been converted to
a plain `string` (`strconv.Itoa`), so a `ProductID`/`SKU` distinction in
*this* package would be erased before reaching the one spot in the
consumer where the two identifiers sit side by side and could plausibly be
swapped. The types would protect the short decode-time hop, not the
call site where a mistake is actually plausible.

## Recommendation

Do not re-attempt this without evidence of an actual mixup (a real bug
report, or a new call site where `ID`/`SKU` are used in a way the prior
analysis didn't cover). If the swap risk at `InventoryEntry`
construction is the real concern, that's a `go-mtgban`-side fix (e.g. typed
fields on `InventoryEntry`/`BuylistEntry` there), not something this
package's field types can enforce from a distance.
