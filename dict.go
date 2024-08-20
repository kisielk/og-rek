package ogórek
// Python-like Dict that handles keys by Python-like equality on access.
//
// For example Dict.Get() will access the same element for all keys int(1), float64(1.0) and big.Int(1).

import (
	"encoding/binary"
	"fmt"
	"hash/maphash"
	"math"
	"math/big"
	"reflect"
	"sort"

	"github.com/aristanetworks/gomap"
)

// Dict represents dict from Python in PyDict mode.
//
// It mirrors Python with respect to which types are allowed to be used as
// keys, and with respect to keys equality. For example Tuple is allowed to be
// used as key, and all int(1), float64(1.0) and big.Int(1) are considered to be
// equal.
//
// For strings, similarly to Python3, [Bytes] and string are considered to be not
// equal, even if their underlying content is the same. However with same
// underlying content [ByteString], because it represents str type from Python2,
// is treated equal to both [Bytes] and string.
//
// See PyDict mode documentation in top-level package overview for details.
//
// Note: similarly to builtin map Dict is pointer-like type: its zero-value
// represents nil dictionary that is empty and invalid to use Set on.
type Dict struct {
	m *gomap.Map[any, any]
}

// NewDict returns new empty dictionary.
func NewDict() Dict {
	return NewDictWithSizeHint(0)
}

// NewDictWithSizeHint returns new empty dictionary with preallocated space for size items.
func NewDictWithSizeHint(size int) Dict {
	return Dict{m: gomap.NewHint[any, any](size, equal, hash)}
}

// NewDictWithData returns new dictionary with preset data.
//
// kv should be key₁, value₁, key₂, value₂, ...
func NewDictWithData(kv ...any) Dict {
	l := len(kv)
	if l % 2 != 0 {
		panic("odd number of arguments")
	}
	l /= 2
	d := NewDictWithSizeHint(l)
	for i := 0; i < l; i++ {
		k := kv[2*i]
		v := kv[2*i+1]
		d.Set(k, v)
	}
	return d
}

// Get returns value associated with equal key.
//
// An entry with key equal to the query is looked up and corresponding value
// is returned.
//
// nil is returned if no matching key is present in the dictionary.
//
// Get panics if key's type is not allowed to be used as Dict key.
func (d Dict) Get(key any) any {
	value, _ := d.Get_(key)
	return value
}

// Get_ is comma-ok version of Get.
func (d Dict) Get_(key any) (value any, ok bool) {
	return d.m.Get(key)
}

// Set sets key to be associated with value.
//
// Any previous keys, equal to the new key, are removed from the dictionary
// before the assignment.
//
// Set panics if key's type is not allowed to be used as Dict key.
func (d Dict) Set(key, value any) {
	// ByteString and container(with ByteString) are non-transitive equal types
	// so  Set(ByteString)       should first remove Bytes and string,
	// and Set(Tuple{ByteString) should first remove Tuple{Bytes} and Tuple{string}
	d.Del(key)
	d.m.Set(key, value)
}

// Del removes equal keys from the dictionary.
//
// All entries with key equal to the query are looked up and removed.
//
// Del panics if key's type is not allowed to be used as Dict key.
func (d Dict) Del(key any) {
	// see comment in Set about ByteString and container(with ByteString)
	for {
		d.m.Delete(key)
		_, have := d.Get_(key)
		if !have {
			break
		}
	}
}

// Len returns the number of items in the dictionary.
func (d Dict) Len() int {
	return d.m.Len()
}

// Iter returns iterator over all elements in the dictionary.
//
// The order to visit entries is arbitrary.
func (d Dict) Iter() /* iter.Seq2 */ func(yield func(any, any) bool) {
	it := d.m.Iter()
	return func(yield func(any, any) bool) {
		for it.Next() {
			cont := yield(it.Key(), it.Elem())
			if !cont {
				break
			}
		}
	}
}

// String returns human-readable representation of the dictionary.
func (d Dict) String() string {
	return d.sprintf("%v")
}

// GoString returns detailed human-readable representation of the dictionary.
func (d Dict) GoString() string {
	return fmt.Sprintf("%T%s", d, d.sprintf("%#v"))
}

