package ogórek

import (
	"strconv"
	"unicode/utf8"
)

// pyquote, similarly to strconv.Quote, quotes s with " but does not use "\u" and "\U" inside.
//
// We need to avoid \u and friends, since for regular strings Python translates
// \u to \\u, not an UTF-8 character.
//
// Dumping strings in a way that is possible to copy/paste into Python and use
// pickletools.dis and pickle.loads there to verify a pickle is handy.
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
