package ogórek

import (
	"bytes"
	"math/big"
	"reflect"
	"testing"
)

func bigInt(s string) *big.Int {
	i := new(big.Int)
	i.SetString(s, 10)
	return i
}

func equal(a, b interface{}) bool {
	if reflect.TypeOf(a) != reflect.TypeOf(b) {
		return false
	}

	switch a.(type) {
	case []interface{}:
		ia := a.([]interface{})
		ib := b.([]interface{})
		if len(ia) != len(ib) {
			return false
		}
		for i := 0; i < len(ia); i++ {
			if !equal(ia[i], ib[i]) {
				return false
			}
		}
		return true
	case map[interface{}]interface{}:
		ia := a.(map[interface{}]interface{})
		ib := b.(map[interface{}]interface{})
		if len(ia) != len(ib) {
			return false
		}
		for k := range ia {
			if !equal(ia[k], ib[k]) {
				return false
			}
		}
		return true
	case *big.Int:
		return a.(*big.Int).Cmp(b.(*big.Int)) == 0
	default:
		return a == b
	}
}

func TestMarker(t *testing.T) {
	buf := bytes.Buffer{}
	dec := NewDecoder(&buf)
	dec.mark()
	if dec.marker() != 0 {
		t.Error("no marker found")
	}
}

func TestDecode(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected interface{}
	}{
		{"int", "I5\n.", int64(5)},
		{"float", "F1.23\n.", float64(1.23)},
		{"long", "L12321231232131231231L\n.", bigInt("12321231232131231231")},
		{"None", "N.", None{}},
		{"empty tuple", "(t.", []interface{}{}},
		{"tuple of two ints", "(I1\nI2\ntp0\n.", []interface{}{int64(1), int64(2)}},
		{"nested tuples", "((I1\nI2\ntp0\n(I3\nI4\ntp1\ntp2\n.",
			[]interface{}{[]interface{}{int64(1), int64(2)}, []interface{}{int64(3), int64(4)}}},
		{"empty list", "(lp0\n.", []interface{}{}},
		{"list of numbers", "(lp0\nI1\naI2\naI3\naI4\na.", []interface{}{int64(1), int64(2), int64(3), int64(4)}},
		{"string", "S'abc'\np0\n.", string("abc")},
		{"unicode", "V\\u65e5\\u672c\\u8a9e\np0\n.", string("日本語")},
		{"empty dict", "(dp0\n.", make(map[interface{}]interface{})},
		{"dict with strings", "(dp0\nS'a'\np1\nS'1'\np2\nsS'b'\np3\nS'2'\np4\ns.", map[interface{}]interface{}{"a": "1", "b": "2"}},
	}
	for _, test := range tests {
		buf := bytes.NewBufferString(test.input)
		dec := NewDecoder(buf)
		v, err := dec.Decode()
		if err != nil {
			t.Error(err)
		}
		if !equal(v, test.expected) {
			t.Errorf("%s: got %q expected %q", test.name, v, test.expected)
		}
	}
}
