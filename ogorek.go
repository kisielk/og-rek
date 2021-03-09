package ogórek

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"math"
	"math/big"
	"strconv"
)

// Opcodes
const (
	// Protocol 0

	opMark           byte = '(' // push special markobject on stack
	opStop           byte = '.' // every pickle ends with STOP
	opPop            byte = '0' // discard topmost stack item
	opDup            byte = '2' // duplicate top stack item
	opFloat          byte = 'F' // push float object; decimal string argument
	opInt            byte = 'I' // push integer or bool; decimal string argument
	opLong           byte = 'L' // push long; decimal string argument
	opNone           byte = 'N' // push None
	opPersid         byte = 'P' // push persistent object; id is taken from string arg
	opReduce         byte = 'R' // apply callable to argtuple, both on stack
	opString         byte = 'S' // push string; NL-terminated string argument
	opUnicode        byte = 'V' // push Unicode string; raw-unicode-escaped"d argument
	opAppend         byte = 'a' // append stack top to list below it
	opBuild          byte = 'b' // call __setstate__ or __dict__.update()
	opGlobal         byte = 'c' // push self.find_class(modname, name); 2 string args
	opDict           byte = 'd' // build a dict from stack items
	opGet            byte = 'g' // push item from memo on stack; index is string arg
	opInst           byte = 'i' // build & push class instance
	opList           byte = 'l' // build list from topmost stack items
	opPut            byte = 'p' // store stack top in memo; index is string arg
	opSetitem        byte = 's' // add key+value pair to dict
	opTuple          byte = 't' // build tuple from topmost stack items

	opTrue  = "I01\n" // not an opcode; see INT docs in pickletools.py
	opFalse = "I00\n" // not an opcode; see INT docs in pickletools.py

	// Protocol 1

	opPopMark        byte = '1' // discard stack top through topmost markobject
	opBinint         byte = 'J' // push four-byte signed int
	opBinint1        byte = 'K' // push 1-byte unsigned int
	opBinint2        byte = 'M' // push 2-byte unsigned int
	opBinpersid      byte = 'Q' // push persistent object; id is taken from stack
	opBinstring      byte = 'T' // push string; counted binary string argument
	opShortBinstring byte = 'U' //  "     "   ;    "      "       "      " < 256 bytes
	opBinunicode     byte = 'X' // push Unicode string; counted UTF-8 string argument
	opAppends        byte = 'e' // extend list on stack by topmost stack slice
	opBinget         byte = 'h' // push item from memo on stack; index is 1-byte arg
	opLongBinget     byte = 'j' //  "    "    "    "    "   "  ;   "    " 4-byte arg
	opEmptyList      byte = ']' // push empty list
	opEmptyTuple     byte = ')' // push empty tuple
	opEmptyDict      byte = '}' // push empty dict
	opObj            byte = 'o' // build & push class instance
	opBinput         byte = 'q' // store stack top in memo; index is 1-byte arg
	opLongBinput     byte = 'r' //   "     "    "   "   " ;   "    " 4-byte arg
	opSetitems       byte = 'u' // modify dict by adding topmost key+value pairs
	opBinfloat       byte = 'G' // push float; arg is 8-byte float encoding

	// Protocol 2

	opProto    byte = '\x80' // identify pickle protocol
	opNewobj   byte = '\x81' // build object by applying cls.__new__ to argtuple
	opExt1     byte = '\x82' // push object from extension registry; 1-byte index
	opExt2     byte = '\x83' // ditto, but 2-byte index
	opExt4     byte = '\x84' // ditto, but 4-byte index
	opTuple1   byte = '\x85' // build 1-tuple from stack top
	opTuple2   byte = '\x86' // build 2-tuple from two topmost stack items
	opTuple3   byte = '\x87' // build 3-tuple from three topmost stack items
	opNewtrue  byte = '\x88' // push True
	opNewfalse byte = '\x89' // push False
	opLong1    byte = '\x8a' // push long from < 256 bytes
	opLong4    byte = '\x8b' // push really big long

	// Protocol 3

	opBinbytes      byte = 'B' // push a Python bytes object (len ule32; [len]data)
	opShortBinbytes byte = 'C' //  "     "      "      "     (len ule8; [len]data)

	// Protocol 4

	opShortBinUnicode byte = '\x8c' // push short string; UTF-8 length < 256 bytes
	opBinunicode8     byte = '\x8d' // push Unicode string (len ule64; [len]data)
	opBinbytes8       byte = '\x8e' // push a Python bytes object (len ule64; [len]data)
	opEmptySet        byte = '\x8f' // push empty set
	opAddItems        byte = '\x90' // add items to existing set
	opFrozenSet       byte = '\x91' // build a frozenset out of mark..top
	opNewobjEx        byte = '\x92' // build object: cls argv kw -> cls.__new__(*argv, **kw)
	opStackGlobal     byte = '\x93' // same as OpGlobal but using names on the stacks
	opMemoize         byte = '\x94' // store top of the stack in memo
	opFrame           byte = '\x95' // indicate the beginning of a new frame

	// Protocol 5

	opBytearray8     byte = '\x96' // push a Python bytearray object (len ule64; [len]data)
	opNextBuffer     byte = '\x97' // push next out-of-band buffer
	opReadOnlyBuffer byte = '\x98' // turn out-of-band buffer at stack top to be read-only
)

