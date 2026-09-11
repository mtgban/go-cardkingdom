package cardkingdom

import "testing"

func TestConditionLookup(t *testing.T) {
	cv := ConditionValue{NMPrice: 4, NMQty: 40, EXPrice: 3, EXQty: 30, VGPrice: 2, VGQty: 20, GPrice: 1, GQty: 10}
	for _, tc := range []struct {
		c     Condition
		price float64
		qty   int
	}{{NM, 4, 40}, {EX, 3, 30}, {VG, 2, 20}, {G, 1, 10}} {
		p, q, ok := cv.Lookup(tc.c)
		if !ok || p != tc.price || q != tc.qty {
			t.Errorf("Lookup(%q) = %v, %v, %v; want %v, %v, true", tc.c, p, q, ok, tc.price, tc.qty)
		}
	}
	for _, c := range []Condition{"", "invalid", "nm"} {
		if p, q, ok := cv.Lookup(c); ok || p != 0 || q != 0 {
			t.Errorf("Lookup(%q) = %v, %v, %v; want 0, 0, false", c, p, q, ok)
		}
	}
	if _, _, ok := (ConditionValue{}).Lookup(NM); !ok {
		t.Error("zero stock must remain a known condition")
	}
}

func TestConditionsReturnsCopy(t *testing.T) {
	got := Conditions()
	want := [4]Condition{NM, EX, VG, G}
	if got != want {
		t.Fatalf("Conditions() = %v, want %v", got, want)
	}
	got[0] = G
	if Conditions() != want {
		t.Error("caller mutated canonical order")
	}
}
