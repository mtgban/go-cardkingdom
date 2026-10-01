package cardkingdom

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"strings"
	"testing"
)

func TestExplicitSources(t *testing.T) {
	t.Run("file", func(t *testing.T) {
		for _, tc := range []struct {
			path string
			want []Product
		}{
			{singlesFixture, wantSingles},
			{sealedFixture, wantSealed},
		} {
			products, _, err := PricelistFromFile(tc.path)
			if err != nil {
				t.Fatalf("PricelistFromFile(%q): %v", tc.path, err)
			}
			assertProducts(t, products, tc.want)
		}
	})

	t.Run("file missing", func(t *testing.T) {
		_, _, err := PricelistFromFile("missing-file.json")
		if !errors.Is(err, os.ErrNotExist) {
			t.Fatalf("file error = %v, want ErrNotExist", err)
		}
	})

	t.Run("reader", func(t *testing.T) {
		file, err := os.Open(singlesFixture)
		if err != nil {
			t.Fatal(err)
		}
		defer file.Close()
		products, _, err := DecodePricelist(file)
		if err != nil {
			t.Fatalf("DecodePricelist: %v", err)
		}
		assertProducts(t, products, wantSingles)
	})

	t.Run("reader syntax error", func(t *testing.T) {
		_, _, err := DecodePricelist(strings.NewReader("{bad"))
		var syntax *json.SyntaxError
		if !errors.As(err, &syntax) {
			t.Fatalf("decode error = %v, want SyntaxError", err)
		}
	})

	t.Run("url rejects non-HTTP", func(t *testing.T) {
		for _, link := range []string{singlesFixture, "ftp://example.com/list"} {
			if _, _, err := PricelistFromURL(context.Background(), nil, link); err == nil {
				t.Errorf("PricelistFromURL(%q) accepted a non-HTTP source", link)
			}
		}
	})
}
