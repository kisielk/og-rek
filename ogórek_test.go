package ogórek

import (
	"bytes"
	"encoding/hex"
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

var graphitePickle1, _ = hex.DecodeString("80025d71017d710228550676616c75657371035d71042847407d90000000000047407f100000000000474080e0000000000047409764000000000047409c40000000000047409d88000000000047409f74000000000047409c74000000000047409cdc00000000004740a10000000000004740a0d800000000004740938800000000004740a00e00000000004740988800000000004e4e655505737461727471054a00d87a5255047374657071064a805101005503656e6471074a00f08f5255046e616d657108552d5a5a5a5a2e55555555555555552e43434343434343432e4d4d4d4d4d4d4d4d2e5858585858585858582e545454710975612e")
var graphitePickle2, _ = hex.DecodeString("286c70300a286470310a53277374617274270a70320a49313338333738323430300a73532773746570270a70330a4938363430300a735327656e64270a70340a49313338353136343830300a73532776616c756573270a70350a286c70360a463437332e300a61463439372e300a61463534302e300a6146313439372e300a6146313830382e300a6146313839302e300a6146323031332e300a6146313832312e300a6146313834372e300a6146323137362e300a6146323135362e300a6146313235302e300a6146323035352e300a6146313537302e300a614e614e617353276e616d65270a70370a5327757365722e6c6f67696e2e617265612e6d616368696e652e6d65747269632e6d696e757465270a70380a73612e")

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
		{"graphite message1", string(graphitePickle1), []interface{}{map[interface{}]interface{}{"values": []interface{}{float64(473), float64(497), float64(540), float64(1497), float64(1808), float64(1890), float64(2013), float64(1821), float64(1847), float64(2176), float64(2156), float64(1250), float64(2055), float64(1570), None{}, None{}}, "start": int64(1383782400), "step": int64(86400), "end": int64(1385164800), "name": "ZZZZ.UUUUUUUU.CCCCCCCC.MMMMMMMM.XXXXXXXXX.TTT"}}},
		{"graphite message2", string(graphitePickle2), []interface{}{map[interface{}]interface{}{"values": []interface{}{float64(473), float64(497), float64(540), float64(1497), float64(1808), float64(1890), float64(2013), float64(1821), float64(1847), float64(2176), float64(2156), float64(1250), float64(2055), float64(1570), None{}, None{}}, "start": int64(1383782400), "step": int64(86400), "end": int64(1385164800), "name": "user.login.area.machine.metric.minute"}}},
	}
	for _, test := range tests {
		buf := bytes.NewBufferString(test.input)
		dec := NewDecoder(buf)
		v, err := dec.Decode()
		if err != nil {
			t.Error(err)
		}
		if !equal(v, test.expected) {
			t.Errorf("%s: got\n%q\n expected\n%q", test.name, v, test.expected)
		}
	}
}
