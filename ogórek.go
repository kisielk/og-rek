// Package ogórek is a library for decoding Python's pickle format.
//
// ogórek is Polish for "pickle".
package ogórek

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"math/big"
	"strconv"
)

// Opcodes
const (
	opMark           = "(" // push special markobject on stack
	opStop           = "." // every pickle ends with STOP
	opPop            = "0" // discard topmost stack item
	opPopMark        = "1" // discard stack top through topmost markobject
	opDup            = "2" // duplicate top stack item
	opFloat          = "F" // push float object; decimal string argument
	opInt            = "I" // push integer or bool; decimal string argument
	opBinint         = "J" // push four-byte signed int
	opBinint1        = "K" // push 1-byte unsigned int
	opLong           = "L" // push long; decimal string argument
	opBinint2        = "M" // push 2-byte unsigned int
	opNone           = "N" // push None
	opPersid         = "P" // push persistent object; id is taken from string arg
	opBinpersid      = "Q" //  "       "         "  ;  "  "   "     "  stack
	opReduce         = "R" // apply callable to argtuple, both on stack
	opString         = "S" // push string; NL-terminated string argument
	opBinstring      = "T" // push string; counted binary string argument
	opShortBinstring = "U" //  "     "   ;    "      "       "      " < 256 bytes
	opUnicode        = "V" // push Unicode string; raw-unicode-escaped"d argument
	opBinunicode     = "X" //   "     "       "  ; counted UTF-8 string argument
	opAppend         = "a" // append stack top to list below it
	opBuild          = "b" // call __setstate__ or __dict__.update()
	opGlobal         = "c" // push self.find_class(modname, name); 2 string args
	opDict           = "d" // build a dict from stack items
	opEmptyDict      = "}" // push empty dict
	opAppends        = "e" // extend list on stack by topmost stack slice
	opGet            = "g" // push item from memo on stack; index is string arg
	opBinget         = "h" //   "    "    "    "   "   "  ;   "    " 1-byte arg
	opInst           = "i" // build & push class instance
	opLongBinget     = "j" // push item from memo on stack; index is 4-byte arg
	opList           = "l" // build list from topmost stack items
	opEmptyList      = "]" // push empty list
	opObj            = "o" // build & push class instance
	opPut            = "p" // store stack top in memo; index is string arg
	opBinput         = "q" //   "     "    "   "   " ;   "    " 1-byte arg
	opLongBinput     = "r" //   "     "    "   "   " ;   "    " 4-byte arg
	opSetitem        = "s" // add key+value pair to dict
	opTuple          = "t" // build tuple from topmost stack items
	opEmptyTuple     = ")" // push empty tuple
	opSetitems       = "u" // modify dict by adding topmost key+value pairs
	opBinfloat       = "G" // push float; arg is 8-byte float encoding

	opTrue  = "I01\n" // not an opcode; see INT docs in pickletools.py
	opFalse = "I00\n" // not an opcode; see INT docs in pickletools.py

	// Protocol 2

	opProto    = "\x80" // identify pickle protocol
	opNewobj   = "\x81" // build object by applying cls.__new__ to argtuple
	opExt1     = "\x82" // push object from extension registry; 1-byte index
	opExt2     = "\x83" // ditto, but 2-byte index
	opExt4     = "\x84" // ditto, but 4-byte index
	opTuple1   = "\x85" // build 1-tuple from stack top
	opTuple2   = "\x86" // build 2-tuple from two topmost stack items
	opTuple3   = "\x87" // build 3-tuple from three topmost stack items
	opNewtrue  = "\x88" // push True
	opNewfalse = "\x89" // push False
	opLong1    = "\x8a" // push long from < 256 bytes
	opLong4    = "\x8b" // push really big long
)

// special marker
type mark struct{}

// None is a representation of Python's None.
type None struct{}

// Decoder is a decoder for pickle streams.
type Decoder struct {
	r     *bufio.Reader
	stack []interface{}
	memo  map[string]interface{}
}

// NewDecoder constructs a new Decoder which will decode the pickle stream in r.
func NewDecoder(r io.Reader) Decoder {
	reader := bufio.NewReader(r)
	return Decoder{reader, make([]interface{}, 0), make(map[string]interface{})}
}