var errNotImplemented = errors.New("unimplemented opcode")
var ErrInvalidPickleVersion = errors.New("invalid pickle version")
var errNoMarker = errors.New("no marker in stack")
var errNoMarkUse = errors.New("pickle: MARK object cannot be exposed")
var errStackUnderflow = errors.New("pickle: stack underflow")

// OpcodeError is the error that Decode returns when it sees unknown pickle opcode.
type OpcodeError struct {
	Key byte
	Pos int
}

func (e OpcodeError) Error() string {
	return fmt.Sprintf("Unknown opcode %d (%c) at position %d: %q", e.Key, e.Key, e.Pos, e.Key)
}

// special marker
type mark struct{}

// None is a representation of Python's None.
type None struct{}

// Tuple is a representation of Python's tuple.
type Tuple []interface{}

// Bytes represents Python's bytes.
type Bytes string

// Decoder is a decoder for pickle streams.
type Decoder struct {
	r      *bufio.Reader
	config *DecoderConfig
	stack  []interface{}
	memo   map[string]interface{}

	// a reusable buffer that can be used by the various decoding functions
	// functions using this should call buf.Reset to clear the old contents
	buf bytes.Buffer

	// reusable buffer for readLine
	line  []byte

	// protocol version seen in last PROTO opcode; 0 by default.
	protocol int
}

// DecoderConfig allows to tune Decoder.
type DecoderConfig struct {
	// PersistentLoad, if !nil, will be used by decoder to handle persistent references.
	//
	// Whenever the decoder finds an object reference in the pickle stream
	// it will call PersistentLoad. If PersistentLoad returns !nil object
	// without error, the decoder will use that object instead of Ref in
	// the resulted built Go object.
	//
	// An example use-case for PersistentLoad is to transform persistent
	// references in a ZODB database of form (type, oid) tuple, into
	// equivalent-to-type Go ghost object, e.g. equivalent to zodb.BTree.
	//
	// See Ref documentation for more details.
	PersistentLoad func(ref Ref) (interface{}, error)
}

// NewDecoder constructs a new Decoder which will decode the pickle stream in r.
func NewDecoder(r io.Reader) *Decoder {
	return NewDecoderWithConfig(r, &DecoderConfig{})
}

// NewDecoderWithConfig is similar to NewDecoder, but allows specifying decoder configuration.
func NewDecoderWithConfig(r io.Reader, config *DecoderConfig) *Decoder {
	reader := bufio.NewReader(r)
	return &Decoder{
		r:        reader,
		config:   config,
		stack:    make([]interface{}, 0),
		memo:     make(map[string]interface{}),
		protocol: 0,
	}
}

