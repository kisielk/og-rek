// +build gofuzz

package ogórek

import (
	"fmt"
	"io/ioutil"
	"log"
	"testing"
)

// TestFuzzGenerate is not a test - it's a program that puts all tests pickles
// from main tests into fuzz/corpus. It is implemented as test because we need
// *_test.go files to be linked in to get to test data defined there.
//
// It is triggered to be run by go:generate from ogorek_test.go .
func TestFuzzGenerate(t *testing.T) {
	for i, test := range tests {
		j := 0
		for _, pickle := range test.picklev {
			if pickle.err != nil {
				continue
			}
			j++

			err := ioutil.WriteFile(
				fmt.Sprintf("fuzz/corpus/test-%d-%d.pickle", i, j),
				[]byte(pickle.data), 0666)
			if err != nil {
				log.Fatal(err)
			}
		}
	}
}
