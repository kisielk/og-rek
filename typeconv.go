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
func AsInt64(x any) (int64, error) {
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


// AsBytes tries to represent unpickled value as Bytes.
//
// It succeeds only if the value is either [Bytes], or [ByteString].
// It does not succeed if the value is string or any other type.
//
// [ByteString] is treated related to [Bytes] because [ByteString] represents str
// type from py2 which can contain both string and binary data.
func AsBytes(x any) (Bytes, error) {
	switch x := x.(type) {
	case Bytes:
		return x, nil
	case ByteString:
		return Bytes(x), nil
	}
	return "", fmt.Errorf("expect bytes|bytestr; got %T", x)
}

// AsString tries to represent unpickled value as string.
//
// It succeeds only if the value is either string, or [ByteString].
// It does not succeed if the value is [Bytes] or any other type.
//
// [ByteString] is treated related to string because [ByteString] represents str
// type from py2 which can contain both string and binary data.
func AsString(x any) (string, error) {
	switch x := x.(type) {
	case string:
		return x, nil
	case ByteString:
		return string(x), nil
	}
	return "", fmt.Errorf("expect unicode|bytestr; got %T", x)
}


// stringEQ compares arbitrary x to string y.
//
// It succeeds only if AsString(x) succeeds and string data of x equals to y.
func stringEQ(x any, y string) bool {
	s, err := AsString(x)
	if err != nil {
		return false
	}
	return s == y
}
