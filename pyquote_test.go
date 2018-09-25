package ogórek

import (
	"testing"
)

// CodecTestCase represents 1 test case of a coder or decoder.
//
// Under the given transformation function in must be transformed to out.
type CodecTestCase struct {
	in, out string
}

// testCodec tests transform func applied to all test cases from testv.
func testCodec(t *testing.T, transform func(in string)(string, error), testv []CodecTestCase) {
	for _, tt := range testv {
		s, err := transform(tt.in)
		if err != nil {
			t.Errorf("%q -> error: %s", tt.in, err)
			continue
		}

		if s != tt.out {
			t.Errorf("%q -> unexpected:\nhave: %q\nwant: %q", tt.in, s, tt.out)
		}
	}
}

func TestPyDecodeStringEscape(t *testing.T) {
	testCodec(t, pydecodeStringEscape, []CodecTestCase{
		{`hello`, "hello"},
		{"hello\\\nworld", "helloworld"},
		{`\\`, `\`},
		{`\'\"`, `'"`},
		{`\b\f\t\n\r\v\a`, "\b\f\t\n\r\v\a"},
		{`\000\001\376\377`, "\000\001\376\377"},
		{`\x00\x01\x7f\x80\xfe\xff`, "\x00\x01\x7f\x80\xfe\xff"},
		// vvv stays as is
		{`\u1234\U00001234\c`, `\u1234\U00001234\c`},
	})
}