// sprintf serves String and GoString.
func (d Dict) sprintf(format string) string {
	type KV struct { k,v string }
	vkv := make([]KV, 0, d.Len())
	d.Iter()(func(k, v any) bool {
		vkv = append(vkv, KV{
			k: fmt.Sprintf(format, k),
			v: fmt.Sprintf(format, v),
		})
		return true
	})

	sort.Slice(vkv, func(i, j int) bool {
		return vkv[i].k < vkv[j].k
	})

	s := "{"
	for i, kv := range vkv {
		if i > 0 {
			s += ", "
		}
		s += kv.k + ": " + kv.v
	}

	s += "}"
	return s
}


// ---- equal ----

// kind represents to which category a type belongs.
//
// It primarily classifies bool, numbers, slices, structs and maps, and puts
// everything else into "other" category.
type kind uint
const (
	kBool = iota
	kInt     // int + intX
	kUint    // uint + uintX
	kFloat   // floatX
	kComplex // complexX
	kBigInt  // *big.Int

	kSlice   // slice + array
	kMap     // map
	kStruct  // struct
	kPointer // pointer
	kOther   // everything else
)

// kindOf returns kind of x.
func kindOf(x any) kind {
	r := reflect.ValueOf(x)

	switch r.Kind() {
	case reflect.Bool:
		return kBool
	case reflect.Int, reflect.Int64, reflect.Int32, reflect.Int16, reflect.Int8:
		return kInt
	case reflect.Uint, reflect.Uint64, reflect.Uint32, reflect.Uint16, reflect.Uint8:
		return kUint
	case reflect.Float64, reflect.Float32:
		return kFloat
	case reflect.Complex128, reflect.Complex64:
		return kComplex

	case reflect.Slice, reflect.Array:
		return kSlice
	case reflect.Map:
		return kMap
	case reflect.Struct:
		return kStruct
	}

	switch x.(type) {
	case *big.Int:
		return kBigInt
	}

	switch r.Kind() {
	case reflect.Pointer:
		return kPointer
	}

	return kOther
}