// Decode decodes the pickle stream and returns the result or an error.
func (d *Decoder) Decode() (interface{}, error) {

	insn := 0
loop:
	for {
		key, err := d.r.ReadByte()
		if err != nil {
			if err == io.EOF && insn != 0 {
				err = io.ErrUnexpectedEOF
			}
			return nil, err
		}

		insn++

		switch key {
		case opMark:
			d.mark()
		case opStop:
			break loop
		case opPop:
			_, err = d.pop()
		case opPopMark:
			d.popMark()
		case opDup:
			err = d.dup()
		case opFloat:
			err = d.loadFloat()
		case opInt:
			err = d.loadInt()
		case opBinint:
			err = d.loadBinInt()
		case opBinint1:
			err = d.loadBinInt1()
		case opLong:
			err = d.loadLong()
		case opBinint2:
			err = d.loadBinInt2()
		case opNone:
			err = d.loadNone()
		case opPersid:
			err = d.loadPersid()
		case opBinpersid:
			err = d.loadBinPersid()
		case opReduce:
			err = d.reduce()
		case opString:
			err = d.loadString()
		case opBinstring:
			err = d.loadBinString()
		case opShortBinstring:
			err = d.loadShortBinString()
		case opUnicode:
			err = d.loadUnicode()
		case opBinunicode:
			err = d.loadBinUnicode()
		case opAppend:
			err = d.loadAppend()
		case opBuild:
			err = d.build()
		case opGlobal:
			err = d.global()
		case opDict:
			err = d.loadDict()
		case opEmptyDict:
			err = d.loadEmptyDict()
		case opAppends:
			err = d.loadAppends()
		case opGet:
			err = d.get()
		case opBinget:
			err = d.binGet()
		case opInst:
			err = d.inst()
		case opLong1:
			err = d.loadLong1()
		case opNewfalse:
			err = d.loadBool(false)
		case opNewtrue:
			err = d.loadBool(true)
		case opLongBinget:
			err = d.longBinGet()
		case opList:
			err = d.loadList()
		case opEmptyList:
			d.push([]interface{}{})
		case opObj:
			err = d.obj()
		case opPut:
			err = d.loadPut()
		case opBinput:
			err = d.binPut()
		case opLongBinput:
			err = d.longBinPut()
		case opSetitem:
			err = d.loadSetItem()
		case opTuple:
			err = d.loadTuple()
		case opTuple1:
			err = d.loadTuple1()
		case opTuple2:
			err = d.loadTuple2()
		case opTuple3:
			err = d.loadTuple3()
		case opEmptyTuple:
			d.push(Tuple{})
		case opSetitems:
			err = d.loadSetItems()
		case opBinfloat:
			err = d.binFloat()
		case opBinbytes:
			err = d.loadBinBytes()
		case opShortBinbytes:
			err = d.loadShortBinBytes()
		case opFrame:
			err = d.loadFrame()
		case opShortBinUnicode:
			err = d.loadShortBinUnicode()
		case opStackGlobal:
			err = d.stackGlobal()
		case opMemoize:
			err = d.loadMemoize()
		case opBytearray8:
			err = d.loadBytearray8()
		case opNextBuffer:
			err = d.loadNextBuffer()
		case opReadOnlyBuffer:
			err = d.readOnlyBuffer()
		case opProto:
			var v byte
			v, err = d.r.ReadByte()
			if err == nil && !(0 <= v && v <= 5) {
				// We support protocol opcodes for up to protocol 5.
				//
				// The PROTO opcode documentation says protocol version must be in [2, 256).
				// However CPython also loads PROTO with version 0 and 1 without error.
				// So we allow all supported versions as PROTO argument.
				err = ErrInvalidPickleVersion
			}
			if err == nil {
				d.protocol = int(v)
			}

		default:
			return nil, OpcodeError{key, insn}
		}

		if err != nil {
			if err == errNotImplemented {
				return nil, OpcodeError{key, insn}
			}
			// EOF from individual opcode decoder is unexpected end of stream
			if err == io.EOF {
				err = io.ErrUnexpectedEOF
			}
			return nil, err
		}
	}

	return d.popUser()
}

// readLine reads next line from pickle stream.
//
// returned line does not contain \n.
// returned line is valid only till next call to readLine.
func (d *Decoder) readLine() ([]byte, error) {
	var (
		data     []byte
		err      error
	)
	d.line = d.line[:0]
	for {
		data, err = d.r.ReadSlice('\n')
		d.line = append(d.line, data...)

		// either have read till \n or got another error
		if err != bufio.ErrBufferFull {
			break
		}
	}

	// trim trailing \n
	if l := len(d.line); l > 0 && d.line[l-1] == '\n' {
		d.line = d.line[:l-1]
	}

	return d.line, err
}

// userOK tells whether it is ok to return all objects to user.
//
// for example it is not ok to return the mark object.
func userOK(objv ...interface{}) error {
	for _, obj := range objv {
		switch obj.(type) {
		case mark:
			return errNoMarkUse
		}
	}

	return nil
}

// Push a marker
func (d *Decoder) mark() {
	d.push(mark{})
}

// Return the position of the topmost marker
func (d *Decoder) marker() (int, error) {
	m := mark{}
	for k := len(d.stack) - 1; k >= 0; k-- {
		if d.stack[k] == m {
			return k, nil
		}
	}
	return 0, errNoMarker
}

// Append a new value
func (d *Decoder) push(v interface{}) {
	d.stack = append(d.stack, v)
}

// Pop a value
// The returned error is errStackUnderflow if decoder stack is empty
func (d *Decoder) pop() (interface{}, error) {
	ln := len(d.stack) - 1
	if ln < 0 {
		return nil, errStackUnderflow
	}
	v := d.stack[ln]
	d.stack = d.stack[:ln]
	return v, nil
}

// Pop a value (when you know for sure decoder stack is not empty)
func (d *Decoder) xpop() interface{} {
	v, err := d.pop()
	if err != nil {
		panic(err)
	}
	return v
}

// popUser pops stack value and checks whether it is ok to return to user.
func (d *Decoder) popUser() (interface{}, error) {
	v, err := d.pop()
	if err != nil {
		return nil, err
	}
	if err := userOK(v); err != nil {
		return nil, err
	}
	return v, nil
}

// Discard the stack through to the topmost marker
func (d *Decoder) popMark() error {
	return errNotImplemented
}

// Duplicate the top stack item
func (d *Decoder) dup() error {
	if len(d.stack) < 1 {
		return errStackUnderflow
	}
	d.stack = append(d.stack, d.stack[len(d.stack)-1])
	return nil
}

