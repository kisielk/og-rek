package ogórek
// Utilities that complement std reflect package.

import (
	"reflect"
)


// deepEqual is like reflect.DeepEqual but also supports Dict.
//
// It is needed because reflect.DeepEqual considers two Dicts not-equal because
// each Dict is made with its own seed.
//
// XXX only top-level Dict is supported currently.
//     For example comparing Dict inside list with the same won't work.
func deepEqual(a, b any) bool {
	da, ok := a.(Dict)
	if !ok {
		return reflect.DeepEqual(a, b)
	}
	db, ok := b.(Dict)
	if !ok {
		return false // Dict != non-dict
	}

	if da.Len() != db.Len() {
		return false
	}

	// XXX O(n^2) because we want to compare keys exactly and so cannot use
	//     db.Get(ka) because Dict.Get uses general equality that would match e.g. int == int64
	eq := true
	da.Iter()(func(ka, va any) bool {
		keq := false
		db.Iter()(func(kb, vb any) bool {
			// NOTE don't use reflect.Equal(ka,kb) because it does not handle e.g. big.Int
			if reflect.TypeOf(ka) == reflect.TypeOf(kb) && equal(ka,kb) {
				if reflect.DeepEqual(va, vb) {
					keq = true
				}
				return false
			}
			return true
		})
		if !keq {
			eq = false
			return false
		}
		return true
	})

	return eq
}