// equal implements equality matching what Python would return for a == b.
//
// Equality properties:
//
// 1) equality is extension of Go ==
//
//	(a == b) ⇒ equal(a,b)
//
// 2) self equal:
//
//	equal(a,a) = y
//
// 3) equality is symmetrical:
//
//	equal(a,b) = equal(b,a)
//
// 4) equality is mostly transitive:
//
//	EqTransitive = set of all x:
//	  ∀ a,b,c ∈ EqTransitive:
//	    equal(a,b) ^ equal(b,c) ⇒ equal(a,c)
//
//	EqTransitive = all \ {ByteString + containers with ByteString}
func equal(xa, xb any) bool {
	// strings/bytes
	switch a := xa.(type) {
	case string:
		switch b := xb.(type) {
		case string:     return a == b
		case ByteString: return a == string(b)
		case Bytes:      return false
		default:         return false
		}

	case ByteString:
		switch b := xb.(type) {
		case string:     return a == ByteString(b)
		case ByteString: return a == b
		case Bytes:      return a == ByteString(b)
		default:         return false
		}

	case Bytes:
		switch b := xb.(type) {
		case string:     return false
		case ByteString: return a == Bytes(b)
		case Bytes:      return a == b
		default:         return false
		}
	}

	// everything else
	a := reflect.ValueOf(xa)
	b := reflect.ValueOf(xb)

	ak := kindOf(xa)
	bk := kindOf(xb)

	// since equality is symmetric, we can implement only half of comparison matrix
	if ak > bk {
		a, b = b, a
		ak, bk = bk, ak
		xa, xb = xb, xa
	}
	// ak ≤ bk

	handled := true
	switch ak {
	default:
		handled = false

	// numbers
	case kBool:
		// bool compares to numbers as 1 or 0
		//
		// In [1]: 1.0 == True
		// Out[1]: True
		//
		// In [2]: 0.0 == False
		// Out[2]: True
		//
		// In [3]: d = {1: 'abc'}
		//
		// In [4]: d[True]
		// Out[4]: 'abc'
		abint := bint(a.Bool())
		switch bk {
		case kBool:	return eq_Int_Int     (abint, bint(b.Bool()))
		case kInt:	return eq_Int_Int     (abint, b.Int())
		case kUint:	return eq_Int_Uint    (abint, b.Uint())
		case kFloat:	return eq_Int_Float   (abint, b.Float())
		case kComplex:	return eq_Int_Complex (abint, b.Complex())
		case kBigInt:	return eq_Int_BigInt  (abint, xb.(*big.Int))
		}

	case kInt:
		aint := a.Int()
		switch bk {
		// kBool
		case kInt:	return eq_Int_Int     (aint, b.Int())
		case kUint:	return eq_Int_Uint    (aint, b.Uint())
		case kFloat:	return eq_Int_Float   (aint, b.Float())
		case kComplex:	return eq_Int_Complex (aint, b.Complex())
		case kBigInt:	return eq_Int_BigInt  (aint, xb.(*big.Int))
		}

	case kUint:
		auint := a.Uint()
		switch bk {
		// kBool
		// kInt
		case kUint:	return eq_Uint_Uint    (auint, b.Uint())
		case kFloat:	return eq_Uint_Float   (auint, b.Float())
		case kComplex:	return eq_Uint_Complex (auint, b.Complex())
		case kBigInt:	return eq_Uint_BigInt  (auint, xb.(*big.Int))
		}

	case kFloat:
		afloat := a.Float()
		switch bk {
		// kBool
		// kInt
		// kUint
		case kFloat:	return eq_Float_Float   (afloat, b.Float())
		case kComplex:	return eq_Float_Complex (afloat, b.Complex())
		case kBigInt:	return eq_Float_BigInt  (afloat, xb.(*big.Int))
		}

	case kComplex:
		acomplex := a.Complex()
		switch bk {
		// kBool
		// kInt
		// kUint
		// kFloat
		case kComplex:	return eq_Complex_Complex (acomplex, b.Complex())
		case kBigInt:	return eq_Complex_BigInt  (acomplex, xb.(*big.Int))
		}

	case kBigInt:
		switch bk {
		// kBool
		// kInt
		// kUint
		// kFloat
		// kComplex
		case kBigInt:	return eq_BigInt_BigInt  (xa.(*big.Int), xb.(*big.Int))
		}

	// slices
	case kSlice:
		switch bk {
		case kSlice:	return eq_Slice_Slice (a, b)
		}

	// builtin map
	case kMap:
		switch bk {
		case kMap:	return eq_Map_Map  (a, b)
		}
		switch b := xb.(type) {
		case Dict:	return eq_Map_Dict (a, b)
		}
	}

	if handled {
		return false
	}

	// our types that need special handling
	switch a := xa.(type) {
	case Dict:
		switch b := xb.(type) {
		case Dict:	return eq_Dict_Dict(a, b)
		default:        return false
		}
	}

	// structs  (also covers None, Class, Call etc...)
	switch ak {
	case kStruct:
		switch bk {
		case kStruct:	return eq_Struct_Struct (a, b)
		default:        return false
		}
	}

	return (xa == xb) // fallback to builtin equality
}


// equality matrix. nontrivial elements

func eq_Int_Uint(a int64, b uint64) bool {
	if a >= 0 {
		return uint64(a) == b
	}
	return false
}


func eq_Int_BigInt(a int64, b *big.Int) bool {
	if b.IsInt64() {
		return a == b.Int64()
	}
	return false
}

func eq_Uint_BigInt(a uint64, b *big.Int) bool {
	if b.IsUint64() {
		return a == b.Uint64()
	}
	return false
}

func eq_Float_BigInt(a float64, b *big.Int) bool {
	bf, accuracy := bigInt_Float64(b)
	if accuracy == big.Exact {
		return a == bf
	}
	return false
}

func eq_Complex_BigInt(a complex128, b *big.Int) bool {
	if imag(a) == 0 {
		return eq_Float_BigInt(real(a), b)
	}
	return false
}

func eq_BigInt_BigInt(a, b *big.Int) bool {
	return (a.Cmp(b) == 0)
}

func eq_Slice_Slice(a, b reflect.Value) bool {
	al := a.Len()
	bl := b.Len()
	if al != bl {
		return false
	}
	for i := 0; i < al; i++ {
		if !equal(a.Index(i).Interface(), b.Index(i).Interface()) {
			return false
		}
	}
	return true
}

