package ogórek

import (
	"fmt"
	"strconv"
	"unicode/utf8"
)

// pyquote, similarly to strconv.Quote, quotes s with " but does not use "\u" and "\U" inside.
//
// We need to avoid \u and friends, since for regular strings Python translates
// \u to \\u, not an UTF-8 character.
//
// We must use Python - not Go - quoting, when emitting text strings with
// STRING opcode.
//
// Dumping strings in a way that is possible to copy/paste into Python and use
// pickletools.dis and pickle.loads there to verify a pickle is also handy.
func pyquote(s string) string {
	const hexdigits = "0123456789abcdef"
	out := make([]byte, 0, len(s))

	for {
		r, width := utf8.DecodeRuneInString(s)
		if width == 0 {
			break
		}

		emitRaw := false

		switch {
		// invalid & everything else goes in numeric byte escapes
		case r == utf8.RuneError:
			fallthrough
		default:
			emitRaw = true

		case r == '\\' || r == '"':
			out = append(out, '\\', byte(r))

		case strconv.IsPrint(r):
			out = append(out, s[:width]...)

		case r < ' ':
			rq := strconv.QuoteRune(r) // e.g. "'\n'"
			rq = rq[1:len(rq)-1]       // ->   `\n`
			out = append(out, rq...)
		}

		if emitRaw {
			for i := 0; i < width; i++ {
				out = append(out, '\\', 'x', hexdigits[s[i]>>4], hexdigits[s[i]&0xf])
			}
		}

		s = s[width:]
	}


	return "\"" + string(out) + "\""
}

// pydecodeStringEscape decodes input according to "string-escape" Python codec.
//
// The codec is essentially defined here:
// https://github.com/python/cpython/blob/v2.7.15-198-g69d0bc1430d/Objects/stringobject.c#L600
func pydecodeStringEscape(s string) (string, error) {
	out := make([]byte, 0, len(s))

loop:
	for {
		r, width := utf8.DecodeRuneInString(s)
		if width == 0 {
			break
		}

		// regular UTF-8 character
		if r != '\\' {
			out = append(out, s[:width]...)
			s = s[width:]
			continue
		}

		if len(s) < 2 {
			return "", strconv.ErrSyntax
		}

		switch c := s[1]; c {
		// \ LF -> just skip
		case '\n':
			s = s[2:]
			continue loop

		// \\ -> \
		case '\\':
			out = append(out, '\\')
			s = s[2:]
			continue loop

		// \' \"  (yes, both quotes are allowed to be escaped).
		//
		// also: both quotes are allowed to be _unescaped_ - e.g. Python
		// unpickles "S'hel'lo'\n." as "hel'lo".
		case '\'', '"':
			out = append(out, c)
			s = s[2:]
			continue loop

		// \c (any character without special meaning) -> \ and proceed with C
		default:
			out = append(out, '\\')
			s = s[1:] // not skipping c
			continue loop

		// escapes we handle (NOTE no \u \U for strings)
		case 'b','f','t','n','r','v','a':     // control characters
		case '0','1','2','3','4','5','6','7': // octals
	        case 'x':                             // hex
		}

		// s starts with a good/known string escape prefix -> reuse unquoteChar.
		r, _, tail, err := strconv.UnquoteChar(s, 0)
		if err != nil {
			return "", err
		}

		// all above escapes must produce single byte. This way we can
		// append it directly, not play rune -> string UTF-8 encoding
		// games (which break on e.g. "\x80" -> "\u0080" (= "\xc2x80").
		c := byte(r)
		if r != rune(c) {
			panic(fmt.Sprintf("pydecode: string-escape: non-byte escaped rune %q (% x  ; from %q)",
				r, r, s))
		}

		out = append(out, c)
		s = tail
	}

	return string(out), nil
}
