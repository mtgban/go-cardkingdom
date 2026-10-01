# todo

Scoped, evidence-based improvement proposals for `go-cardkingdom`. Each file
is self-contained enough to become its own PR, per `AGENTS.md`'s
one-focused-change convention. Check the status column before starting work
— two ideas below (`002`, `003`) were already proposed as PRs `#2` and `#1`
and closed without merging; re-proposing them needs new evidence, not a
restated version of the same argument.

| # | Item | Status |
|---|---|---|
| [002](002-condition-type-revisit.md) | Typed per-condition accessor | **Closed without merging (PR #2)** — revisit only with new evidence |
| [003](003-domain-id-types-not-recommended.md) | `ProductID`/`SKU` domain types | **Closed without merging (PR #1)** — not recommended |
| [004](004-money-representation-float64.md) | `float64` for currency | Documented limitation, no action recommended |

## Resolved (removed from this list)

- **Sealed test fixture matching the real feed shape** (`001`) — done:
  both fixtures are real feed records, and the sealed path is tested
  against its own.
- **Decode `ships_internationally` on the sealed feed** — done: `Product`
  now has `ShipsInternationally`.
- **`Pricelist`'s file-path branch ignoring `ctx`** — done: `Pricelist`
  checks `ctx.Err()` before opening a local file (the read itself still
  can't be cancelled — a deliberate, documented choice; see
  `SPECIFICATIONS.md`).