// Decode decodes the pickle stream and returns the result or an error.
func (d Decoder) Decode() (interface{}, error) {
	for {
		key, err := d.r.ReadByte()
		if err == io.EOF {
			break
		} else if err != nil {
			return nil, err
		}

		switch string(key) {
		case opMark:
			d.mark()
		case opStop:
			break
		case opPop:
			d.pop()
		case opPopMark:
			d.popMark()
		case opDup:
			d.dup()
		case opFloat:
			err = d.loadFloat()
		case opInt:
			err = d.loadInt()
		case opBinint:
			d.loadBinInt()
		case opBinint1:
			d.loadBinInt1()
		case opLong:
			err = d.loadLong()
		case opBinint2:
			d.loadBinInt2()
		case opNone:
			d.loadNone()
		case opPersid:
			d.loadPersid()
		case opBinpersid:
			d.loadBinPersid()
		case opReduce:
			d.reduce()
		case opString:
			err = d.loadString()
		case opBinstring:
			d.loadBinString()
		case opShortBinstring:
			d.loadShortBinString()
		case opUnicode:
			err = d.loadUnicode()
		case opBinunicode:
			d.loadBinUnicode()
		case opAppend:
			d.loadAppend()
		case opBuild:
			d.build()
		case opGlobal:
			d.global()
		case opDict:
			d.loadDict()
		case opEmptyDict:
			d.loadEmptyDict()
		case opAppends:
			err = d.loadAppends()
		case opGet:
			d.get()
		case opBinget:
			d.binGet()
		case opInst:
			d.inst()
		case opLongBinget:
			d.longBinGet()
		case opList:
			d.loadList()
		case opEmptyList:
			d.append([]interface{}{})
		case opObj:
			d.obj()
		case opPut:
			err = d.loadPut()
		case opBinput:
			d.binPut()
		case opLongBinput:
			d.longBinPut()
		case opSetitem:
			err = d.loadSetItem()
		case opTuple:
			d.loadTuple()
		case opEmptyTuple:
			d.append([]interface{}{})
		case opSetitems:
			d.setItems()
		case opBinfloat:
			d.binFloat()
		default:
			return nil, fmt.Errorf("Unknown opcode: %q", key)
		}

		if err != nil {
			return nil, err
		}
	}
	return d.pop(), nil
}

// Push a marker
func (d *Decoder) mark() {
	d.append(mark{})
}

// Return the position of the topmost marker
func (d *Decoder) marker() int {
	m := mark{}
	var k int
	for k = len(d.stack) - 1; d.stack[k] != m && k > 0; k-- {
	}
	if k >= 0 {
		return k
	}
	panic("no marker in stack")
}

// Append a new value
func (d *Decoder) append(v interface{}) {
	d.stack = append(d.stack, v)
}

// Pop a value
func (d *Decoder) pop() interface{} {
	v := d.stack[len(d.stack)-1]
	d.stack = d.stack[:len(d.stack)-1]
	return v
}

// Discard the stack through to the topmost marker
func (d *Decoder) popMark() {

}

// Duplicate the top stack item
func (d *Decoder) dup() {
	d.stack = append(d.stack, d.stack[len(d.stack)-1])
}

// Push a float
func (d *Decoder) loadFloat() error {
	line, _, err := d.r.ReadLine()
	if err != nil {
		return err
	}
	v, err := strconv.ParseFloat(string(line), 64)
	if err != nil {
		return err
	}
	d.append(interface{}(v))
	return nil
}

// Push an int
func (d *Decoder) loadInt() error {
	line, _, err := d.r.ReadLine()
	if err != nil {
		return err
	}

	var val interface{}

	switch string(line) {
	case opFalse[1:3]:
		val = false
	case opTrue[1:3]:
		val = true
	default:
		i, err := strconv.ParseInt(string(line), 10, 64)
		if err != nil {
			return err
		}
		val = i
	}

	d.append(val)
	return nil
}

// Push a four-byte signed int
func (d *Decoder) loadBinInt() {
}

// Push a 1-byte unsigned int
func (d *Decoder) loadBinInt1() {
}

// Push a long
func (d *Decoder) loadLong() error {
	line, _, err := d.r.ReadLine()
	if err != nil {
		return err
	}
	v := new(big.Int)
	v.SetString(string(line[:len(line)-1]), 10)
	d.append(v)
	return nil
}

// Push a 2-byte unsigned int
func (d *Decoder) loadBinInt2() {

}

