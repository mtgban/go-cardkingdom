package cardkingdom

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"
)

// The fixtures are real records from both live feeds, trimmed to a few each:
// a foil single, a non-foil variation, a single whose scryfall_id is null, and
// sealed product that ships internationally and one that does not.
const (
	singlesFixture = "testdata/pricelist.json"
	sealedFixture  = "testdata/sealed_pricelist.json"
)

var wantSingles = []Product{
	{
		ID:          46329,
		SKU:         "FINV-044",
		ScryfallID:  "57e45de5-0e8b-41d3-979b-ec5a29cac682",
		URL:         "mtg/invasion/wayfaring-giant-foil",
		Name:        "Wayfaring Giant",
		Edition:     "Invasion",
		IsFoil:      true,
		PriceRetail: 1.49,
		QtyRetail:   6,
		PriceBuy:    0.45,
		QtyBuying:   10,
		ConditionValues: ConditionValue{
			NMPrice: 1.49, NMQty: 1,
			EXPrice: 1.19, EXQty: 3,
			VGPrice: 0.89, VGQty: 1,
			GPrice: 0.60, GQty: 1,
		},
	},
	{
		ID:          14043,
		SKU:         "ATQ-080C",
		ScryfallID:  "a696c5b6-f216-454d-8029-74e84bbd1428",
		URL:         "mtg/antiquities/mishras-factory-autumn",
		Name:        "Mishra's Factory",
		Variation:   "Autumn",
		Edition:     "Antiquities",
		PriceRetail: 249.99,
		QtyRetail:   2,
		PriceBuy:    165.00,
		QtyBuying:   15,
		ConditionValues: ConditionValue{
			NMPrice: 249.99, EXPrice: 224.99, EXQty: 2,
			VGPrice: 199.99, GPrice: 174.99,
		},
	},
	{
		ID:          333921,
		SKU:         "FRA-0001P",
		URL:         "mtg/promo-pack/emrakul-the-exigent-doom-promo-pack",
		Name:        "Emrakul, the Exigent Doom",
		Variation:   "Promo Pack",
		Edition:     "Promo Pack",
		PriceRetail: 34.99,
		PriceBuy:    22.00,
		QtyBuying:   5,
		ConditionValues: ConditionValue{
			NMPrice: 34.99, EXPrice: 29.74, VGPrice: 26.24, GPrice: 22.74,
		},
	},
}

var wantSealed = []Product{
	{
		ID:                   1132,
		URL:                  "mtg-sealed/mercadian-masques-booster-box",
		Name:                 "Mercadian Masques Booster Box",
		Edition:              "Mercadian Masques",
		PriceRetail:          2699.99,
		QtyRetail:            1,
		PriceBuy:             1755.00,
		QtyBuying:            1,
		ShipsInternationally: true,
	},
	{
		ID:          230638,
		URL:         "mtg-sealed/secret-lair-drop-bitterblossom-dreams",
		Name:        "Secret Lair Drop - Bitterblossom Dreams",
		Edition:     "Secret Lair",
		PriceRetail: 74.99,
		PriceBuy:    33.00,
		QtyBuying:   9,
	},
}

func assertProducts(t *testing.T, got, want []Product) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("got %d products, want %d", len(got), len(want))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("product %d:\n got  %+v\n want %+v", i, got[i], want[i])
		}
	}
}

func TestPricelistFromFile(t *testing.T) {
	products, meta, err := Pricelist(context.Background(), nil, singlesFixture)
	if err != nil {
		t.Fatalf("Pricelist: %v", err)
	}

	assertProducts(t, products, wantSingles)

	if meta.BaseURL != "https://www.cardkingdom.com/" {
		t.Errorf("BaseURL = %q, want %q", meta.BaseURL, "https://www.cardkingdom.com/")
	}

	got, err := meta.CreatedAtTime()
	if err != nil {
		t.Fatalf("CreatedAtTime: %v", err)
	}
	want := time.Date(2026, 10, 1, 4, 4, 53, 0, time.UTC)
	if !got.Equal(want) {
		t.Errorf("CreatedAtTime = %v, want %v", got, want)
	}
}

