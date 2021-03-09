package ogórek

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"math"
	"math/big"
	"reflect"
	"strings"
)

const highestProtocol = 5 // highest protocol version we support generating

// unicode is string that always encodes as unicode pickle object.
// (regular string encodes to unicode pickle object only for protocol >= 3)
type unicode string

type TypeError struct {
	typ string
}

func (te *TypeError) Error() string {
	return fmt.Sprintf("no support for type '%s'", te.typ)
}

// An Encoder encodes Go data structures into pickle byte stream
type Encoder struct {
	w      io.Writer
	config *EncoderConfig
}

// EncoderConfig allows to tune Encoder.
type EncoderConfig struct {
	// Protocol specifies which pickle protocol version should be used.
	Protocol int

	// PersistentRef, if !nil, will be used by encoder to encode objects as persistent references.
	//
	// Whenever the encoders sees pointer to a Go struct object, it will call
	// PersistentRef to find out how to encode that object. If PersistentRef
	// returns nil, the object is encoded regularly. If !nil - the object
	// will be encoded as an object reference.
	//
	// See Ref documentation for more details.
	PersistentRef func(obj interface{}) *Ref
}

// NewEncoder returns a new Encoder struct with default values
func NewEncoder(w io.Writer) *Encoder {
	return NewEncoderWithConfig(w, &EncoderConfig{
		// allow both Python2 and Python3 to decode what ogórek produces by default
		Protocol: 2,
	})
}

// NewEncoderWithConfig is similar to NewEncoder, but allows specifying the encoder configuration.
func NewEncoderWithConfig(w io.Writer, config *EncoderConfig) *Encoder {
	return &Encoder{w: w, config: config}
}

// Encode writes the pickle encoding of v to w, the encoder's writer
func (e *Encoder) Encode(v interface{}) error {
	proto := e.config.Protocol
	if !(0 <= proto && proto <= highestProtocol) {
		return fmt.Errorf("pickle: encode: invalid protocol %d", proto)
	}
	// protocol >= 2  -> emit PROTO <protocol>
	if proto >= 2 {
		err := e.emit(opProto, byte(proto))
		if err != nil {
			return err
		}
	}

	rv := reflectValueOf(v)
	err := e.encode(rv)
	if err != nil {
		return err
	}
	return e.emit(opStop)
}

// emit writes byte vector into encoder output.
func (e *Encoder) emitb(b []byte) error {
	_, err := e.w.Write(b)
	return err
}

// emits writes string into encoder output.
func (e *Encoder) emits(s string) error {
	return e.emitb([]byte(s))
}

// emit writes byte arguments into encoder output.
func (e *Encoder) emit(bv ...byte) error {
	return e.emitb(bv)
}

// emitf writes formatted string into encoder output.
func (e *Encoder) emitf(format string, argv ...interface{}) error {
	_, err := fmt.Fprintf(e.w, format, argv...)
	return err
}

func (e *Encoder) encode(rv reflect.Value) error {

	switch rk := rv.Kind(); rk {

	case reflect.Bool:
		return e.encodeBool(rv.Bool())
	case reflect.Int, reflect.Int8, reflect.Int64, reflect.Int32, reflect.Int16:
		return e.encodeInt(reflect.Int, rv.Int())
	case reflect.Uint8, reflect.Uint64, reflect.Uint, reflect.Uint32, reflect.Uint16:
		return e.encodeInt(reflect.Uint, int64(rv.Uint()))
	case reflect.String:
		switch rv.Interface().(type) {
		case unicode:
			return e.encodeUnicode(rv.String())
		case Bytes:
			return e.encodeBytes(Bytes(rv.String()))
		default:
			return e.encodeString(rv.String())
		}
	case reflect.Array, reflect.Slice:
		if rv.Type().Elem().Kind() == reflect.Uint8 {
			return e.encodeByteArray(rv.Bytes())
		} else if _, ok := rv.Interface().(Tuple); ok {
			return e.encodeTuple(rv.Interface().(Tuple))
		} else {
			return e.encodeArray(rv)
		}

	case reflect.Map:
		return e.encodeMap(rv)

	case reflect.Struct:
		return e.encodeStruct(rv)

	case reflect.Float32, reflect.Float64:
		return e.encodeFloat(float64(rv.Float()))

	case reflect.Interface:
		// recurse until we get a concrete type
		// could be optmized into a tail call
		return e.encode(rv.Elem())

	case reflect.Ptr:

		if rv.Elem().Kind() == reflect.Struct {
			// check if we have to encode this object as persistent reference.
			if getref := e.config.PersistentRef; getref != nil {
				ref := getref(rv.Interface())
				if ref != nil {
					return e.encodeRef(ref)
				}
			}

			switch rv.Elem().Interface().(type) {
			case None:
				return e.encodeStruct(rv.Elem())
			}
		}

		return e.encode(rv.Elem())

	case reflect.Invalid:
		return e.emit(opNone)
	default:
		return &TypeError{typ: rk.String()}
	}

	return nil
}