// Push a float
func (d *Decoder) loadFloat() error {
	line, err := d.readLine()
	if err != nil {
		return err
	}
	v, err := strconv.ParseFloat(string(line), 64)
	if err != nil {
		return err
	}
	d.push(v)
	return nil
}

// Push an int
func (d *Decoder) loadInt() error {
	line, err := d.readLine()
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

	d.push(val)
	return nil
}

// Push a four-byte signed int
func (d *Decoder) loadBinInt() error {
	var b [4]byte
	_, err := io.ReadFull(d.r, b[:])
	if err != nil {
		return err
	}
	v := binary.LittleEndian.Uint32(b[:])
	d.push(int64(int32(v))) // NOTE signed: uint32 -> int32, and only then -> int64
	return nil
}

// Push a 1-byte unsigned int
func (d *Decoder) loadBinInt1() error {
	b, err := d.r.ReadByte()
	if err != nil {
		return err
	}
	d.push(int64(b))
	return nil
}

// Push a long
func (d *Decoder) loadLong() error {
	line, err := d.readLine()
	if err != nil {
		return err
	}
	l := len(line)
	if l < 1 || line[l-1] != 'L' {
		return io.ErrUnexpectedEOF
	}
	v := new(big.Int)
	_, ok := v.SetString(string(line[:l-1]), 10)
	if !ok {
		return fmt.Errorf("pickle: loadLong: invalid string")
	}
	d.push(v)
	return nil
}

// Push a long1
func (d *Decoder) loadLong1() error {
	rawNum := []byte{}
	b, err := d.r.ReadByte()
	if err != nil {
		return err
	}
	length, err := decodeLong(string(b))
	if err != nil {
		return err
	}
	for i := 0; int64(i) < length.Int64(); i++ {
		b2, err := d.r.ReadByte()
		if err != nil {
			return err
		}
		rawNum = append(rawNum, b2)
	}
	decodedNum, err := decodeLong(string(rawNum))
	d.push(decodedNum)
	return nil
}

// Push a 2-byte unsigned int
func (d *Decoder) loadBinInt2() error {
	var b [2]byte
	_, err := io.ReadFull(d.r, b[:])
	if err != nil {
		return err
	}
	v := binary.LittleEndian.Uint16(b[:])
	d.push(int64(v))
	return nil
}

// Push None
func (d *Decoder) loadNone() error {
	d.push(None{})
	return nil
}


// Ref is the default representation for a Python persistent reference.
//
// Such references are used when one pickle somehow references another pickle
// in e.g. a database.
//
// See https://docs.python.org/3/library/pickle.html#pickle-persistent for details.
//
// See DecoderConfig.PersistentLoad and EncoderConfig.PersistentRef for ways to
// tune Decoder and Encoder to handle persistent references with user-specified
// application logic.
type Ref struct {
	// persistent ID of referenced object.
	//
	// used to be string for protocol 0, but "upgraded" to be arbitrary
	// object for later protocols.
	Pid interface{}
}

// Push a persistent object id
func (d *Decoder) loadPersid() error {
	pid, err := d.readLine()
	if err != nil {
		return err
	}

	return d.handleRef(Ref{Pid: string(pid)})
}

// Push a persistent object id from items on the stack
func (d *Decoder) loadBinPersid() error {
	pid, err := d.popUser()
	if err != nil {
		return err
	}
	return d.handleRef(Ref{Pid: pid})
}

// handleRef is common place to handle Refs.
func (d *Decoder) handleRef(ref Ref) error {
	if load := d.config.PersistentLoad; load != nil {
		obj, err := load(ref)
		if err != nil {
			return fmt.Errorf("pickle: handleRef: %s", err)
		}
		if obj == nil {
			// PersistentLoad asked to leave the reference as is.
			obj = ref
		}
		d.push(obj)
	} else {
		d.push(ref)
	}
	return nil
}

// Call represents Python's call.
type Call struct {
	Callable Class
	Args     Tuple
}

func (d *Decoder) reduce() error {
	if len(d.stack) < 2 {
		return errStackUnderflow
	}
	xargs := d.xpop()
	xclass := d.xpop()
	args, ok := xargs.(Tuple)
	if !ok {
		return fmt.Errorf("pickle: reduce: invalid args: %T", xargs)
	}
	class, ok := xclass.(Class)
	if !ok {
		return fmt.Errorf("pickle: reduce: invalid class: %T", xclass)
	}

	// try to handle the call.
	// If the call is unknown - represent it symbolically with Call{...} .
	err := d.handleCall(class, args)
	if err == errCallNotHandled {
		d.push(Call{Callable: class, Args: args})
		err = nil
	}

	return err
}

// errCallNotHandled is internal error via which handleCall signals that it did
// not handled the call.
var errCallNotHandled = errors.New("handleCall: call not handled")

