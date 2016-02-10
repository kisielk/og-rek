package ogórek

import (
	"bytes"
	"testing"
)

func TestFuzzCrashers(t *testing.T) {
	crashers := []string{
		"\x94",
		"R",
		"l",
		"q0",
		"NNd",
		"S'",
		"r0000",
		"a",
		"(]R",
		"]NR",
		"s",
		"Nu",
		"L\n",
		"Lq\n",
		"\x85",
		"N\x86",
		"NN\x87",
		"S\n",
		"p0\n",

		// TODO: fix this by creating some sort of hashable-check

		// // runtime error: hash of unhashable type map[interface {}]interface {}
		// // ogorek.go:823
		// "}((dU-00000000000000" +
		// 	"00000000000000000000" +
		// 	"00000000000u",
		// // runtime error: hash of unhashable type []interface {}
		// // ogorek.go:627
		// "((t(td",
		// // panic: runtime error: hash of unhashable type map[interface {}]interface {}
		// // ogorek.go:803
		// "(d(d(s",
		// "(dg0\n(lp\nsg\nS''\ns",
	}
	for _, s := range crashers {
		_, err := NewDecoder(bytes.NewBufferString(s)).Decode()
		if err == nil {
			// Crashers that were actually valid pickle are not
			// included here, so we're expecting errors.
			t.Errorf("expected error for %q, got none", s)
		}
	}
}
