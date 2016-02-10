// +build gofuzz

package ogórek

import (
	"bytes"
)

func Fuzz(data []byte) int {
	_, err := NewDecoder(bytes.NewBuffer(data)).Decode()
	if err != nil {
		return 0
	}
	return 1
}