// handleCall translates known python calls to appropriate Go objects.
//
// for example _codecs.encode(..., 'latin1') is handled as conversion to []byte.
func (d *Decoder) handleCall(class Class, argv Tuple) error {
	// for protocols <= 2 Python3 encodes bytes as `_codecs.encode(byt.decode('latin1'), 'latin1')`
	if class.Module == "_codecs" && class.Name == "encode" &&
		len(argv) == 2 && argv[1] == "latin1" {

		// bytes as latin1-decoded unicode
		data, err := decodeLatin1Bytes(argv[0])
		if err != nil {
			return fmt.Errorf("_codecs.encode: %s", err)
		}

		d.push(Bytes(data))
		return nil
	}

	// handle bytearray(...) -> []byte(...)
	if class == pybuiltin(d.protocol, "bytearray") {
		// bytearray(bytes(...))
		if len(argv) == 1 {
			data, ok := argv[0].(Bytes)
			if !ok {
				return fmt.Errorf("bytearray: want (bytes,)  ; got (%T,)", argv[0])
			}

			d.push([]byte(data))
			return nil
		}

		// bytearray(unicode, encoding)
		if len(argv) == 2 && argv[1] == "latin-1" {
			// bytes as latin1-decode unicode
			data, err := decodeLatin1Bytes(argv[0])
			if err != nil {
				return fmt.Errorf("bytearray: %s", err)
			}

			d.push([]byte(data))
			return nil
		}
	}

	return errCallNotHandled
}

// Push a string
func (d *Decoder) loadString() error {
	line, err := d.readLine()
	if err != nil {
		return err
	}

	if len(line) < 2 {
		return io.ErrUnexpectedEOF
	}

	var delim byte
	switch line[0] {
	case '\'':
		delim = '\''
	case '"':
		delim = '"'
	default:
		return fmt.Errorf("invalid string delimiter: %c", line[0])
	}

	if line[len(line)-1] != delim {
		return io.ErrUnexpectedEOF
	}

	s, err := pydecodeStringEscape(string(line[1 : len(line)-1]))
	if err != nil {
		return err
	}

	d.push(s)
	return nil
}

// bufLoadBinData4 decodes `len(LE32) [len]data` into d.buf .
// it serves loadBin{String,Bytes}.
func (d *Decoder) bufLoadBinData4() error {
	var b [4]byte
	_, err := io.ReadFull(d.r, b[:])
	if err != nil {
		return err
	}
	v := binary.LittleEndian.Uint32(b[:])
	return d.bufLoadBytesData(uint64(v))
}

// bufLoadBinData8 decodes `len(LE64) [len]data into d.buf .
// it serves loadBytearray8 (and TODO loadBinBytes8, loadBinUnicode8)
func (d *Decoder) bufLoadBinData8() error {
	var b [8]byte
	_, err := io.ReadFull(d.r, b[:])
	if err != nil {
		return err
	}
	v := binary.LittleEndian.Uint64(b[:])
	return d.bufLoadBytesData(v)
}

// bufLoadBytesData fetches [lel]data into d.buf.
// it serves bufloadBinBytes{4,8}
func (d *Decoder) bufLoadBytesData(l uint64) error {
	d.buf.Reset()
	// don't allow malicious `BINSTRING <bigsize> nodata` to make us out of memory
	prealloc := int(l)
	if maxgrow := 0x10000; prealloc > maxgrow {
		prealloc = maxgrow
	}
	d.buf.Grow(prealloc)
	if l > math.MaxInt64 {
		return fmt.Errorf("size([]data) > maxint64")
	}
	_, err := io.CopyN(&d.buf, d.r, int64(l))
	if err != nil {
		return err
	}
	return nil
}

func (d *Decoder) loadBinString() error {
	err := d.bufLoadBinData4()
	if err != nil {
		return err
	}
	d.push(d.buf.String())
	return nil
}

func (d *Decoder) loadBinBytes() error {
	err := d.bufLoadBinData4()
	if err != nil {
		return err
	}
	d.push(Bytes(d.buf.Bytes()))
	return nil
}

// bufLoadShortBinBytes decodes `len(U8) [len]data` into d.buf .
// it serves loadShortBin{String,Bytes} .
func (d *Decoder) bufLoadShortBinBytes() error {
	b, err := d.r.ReadByte()
	if err != nil {
		return err
	}

	d.buf.Reset()
	d.buf.Grow(int(b))
	_, err = io.CopyN(&d.buf, d.r, int64(b))
	if err != nil {
		return err
	}
	return nil
}

func (d *Decoder) loadShortBinString() error {
	err := d.bufLoadShortBinBytes()
	if err != nil {
		return err
	}
	d.push(d.buf.String())
	return nil
}

