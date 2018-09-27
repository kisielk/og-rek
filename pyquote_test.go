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

func TestPyQuote(t *testing.T) {
	testCodec(t, func(in string) (string, error) {
		return pyquote(in), nil
	}, []CodecTestCase{
		{`\"'`, `"\\\"'"`},
		{"\x80hello мир", `"\x80hello мир"`},
		{"\n\r\x01", `"\n\r\x01"`},
	})
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

func TestPyEncodeRawUnicodeEscape(t *testing.T) {
	testCodec(t, func(in string) (string, error) {
		return pyencodeRawUnicodeEscape(in), nil
	}, []CodecTestCase{
		{"\xc3\x28", "\xc3\x28"}, // invalid UTF-8
		{"\x00\x01\x80\xfe\xffabc", "\x00\x01\x80\xfe\xffabc"},
		{`\`, `\u005c`},
		{"\n", `\u000a`},
		{`"'`, `"'`},
		{"hello\nмир", `hello\u000a\u043c\u0438\u0440`},
		{"hello\nÐ¼Ð¸Ñ\u0080\x01", `hello\u000aмир`+"\x01"},
		{"\u1234\U00004321", `\u1234\u4321`},
		{"\U00012345", `\U00012345`},
	})
}

func TestPyDecodeRawUnicodeEscape(t *testing.T) {
	testCodec(t, pydecodeRawUnicodeEscape, []CodecTestCase{
		{`hello`, "hello"},
		{"\x00\x01\x80\xfe\xff", "\u0000\u0001\u0080\u00fe\u00ff"},
		{`\`, `\`},
		{`\\`, `\\`},
		{`\\\`, `\\\`},
		{`\\\\`, `\\\\`},
		{`\u1234\U00004321`, "\u1234\U00004321"},
		{`\\u1234\\U00004321`, `\\u1234\\U00004321`},
		{`\\\u1234\\\U00004321`, "\\\\\u1234\\\\\U00004321"},
		{`\\\\u1234\\\\U00004321`, `\\\\u1234\\\\U00004321`},
		{`\\\\\u1234\\\\\U00004321`, "\\\\\\\\\u1234\\\\\\\\\U00004321"},
		// vvv stays as is
		{"hello\\\nworld", "hello\\\nworld"},
		{`\'\"`, `\'\"`},
		{`\b\f\t\n\r\v\a`, `\b\f\t\n\r\v\a`},
		{`\000\001\376\377`, `\000\001\376\377`},
		{`\x00\x01\x7f\x80\xfe\xff`, `\x00\x01\x7f\x80\xfe\xff`},
	})
}
