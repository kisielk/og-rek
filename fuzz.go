// +build gofuzz

package ogórek

import (
	"bytes"
	"fmt"
	"reflect"
)

func Fuzz(data []byte) int {
	// obj = decode(data) - this tests things like stack overflow in Decoder
	buf := bytes.NewBuffer(data)
	dec := NewDecoder(buf)
	obj, err := dec.Decode()
	if err != nil {
		return 0
	}

	// assert decode(encode(obj)) == obj
	//
	// this tests that Encoder and Decoder are consistent: sometimes
	// Encoder is right and Decoder succeeds, but decodes data incorectly;
	// sometimes vice versa. We can be safe to test for idempotency here
	// because obj - as we got it as decoding from input - is known not to
	// contain arbitrary Go structs.
	for proto := 0; proto <= highestProtocol; proto++ {
		buf.Reset()
		enc := NewEncoderWithConfig(buf, &EncoderConfig{
			Protocol: proto,
		})
		err = enc.Encode(obj)
		if err != nil {
			// must succeed, as obj was obtained via successful decode
			// the only exception is that we cannot encode non-string Ref at proto=0
			if proto == 0 && err == errP0PersIDStringLineOnly {
				continue
			}
			panic(fmt.Sprintf("protocol %d: encode error: %s", proto, err))
		}
		encoded := buf.String()

		dec = NewDecoder(bytes.NewStringReader(encoded))
		obj2, err := dec.Decode()
		if err != nil {
			// must succeed, as buf should contain valid pickle from encoder
			panic(fmt.Sprintf("protocol %d: decode back error: err\npickle: %q", proto, encoded)
		}

		if !reflect.DeepEqual(obj, obj2) {
			panic(fmt.Sprintf("protocol %d: decode·encode != identity:\nhave: %#v\nwant: %#v", proto, obj2, obj))
		}
	}

	return 1
}
