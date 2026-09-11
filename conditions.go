package cardkingdom

// Condition identifies a Card Kingdom condition grade.
// Values outside the declared constants are rejected by ConditionValue.Lookup.
type Condition string

const (
	// NM denotes Near Mint.
	NM Condition = "NM"
	// EX denotes Excellent.
	EX Condition = "EX"
	// VG denotes Very Good.
	VG Condition = "VG"
	// G denotes Good.
	G Condition = "G"
)

// Conditions returns the grades in order from Near Mint to Good.
// Each call returns an independent array.
func Conditions() [4]Condition { return [4]Condition{NM, EX, VG, G} }

// Lookup returns a condition's price and quantity together.
// The boolean is false for an unknown condition, including the zero value.
// A known condition with zero price or quantity still returns true.
func (cv ConditionValue) Lookup(condition Condition) (price float64, quantity int, ok bool) {
	switch condition {
	case NM:
		return cv.NMPrice, cv.NMQty, true
	case EX:
		return cv.EXPrice, cv.EXQty, true
	case VG:
		return cv.VGPrice, cv.VGQty, true
	case G:
		return cv.GPrice, cv.GQty, true
	default:
		return 0, 0, false
	}
}
