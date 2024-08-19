//go:build !go1.19

package ogórek

import (
	"hash/maphash"
)

func maphash_String(seed maphash.Seed, s string) uint64 {
	var h maphash.Hash
	h.SetSeed(seed)
	h.WriteString(s)
	return h.Sum64()
}
