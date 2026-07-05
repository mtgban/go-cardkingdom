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

const fixturePath = "testdata/pricelist.json"

// assertFixtureProducts verifies that products decoded from testdata match the
// expected values, exercising the ",string" JSON tags in particular.
func assertFixtureProducts(t *testing.T, products []Product) {
	t.Helper()

	if len(products) != 2 {
		t.Fatalf("got %d products, want 2", len(products))
	}

	lotus := products[0]
	if lotus.ID != 1234 {
		t.Errorf("ID = %d, want 1234", lotus.ID)
	}
	if lotus.Name != "Black Lotus" {
		t.Errorf("Name = %q, want %q", lotus.Name, "Black Lotus")
	}
	if !lotus.IsFoil {
		t.Errorf("IsFoil = false, want true (parsed from string)")
	}
	if lotus.PriceRetail != 12345.67 {
		t.Errorf("PriceRetail = %v, want 12345.67", lotus.PriceRetail)
	}
	if lotus.PriceBuy != 9000.00 {
		t.Errorf("PriceBuy = %v, want 9000.00", lotus.PriceBuy)
	}
	if lotus.QtyRetail != 2 {
		t.Errorf("QtyRetail = %d, want 2", lotus.QtyRetail)
	}
	if lotus.ConditionValues.ExPrice != 8000.50 {
		t.Errorf("ExPrice = %v, want 8000.50", lotus.ConditionValues.ExPrice)
	}
	if lotus.ConditionValues.GQty != 2 {
		t.Errorf("GQty = %d, want 2", lotus.ConditionValues.GQty)
	}

	box := products[1]
	if box.IsFoil {
		t.Errorf("IsFoil = true, want false (parsed from string)")
	}
	if box.PriceBuy != 0 {
		t.Errorf("PriceBuy = %v, want 0", box.PriceBuy)
	}
	if box.ScryfallID != "" {
		t.Errorf("ScryfallID = %q, want empty", box.ScryfallID)
	}
}

func TestPricelistFromFile(t *testing.T) {
	products, meta, err := Pricelist(context.Background(), nil, fixturePath)
	if err != nil {
		t.Fatalf("Pricelist: %v", err)
	}

	assertFixtureProducts(t, products)

	if meta.BaseURL != "https://www.cardkingdom.com/" {
		t.Errorf("BaseURL = %q, want %q", meta.BaseURL, "https://www.cardkingdom.com/")
	}

	got, err := meta.CreatedAtTime()
	if err != nil {
		t.Fatalf("CreatedAtTime: %v", err)
	}
	want := time.Date(2025, 9, 21, 14, 30, 0, 0, time.UTC)
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

func TestPricelistHTTP(t *testing.T) {
	body, err := os.ReadFile(fixturePath)
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
	assertFixtureProducts(t, products)

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

// fixtureTransport serves the testdata fixture for any request, recording the
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
	body, err := os.ReadFile(fixturePath)
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}

	for _, tc := range []struct {
		name    string
		call    func(ctx context.Context, c *http.Client) ([]Product, error)
		wantURL string
	}{
		{"singles", SinglesPricelist, PricelistURL},
		{"sealed", SealedPricelist, SealedListURL},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ft := &fixtureTransport{body: body}
			client := &http.Client{Transport: ft}

			products, err := tc.call(context.Background(), client)
			if err != nil {
				t.Fatalf("%s: %v", tc.name, err)
			}
			assertFixtureProducts(t, products)

			if ft.lastURL != tc.wantURL {
				t.Errorf("requested URL = %q, want %q", ft.lastURL, tc.wantURL)
			}
		})
	}
}
