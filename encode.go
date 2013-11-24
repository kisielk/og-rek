package ogórek

import (
	"encoding/binary"
	"fmt"
	"io"
	"math"
	"reflect"
)

// An Encoder encodes Go data structures into pickle byte stream
type Encoder struct {
	w io.Writer
}

// NewEncoder returns a new Encoder struct with default values
func NewEncoder(w io.Writer) *Encoder {
	return &Encoder{w: w}
}

// Encode writes the pickle encoding of v to w, the encoder's writer
func (e *Encoder) Encode(v interface{}) error {
	rv := reflectValueOf(v)
	e.encode(rv)
	e.w.Write([]byte{opStop})
	return nil
}

func (e *Encoder) encode(rv reflect.Value) error {

	switch rk := rv.Kind(); rk {

	case reflect.Bool:
		e.encodeBool(rv.Bool())
	case reflect.Int, reflect.Int8, reflect.Int64, reflect.Int32, reflect.Int16:
		e.encodeInt(reflect.Int, rv.Int())
	case reflect.Uint8, reflect.Uint64, reflect.Uint, reflect.Uint32, reflect.Uint16:
		e.encodeInt(reflect.Uint, int64(rv.Uint()))
	case reflect.String:
		e.encodeString(rv.String())
	case reflect.Array, reflect.Slice:
		if rv.Type().Elem().Kind() == reflect.Uint8 {
			e.encodeBytes(rv.Bytes())
		} else {
			e.encodeArray(rv)
		}

	case reflect.Map:
		e.encodeMap(rv)

	case reflect.Struct:
		e.encodeStruct(rv)

	case reflect.Float32, reflect.Float64:
		e.encodeFloat(float64(rv.Float()))

	case reflect.Interface:
		// recurse until we get a concrete type
		// could be optmized into a tail call
		var err error
		err = e.encode(rv.Elem())
		if err != nil {
			return err
		}

	case reflect.Ptr:

		if rv.Elem().Kind() == reflect.Struct {
			switch rv.Elem().Interface().(type) {
			case None:
				e.encodeStruct(rv.Elem())
				return nil
			}
		}

		e.encode(rv.Elem())

	case reflect.Invalid:
		e.w.Write([]byte{opNone})

	default:
		panic(fmt.Sprintf("no support for type '%s'", rk.String()))
	}

	return nil
}

func (e *Encoder) encodeArray(arr reflect.Value) {

	l := arr.Len()

	e.w.Write([]byte{opEmptyList, opMark})

	for i := 0; i < l; i++ {
		v := arr.Index(i)
		e.encode(v)
	}
	e.w.Write([]byte{opAppends})
}

func (e *Encoder) encodeBool(b bool) {
	if b {
		e.w.Write([]byte(opTrue))
	} else {
		e.w.Write([]byte(opFalse))
	}
}

func (e *Encoder) encodeBytes(byt []byte) {

	l := len(byt)

	if l < 256 {
		e.w.Write([]byte{opShortBinstring, byte(l)})
	} else {
		e.w.Write([]byte{opBinstring})
		binary.Write(e.w, binary.LittleEndian, int32(l))
	}

	e.w.Write(byt)
}

func (e *Encoder) encodeFloat(f float64) {
	var u uint64
	u = math.Float64bits(f)
	e.w.Write([]byte{opBinfloat})
	binary.Write(e.w, binary.BigEndian, u)
}

func (e *Encoder) encodeInt(k reflect.Kind, i int64) {

	// FIXME: need support for 64-bit ints

	switch {
	case i > 0 && i < math.MaxUint8:
		e.w.Write([]byte{opBinint1, byte(i)})
	case i > 0 && i < math.MaxUint16:
		e.w.Write([]byte{opBinint2, byte(i), byte(i >> 8)})
	case i >= math.MinInt32 && i <= math.MaxInt32:
		e.w.Write([]byte{opBinint})
		binary.Write(e.w, binary.LittleEndian, int32(i))
	default: // int64, but as a string :/
		e.w.Write([]byte{opInt})
		fmt.Fprintf(e.w, "%d\n", i)
	}
}

func (e *Encoder) encodeScalar(rv reflect.Value) {

	switch rv.Kind() {
	case reflect.Array, reflect.Slice:
		if rv.Type().Elem().Kind() == reflect.Uint8 {
			e.encodeBytes(rv.Bytes())
		} else {
			e.encodeArray(rv)
		}
	case reflect.Map:
		e.encodeMap(rv)
	case reflect.Interface:
		e.encodeScalar(rv.Elem())
	default:
		e.encode(rv)
	}
}

func (e *Encoder) encodeMap(m reflect.Value) {

	keys := m.MapKeys()

	l := len(keys)

	e.w.Write([]byte{opEmptyDict})

	if l > 0 {
		e.w.Write([]byte{opMark})

		for _, k := range keys {
			e.encode(k)
			v := m.MapIndex(k)
			e.encode(v)
		}
		e.w.Write([]byte{opSetitems})
	}
}

func (e *Encoder) encodeString(s string) {
	e.encodeBytes([]byte(s))
}

func (e *Encoder) encodeStruct(st reflect.Value) {

	typ := st.Type()

	// first test if it's one of our internal python structs
	if _, ok := st.Interface().(None); ok {
		e.w.Write([]byte{opNone})
		return
	}

	e.w.Write([]byte{opEmptyDict})

	l := typ.NumField()

	structTags := getStructTags(st)

	e.w.Write([]byte{opEmptyDict, opMark})

	if structTags != nil {
		for f, i := range structTags {
			e.encodeString(f)
			e.encode(st.Field(i))
		}
	} else {
		for i := 0; i < l; i++ {
			fty := typ.Field(i)
			if fty.PkgPath != "" {
				continue // skip unexported names
			}
			e.encodeString(fty.Name)
			e.encode(st.Field(i))
		}
	}

	e.w.Write([]byte{opSetitems})
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