func (d *Decoder) loadShortBinBytes() error {
	err := d.bufLoadShortBinBytes()
	if err != nil {
		return err
	}
	d.push(Bytes(d.buf.Bytes()))
	return nil
}

func (d *Decoder) loadUnicode() error {
	line, err := d.readLine()
	if err != nil {
		return err
	}

	text, err := pydecodeRawUnicodeEscape(string(line))
	if err != nil {
		return err
	}

	d.push(text)
	return nil
}

func (d *Decoder) loadBinUnicode() error {
	var length int32
	for i := 0; i < 4; i++ {
		t, err := d.r.ReadByte()
		if err != nil {
			return err
		}
		length = length | (int32(t) << uint(8*i))
	}
	rawB := []byte{}
	for z := 0; int32(z) < length; z++ {
		n, err := d.r.ReadByte()
		if err != nil {
			return err
		}
		rawB = append(rawB, n)
	}
	d.push(string(rawB))
	return nil
}

func (d *Decoder) loadAppend() error {
	if len(d.stack) < 2 {
		return errStackUnderflow
	}
	v := d.xpop()
	l := d.stack[len(d.stack)-1]
	if err := userOK(v); err != nil {
		return err
	}
	switch l.(type) {
	case []interface{}:
		l := l.([]interface{})
		d.stack[len(d.stack)-1] = append(l, v)
	default:
		return fmt.Errorf("pickle: loadAppend: expected a list, got %T", l)
	}
	return nil
}

func (d *Decoder) build() error {
	return errNotImplemented
}

// Class represents a Python class.
type Class struct {
	Module, Name string
}

func (d *Decoder) global() error {
	module, err := d.readLine()
	if err != nil {
		return err
	}
	smodule := string(module)
	name, err := d.readLine()
	if err != nil {
		return err
	}
	sname := string(name)
	d.push(Class{Module: smodule, Name: sname})
	return nil
}

// mapTryAssign tries to do `m[key] = value`.
//
// It checks whether key is of appropriate type, and if yes - succeeds.
// If key is not appropriate - the map stays unchanged and false is returned.
func mapTryAssign(m map[interface{}]interface{}, key, value interface{}) (ok bool) {
	// use panic/recover to detect inappropriate keys.
	//
	// We could try to use reflect.TypeOf(key).Comparable() instead, but that
	// is not generally enough: with Comparable, key type structure has to
	// be manually walked recursively and each subfield checked for
	// comparability. -> panic/recover is simpler to use instead.
	//
	// See https://github.com/kisielk/og-rek/issues/30#issuecomment-423803200
	// for details.
	defer func() {
		// on invalid dynamic key type runtime panics like below:
		//
		//	`panic: runtime error: hash of unhashable type []interface {}`
		//
		// we don't try to detect the exact message as mapTryAssign does
		// only 1 operation.
		if r := recover(); r != nil {
			ok = false
		}
	}()

	m[key] = value
	ok = true
	return
}

func (d *Decoder) loadDict() error {
	k, err := d.marker()
	if err != nil {
		return err
	}

	m := make(map[interface{}]interface{}, 0)
	items := d.stack[k+1:]
	if len(items) % 2 != 0 {
		return fmt.Errorf("pickle: loadDict: odd # of elements")
	}
	for i := 0; i < len(items); i += 2 {
		key := items[i]
		if !mapTryAssign(m, key, items[i+1]) {
			return fmt.Errorf("pickle: loadDict: invalid key type %T", key)
		}
	}
	d.stack = append(d.stack[:k], m)
	return nil
}

func (d *Decoder) loadEmptyDict() error {
	m := make(map[interface{}]interface{}, 0)
	d.push(m)
	return nil
}

func (d *Decoder) loadAppends() error {
	k, err := d.marker()
	if err != nil {
		return err
	}
	if k < 1 {
		return errStackUnderflow
	}

	l := d.stack[k-1]
	switch l.(type) {
	case []interface{}:
		l := l.([]interface{})
		for _, v := range d.stack[k+1 : len(d.stack)] {
			l = append(l, v)
		}
		d.stack = append(d.stack[:k-1], l)
	default:
		return fmt.Errorf("pickle: loadAppends: expected a list, got %T", l)
	}
	return nil
}

func (d *Decoder) get() error {
	line, err := d.readLine()
	if err != nil {
		return err
	}
	v, ok := d.memo[string(line)]
	if !ok {
		return fmt.Errorf("pickle: memo: key error %q", line)
	}
	d.push(v)
	return nil
}

func (d *Decoder) binGet() error {
	b, err := d.r.ReadByte()
	if err != nil {
		return err
	}

	v, ok := d.memo[strconv.Itoa(int(b))]
	if !ok {
		return fmt.Errorf("pickle: memo: key error %d", b)
	}
	d.push(v)
	return nil
}

func (d *Decoder) inst() error {
	return errNotImplemented
}