func eq_Struct_Struct(a, b reflect.Value) bool {
	if a.Type() != b.Type() {
		return false
	}

	typ := a.Type()
	l := typ.NumField()
	for i := 0; i < l; i++ {
		af := a.Field(i)
		bf := b.Field(i)

		// .Interface() is not allowed if the field is private.
		// Work around the protection via unsafe. We may need to switch
		// to struct copy if it is not addressable because Addr() is
		// used in the workaround. https://stackoverflow.com/a/43918797/9456786
		ftyp := typ.Field(i)
		if !ftyp.IsExported() {
			if !af.CanAddr() {
				// switch a to addressable copy
				a_ := reflect.New(typ).Elem()
				a_.Set(a)
				a = a_
				af = a.Field(i)
			}

			if !bf.CanAddr() {
				// switch b to addressable copy
				b_ := reflect.New(typ).Elem()
				b_.Set(b)
				b = b_
				bf = b.Field(i)
			}

			af = reflect.NewAt(ftyp.Type, af.Addr().UnsafePointer()).Elem()
			bf = reflect.NewAt(ftyp.Type, bf.Addr().UnsafePointer()).Elem()
		}

		if !equal(af.Interface(), bf.Interface()) {
			return false
		}
	}
	return true
}

func eq_Dict_Dict(a Dict, b Dict) bool {
	// dicts D₁ and D₂ are considered equal if the following is true:
	//
	//     - len(D₁) = len(D₂)
	//     - ∀ k ∈ D₁  equal(D₁[k], D₂[k]) = y
	//     - ∀ k ∈ D₂  equal(D₁[k], D₂[k]) = y
	//
	// this definition is reasonable and fast to implement without additional memory.
	// Also if D₁ and D₂ have keys only from equal-transitive subset of all
	// keys (i.e. anything without ByteString), it becomes equivalent to the
	// following definition:
	//
	//     - (k₁i, v₁i) is set of all key/values from D₁
	//     - (k₂j, v₂j) is set of all key/values from D₂
	//     - equal(D₁,D₂):
	//
	//         ∃ 1-1 mapping in between i<->j: equal(k₁i, k₂j) ^ equal(v₁i, v₂j)

	if a.Len() != b.Len() {
		return false
	}

	eq := true
	a.Iter()(func(k,va any) bool {
		vb, ok := b.Get_(k)
		if !ok || !equal(va, vb) {
			eq = false
			return false
		}
		return true
	})
	if !eq {
		return false
	}

	b.Iter()(func(k,vb any) bool {
		va, ok := a.Get_(k)
		if !ok || !equal(va, vb) {
			eq = false
			return false
		}
		return true
	})
	return eq
}

// equal(Map, Dict) and equal(Map, Map) follow semantic of equal(Dict, Dict)

func eq_Map_Dict(a reflect.Value, b Dict) bool {
	if a.Len() != b.Len() {
		return false
	}

	aKeyType := a.Type().Key()

	ai := a.MapRange()
	for ai.Next() {
		k  := ai.Key().Interface()
		va := ai.Value().Interface()
		vb, ok := b.Get_(k)
		if !ok || !equal(va, vb) {
			return false
		}
	}

	eq := true
	b.Iter()(func(k,vb any) bool {
		xk := reflect.ValueOf(k)
		if !xk.Type().AssignableTo(aKeyType) {
			eq = false
			return false
		}
		xva := a.MapIndex(xk)
		if !(xva.IsValid() && equal(xva.Interface(), vb)) {
			eq = false
			return false
		}
		return true
	})
	return eq
}

func eq_Map_Map(a reflect.Value, b reflect.Value) bool {
	if a.Len() != b.Len() {
		return false
	}

	aKeyType := a.Type().Key()
	bKeyType := b.Type().Key()

	ai := a.MapRange()
	for ai.Next() {
		k  := ai.Key().Interface() // NOTE xk != ai.Key() because that might have type any
		xk := reflect.ValueOf(k)   //      while xk has type of particular contained value
		va := ai.Value().Interface()
		if !xk.Type().AssignableTo(bKeyType) {
			return false
		}
		xvb := b.MapIndex(xk)
		if !(xvb.IsValid() && equal(va, xvb.Interface())) {
			return false
		}
	}

	bi := b.MapRange()
	for bi.Next() {
		k  := bi.Key().Interface() // see ^^^
		xk := reflect.ValueOf(k)
		vb := bi.Value().Interface()
		if !xk.Type().AssignableTo(aKeyType) {
			return false
		}
		xva := a.MapIndex(xk)
		if !(xva.IsValid() && equal(xva.Interface(), vb)) {
			return false
		}
	}

	return true
}


// equality matrix. trivial elements

func eq_Int_Int     (a int64, b int64)      bool { return a == b }
func eq_Int_Float   (a int64, b float64)    bool { return float64(a) == b }
func eq_Int_Complex (a int64, b complex128) bool { return complex(float64(a), 0) == b }

