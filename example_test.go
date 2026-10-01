package cardkingdom_test

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/mtgban/go-cardkingdom"
)

func ExampleSinglesPricelist() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	singles, err := cardkingdom.SinglesPricelist(ctx, nil)
	if err != nil {
		panic(err)
	}
	fmt.Printf("found %d products\n", len(singles))
}

func ExampleSealedPricelist() {
	client := &http.Client{Timeout: 10 * time.Second}
	sealed, err := cardkingdom.SealedPricelist(context.Background(), client)
	if err != nil {
		panic(err)
	}
	fmt.Printf("found %d sealed products\n", len(sealed))
}

func ExampleDecodePricelist() {
	body := `{
		"meta": {"created_at": "2026-10-01 04:05:00", "base_url": "https://www.cardkingdom.com/"},
		"data": [{
			"id": 1132,
			"url": "mtg-sealed/mercadian-masques-booster-box",
			"name": "Mercadian Masques Booster Box",
			"edition": "Mercadian Masques",
			"ships_internationally": true,
			"price_retail": "2699.99",
			"qty_retail": 1,
			"price_buy": "1755.00",
			"qty_buying": 1
		}]
	}`
	products, meta, err := cardkingdom.DecodePricelist(strings.NewReader(body))
	if err != nil {
		panic(err)
	}
	created, err := meta.CreatedAtTime()
	if err != nil {
		panic(err)
	}
	for _, p := range products {
		fmt.Printf("%s: $%.2f retail, $%.2f buylist\n", p.Name, p.PriceRetail, p.PriceBuy)
		fmt.Println(meta.BaseURL + p.URL)
	}
	fmt.Println(created.Format(time.RFC3339))
	// Output:
	// Mercadian Masques Booster Box: $2699.99 retail, $1755.00 buylist
	// https://www.cardkingdom.com/mtg-sealed/mercadian-masques-booster-box
	// 2026-10-01T04:05:00Z
}