func (e *Encoder) encodeTuple(t Tuple) error {
	l := len(t)

	// protocol >= 2: [1-3]() -> TUPLE{1-3}
	if e.config.Protocol >= 2 && (1 <= l && l <= 3) {
		for i := range t {
			err := e.encode(reflectValueOf(t[i]))
			if err != nil {
				return err
			}
		}

		var op byte
		switch l {
		case 1:
			op = opTuple1
		case 2:
			op = opTuple2
		case 3:
			op = opTuple3
		}

		return e.emit(op)
	}

	// protocol >= 1: ø tuple -> EMPTY_TUPLE
	if e.config.Protocol >= 1 && l == 0 {
		return e.emit(opEmptyTuple)
	}

	// general case: MARK ... TUPLE
	// TODO detect cycles and double references to the same object
	err := e.emit(opMark)
	if err != nil {
		return err
	}

	for i := 0; i < l; i++ {
		err = e.encode(reflectValueOf(t[i]))
		if err != nil {
			return err
		}
	}

	return e.emit(opTuple)
}

func (e *Encoder) encodeArray(arr reflect.Value) error {

	l := arr.Len()

	// protocol >= 1: ø list -> EMPTY_LIST
	if e.config.Protocol >= 1 && l == 0 {
		return e.emit(opEmptyList)
	}

	// MARK + ... + LIST
	// TODO detect cycles and double references to the same object
	err := e.emit(opMark)
	if err != nil {
		return err
	}

	for i := 0; i < l; i++ {
		v := arr.Index(i)
		err = e.encode(v)
		if err != nil {
			return err
		}
	}

	return e.emit(opList)
}

func (e *Encoder) encodeBool(b bool) error {
	// protocol >= 2  ->  NEWTRUE/NEWFALSE
	if e.config.Protocol >= 2 {
		op := opNewfalse
		if b {
			op = opNewtrue
		}
		return e.emit(op)
	}

	// INT(01 | 00)
	var err error
	if b {
		err = e.emits(opTrue)
	} else {
		err = e.emits(opFalse)
	}

	return err
}

func (e *Encoder) encodeBytes(byt Bytes) error {
	l := len(byt)

	// protocol >= 3  ->  BINBYTES*
	if e.config.Protocol >= 3 {
		if l < 256 {
			err := e.emit(opShortBinbytes, byte(l))
			if err != nil {
				return err
			}
		} else {
			var b = [1+4]byte{opBinbytes}

			binary.LittleEndian.PutUint32(b[1:], uint32(l))
			err := e.emitb(b[:])
			if err != nil {
				return err
			}
		}

		return e.emits(string(byt))
	}

	// protocol 0..2 -> emit as `_codecs.encode(byt.decode('latin1'), 'latin1')`
	// (as python3 does)
	rlatin1 := make([]rune, len(byt))
	for i := 0; i < l; i++ {
		rlatin1[i] = rune(byt[i]) // decode as latin1
	}
	ulatin1 := unicode(rlatin1) // -> UTF8

	return e.encodeCall(&Call{
		Callable: Class{Module: "_codecs", Name: "encode"},
		Args:     Tuple{ulatin1, "latin1"},
	})
}