func eq_Uint_Uint    (a uint64, b uint64)     bool { return a == b }
func eq_Uint_Float   (a uint64, b float64)    bool { return float64(a) == b }
func eq_Uint_Complex (a uint64, b complex128) bool { return complex(float64(a), 0) == b }

func eq_Float_Float    (a float64, b float64)     bool { return a == b }
func eq_Float_Complex  (a float64, b complex128)  bool { return complex(a, 0) == b }

func eq_Complex_Complex (a complex128, b complex128)  bool { return a == b }


// ---- hash ----

// hash returns hash of x consistent with equality implemented by equal.
//
//	equal(a,b)  ⇒  hash(a) = hash(b)
//
// hash panics with "unhashable type: ..." if x is not allowed to be used as Dict key.
func hash(seed maphash.Seed, x any) uint64 {
	// strings/bytes use standard hash of string
	switch v := x.(type) {
	case string:     return maphash_String(seed, v)
	case ByteString: return maphash_String(seed, string(v))
	case Bytes:      return maphash_String(seed, string(v))
	}

	// for everything else we implement custom hashing ourselves to match equal
	var h maphash.Hash
	h.SetSeed(seed)

	hash_Uint := func(u uint64) {
		var b [8]byte
		binary.BigEndian.PutUint64(b[:], u)
		h.Write(b[:])
	}

	hash_Int := func(i int64) {
		hash_Uint(uint64(i))
	}

	hash_Float := func(f float64) {
		// if float is in int range and is integer number - hash it as integer
		i  := int64(f)
		f_ := float64(i)
		if f_ == f {
			hash_Int(i)

		// else use raw float64 bytes representation for hashing
		} else {
			hash_Uint(math.Float64bits(f))
		}
	}


	// numbers
	r := reflect.ValueOf(x)
	k := kindOf(x)

	handled := true
	switch k {
	default:
		handled = false

	case kBool:   hash_Int(bint(r.Bool()))
	case kInt:    hash_Int(r.Int())
	case kUint:   hash_Uint(r.Uint())
	case kFloat:  hash_Float(r.Float())

	case kComplex:
		c := r.Complex()
		hash_Float(real(c))
		if imag(c) != 0 {
			hash_Float(imag(c))
		}

	case kBigInt:
		b := x.(*big.Int)
		switch {
		case b.IsInt64():  hash_Int(b.Int64())
		case b.IsUint64(): hash_Uint(b.Uint64())
		default:
			f, accuracy := bigInt_Float64(b)
			if accuracy == big.Exact {
				hash_Float(f)
			} else {
				h.WriteString("bigInt")
				h.Write(b.Bytes())
			}
		}

	// kSlice  - skip
	// kStruct - skip

	case kPointer:  hash_Uint(uint64(r.Elem().UnsafeAddr()))
	}

	if handled {
		return h.Sum64()
	}

	// tuple
	switch v := x.(type) {
	case Tuple:
		h.WriteString("tuple")
		for _, item := range v {
			hash_Uint(hash(seed, item))
		}
		return h.Sum64()
	}

	// structs  (also covers None, Class, Call etc)
	switch k {
	case kStruct:
		// our types that are handled specially by equal
		switch x.(type) {
		case Dict:
			goto unhashable
		}

		typ := r.Type()
		h.WriteString(typ.Name())
		l := typ.NumField()
		for i := 0; i < l; i++ {
			f := r.Field(i)

			// .Interface() is not allowed if the field is private.
			// Work it around via unsafe. See eq_Struct_Struct for details.
			ftyp := typ.Field(i)
			if !ftyp.IsExported() {
				if !f.CanAddr() {
					// switch r to addressable copy
					r_ := reflect.New(typ).Elem()
					r_.Set(r)
					r = r_
					f = r.Field(i)
				}
				f = reflect.NewAt(ftyp.Type, f.Addr().UnsafePointer()).Elem()
			}

			hash_Uint(hash(seed, f.Interface()))
		}
		return h.Sum64()
	}

unhashable:
	panic(fmt.Sprintf("unhashable type: %T", x))
}


// ---- misc ----

// bint returns int corresponding to bool.
//
// true  -> 1
// false -> 0
func bint(x bool) int64 {
	if x {
		return 1
	}
	return 0
}
