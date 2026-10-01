// Package cardkingdom provides a client for the Card Kingdom public price list API.
//
// It supports fetching both singles and sealed-product price lists, decoding
// them into typed Go structs. Requests are made over HTTP using a standard
// [net/http.Client], and every function accepts a [context.Context] for
// cancellation and deadline control.
//
// [SinglesPricelist], [SealedPricelist] and [Pricelist] return the products.
// Their counterparts [SinglesPricelistFile], [SealedPricelistFile] and
// [LoadPricelistFile] return the whole [PricelistFile], with the time Card
// Kingdom built it. [Pricelist] and [LoadPricelistFile] also read a local
// file.
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

	// DefaultTimeout is the request timeout applied to the default HTTP
	// client used when a nil client is passed. It only affects the default
	// client; a caller-supplied client is used as-is.
	DefaultTimeout = 30 * time.Second
)

// PricelistFile is a published price list as Card Kingdom sends it: its
// metadata and its products.
type PricelistFile struct {
	Meta Metadata  `json:"meta"`
	Data []Product `json:"data"`
}

// Metadata is the header Card Kingdom sends alongside the products.
type Metadata struct {
	// CreatedAt is when Card Kingdom built the list, formatted as
	// "2006-01-02 15:04:05" with no timezone. Use [Metadata.CreatedAtTime]
	// to parse it into a [time.Time].
	CreatedAt string `json:"created_at"`

	// BaseURL is the base that each [Product.URL] is relative to.
	BaseURL string `json:"base_url"`
}

// CreatedAtTime parses the CreatedAt field into a [time.Time].
// The expected layout is "2006-01-02 15:04:05". An error is returned if the
// value does not match that format. The timezone-free value is interpreted
// as UTC; the feed does not identify its source timezone.
func (m Metadata) CreatedAtTime() (time.Time, error) {
	return time.Parse("2006-01-02 15:04:05", m.CreatedAt)
}

// Product represents a single purchasable item in the Card Kingdom catalog,
// either a trading card or a sealed product.
type Product struct {
	// ID is the Card Kingdom internal product identifier.
	ID int `json:"id"`

	// SKU is the stock-keeping unit code for this listing.
	SKU string `json:"sku"`

	// ScryfallID is the Scryfall UUID for this card, suitable for
	// cross-referencing with the Scryfall API. Empty for sealed products, and
	// for the singles the feed sends with a null scryfall_id.
	ScryfallID string `json:"scryfall_id"`

	// URL is a product path relative to [Metadata.BaseURL].
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

	// ConditionValues holds per-condition retail prices and stock quantities.
	ConditionValues ConditionValue `json:"condition_values"`

	// ShipsInternationally reports whether Card Kingdom ships this item
	// outside the United States. Only meaningful for sealed product; always
	// false (and absent from the feed) for singles.
	ShipsInternationally bool `json:"ships_internationally"`
}

// ConditionValue holds retail prices and stock quantities broken down by
// card condition. Prices are in USD.
type ConditionValue struct {
	// NMPrice is the retail price for Near Mint copies.
	NMPrice float64 `json:"nm_price,string"`
	// NMQty is the number of Near Mint copies in stock.
	NMQty int `json:"nm_qty"`

	// EXPrice is the retail price for Excellent copies.
	EXPrice float64 `json:"ex_price,string"`
	// EXQty is the number of Excellent copies in stock.
	EXQty int `json:"ex_qty"`

	// VGPrice is the retail price for Very Good copies.
	VGPrice float64 `json:"vg_price,string"`
	// VGQty is the number of Very Good copies in stock.
	VGQty int `json:"vg_qty"`

	// GPrice is the retail price for Good (heavily played) copies.
	GPrice float64 `json:"g_price,string"`
	// GQty is the number of Good copies in stock.
	GQty int `json:"g_qty"`
}

// SinglesPricelist fetches the current singles price list from Card Kingdom.
// Passing nil for client uses a default clean HTTP client; see [Pricelist].
func SinglesPricelist(ctx context.Context, client *http.Client) ([]Product, error) {
	return Pricelist(ctx, client, PricelistURL)
}

// SealedPricelist fetches the current sealed-product price list from Card
// Kingdom. Passing nil for client uses a default clean HTTP client; see
// [Pricelist].
func SealedPricelist(ctx context.Context, client *http.Client) ([]Product, error) {
	return Pricelist(ctx, client, SealedListURL)
}

// SinglesPricelistFile is [SinglesPricelist] with the time Card Kingdom
// built the list; see [LoadPricelistFile].
func SinglesPricelistFile(ctx context.Context, client *http.Client) (*PricelistFile, error) {
	return LoadPricelistFile(ctx, client, PricelistURL)
}

// SealedPricelistFile is [SealedPricelist] with the time Card Kingdom built
// the list; see [LoadPricelistFile].
func SealedPricelistFile(ctx context.Context, client *http.Client) (*PricelistFile, error) {
	return LoadPricelistFile(ctx, client, SealedListURL)
}

// Pricelist fetches and decodes a Card Kingdom price list from the given link.
//
// If link begins with "http://" or "https://", an HTTP GET request is made
// using the provided client. Passing nil for client will use a default clean
// HTTP client from [github.com/hashicorp/go-cleanhttp] with a [DefaultTimeout]
// request timeout. Otherwise link is treated as a local file path, which is
// useful for testing or processing cached snapshots. In that case ctx is
// only checked before the read starts; the read itself cannot be cancelled.
//
// On a non-200 response, the error includes the status and up to 4 KB of the
// response body. JSON decode errors are wrapped with the source link for
// easier diagnosis.
func Pricelist(ctx context.Context, client *http.Client, link string) ([]Product, error) {
	file, err := load(ctx, client, link)
	if err != nil {
		return nil, err
	}
	return file.Data, nil
}

// LoadPricelistFile is [Pricelist] returning the whole [PricelistFile], with
// the [Metadata] that says when Card Kingdom built the list, so a caller can
// refuse one that has stopped updating. How old is too old is the caller's to
// say.
func LoadPricelistFile(ctx context.Context, client *http.Client, link string) (*PricelistFile, error) {
	return load(ctx, client, link)
}

func load(ctx context.Context, client *http.Client, link string) (*PricelistFile, error) {
	if strings.HasPrefix(link, "http://") || strings.HasPrefix(link, "https://") {
		return fetch(ctx, client, link)
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	file, err := os.Open(link)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	return decode(file, link)
}

func fetch(ctx context.Context, client *http.Client, link string) (*PricelistFile, error) {
	if client == nil {
		client = cleanhttp.DefaultClient()
		client.Timeout = DefaultTimeout
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, link, http.NoBody)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", UserAgent)
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		ret, _ := io.ReadAll(io.LimitReader(resp.Body, 4<<10))
		return nil, fmt.Errorf("GET %q: %s: %s", link, resp.Status, string(ret))
	}
	return decode(resp.Body, link)
}

func decode(reader io.Reader, source string) (*PricelistFile, error) {
	var file PricelistFile
	if err := json.NewDecoder(reader).Decode(&file); err != nil {
		return nil, fmt.Errorf("decode %q: %w", source, err)
	}
	return &file, nil
}
