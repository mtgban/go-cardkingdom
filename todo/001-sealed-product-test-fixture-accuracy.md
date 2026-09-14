# Fix the sealed-product test fixture to match the real feed shape

**Status:** partially done — `ships_internationally` was added to the
fixture and is now asserted (see the commit that added
`Product.ShipsInternationally`). The rest of this item is still open.

## Evidence

`testdata/pricelist.json`'s second record ("Booster Box") is meant to model
a sealed product and now correctly carries `ships_internationally: true`
(asserted by `assertFixtureProducts`), but it still also carries `sku`,
`scryfall_id`, `variation`, `is_foil`, and `condition_values` (all as
zero/empty values). A real sealed record never carries any of those five
keys — see `SPECIFICATIONS.md`'s field-presence table, verified against the
live feed.

`TestSinglesAndSealedURLs` in `cardkingdom_test.go` serves this same fixture
body for both the singles and sealed code paths via a fake
`http.RoundTripper`, so the sealed path is still never tested against data
that has the real disjoint shape — only against data with the right values
for the fields that do apply.

## Recommendation

- Add a second, sealed-shaped fixture (either a new file, e.g.
  `testdata/sealed_pricelist.json`, or a second record in the existing
  fixture used only by the sealed-path test) containing exactly the keys
  the live sealed feed sends: `id`, `url`, `name`, `edition`, `price_retail`,
  `qty_retail`, `price_buy`, `qty_buying`, `ships_internationally` — and
  none of the singles-only keys.
- Update `TestSinglesAndSealedURLs` (or add a dedicated test) to serve the
  sealed fixture only for the sealed request, so the test actually exercises
  decoding a record that lacks the singles-only fields, rather than one
  that merely has zero values for them.