func (e *Encoder) encodeByteArray(bv []byte) error {
	// protocol >= 5  ->  BYTEARRAY8
	if e.config.Protocol >= 5 {
		var b = [1+8]byte{opBytearray8}

		binary.LittleEndian.PutUint64(b[1:], uint64(len(bv)))
		err := e.emitb(b[:])
		if err != nil {
			return err
		}
		return e.emitb(bv)
	}

	// TODO protocol <= 2: pickle can be shorter if we emit -> bytearray(unicode, encoding)
	// instead of bytearray(_codecs.encode(unicode, encoding))

	return e.encodeCall(&Call{
		Callable: pybuiltin(e.config.Protocol, "bytearray"),
		Args:     Tuple{Bytes(bv)},
	})
}

func (e *Encoder) encodeString(s string) error {
	// protocol >= 3 -> encode string as unicode object
	// (as python3 does)
	if e.config.Protocol >= 3 {
		return e.encodeUnicode(s)
	}

	l := len(s)

	// protocol >= 1  ->  BINSTRING*
	if e.config.Protocol >= 1 {
		if l < 256 {
			err := e.emit(opShortBinstring, byte(l))
			if err != nil {
				return err
			}
		} else {
			var b = [1+4]byte{opBinstring}

			binary.LittleEndian.PutUint32(b[1:], uint32(l))
			err := e.emitb(b[:])
			if err != nil {
				return err
			}
		}

		return e.emits(s)
	}

	// protocol 0: STRING
	// XXX Python uses both ' and " for quoting - we quote with " only.
	// XXX -> use https://godoc.org/lab.nexedi.com/kirr/go123/xfmt#AppendQuotePy ?
	//
	// don't use %q - that will use \u and \U in quoting which python won't
	// interpret when decoding string literals.
	return e.emitf("%c%s\n", opString, pyquote(s))
}

// encodeUnicode emits UTF-8 encoded string s as unicode pickle object.
func (e *Encoder) encodeUnicode(s string) error {
	// protocol >= 1  -> BINUNICODE*
	if e.config.Protocol >= 1 {
		l := len(s)

		// protocol >= 4  -> SHORT_BINUNICODE
		if l < 256 && e.config.Protocol >= 4 {
			err := e.emit(opShortBinUnicode, byte(l))
			if err != nil {
				return err
			}
		} else {
			var b = [1+4]byte{opBinunicode}

			binary.LittleEndian.PutUint32(b[1:], uint32(l))
			err := e.emitb(b[:])
			if err != nil {
				return err
			}
		}

		return e.emits(s)
	}

	// protocol 0: UNICODE
	return e.emitf("%c%s\n", opUnicode, pyencodeRawUnicodeEscape(s))
}

func (e *Encoder) encodeFloat(f float64) error {
	// protocol >= 1: BINFLOAT
	if e.config.Protocol >= 1 {
		u := math.Float64bits(f)
		var b = [1+8]byte{opBinfloat}
		binary.BigEndian.PutUint64(b[1:], u)
		return e.emitb(b[:])
	}

	// protocol 0: FLOAT
	return e.emitf("%c%g\n", opFloat, f)
}

func (e *Encoder) encodeInt(k reflect.Kind, i int64) error {
	// FIXME: need support for 64-bit ints

	// protocol >= 1: BININT*
	if e.config.Protocol >= 1 {
		switch {
		case i > 0 && i < math.MaxUint8:
			return e.emit(opBinint1, byte(i))

		case i > 0 && i < math.MaxUint16:
			return e.emit(opBinint2, byte(i), byte(i >> 8))

		case i >= math.MinInt32 && i <= math.MaxInt32:
			var b = [1+4]byte{opBinint}
			binary.LittleEndian.PutUint32(b[1:], uint32(i))
			return e.emitb(b[:])
		}
	}

	// protocol 0: INT
	return e.emitf("%c%d\n", opInt, i)
}

func (e *Encoder) encodeLong(b *big.Int) error {
	// TODO if e.protocol >= 2 use opLong1 & opLong4
	return e.emitf("%c%dL\n", opLong, b)
}

