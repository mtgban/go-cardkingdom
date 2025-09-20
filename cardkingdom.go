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
	PricelistURL  = "https://api.cardkingdom.com/api/v2/pricelist"
	SealedListURL = "https://api.cardkingdom.com/api/sealed_pricelist"

	UserAgent = "go-cardkingdom"
)

type Response struct {
	Meta Metadata  `json:"meta"`
	Data []Product `json:"data"`
}

type Metadata struct {
	CreatedAt string `json:"created_at"`
	BaseURL   string `json:"base_url"`
}

func (m Metadata) CreatedAtTime() (time.Time, error) {
	return time.Parse("2006-01-02 15:04:05", m.CreatedAt)
}

type Product struct {
	ID              int            `json:"id"`
	Sku             string         `json:"sku"`
	ScryfallID      string         `json:"scryfall_id"`
	URL             string         `json:"url"`
	Name            string         `json:"name"`
	Variation       string         `json:"variation"`
	Edition         string         `json:"edition"`
	IsFoil          bool           `json:"is_foil,string"`
	PriceRetail     float64        `json:"price_retail,string"`
	QtyRetail       int            `json:"qty_retail"`
	PriceBuy        float64        `json:"price_buy,string"`
	QtyBuying       int            `json:"qty_buying"`
	ConditionValues ConditionValue `json:"condition_values"`
}

type ConditionValue struct {
	NmPrice float64 `json:"nm_price,string"`
	NmQty   int     `json:"nm_qty"`
	ExPrice float64 `json:"ex_price,string"`
	ExQty   int     `json:"ex_qty"`
	VgPrice float64 `json:"vg_price,string"`
	VgQty   int     `json:"vg_qty"`
	GPrice  float64 `json:"g_price,string"`
	GQty    int     `json:"g_qty"`
}

func SinglesPricelist(ctx context.Context, client *http.Client) ([]Product, error) {
	products, _, err := Pricelist(ctx, client, PricelistURL)
	return products, err
}

func SealedPricelist(ctx context.Context, client *http.Client) ([]Product, error) {
	products, _, err := Pricelist(ctx, client, SealedListURL)
	return products, err
}

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
