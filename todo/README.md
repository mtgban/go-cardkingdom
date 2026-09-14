# todo

Scoped, evidence-based improvement proposals for `go-cardkingdom`. Each file
is self-contained enough to become its own PR, per `AGENTS.md`'s
one-focused-change convention. Check the status column before starting work
— two ideas below (`002`, `003`) were already proposed as PRs `#2` and `#1`
and closed without merging; re-proposing them needs new evidence, not a
restated version of the same argument.

| # | Item | Status |
|---|---|---|
| [001](001-sealed-product-test-fixture-accuracy.md) | Fix the sealed-product test fixture to fully match the real feed shape | Partially done — `ships_internationally` added; singles-only fields still leak into the fixture |
| [002](002-condition-type-revisit.md) | Typed per-condition accessor | **Closed without merging (PR #2)** — revisit only with new evidence |
| [003](003-domain-id-types-not-recommended.md) | `ProductID`/`SKU` domain types | **Closed without merging (PR #1)** — not recommended |
| [004](004-money-representation-float64.md) | `float64` for currency | Documented limitation, no action recommended |
| [005](005-changelog-and-github-releases.md) | No `CHANGELOG.md` / no GitHub Releases | Not started — low effort |

## Resolved (removed from this list)

- **Decode `ships_internationally` on the sealed feed** — done: `Product`
  now has `ShipsInternationally`.
- **`Pricelist`'s file-path branch ignoring `ctx`** — done: `Pricelist`
  checks `ctx.Err()` before delegating to `PricelistFromFile` (which itself
  still doesn't support cancellation mid-read — a deliberate, documented
  choice, not a gap; see `SPECIFICATIONS.md`).
- **Review/merge decision on the explicit source APIs** — done: `PR #3`
  merged, adding `PricelistFromURL`/`PricelistFromFile`/`DecodePricelist`.
  See `SPECIFICATIONS.md`'s API surface section.
