//go:build go1.21

package ogórek

import (
	"math/big"
)

func bigInt_Float64(b *big.Int) (float64, big.Accuracy) {
	return b.Float64()
}
