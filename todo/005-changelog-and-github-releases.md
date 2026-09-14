# No `CHANGELOG.md` / no GitHub Releases

**Status:** not started, low effort.

## Evidence

Four tags exist (`v0.0.1` retracted, `v0.0.2`, `v0.0.3`, `v0.1.0`), but
there is no `CHANGELOG.md` in the repo and no GitHub Releases published
(`gh release list` returns empty) for any of them. `v0.1.0` in particular
was a breaking change (exported-identifier renames) — a consumer bumping
past it with no changelog or release notes has to diff the tags themselves
to find out what broke.

## Recommendation

- Add a `CHANGELOG.md` (Keep a Changelog format is fine) covering the four
  existing tags retroactively, then update it alongside each future tag.
- Publish a GitHub Release per tag going forward (`gh release create`),
  at minimum for `v0.1.0` since it's the one breaking change so far — this
  is what shows up when a consumer runs `go list -m -u` or checks the repo
  before upgrading.

Low effort, no code risk; good first task for whoever picks up this list.
