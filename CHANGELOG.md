# Changelog

All notable changes to `go-cardkingdom`. The format follows
[Keep a Changelog](https://keepachangelog.com/en/1.1.0/). Before `v1`, a
breaking change bumps the minor version.

## [Unreleased]

### Added

- `PricelistFromURL`, `PricelistFromFile` and `DecodePricelist`: explicit
  HTTP(S), file and `io.Reader` sources alongside `Pricelist`.
- `Product.ShipsInternationally`, decoded from the sealed feed.

### Changed

- Requires Go 1.26 (`go 1.26.0`); a module that upgrades gets its own `go`
  line raised to match.
- `Pricelist` returns the context's error before reading a local file when
  the context is already cancelled or expired.
- Error messages quote the source link with `%q`.
- `ConditionValue` and `Product.URL` documentation: condition prices and
  quantities are retail stock, and `URL` is a path relative to
  `Metadata.BaseURL`.

## [0.1.0] - 2026-08-13

### Changed

- **Breaking:** exported field initialisms follow the Go style guide:
  `Product.Sku` → `SKU`; `ConditionValue.NmPrice`/`NmQty`,
  `ExPrice`/`ExQty`, `VgPrice`/`VgQty` → `NMPrice`/`NMQty`,
  `EXPrice`/`EXQty`, `VGPrice`/`VGQty`. JSON field names are unchanged.

## [0.0.3] - 2026-07-06

### Added

- `DefaultTimeout` (30s), applied to the client used when `nil` is passed.
- Godoc for every exported identifier.

### Changed

- `Pricelist` treats `link` as a URL only when it starts with `http://` or
  `https://`; anything else is a file path.

### Fixed

- README import path.

### Removed

- `v0.0.1` is retracted.

## [0.0.2] - 2025-09-21

### Added

- `Metadata` and `Metadata.CreatedAtTime`.
- `go.mod` and `go.sum`.

### Changed

- **Breaking:** `Pricelist` also returns the price list's `Metadata`.

## [0.0.1] - 2025-09-21 [RETRACTED]

Contains a compilation error. Do not use.

[Unreleased]: https://github.com/mtgban/go-cardkingdom/compare/v0.1.0...HEAD
[0.1.0]: https://github.com/mtgban/go-cardkingdom/compare/v0.0.3...v0.1.0
[0.0.3]: https://github.com/mtgban/go-cardkingdom/compare/v0.0.2...v0.0.3
[0.0.2]: https://github.com/mtgban/go-cardkingdom/compare/v0.0.1...v0.0.2
[0.0.1]: https://github.com/mtgban/go-cardkingdom/releases/tag/v0.0.1
