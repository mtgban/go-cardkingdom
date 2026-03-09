// Package cardkingdom provides a client for the Card Kingdom public price list API.
//
// It supports fetching both singles and sealed-product price lists, decoding
// them into typed Go structs. Requests are made over HTTP using a standard
// [net/http.Client], and all functions accept a [context.Context] for
// cancellation and deadline control.
//
// The two main entry points are [SinglesPricelist] and [SealedPricelist].
// For lower-level access — including reading from a local file — use [Pricelist]
// directly.
package cardkingdom

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/hashicorp/go-cleanhttp"
)

const (
	// PricelistURL is the Card Kingdom API endpoint for singles prices.
	PricelistURL = "https://api.cardkingdom.com/api/v2/pricelist"

	// SealedListURL is the Card Kingdom API endpoint for sealed product prices.
	SealedListURL = "https://api.cardkingdom.com/api/sealed_pricelist"

	// UserAgent is the HTTP User-Agent header sent with every request.
	UserAgent = "go-cardkingdom"
)

// Response is the top-level envelope returned by the Card Kingdom API.
// Most callers should use [SinglesPricelist], [SealedPricelist], or [Pricelist]
// instead of working with Response directly.
type Response struct {
	Meta Metadata  `json:"meta"`
	Data []Product `json:"data"`
}

// Metadata contains header information returned alongside the price list.
type Metadata struct {
	// CreatedAt is the timestamp at which the price list was generated,
	// formatted as "2006-01-02 15:04:05". Use [Metadata.CreatedAtTime] to
	// parse it into a [time.Time].
	CreatedAt string `json:"created_at"`

	// BaseURL is the base URL used to construct product page links.
	BaseURL string `json:"base_url"`
}

// CreatedAtTime parses the CreatedAt field into a [time.Time].
// The expected layout is "2006-01-02 15:04:05". An error is returned if the
// value does not match that format.
func (m Metadata) CreatedAtTime() (time.Time, error) {
	return time.Parse("2006-01-02 15:04:05", m.CreatedAt)
}

// Product represents a single purchasable item in the Card Kingdom catalog,
// either a trading card or a sealed product.
type Product struct {
	// ID is the Card Kingdom internal product identifier.
	ID int `json:"id"`

	// Sku is the stock-keeping unit code for this listing.
	Sku string `json:"sku"`

	// ScryfallID is the Scryfall UUID for this card, suitable for
	// cross-referencing with the Scryfall API. Empty for sealed products.
	ScryfallID string `json:"scryfall_id"`

	// URL is the direct link to this product's page on cardkingdom.com.
	URL string `json:"url"`

	// Name is the card or product name.
	Name string `json:"name"`

	// Variation describes the specific printing (e.g. "Borderless", "Extended Art").
	Variation string `json:"variation"`

	// Edition is the set name or product line this item belongs to.
	Edition string `json:"edition"`

	// IsFoil reports whether this listing is for a foil printing.
	IsFoil bool `json:"is_foil,string"`

	// PriceRetail is the current price Card Kingdom charges buyers, in USD.
	PriceRetail float64 `json:"price_retail,string"`

	// QtyRetail is the number of units currently in stock for sale.
	QtyRetail int `json:"qty_retail"`

	// PriceBuy is the current buylist price Card Kingdom will pay, in USD.
	// A value of 0 means Card Kingdom is not purchasing this item.
	PriceBuy float64 `json:"price_buy,string"`

	// QtyBuying is the number of additional copies Card Kingdom is willing
	// to purchase. A value of 0 means they are not currently buying.
	QtyBuying int `json:"qty_buying"`

	// ConditionValues holds per-condition buy prices and quantities.
	ConditionValues ConditionValue `json:"condition_values"`
}

// ConditionValue holds buylist prices and purchase quantities broken down by
// card condition. Prices are in USD.
type ConditionValue struct {
	// NmPrice is the buy price for Near Mint copies.
	NmPrice float64 `json:"nm_price,string"`
	// NmQty is the number of Near Mint copies Card Kingdom will purchase.
	NmQty int `json:"nm_qty"`

	// ExPrice is the buy price for Excellent copies.
	ExPrice float64 `json:"ex_price,string"`
	// ExQty is the number of Excellent copies Card Kingdom will purchase.
	ExQty int `json:"ex_qty"`

	// VgPrice is the buy price for Very Good copies.
	VgPrice float64 `json:"vg_price,string"`
	// VgQty is the number of Very Good copies Card Kingdom will purchase.
	VgQty int `json:"vg_qty"`

	// GPrice is the buy price for Good (heavily played) copies.
	GPrice float64 `json:"g_price,string"`
	// GQty is the number of Good copies Card Kingdom will purchase.
	GQty int `json:"g_qty"`
}

// SinglesPricelist fetches the current singles price list from Card Kingdom.
//
// It is a convenience wrapper around [Pricelist] that discards the [Metadata].
// Passing nil for client will use a default clean HTTP client.
func SinglesPricelist(ctx context.Context, client *http.Client) ([]Product, error) {
	products, _, err := Pricelist(ctx, client, PricelistURL)
	return products, err
}

// SealedPricelist fetches the current sealed-product price list from Card Kingdom.
//
// It is a convenience wrapper around [Pricelist] that discards the [Metadata].
// Passing nil for client will use a default clean HTTP client.
func SealedPricelist(ctx context.Context, client *http.Client) ([]Product, error) {
	products, _, err := Pricelist(ctx, client, SealedListURL)
	return products, err
}

// Pricelist fetches and decodes a Card Kingdom price list from the given link.
//
// If link begins with "http", an HTTP GET request is made using the provided
// client. Passing nil for client will use a default clean HTTP client from
// [github.com/hashicorp/go-cleanhttp]. If link does not begin with "http", it
// is treated as a local file path, which is useful for testing or processing
// cached snapshots.
//
// On a non-200 response, the error includes the status and up to 4 KB of the
// response body. JSON decode errors are wrapped with the source link for
// easier diagnosis.
func Pricelist(ctx context.Context, client *http.Client, link string) ([]Product, Metadata, error) {
	var reader io.Reader
	if strings.HasPrefix(link, "http") {
		if client == nil {
			client = cleanhttp.DefaultClient()
		}

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, link, http.NoBody)
		if err != nil {
			return nil, Metadata{}, err
		}
		req.Header.Set("User-Agent", UserAgent)

		resp, err := client.Do(req)
		if err != nil {
			return nil, Metadata{}, err
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			// Try reading something from the body
			ret, _ := io.ReadAll(io.LimitReader(resp.Body, 4<<10))
			return nil, Metadata{}, fmt.Errorf("GET %s: %s: %s", link, resp.Status, string(ret))
		}

		reader = resp.Body
	} else {
		file, err := os.Open(link)
		if err != nil {
			return nil, Metadata{}, err
		}
		defer file.Close()

		reader = file
	}

	var pricelist Response
	err := json.NewDecoder(reader).Decode(&pricelist)
	if err != nil {
		return nil, Metadata{}, fmt.Errorf("decode %s: %w", link, err)
	}

	return pricelist.Data, pricelist.Meta, nil
}
