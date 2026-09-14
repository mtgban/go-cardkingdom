# `float64` for currency — documented limitation, no action recommended

**Status:** documented in `SPECIFICATIONS.md` under Known Limitations; not
proposed as a change.

## The concern

All price fields (`PriceRetail`, `PriceBuy`, and the four `ConditionValue`
prices) are `float64`, decoded from vendor-sent JSON strings like
`"12345.67"`. Binary floating point cannot represent most decimal fractions
exactly, which is the usual argument for an integer-cents or
`decimal.Decimal`-style type for money.

## Why no change is recommended here

- The vendor's own JSON representation is a decimal string with two
  fraction digits; this package's `float64` already matches that precision
  for every value actually observed on the feed. There's no evidence of a
  precision bug in practice.
- The real consumer (`go-mtgban`) already does float arithmetic directly on
  these values (e.g. `buyPrice := card.PriceBuy * retailPrices[i] /
  retailPrices[0]`, per `SPECIFICATIONS.md`'s downstream-consumer section) —
  changing the type here would either break that call site outright or
  require the consumer to convert back to `float64` immediately, buying
  nothing.
- All eight price fields use the `json:",string"` tag specifically because
  the vendor sends them as strings; switching the Go type would touch every
  price field at once and is a breaking change with no concrete bug driving
  it.

## Recommendation

Leave as `float64`, documented as a known limitation (done, in
`SPECIFICATIONS.md`). Revisit only if a real precision-loss bug is observed
in practice — at that point the right fix is likely a caller-side
conversion at the application boundary (as the existing docs already
suggest), not a change to this package's wire types.