// Push None
func (d *Decoder) loadNone() {
	d.append(None{})
}

// Push a persistent object id
func (d *Decoder) loadPersid() {
}

// Push a persistent object id from items on the stack
func (d *Decoder) loadBinPersid() {
}

func (d *Decoder) reduce() {

}

func decodeStringEscape(b []byte) string {
	// TODO
	return string(b)
}

// Push a string
func (d *Decoder) loadString() error {
	line, _, err := d.r.ReadLine()
	if err != nil {
		return err
	}

	var delim byte
	switch line[0] {
	case '\'':
		delim = '\''
	case '"':
		delim = '"'
	default:
		return fmt.Errorf("invalid string delimiter: %s", line[0])
	}

	if line[len(line)-1] != delim {
		return fmt.Errorf("insecure string")
	}

	d.append(decodeStringEscape(line[1 : len(line)-1]))
	return nil
}

func (d *Decoder) loadBinString() {
}

func (d *Decoder) loadShortBinString() {

}

func (d *Decoder) loadUnicode() error {
	line, _, err := d.r.ReadLine()
	if err != nil {
		return err
	}
	sline := string(line)

	buf := bytes.Buffer{}

	for len(sline) >= 6 {
		var r rune
		var err error
		r, _, sline, err = strconv.UnquoteChar(sline, '\'')
		if err != nil {
			return err
		}
		_, err = buf.WriteRune(r)
		if err != nil {
			return err
		}
	}
	if len(sline) > 0 {
		return fmt.Errorf("characters remaining after loadUnicode operation: %s", sline)
	}

	d.append(buf.String())
	return nil
}

func (d *Decoder) loadBinUnicode() {

}

func (d *Decoder) loadAppend() error {
	v := d.pop()
	l := d.stack[len(d.stack)-1]
	switch l.(type) {
	case []interface{}:
		l := l.([]interface{})
		d.stack[len(d.stack)-1] = append(l, v)
	default:
		return fmt.Errorf("loadAppend expected a list, got %t", l)
	}
	return nil
}

func (d *Decoder) build() {

}

func (d *Decoder) global() {

}

func (d *Decoder) loadDict() {
	k := d.marker()
	m := make(map[interface{}]interface{}, 0)
	items := d.stack[k+1:]
	for i := 0; i < len(items); i += 2 {
		m[items[i]] = items[i+1]
	}
	d.stack = append(d.stack[:k], m)
}

func (d *Decoder) loadEmptyDict() {
	m := make(map[interface{}]interface{}, 0)
	d.append(m)
}

func (d *Decoder) loadAppends() error {
	k := d.marker()
	l := d.stack[len(d.stack)-1]
	switch l.(type) {
	case []interface{}:
		l := l.([]interface{})
		for _, v := range d.stack[k:len(d.stack)] {
			l = append(l, v)
		}
		d.stack = append(d.stack[:k], l)
	default:
		return fmt.Errorf("loadAppends expected a list, got %t", l)
	}
	return nil
}

func (d *Decoder) get() {

}

func (d *Decoder) binGet() {

}

func (d *Decoder) inst() {

}

func (d *Decoder) longBinGet() {

}

func (d *Decoder) loadList() {
	k := d.marker()
	v := append([]interface{}{}, d.stack[k+1:]...)
	d.stack = append(d.stack[:k], v)
}

func (d *Decoder) loadTuple() {
	k := d.marker()
	v := append([]interface{}{}, d.stack[k+1:]...)
	d.stack = append(d.stack[:k], v)
}

func (d *Decoder) obj() {

}

func (d *Decoder) loadPut() error {
	line, _, err := d.r.ReadLine()
	if err != nil {
		return err
	}
	d.memo[string(line)] = d.stack[len(d.stack)-1]
	return nil
}

func (d *Decoder) binPut() {

}

func (d *Decoder) longBinPut() {

}

func (d *Decoder) loadSetItem() error {
	v := d.pop()
	k := d.pop()
	m := d.stack[len(d.stack)-1]
	switch m.(type) {
	case map[interface{}]interface{}:
		m := m.(map[interface{}]interface{})
		m[k] = v
	default:
		return fmt.Errorf("loadSetItem expected a map, got %t", m)
	}
	return nil
}

func (d *Decoder) setItems() {

}

func (d *Decoder) binFloat() {

}
