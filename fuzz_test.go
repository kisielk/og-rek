// +build gofuzz

package ogórek

import (
	"crypto/sha1"
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
	for _, test := range tests {
		for _, pickle := range test.picklev {
			if pickle.err != nil {
				continue
			}

			err := ioutil.WriteFile(
				fmt.Sprintf("fuzz/corpus/test-%x.pickle", sha1.Sum([]byte(pickle.data))),
				[]byte(pickle.data), 0666)
			if err != nil {
				log.Fatal(err)
			}
		}
	}
}
