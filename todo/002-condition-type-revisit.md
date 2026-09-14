# Typed per-condition accessor

**Status:** closed without merging —
[PR #2 "Add typed condition lookup returning price and quantity"](https://github.com/mtgban/go-cardkingdom/pull/2).
No review comments are recorded on the PR; this file documents what was
proposed and the evidence for the problem it targeted, so a future attempt
starts from that instead of re-deriving it.

## What it proposed

Per the PR description: `Condition` string-constants (`NM`, `EX`, `VG`, `G`),
a `Conditions()` function returning an independent array of them, and
`ConditionValue.Lookup(Condition) (price float64, qty int, ok bool)` —
additive, existing fields and JSON untouched. `ok` distinguishes an unknown
condition from a known one with zero stock.

## The problem it targeted (real, not hypothetical)

`go-mtgban`'s `cardkingdom/cardkingdom.go` (`Load`) builds two parallel
slices to read `ConditionValue`'s four grades:

```go
retailPrices := []float64{
    card.ConditionValues.NMPrice, card.ConditionValues.EXPrice,
    card.ConditionValues.VGPrice, card.ConditionValues.GPrice,
}
qtys := []int{
    card.ConditionValues.NMQty, card.ConditionValues.EXQty,
    card.ConditionValues.VGQty, card.ConditionValues.GQty,
}
for i, cond := range mtgban.DefaultGradeTags {
    // retailPrices[i] / qtys[i] assumed aligned with cond
    ...
}
```

Three sequences — two locally built slices and one externally-defined
ordering (`mtgban.DefaultGradeTags`, in a different repo) — are zipped by
positional index. Reordering any one silently misattributes a price/quantity
to the wrong condition; nothing in the type system catches it. This is
documented in more detail in `SPECIFICATIONS.md`'s downstream-consumer
section.

## Why this is a different case from `006`

Unlike the `ProductID`/`SKU` proposal (`006`, also closed), this one targets
a concretely identified failure mode at the actual point of use, not a
speculative one. A `Lookup`/`Price`/`Qty`-style accessor keyed by `Condition`
would let the consumer iterate `Conditions()` once instead of hand-aligning
three sequences — removing the index-alignment risk at its source rather
than adding a cast at a boundary that doesn't guard it (contrast with `006`).

## Recommendation

If revisited: keep the design additive (existing fields/JSON untouched, as
originally proposed), and consider driving the API from what the consumer
actually needs — e.g. an iterator or a small struct slice
(`[]struct{Condition; Price float64; Qty int}`) might read more naturally
at the call site than four `Lookup` calls. Worth a short design pass against
the real `go-mtgban` call site before re-proposing, given the prior attempt
didn't get review feedback to build on.
