# AGENTS.md

Operating instructions for any agent (human or AI) contributing to
`go-cardkingdom`. Read [SPECIFICATIONS.md](SPECIFICATIONS.md) first for what
the package does and the exact wire format it decodes; this file covers how
to work in the repo.

## What this repo is

A small, dependency-light Go client (`cardkingdom.go`, one file) for Card
Kingdom's two public price-list JSON feeds. It has one real dependency
(`go-cleanhttp`), a test suite (`cardkingdom_test.go` + `testdata/`), and no
CLI or server component. `github.com/mtgban/go-mtgban` is a real, live
consumer pinned to a specific tagged version — treat its `cardkingdom/`
package as a compatibility check, not a hypothetical.

## Before changing anything

- If the change touches how a JSON field is parsed or what fields exist,
  fetch one raw record from the actual endpoint first (`PricelistURL` /
  `SealedListURL`) and check every key it carries, not just the ones the
  `Product` struct already decodes. The two endpoints return **disjoint**
  field sets from a single shared struct (see SPECIFICATIONS.md) — this
  already produced one field (`ships_internationally`, since fixed as
  `Product.ShipsInternationally`) that shipped undecoded and unnoticed for
  a release, because nobody had checked a sealed record's full key list
  against the struct. Go's JSON decoder drops unknown keys silently;
  absence of an error is not evidence the struct is complete.
- If the change renames, removes, or retypes an exported identifier, check
  `go-mtgban`'s `cardkingdom/` package (or ask for its current state) before
  assuming the blast radius is zero. Two prior proposals in this repo's PR
  history (`#1`, `#2` — see `todo/003` and `todo/002`) were evaluated this
  way; one was measured against a real `go build` of the consumer with a
  `replace` directive before a decision was made.
- Check `todo/README.md` before starting new work. Several ideas that sound
  obvious in isolation (typed IDs, a `Condition` enum) have already been
  tried here and closed without merging — the file explains why, so you
  don't re-derive the same tradeoff from scratch.

## Build / test / lint gate

Run all of this before committing; it is exactly what CI
(`.github/workflows/ci.yml`) runs, in the same order:

```bash
gofmt -l .                                                            # must be empty
go vet ./...
go run github.com/mgechev/revive@v1.13.0 -set_exit_status -config .revive.toml ./...
go run honnef.co/go/tools/cmd/staticcheck@2025.1.1 ./...
go build ./...
go test -race -coverprofile=coverage.out ./...
```

Staticcheck's version is pinned deliberately — a newer release finding new
issues should not fail a build that changed nothing. Bump the pin in its own
commit if you need the newer checks.

## Style

- Every exported symbol needs a godoc comment starting with its own name
  (enforced by `revive`'s `exported` rule).
- Initialisms are cased per the [Google Go Style Guide](https://google.github.io/styleguide/go/decisions#initialisms):
  `SKU`, `URL`, `ID`, `NM`/`EX`/`VG` — never `Sku`, `Url`, `NmPrice`. This was
  a breaking rename in `v0.1.0`; don't reintroduce mixed-case initialisms.
- Error strings are lowercase, no trailing punctuation (unless they start
  with an acronym like `GET`), wrapped with `%w`, and quote untrusted/opaque
  values with `%q` rather than `%s` so an empty value stays visible in the
  message.
- `.revive.toml`'s `unhandled-error` exclusion list is a specific, commented
  allowlist of calls whose error return is genuinely not actionable (e.g. a
  test handler's `ResponseWriter.Write`). Add to it only with the same kind
  of justification already in the file — it is not a general escape hatch.

## Versioning

Pre-1.0 (`go.mod` declares no `v1` yet), so the practical convention already
in use:

- A breaking change (renamed/removed exported identifier, changed field
  semantics) bumps the **minor** version (`v0.1.0` did this for the
  initialism renames). A purely additive or internal change can go out
  un-tagged or as a patch-style bump.
- A published version that turns out to be broken is marked with `retract`
  in `go.mod`, with a one-line comment explaining why (see `v0.0.1`), and
  the retraction only takes effect once a *newer* tag carrying it is pushed.
- Tag after the commit is on `master` and pushed, never before.

## Git / PR workflow

The existing PR history (`#1`–`#4` on this repo) already establishes the
house style — follow it:

- One focused change per PR, branched from `master`, not from another PR's
  branch — even when the work depends on something an open PR added. Wait
  for the dependency to merge, or accept the diff will show its changes
  until it does.
- State what was validated in the PR description (the gate above) so a
  reviewer doesn't have to re-derive it.
- Small is normal here: several past PRs are under 150 lines including
  tests. Don't bundle unrelated cleanups into a feature change.