func TestCreatedAtTimeInvalid(t *testing.T) {
	m := Metadata{CreatedAt: "not-a-date"}
	if _, err := m.CreatedAtTime(); err == nil {
		t.Fatal("CreatedAtTime: expected error for malformed timestamp, got nil")
	}
}

func TestPricelistFromFileMissing(t *testing.T) {
	_, _, err := Pricelist(context.Background(), nil, "testdata/does-not-exist.json")
	if err == nil {
		t.Fatal("Pricelist: expected error for missing file, got nil")
	}
}

func TestPricelistFileContextCancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel before the read is attempted

	_, _, err := Pricelist(ctx, nil, singlesFixture)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("Pricelist error = %v, want context.Canceled", err)
	}
}

func TestPricelistHTTP(t *testing.T) {
	body, err := os.ReadFile(singlesFixture)
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}

	var gotUA string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotUA = r.Header.Get("User-Agent")
		w.Write(body)
	}))
	defer srv.Close()

	products, _, err := Pricelist(context.Background(), srv.Client(), srv.URL)
	if err != nil {
		t.Fatalf("Pricelist: %v", err)
	}
	assertProducts(t, products, wantSingles)

	if gotUA != UserAgent {
		t.Errorf("User-Agent = %q, want %q", gotUA, UserAgent)
	}
}

func TestPricelistHTTPNon200(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "rate limited, slow down", http.StatusTooManyRequests)
	}))
	defer srv.Close()

	_, _, err := Pricelist(context.Background(), srv.Client(), srv.URL)
	if err == nil {
		t.Fatal("Pricelist: expected error for non-200 response, got nil")
	}
	// The error should surface both the status and a body preview.
	if !strings.Contains(err.Error(), "429") {
		t.Errorf("error %q does not mention status 429", err)
	}
	if !strings.Contains(err.Error(), "slow down") {
		t.Errorf("error %q does not include body preview", err)
	}
}

func TestPricelistHTTPBadJSON(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("{not json"))
	}))
	defer srv.Close()

	_, _, err := Pricelist(context.Background(), srv.Client(), srv.URL)
	if err == nil {
		t.Fatal("Pricelist: expected decode error, got nil")
	}
	if !strings.Contains(err.Error(), "decode") {
		t.Errorf("error %q not wrapped with decode context", err)
	}
}

func TestPricelistContextCancelled(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("{}"))
	}))
	defer srv.Close()

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel before the request is made

	_, _, err := Pricelist(ctx, srv.Client(), srv.URL)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("Pricelist error = %v, want context.Canceled", err)
	}
}

// fixtureTransport serves one testdata fixture for any request, recording the
// URL it was asked for. This lets us exercise SinglesPricelist / SealedPricelist,
// which target hardcoded live endpoints, without any network access.
type fixtureTransport struct {
	body    []byte
	lastURL string
}

func (ft *fixtureTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	ft.lastURL = r.URL.String()
	return &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(bytes.NewReader(ft.body)),
		Header:     make(http.Header),
		Request:    r,
	}, nil
}

func TestSinglesAndSealedURLs(t *testing.T) {
	for _, tc := range []struct {
		name    string
		call    func(ctx context.Context, c *http.Client) ([]Product, error)
		fixture string
		want    []Product
		wantURL string
	}{
		{"singles", SinglesPricelist, singlesFixture, wantSingles, PricelistURL},
		{"sealed", SealedPricelist, sealedFixture, wantSealed, SealedListURL},
	} {
		t.Run(tc.name, func(t *testing.T) {
			body, err := os.ReadFile(tc.fixture)
			if err != nil {
				t.Fatalf("read fixture: %v", err)
			}
			ft := &fixtureTransport{body: body}
			client := &http.Client{Transport: ft}

			products, err := tc.call(context.Background(), client)
			if err != nil {
				t.Fatalf("%s: %v", tc.name, err)
			}
			assertProducts(t, products, tc.want)

			if ft.lastURL != tc.wantURL {
				t.Errorf("requested URL = %q, want %q", ft.lastURL, tc.wantURL)
			}
		})
	}
}