func (d *Decoder) longBinGet() error {
	var b [4]byte
	_, err := io.ReadFull(d.r, b[:])
	if err != nil {
		return err
	}
	v := binary.LittleEndian.Uint32(b[:])
	vv, ok := d.memo[strconv.Itoa(int(v))]
	if !ok {
		return fmt.Errorf("pickle: memo: key error %d", v)
	}
	d.push(vv)
	return nil
}

func (d *Decoder) loadBool(b bool) error {
	d.push(b)
	return nil
}

func (d *Decoder) loadList() error {
	k, err := d.marker()
	if err != nil {
		return err
	}

	v := append([]interface{}{}, d.stack[k+1:]...)
	d.stack = append(d.stack[:k], v)
	return nil
}

func (d *Decoder) loadTuple() error {
	k, err := d.marker()
	if err != nil {
		return err
	}

	v := append(Tuple{}, d.stack[k+1:]...)
	d.stack = append(d.stack[:k], v)
	return nil
}

// tupleN(n) creates tuple from top n stack objects.
// it serves TUPLE{1,2,3} opcode handlers.
func (d *Decoder) tupleN(n int) error {
	if len(d.stack) < n {
		return errStackUnderflow
	}
	k := len(d.stack) - n
	if err := userOK(d.stack[k:]...); err != nil {
		return err
	}
	v := append(Tuple{}, d.stack[k:]...)
	d.stack = append(d.stack[:k], v)
	return nil
}

func (d *Decoder) loadTuple1() error {
	return d.tupleN(1)
}

func (d *Decoder) loadTuple2() error {
	return d.tupleN(2)
}

func (d *Decoder) loadTuple3() error {
	return d.tupleN(3)
}

func (d *Decoder) obj() error {
	return errNotImplemented
}

// memoTop puts top of the stack into memo[key]; the stack is not changed.
// it is the worker for handling PUT, BINPUT, ... opcodes
func (d *Decoder) memoTop(key string) error {
	if len(d.stack) < 1 {
		return errStackUnderflow
	}

	obj := d.stack[len(d.stack)-1]
	if err := userOK(obj); err != nil {
		return err
	}

	d.memo[key] = obj
	return nil
}

func (d *Decoder) loadPut() error {
	line, err := d.readLine()
	if err != nil {
		return err
	}
	return d.memoTop(string(line))
}

func (d *Decoder) binPut() error {
	b, err := d.r.ReadByte()
	if err != nil {
		return err
	}
	return d.memoTop(strconv.Itoa(int(b)))
}

func (d *Decoder) longBinPut() error {
	var b [4]byte
	_, err := io.ReadFull(d.r, b[:])
	if err != nil {
		return err
	}
	v := binary.LittleEndian.Uint32(b[:])
	return d.memoTop(strconv.Itoa(int(v)))
}

func (d *Decoder) loadSetItem() error {
	if len(d.stack) < 3 {
		return errStackUnderflow
	}
	v := d.xpop()
	k := d.xpop()
	if err := userOK(k, v); err != nil {
		return err
	}
	m := d.stack[len(d.stack)-1]
	switch m := m.(type) {
	case map[interface{}]interface{}:
		if !mapTryAssign(m, k, v) {
			return fmt.Errorf("pickle: loadSetItem: invalid key type %T", k)
		}
	default:
		return fmt.Errorf("pickle: loadSetItem: expected a map, got %T", m)
	}
	return nil
}

func (d *Decoder) loadSetItems() error {
	k, err := d.marker()
	if err != nil {
		return err
	}
	if k < 1 {
		return errStackUnderflow
	}

	l := d.stack[k-1]
	switch m := l.(type) {
	case map[interface{}]interface{}:
		if (len(d.stack) - (k + 1)) % 2 != 0 {
			return fmt.Errorf("pickle: loadSetItems: odd # of elements")
		}
		for i := k + 1; i < len(d.stack); i += 2 {
			key := d.stack[i]
			if !mapTryAssign(m, key, d.stack[i+1]) {
				return fmt.Errorf("pickle: loadSetItems: invalid key type %T", key)
			}
		}
		d.stack = append(d.stack[:k-1], m)
	default:
		return fmt.Errorf("pickle: loadSetItems: expected a map, got %T", m)
	}
	return nil
}

func (d *Decoder) binFloat() error {
	var b [8]byte
	_, err := io.ReadFull(d.r, b[:])
	if err != nil {
		return err
	}
	u := binary.BigEndian.Uint64(b[:])
	d.push(math.Float64frombits(u))
	return nil
}

// loadFrame discards the framing opcode+information, this information is useful to do one large read (instead of many small reads)
// https://www.python.org/dev/peps/pep-3154/#framing
func (d *Decoder) loadFrame() error {
	var b [8]byte
	_, err := io.ReadFull(d.r, b[:])
	if err != nil {
		return err
	}
	return nil
}

