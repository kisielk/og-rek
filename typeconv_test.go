package ogórek

import (
	"fmt"
	"reflect"
	"testing"
)

func TestAsInt64(t *testing.T) {
	Etype := func(typename string) error {
		return fmt.Errorf("expect int64|long; got %s", typename)
	}
	Erange := fmt.Errorf("long outside of int64 range")

	testv := []struct {
		in    interface{}
		outOK interface{}
	}{
		{int64(0),                       int64(0)},
		{int64(1),                       int64(1)},
		{int64(2),                       int64(2)},
		{int64(123),                     int64(123)},
		{int64(0x7fffffffffffffff),      int64(0x7fffffffffffffff)},
		{int64(-0x8000000000000000),     int64(-0x8000000000000000)},
		{bigInt("0"),                    int64(0)},
		{bigInt("1"),                    int64(1)},
		{bigInt("2"),                    int64(2)},
		{bigInt("123"),                  int64(123)},
		{bigInt("9223372036854775807"),  int64(0x7fffffffffffffff)},
		{bigInt("9223372036854775808"),  Erange},
		{bigInt("-9223372036854775808"), int64(-0x8000000000000000)},
		{bigInt("-9223372036854775809"), Erange},
		{1.0,                            Etype("float64")},
		{"a",                            Etype("string")},
	}

	for _, tt := range testv {
		iout, err := AsInt64(tt.in)
		var out interface{} = iout
		if err != nil {
			out = err
			if iout != 0 {
				t.Errorf("%T %#v -> err, but ret int64 = %d  ; want 0",
					tt.in, tt.in, iout)
			}
		}

		if !deepEqual(out, tt.outOK) {
			t.Errorf("%T %#v -> %T %#v  ; want %T %#v",
				tt.in, tt.in, out, out, tt.outOK, tt.outOK)
		}
	}

}

func TestAsBytesString(t *testing.T) {
	Ebytes := func(x interface{}) error {
		return fmt.Errorf("expect bytes|bytestr; got %T", x)
	}
	Estring := func(x interface{}) error {
		return fmt.Errorf("expect unicode|bytestr; got %T", x)
	}

	const y = true
	const n = false

	testv := []struct {
		in  interface{}
		bok bool // AsBytes  succeeds
		sok bool // AsString succeeds
	}{
		{"мир",             n, y},
		{Bytes("мир"),      y, n},
		{ByteString("мир"), y, y},
		{1.0,               n, n},
		{None{},            n, n},
	}

	for _, tt := range testv {
		bout, berr := AsBytes(tt.in)
		sout, serr := AsString(tt.in)

		sin := ""
		xin := reflect.ValueOf(tt.in)
		if xin.Kind() == reflect.String {
			sin = xin.String()
		}

		boutOK := Bytes(sin)
		var berrOK error
		if !tt.bok {
			boutOK = ""
			berrOK = Ebytes(tt.in)
		}

		soutOK := sin
		var serrOK error
		if !tt.sok {
			soutOK = ""
			serrOK = Estring(tt.in)
		}

		if !(bout == boutOK && deepEqual(berr, berrOK)) {
			t.Errorf("%#v: AsBytes:\nhave %#v %#v\nwant %#v %#v",
				tt.in, bout, berr, boutOK, berrOK)
		}

		if !(sout == soutOK && deepEqual(serr, serrOK)) {
			t.Errorf("%#v: AsString:\nhave %#v %#v\nwant %#v %#v",
				tt.in, sout, serr, soutOK, serrOK)
		}
	}
}
