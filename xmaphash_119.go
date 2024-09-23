//go:build go1.19

package ogórek

import (
	"hash/maphash"
)

func maphash_String(seed maphash.Seed, s string) uint64 {
	return maphash.String(seed, s)
}
