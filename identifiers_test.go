package cardkingdom

import (
	"encoding/json"
	"testing"
)

func TestIdentifierWireFormat(t *testing.T) {
	var got struct {
		ID  ProductID `json:"id"`
		SKU SKU       `json:"sku"`
	}
	const wire = `{"id":123,"sku":"ABC-1"}`
	if err := json.Unmarshal([]byte(wire), &got); err != nil {
		t.Fatal(err)
	}
	if got.ID != 123 || got.SKU != "ABC-1" {
		t.Fatalf("decoded = %+v", got)
	}
	data, err := json.Marshal(got)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != wire {
		t.Errorf("JSON = %s, want %s", data, wire)
	}
}