func (d *Decoder) loadShortBinUnicode() error {
	b, err := d.r.ReadByte()
	if err != nil {
		return err
	}

	d.buf.Reset()
	d.buf.Grow(int(b))
	_, err = io.CopyN(&d.buf, d.r, int64(b))
	if err != nil {
		return err
	}
	d.push(d.buf.String())
	return nil
}

func (d *Decoder) stackGlobal() error {
	if len(d.stack) < 2 {
		return errStackUnderflow
	}
	xname := d.xpop()
	xmodule := d.xpop()

	name, ok := xname.(string)
	if !ok {
		return fmt.Errorf("pickle: stackGlobal: invalid name: %T", xname)
	}
	module, ok := xmodule.(string)
	if !ok {
		return fmt.Errorf("pickle: stackGlobal: invalid module: %T", xmodule)
	}

	d.push(Class{Module: module, Name: name})
	return nil
}

func (d *Decoder) loadMemoize() error {
	return d.memoTop(strconv.Itoa(len(d.memo)))
}

func (d *Decoder) loadBytearray8() error {
	err := d.bufLoadBinData8()
	if err != nil {
		return err
	}
	d.push(d.buf.Bytes())
	d.buf = bytes.Buffer{} // fully reset .buf to unalias just pushed []byte
	return nil
}

func (d *Decoder) loadNextBuffer() error {
	// TODO consider adding support for out-of-band data in the future
	return fmt.Errorf("next_buffer: no out-of-band data")
}

func (d *Decoder) readOnlyBuffer() error {
	// TODO consider adding support for out-of-band data in the future
	return fmt.Errorf("read_only_buffer: stack top is not buffer")
}

// unquoteChar is like strconv.UnquoteChar, but returns io.ErrUnexpectedEOF
// instead of strconv.ErrSyntax, when input is prematurely terminted.
//
// XXX remove if ever something like https://golang.org/cl/37052 is accepted.
func unquoteChar(s string, quote byte) (value rune, multibyte bool, tail string, err error) {
	// strconv.UnquoteChar("") panics before Go1.11
	if s == "" {
		return 0, false, "", io.ErrUnexpectedEOF
	}

	value, multibyte, tail, err = strconv.UnquoteChar(s, quote)
	if err == nil {
		return
	}

	// now we have to find out whether it was due to input cut.
	if len(s) > 10 { // \U12345678
		return
	}

	// + "0"*9 should make s valid if it was cut, e.g. "\U012" becomes "\U012000000000".
	// On the other hand, if s was invalid, e.g. "\Uz"
	// it will remain invaild even with the suffix.
	_, _, _, err2 := strconv.UnquoteChar(s + "000000000", quote)
	if err2 == nil {
		err = io.ErrUnexpectedEOF
	}

	return
}

// decodeLong takes a byte array of 2's compliment little-endian binary words and converts them
// to a big integer
func decodeLong(data string) (*big.Int, error) {
	decoded := big.NewInt(0)
	var negative bool
	switch x := len(data); {
	case x < 1:
		return decoded, nil
	case x > 1:
		if data[x-1] > 127 {
			negative = true
		}
		for i := x - 1; i >= 0; i-- {
			a := big.NewInt(int64(data[i]))
			for n := i; n > 0; n-- {
				a = a.Lsh(a, 8)
			}
			decoded = decoded.Add(a, decoded)
		}
	default:
		if data[0] > 127 {
			negative = true
		}
		decoded = big.NewInt(int64(data[0]))
	}

	if negative {
		// Subtract 1 from the number
		one := big.NewInt(1)
		decoded.Sub(decoded, one)

		// Flip the bits
		bytes := decoded.Bytes()
		for i := 0; i < len(bytes); i++ {
			bytes[i] = ^bytes[i]
		}
		decoded.SetBytes(bytes)

		// Mark as negative now conversion has been completed
		decoded.Neg(decoded)
	}
	return decoded, nil
}

// decodeLatin1Bytes tries to decode bytes from arg assuming it is latin1-encoded unicode.
//
// Python uses such representation of bytes for protocols <= 2 - where there is
// no BYTES* opcodes.
func decodeLatin1Bytes(arg interface{}) ([]byte, error) {
	// bytes as latin1-decoded unicode
	ulatin1, ok := arg.(string)
	if !ok {
		return nil, fmt.Errorf("latin1: arg must be string, not %T", arg)
	}

	data := make([]byte, 0, len(ulatin1))
	for _, r := range ulatin1 {
		if r >= 0x100 {
			return nil, fmt.Errorf("latin1: cannot encode %q", r)
		}

		data = append(data, byte(r))
	}

	return data, nil
}

// pybuiltin returns Class corresponding to Python builtin name.
func pybuiltin(protocol int, name string) Class {
	module := "builtins" // py3
	if protocol <= 2 {
		module = "__builtin__" // py2
	}

	return Class{Module: module, Name: name}
}
