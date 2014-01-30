package ogórek

import (
	"bytes"
	"reflect"
	"testing"
)

func TestEncode(t *testing.T) {

	tests := []struct {
		name  string
		input interface{}
	}{
		{
			"graphite message",
			[]interface{}{map[interface{}]interface{}{"values": []interface{}{float64(473), float64(497), float64(540), float64(1497), float64(1808), float64(1890), float64(2013), float64(1821), float64(1847), float64(2176), float64(2156), float64(1250), float64(2055), float64(1570), None{}, None{}}, "start": int64(1383782400), "step": int64(86400), "end": int64(1385164800), "name": "ZZZZ.UUUUUUUU.CCCCCCCC.MMMMMMMM.XXXXXXXXX.TTT"}},
		},
		{
			"small types",
			[]interface{}{int64(0), int64(1), int64(258), int64(65537), false, true},
		},
	}

	for _, tt := range tests {
		p := &bytes.Buffer{}
		e := NewEncoder(p)
		e.Encode(tt.input)

		d := NewDecoder(bytes.NewReader(p.Bytes()))
		output, _ := d.Decode()

		if !reflect.DeepEqual(tt.input, output) {
			t.Errorf("%s: got\n%q\n expected\n%q", tt.name, output, tt.input)
		}

	}
}
