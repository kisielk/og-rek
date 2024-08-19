// +build gofuzz

package ogórek

import (
	"bytes"
	"fmt"
)

func Fuzz(data []byte) int {
	f1 := fuzz(data, false)
	f2 := fuzz(data, true)

	f := f1+f2
	if f > 1 {
		f = 1
	}
	return f
}

func fuzz(data []byte, strictUnicode bool) int {
	// obj = decode(data) - this tests things like stack overflow in Decoder
	buf := bytes.NewBuffer(data)
	dec := NewDecoderWithConfig(buf, &DecoderConfig{
		StrictUnicode: strictUnicode,
	})
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
		subj := fmt.Sprintf("strictUnicode %v: protocol %d", strictUnicode, proto)

		buf.Reset()
		enc := NewEncoderWithConfig(buf, &EncoderConfig{
			Protocol:      proto,
			StrictUnicode: strictUnicode,
		})
		err = enc.Encode(obj)
		if err != nil {
			// must succeed, as obj was obtained via successful decode
			// some  exceptions are accounted for first:
			switch {
			case proto == 0 && err == errP0PersIDStringLineOnly:
				// we cannot encode non-string Ref at proto=0
				continue

			case proto == 0 && err == errP0UnicodeUTF8Only:
				// we cannot encode non-UTF8 Unicode at proto=0
				continue

			case proto <= 3 && err == errP0123GlobalStringLineOnly:
				// we cannot encode Class (GLOBAL opcode) with \n at proto <= 4
				continue
			}
			panic(fmt.Sprintf("%s: encode error: %s", subj, err))
		}
		encoded := buf.String()

		dec = NewDecoderWithConfig(bytes.NewBufferString(encoded), &DecoderConfig{
			StrictUnicode: strictUnicode,
		})
		obj2, err := dec.Decode()
		if err != nil {
			// must succeed, as buf should contain valid pickle from encoder
			panic(fmt.Sprintf("%s: decode back error: %s\npickle: %q", subj, err, encoded))
		}

		if !deepEqual(obj, obj2) {
			panic(fmt.Sprintf("%s: decode·encode != identity:\nhave: %#v\nwant: %#v", subj, obj2, obj))
		}
	}

	return 1
}
