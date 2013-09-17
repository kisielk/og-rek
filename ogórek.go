// Package ogórek is a library for handling Python pickles
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
	MARK            = "(" // push special markobject on stack
	STOP            = "." // every pickle ends with STOP
	POP             = "0" // discard topmost stack item
	POP_MARK        = "1" // discard stack top through topmost markobject
	DUP             = "2" // duplicate top stack item
	FLOAT           = "F" // push float object; decimal string argument
	INT             = "I" // push integer or bool; decimal string argument
	BININT          = "J" // push four-byte signed int
	BININT1         = "K" // push 1-byte unsigned int
	LONG            = "L" // push long; decimal string argument
	BININT2         = "M" // push 2-byte unsigned int
	NONE            = "N" // push None
	PERSID          = "P" // push persistent object; id is taken from string arg
	BINPERSID       = "Q" //  "       "         "  ;  "  "   "     "  stack
	REDUCE          = "R" // apply callable to argtuple, both on stack
	STRING          = "S" // push string; NL-terminated string argument
	BINSTRING       = "T" // push string; counted binary string argument
	SHORT_BINSTRING = "U" //  "     "   ;    "      "       "      " < 256 bytes
	UNICODE         = "V" // push Unicode string; raw-unicode-escaped"d argument
	BINUNICODE      = "X" //   "     "       "  ; counted UTF-8 string argument
	APPEND          = "a" // append stack top to list below it
	BUILD           = "b" // call __setstate__ or __dict__.update()
	GLOBAL          = "c" // push self.find_class(modname, name); 2 string args
	DICT            = "d" // build a dict from stack items
	EMPTY_DICT      = "}" // push empty dict
	APPENDS         = "e" // extend list on stack by topmost stack slice
	GET             = "g" // push item from memo on stack; index is string arg
	BINGET          = "h" //   "    "    "    "   "   "  ;   "    " 1-byte arg
	INST            = "i" // build & push class instance
	LONG_BINGET     = "j" // push item from memo on stack; index is 4-byte arg
	LIST            = "l" // build list from topmost stack items
	EMPTY_LIST      = "]" // push empty list
	OBJ             = "o" // build & push class instance
	PUT             = "p" // store stack top in memo; index is string arg
	BINPUT          = "q" //   "     "    "   "   " ;   "    " 1-byte arg
	LONG_BINPUT     = "r" //   "     "    "   "   " ;   "    " 4-byte arg
	SETITEM         = "s" // add key+value pair to dict
	TUPLE           = "t" // build tuple from topmost stack items
	EMPTY_TUPLE     = ")" // push empty tuple
	SETITEMS        = "u" // modify dict by adding topmost key+value pairs
	BINFLOAT        = "G" // push float; arg is 8-byte float encoding

	TRUE  = "I01\n" // not an opcode; see INT docs in pickletools.py
	FALSE = "I00\n" // not an opcode; see INT docs in pickletools.py

	// Protocol 2

	PROTO    = "\x80" // identify pickle protocol
	NEWOBJ   = "\x81" // build object by applying cls.__new__ to argtuple
	EXT1     = "\x82" // push object from extension registry; 1-byte index
	EXT2     = "\x83" // ditto, but 2-byte index
	EXT4     = "\x84" // ditto, but 4-byte index
	TUPLE1   = "\x85" // build 1-tuple from stack top
	TUPLE2   = "\x86" // build 2-tuple from two topmost stack items
	TUPLE3   = "\x87" // build 3-tuple from three topmost stack items
	NEWTRUE  = "\x88" // push True
	NEWFALSE = "\x89" // push False
	LONG1    = "\x8a" // push long from < 256 bytes
	LONG4    = "\x8b" // push really big long
)

// special marker
type mark struct{}

// None is a representation of Python's None
type None struct{}

// Decoder is a decoder for pickle streams
type Decoder struct {
	r     *bufio.Reader
	stack []interface{}
	memo  map[string]interface{}
}

// NewDecoder constructs a new Decoder which will decode the pickle stream in r
func NewDecoder(r io.Reader) Decoder {
	reader := bufio.NewReader(r)
	return Decoder{reader, make([]interface{}, 0), make(map[string]interface{})}
}

// Decode decodes the pickele stream and returns the result or an error
func (d Decoder) Decode() (interface{}, error) {
	for {
		key, err := d.r.ReadByte()
		if err == io.EOF {
			break
		} else if err != nil {
			return nil, err
		}

		switch string(key) {
		case MARK:
			d.mark()
		case STOP:
			break
		case POP:
			d.pop()
		case POP_MARK:
			d.popMark()
		case DUP:
			d.dup()
		case FLOAT:
			err = d.loadFloat()
		case INT:
			err = d.loadInt()
		case BININT:
			d.loadBinInt()
		case BININT1:
			d.loadBinInt1()
		case LONG:
			err = d.loadLong()
		case BININT2:
			d.loadBinInt2()
		case NONE:
			d.append(None{})
		case PERSID:
			d.loadPersid()
		case BINPERSID:
			d.loadBinPersid()
		case REDUCE:
			d.reduce()
		case STRING:
			err = d.loadString()
		case BINSTRING:
			d.loadBinString()
		case SHORT_BINSTRING:
			d.loadShortBinString()
		case UNICODE:
			err = d.loadUnicode()
		case BINUNICODE:
			d.loadBinUnicode()
		case APPEND:
			d.loadAppend()
		case BUILD:
			d.build()
		case GLOBAL:
			d.global()
		case DICT:
			d.loadDict()
		case EMPTY_DICT:
			d.loadEmptyDict()
		case APPENDS:
			err = d.loadAppends()
		case GET:
			d.get()
		case BINGET:
			d.binGet()
		case INST:
			d.inst()
		case LONG_BINGET:
			d.longBinGet()
		case LIST:
			d.loadList()
		case EMPTY_LIST:
			d.append([]interface{}{})
		case OBJ:
			d.obj()
		case PUT:
			err = d.loadPut()
		case BINPUT:
			d.binPut()
		case LONG_BINPUT:
			d.longBinPut()
		case SETITEM:
			err = d.loadSetItem()
		case TUPLE:
			d.loadTuple()
		case EMPTY_TUPLE:
			d.append([]interface{}{})
		case SETITEMS:
			d.setItems()
		case BINFLOAT:
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
	case FALSE[1:3]:
		val = false
	case TRUE[1:3]:
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