func (e *Encoder) encodeMap(m reflect.Value) error {

	keys := m.MapKeys()

	l := len(keys)

	// protocol >= 1: ø dict -> EMPTY_DICT
	if e.config.Protocol >= 1 && l == 0 {
		return e.emit(opEmptyDict)
	}

	// MARK + ... + DICT
	// TODO detect cycles and double references to the same object
	// XXX sort keys, so the output is stable?
	err := e.emit(opMark)
	if err != nil {
		return err
	}

	for _, k := range keys {
		err = e.encode(k)
		if err != nil {
			return err
		}
		v := m.MapIndex(k)

		err = e.encode(v)
		if err != nil {
			return err
		}
	}

	return e.emit(opDict)
}

func (e *Encoder) encodeCall(v *Call) error {
	err := e.encodeClass(&v.Callable)
	if err != nil {
		return err
	}
	err = e.encodeTuple(v.Args)
	if err != nil {
		return err
	}
	return e.emit(opReduce)
}

var errP0123GlobalStringLineOnly = errors.New(`protocol 0-3: global: module & name must be string without \n`)

func (e *Encoder) encodeClass(v *Class) error {
	// PEP 3154: Protocol 4 forbids use of the GLOBAL opcode and replaces
	// it with STACK_GLOBAL.
	if e.config.Protocol >= 4 {
		err := e.encodeString(v.Module)
		if err != nil {
			return err
		}
		err = e.encodeString(v.Name)
		if err != nil {
			return err
		}
		return e.emit(opStackGlobal)
	}

	// else use GLOBAL opcode from protocol 0
	if strings.Contains(v.Module, "\n") || strings.Contains(v.Name, "\n") {
		return errP0123GlobalStringLineOnly
	}
	return e.emitf("%c%s\n%s\n", opGlobal, v.Module, v.Name)
}

var errP0PersIDStringLineOnly = errors.New(`protocol 0: persistent ID must be string without \n`)

func (e *Encoder) encodeRef(v *Ref) error {
	// protocol 0: pid must be string without \n
	if e.config.Protocol == 0 {
		pids, ok := v.Pid.(string)
		if !ok || strings.Contains(pids, "\n") {
			return errP0PersIDStringLineOnly
		}

		return e.emitf("%c%s\n", opPersid, pids)
	}

	// protocol >= 1: we can use opBinpersid which allows arbitrary object as argument
	err := e.encode(reflectValueOf(v.Pid))
	if err != nil {
		return err
	}
	return e.emit(opBinpersid)
}

func (e *Encoder) encodeStruct(st reflect.Value) error {

	typ := st.Type()

	// first test if it's one of our internal python structs
	switch v := st.Interface().(type) {
	case None:
		return e.emit(opNone)
	case Call:
		return e.encodeCall(&v)
	case Class:
		return e.encodeClass(&v)
	case Ref:
		return e.encodeRef(&v)
	case big.Int:
		return e.encodeLong(&v)
	}

	structTags := getStructTags(st)

	err := e.emit(opMark)
	if err != nil {
		return err
	}

	if structTags != nil {
		for f, i := range structTags {
			err := e.encodeString(f)
			if err != nil {
				return err
			}

			err = e.encode(st.Field(i))
			if err != nil {
				return err
			}
		}
	} else {
		l := typ.NumField()
		for i := 0; i < l; i++ {
			fty := typ.Field(i)
			if fty.PkgPath != "" {
				continue // skip unexported names
			}

			err := e.encodeString(fty.Name)
			if err != nil {
				return err
			}

			err = e.encode(st.Field(i))
			if err != nil {
				return err
			}
		}
	}

	return e.emit(opDict)
}

func reflectValueOf(v interface{}) reflect.Value {

	rv, ok := v.(reflect.Value)
	if !ok {
		rv = reflect.ValueOf(v)
	}
	return rv
}

func getStructTags(ptr reflect.Value) map[string]int {
	if ptr.Kind() != reflect.Struct {
		return nil
	}

	m := make(map[string]int)

	t := ptr.Type()

	l := t.NumField()
	numTags := 0
	for i := 0; i < l; i++ {
		field := t.Field(i).Tag.Get("pickle")
		if field != "" {
			m[field] = i
			numTags++
		}
	}

	if numTags == 0 {
		return nil
	}

	return m
}
