package ogórek
// conversion in between Go types to match Python.

import (
	"fmt"
	"math/big"
)


// AsInt64 tries to represent unpickled value to int64.
//
// Python int is decoded as int64, while Python long is decoded as big.Int.
// Go code should use AsInt64 to accept normal-range integers independently of
// their Python representation.
func AsInt64(x interface{}) (int64, error) {
	switch x := x.(type) {
	case int64:
		return x, nil
	case *big.Int:
		if !x.IsInt64() {
			return 0, fmt.Errorf("long outside of int64 range")
		}
		return x.Int64(), nil
	}
	return 0, fmt.Errorf("expect int64|long; got %T", x)
}
