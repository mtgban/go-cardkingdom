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
	products, _, err := PricelistFromFile(fixturePath)
	if err != nil {
		t.Fatal(err)
	}
	assertFixtureProducts(t, products)
	body, err := os.ReadFile(fixturePath)
	if err != nil {
		t.Fatal(err)
	}
	products, _, err = DecodePricelist(strings.NewReader(string(body)))
	if err != nil {
		t.Fatal(err)
	}
	assertFixtureProducts(t, products)
	for _, link := range []string{fixturePath, "ftp://example.com/list"} {
		if _, _, err := PricelistFromURL(context.Background(), nil, link); err == nil {
			t.Errorf("PricelistFromURL(%q) accepted a non-HTTP source", link)
		}
	}
	_, _, err = DecodePricelist(strings.NewReader("{bad"))
	var syntax *json.SyntaxError
	if !errors.As(err, &syntax) {
		t.Fatalf("decode error = %v, want SyntaxError", err)
	}
	_, _, err = PricelistFromFile("missing-file.json")
	if !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("file error = %v, want ErrNotExist", err)
	}
}
