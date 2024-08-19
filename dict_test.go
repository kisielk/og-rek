package ogórek

import (
	"fmt"
	"hash/maphash"
	"reflect"
	"strings"
	"testing"
)


// tStructWithPrivate is used by tests to verify handing of struct with private fields.
type tStructWithPrivate struct {
	x, y any
}


// TestEqual verifies equal and hash.
func TestEqual(t *testing.T) {
	// tEqualSet represents tested set of values:
	// ∀ a ∈ tEqualSet:
	//   ∀ b ∈ tEqualSet ⇒ equal(a,b) = y
	//   ∀ c ∉ tEqualSet ⇒ equal(a,c) = n
	//
	// Intersection in between different tEqualSets is mostly empty: such
	// intersections can contain elements only from all \ EqTransitive, i.e. only ByteString.
	type tAllEqual []any

	// E is shortcut to create tEqualSet
	E := func(v ...any) tAllEqual { return tAllEqual(v) }

	// D and M are shortcuts to create Dict and map[any]any
	D := NewDictWithData
	type M = map[any]any

	// i1  and i1_  are two integer variables equal to 1 but with different address
	// obj and obj_ are similar equal structures located at different memory regions
	i1  := 1;  i1_ := 1
	obj := &Class{"a","b"};  obj_ := &Class{"a","b"}

	// testv is vector of all test-cases
	testv := []tAllEqual {
		// numbers
		E(int(0),
		   int64(0),  int32(0),  int16(0),  int8(0),
		  uint64(0), uint32(0), uint16(0), uint8(0),
		  bigInt("0"),
		  false,
		  float32  (0), float64   (0),
		  complex64(0), complex128(0)),

		E(int(1),
		  int64 (1),  int32(1),  int16(1),  int8(1),
		  uint64(1), uint32(1), uint16(1), uint8(1),
		  bigInt("1"),
		  true,
		  float32  (1), float64   (1),
		  complex64(1), complex128(1)),

		E(int(-1),
		  int64(-1), int32(-1), int16(-1),  int8(-1),
		  // NOTE no uintX because they ≥ 0 only
		  bigInt("-1"),
		  // NOTE no bool because it ∈ {0,1}
		  float32  (-1), float64   (-1),
		  complex64(-1), complex128(-1)),

		// intX/uintX different range
		E(int(0xff),
		   int64(0xff),  int32(0xff),  int16(0xff), // int8(overflow),
		  uint64(0xff), uint32(0xff), uint16(0xff), // uint8(overflow),
		  bigInt("255"),
		  bigInt("255"), // two different *big.Int instances
		  float32  (0xff), float64   (0xff),
		  complex64(0xff), complex128(0xff)),

		E(int(-0x80),
		  int64(-0x80), int32(-0x80), int16(-0x80), int8(-0x80),
		  //uint64(), uint32(), uint16(), uint8(), ≥ 0 only
		  bigInt("-128"),
		  float32  (-0x80), float64   (-0x80),
		  complex64(-0x80), complex128(-0x80)),

		E(int(0xffff),
		   int64(0xffff),  int32(0xffff), // int16(overflow), int8(overflow),
		  uint64(0xffff), uint32(0xffff), uint16(0xffff), // uint8(overflow),
		  bigInt("65535"),
		  float32  (0xffff), float64   (0xffff),
		  complex64(0xffff), complex128(0xffff)),

		E(int(-0x8000),
		  int64(-0x8000), int32(-0x8000), int16(-0x8000), // int8(overflow),
		  //uint64(), uint32(), uint16(), uint8(), ≥ 0 only
		  bigInt("-32768"),
		  float32  (-0x8000), float64   (-0x8000),
		  complex64(-0x8000), complex128(-0x8000)),

		E(int(0xffffffff),
		   int64(0xffffffff), // int32(overflow), int16(overflow), int8(overflow),
		  uint64(0xffffffff), uint32(0xffffffff), // uint16(overflow), uint8(overflow),
		  bigInt("4294967295"),
		  /* float32  (precision loss), */ float64   (0xffffffff),
		  /* complex64(precision loss), */ complex128(0xffffffff)),

		E(int(-0x80000000),
		  int64(-0x80000000), int32(-0x80000000), // int16(overflow), int8(overflow),
		  //uint64(), uint32(), uint16(), uint8(), ≥ 0 only
		  bigInt("-2147483648"),
		  float32  (-0x80000000), float64   (-0x80000000),
		  complex64(-0x80000000), complex128(-0x80000000)),

		E(// int(overflow),
		  // int64(overflow), int32(overflow), int16(overflow), int8(overflow),
		  uint64(0xffffffffffffffff), // uint32(overflow), uint16(overflow), uint8(overflow),
		  bigInt("18446744073709551615")),
		  // float32  (precision loss), float64   (precision loss),
		  // complex64(precision loss), complex128(precision loss)),

		E(int(-0x8000000000000000),
		  int64(-0x8000000000000000), // int32(overflow), int16(overflow), int8(overflow),
		  //uint64(), uint32(), uint16(), uint8(), ≥ 0 only
		  bigInt("-9223372036854775808"),
		  float32  (-0x8000000000000000), float64   (-0x8000000000000000),
		  complex64(-0x8000000000000000), complex128(-0x8000000000000000)),

		E(bigInt("1"+strings.Repeat("0",22)), float64(1e22), complex128(complex(1e22,0))),
		E(complex64(complex(0,1)), complex128(complex(0,1))),
		E(float64(1.25), float32(1.25), complex64(complex(1.25,0)), complex128(complex(1.25,0))),

		// strings/bytes
		E("",    ByteString("")),	E(ByteString(""),    Bytes("")),
		E("a",   ByteString("a")),	E(ByteString("a"),   Bytes("a")),
		E("мир", ByteString("мир")),	E(ByteString("мир"), Bytes("мир")),

		// none / empty tuple|list
		E(None{}),
		E(Tuple{}, []any{}),

		// sequences
		E([]int{}, []float32{}, []any{}, Tuple{}, [0]float64{}),
		E([]int{1,2}, []float32{1,2}, []any{1,2}, Tuple{1,2}, [2]float64{1,2}),
		E([]any{1,"a"}, Tuple{1,"a"}, [2]any{1,"a"}, Tuple{1,ByteString("a")}),

		// Dict, map
		E(D(),
		  M{}, map[int]bool{}),
		E(D(1,bigInt("2")),
		  M{1:2.0}, map[int]int{1:2}),
		E(D(1,"a"),
		  M{1:"a"}, map[int]string{1:"a"}),
		E(D("a",1),
		  M{"a":1}),
		E(D("a",1, None{},2),
		  M{"a":1, None{}:2}),
		E(D("a",1, Bytes("a"),1),
		  M{"a":1, Bytes("a"):1}),
		E(D("a",1, Bytes("a"),2),
		  M{"a":1, Bytes("a"):2}),
		E(D("a",1), D(ByteString("a"),1)),  E(D(ByteString("a"),1), D(Bytes("a"),1)),
		E(D("a",1, Bytes("a"),1, ByteString("b"),2),
		  D(ByteString("a"),1, "b",2, Bytes("b"),2)),

		// structs
		E(Class{"mod","cls"}, Class{"mod","cls"}),
		E(Call{Class{"mod","cls"}, Tuple{"a","b",3}},
		  Call{Class{"mod","cls"}, Tuple{ByteString("a"),"b",bigInt("3")}}),
		E(Ref{1}, Ref{bigInt("1")}, Ref{1.0}),
		E(tStructWithPrivate{"a",1}, tStructWithPrivate{ByteString("a"),bigInt("1")}),
		E(tStructWithPrivate{"b",2}, tStructWithPrivate{"b",2.0}),

		// pointers, as in builtin ==, are compared only by address
		E(&i1), E(&i1_), E(&obj), E(&obj_),

		// nil
		E(nil),
	}
	// automatically test equality on Tuples/list from ^^^ data
	testvAddSequences := func() {
		l := len(testv)
		for i := 0; i < l; i++ {
			Ex := testv[i]
			Ey := testv[(i+1)%l]

			x0 := Ex[0];  x1 := Ex[1%len(Ex)]
			y0 := Ey[0];  y1 := Ey[1%len(Ey)]

			t1 := Tuple{x0,y0};  l1 := []any{x0,y0}
			t2 := Tuple{x1,y1};  l2 := []any{x1,y1}

			testv = append(testv, E(t1, t2, l1, l2))
		}
	}
	testvAddSequences()
	// and sequences of sequences
	testvAddSequences()

	// thash is used to invoke hash.
	// if x is not hashable ok=false is returned instead of panic.
	tseed := maphash.MakeSeed()
	thash := func(x any) (h uint64, ok bool) {
		defer func() {
			r := recover()
			if r != nil {
				s, sok := r.(string)
				if sok && strings.HasPrefix(s, "unhashable type: ") {
					ok = false
					h  = 0
				} else {
					panic(r)
				}
			}
		}()

		return hash(tseed, x), true
	}

	// tequal is used to invoke equal.
	// it automatically checks Go-extension, self-equal, symmetry and hash invariants:
	//
	//	a==b        ⇒  equal(a,b)
	//	equal(a,a)  =  y
	//	equal(a,b)  =  equal(b,a)
	//	equal(a,b)  ⇒  hash(a) = hash(b)
	tequal := func(a, b any) bool {
		aa := equal(a, a)
		bb := equal(b, b)
		if !aa {
			t.Errorf("not self-equal  %T %#v", a,a)
		}
		if !bb {
			t.Errorf("not self-equal  %T %#v", b,b)
		}

		eq := equal(a, b)
		qe := equal(b, a)

		if eq != qe {
			t.Errorf("equal not symmetric:  %T %#v  %T %#v;  a == b: %v  b == a: %v",
				 a,a, b,b, eq, qe)
		}

		ah, ahOk := thash(a)
		bh, bhOk := thash(b)
		if eq && ahOk && bhOk && !(ah == bh) {
			t.Errorf("hash different of equal  %T %#v hash:%x  %T %#v hash:%x",
				 a,a,ah, b,b,bh)
		}

		goeq := false
		func() {
			// a == b can trigger "comparing uncomparable type ..."
			// even if reflect reports both types as comparable
			// (see mapTryAssign for details)
			defer func() {
				recover()
			}()

			goeq = (a == b)
		}()

		if goeq && !eq {
			t.Errorf("equal is not extension of ==  %T %#v  %T %#v",
				a,a, b,b)
		}

		return eq
	}

	// EHas returns whether x ∈ E.
	EHas := func(E tAllEqual, x any) bool {
		for _, a := range E {
			if tequal(a, x) {
				return true
			}
		}
		return false;
	}


	// do the tests
	for i, E1 := range testv {
		// ∀ a,b ∈ tEqualSet ⇒ equal(a,b) = y
		for _, a := range E1 {
			for _, b := range E1 {
				if !tequal(a,b) {
					t.Errorf("not equal  %T %#v  %T %#v", a,a, b,b)
				}
			}
		}

		// ∀ a ∈ tEqualSet
		// ∀ c ∉ tEqualSet ⇒ equal(a,c) = n
		for j, E2 := range testv {
			if j == i {
				continue
			}

			for _, a := range E1 {
				for _, c := range E2 {
					if EHas(E1, c) {
						continue
					}

					if tequal(a,c) {
						t.Errorf("equal  %T %#v  %T %#v", a,a, c,c)
					}
				}
			}
		}
	}
}


// TestDict verifies Dict.
func TestDict(t *testing.T) {
	d := NewDict()

	// assertData asserts that d has data exactly as specified by provided key,value pairs.
	assertData := func(kvok ...any) {
		t.Helper()

		if len(kvok) % 2 != 0 {
			panic("kvok % 2 != 0")
		}
		lok := len(kvok)/2
		kvokGet := func(k any) (any, bool) {
			t.Helper()
			for i := 0; i < lok; i++ {
				kok := kvok[2*i]
				vok := kvok[2*i+1]
				if reflect.TypeOf(k) == reflect.TypeOf(kok) &&
				   equal(k, kok) {
					return vok, true
				}
			}
			return nil, false
		}

		bad := false
		badf := func(format string, argv ...any) {
			t.Helper()
			bad = true
			t.Errorf(format, argv...)
		}

		l := d.Len()
		if l != lok {
			badf("len: have: %d  want: %d", l, lok)
		}

		d.Iter()(func(k,v any) bool {
			t.Helper()
			vok, ok := kvokGet(k)
			if !ok {
				badf("unexpected key %#v", k)
			}
			if v != vok {
				badf("key %T %#v -> value %#T %#v  ;  want %T %#v", k,k, v,v, vok,vok)
			}
			return true
		})

		if bad {
			t.Fatalf("\nd:   %#v\nkvok: %#v", d, kvok)
		}
	}

	// assertGet asserts that d.Get(k) results in exactly vok or any element from vokExtra.
	assertGet := func(k any, vok any, vokExtra ...any) {
		t.Helper()
		v := d.Get(k)
		if v == vok {
			return
		}
		for _, eok := range vokExtra {
			if v == eok {
				return
			}
		}

		emsg := fmt.Sprintf("get %#v: have: %#v  want: %#v", k, v, vok)
		for _, eok := range vokExtra {
			emsg += fmt.Sprintf(" ∪ %#v", eok)
		}
		emsg += fmt.Sprintf("\nd: %#v", d)
		t.Fatal(emsg)
	}

	// numbers
	assertData()

	d.Set(1, "x")
	assertData(1,"x")
	assertGet(1,            "x")
	assertGet(1.0,          "x")
	assertGet(bigInt("1"),  "x")
	assertGet(complex(1,0), "x")

	d.Del(7)
	assertData(1,"x")
	assertGet(1,            "x")
	assertGet(1.0,          "x")
	assertGet(bigInt("1"),  "x")
	assertGet(complex(1,0), "x")

	d.Set(2.5, "y")
	assertData(1,"x", 2.5,"y")
	assertGet(1,              "x")
	assertGet(1.0,            "x")
	assertGet(bigInt("1"),    "x")
	assertGet(complex(1,0),   "x")
	assertGet(2,              nil)
	assertGet(2.5,            "y")
	assertGet(bigInt("2"),    nil)
	assertGet(complex(2.5,0), "y")

	d.Del(1)
	assertData(2.5,"y")
	assertGet(1,              nil)
	assertGet(1.0,            nil)
	assertGet(bigInt("1"),    nil)
	assertGet(complex(1,0),   nil)
	assertGet(2,              nil)
	assertGet(2.5,            "y")
	assertGet(bigInt("2"),    nil)
	assertGet(complex(2.5,0), "y")

	d.Del(2.5)
	assertData()
	assertGet(1,              nil)
	assertGet(1.0,            nil)
	assertGet(bigInt("1"),    nil)
	assertGet(complex(1,0),   nil)
	assertGet(2,              nil)
	assertGet(2.5,            nil)
	assertGet(bigInt("2"),    nil)
	assertGet(complex(2.5,0), nil)

	// strings/bytes
	assertData()
	assertGet("abc", nil)

	d.Set("abc", "a")
	assertData("abc","a")
	assertGet("abc",             "a")
	assertGet(Bytes("abc"),      nil)
	assertGet(ByteString("abc"), "a")

	d.Set(Bytes("abc"), "b")
	assertData("abc","a", Bytes("abc"),"b")
	assertGet("abc",             "a")
	assertGet(Bytes("abc"),      "b")
	assertGet(ByteString("abc"), "a", "b")

	d.Set(ByteString("abc"), "c")
	assertData(ByteString("abc"),"c")
	assertGet("abc",             "c")
	assertGet(Bytes("abc"),      "c")
	assertGet(ByteString("abc"), "c")

	d.Del("abc")
	assertData()
	assertGet("abc",             nil)
	assertGet(Bytes("abc"),      nil)
	assertGet(ByteString("abc"), nil)

	d.Set("abc", "a")
	assertData("abc","a")
	assertGet("abc",             "a")
	assertGet(Bytes("abc"),      nil)
	assertGet(ByteString("abc"), "a")

	d.Set(Bytes("abc"), "b")
	assertData("abc","a", Bytes("abc"),"b")
	assertGet("abc",             "a")
	assertGet(Bytes("abc"),      "b")
	assertGet(ByteString("abc"), "a", "b")

	d.Del(ByteString("abc"))
	assertData()
	assertGet("abc",             nil)
	assertGet(Bytes("abc"),      nil)
	assertGet(ByteString("abc"), nil)

	// None, tuple
	assertData()

	d.Set(None{}, "n")
	assertData(None{},"n")
	assertGet(None{},  "n")
	assertGet(Tuple{}, nil)

	d.Set(Tuple{}, "t")
	assertData(None{},"n", Tuple{},"t")
	assertGet(None{},  "n")
	assertGet(Tuple{}, "t")

	d.Set(Tuple{1,2,"a"}, "t12a")
	assertData(None{},"n", Tuple{},"t", Tuple{1,2,"a"},"t12a")
	assertGet(None{},       "n")
	assertGet(Tuple{},      "t")
	assertGet(Tuple{1,2},   nil)
	assertGet(Tuple{1,2,"a"},             "t12a")
	assertGet(Tuple{1,2,Bytes("a")},      nil)
	assertGet(Tuple{1,2,ByteString("a")}, "t12a")

	d.Set(Tuple{1,2,Bytes("a")}, "t12b")
	assertData(None{},"n", Tuple{},"t", Tuple{1,2,"a"},"t12a", Tuple{1,2,Bytes("a")},"t12b")
	assertGet(None{},       "n")
	assertGet(Tuple{},      "t")
	assertGet(Tuple{1,2},   nil)
	assertGet(Tuple{1,2,"a"},             "t12a")
	assertGet(Tuple{1,2,Bytes("a")},      "t12b")
	assertGet(Tuple{1,2,ByteString("a")}, "t12a", "t12b")

	d.Set(Tuple{1,2,ByteString("a")}, "t12c")
	assertData(None{},"n", Tuple{},"t", Tuple{1,2,ByteString("a")},"t12c")
	assertGet(None{},       "n")
	assertGet(Tuple{},      "t")
	assertGet(Tuple{1,2},   nil)
	assertGet(Tuple{1,2,"a"},             "t12c")
	assertGet(Tuple{1,2,Bytes("a")},      "t12c")
	assertGet(Tuple{1,2,ByteString("a")}, "t12c")

	d.Set(Tuple{1,2,"a"}, "t12a")
	assertData(None{},"n", Tuple{},"t", Tuple{1,2,"a"},"t12a")
	assertGet(None{},       "n")
	assertGet(Tuple{},      "t")
	assertGet(Tuple{1,2},   nil)
	assertGet(Tuple{1,2,"a"},             "t12a")
	assertGet(Tuple{1,2,Bytes("a")},      nil)
	assertGet(Tuple{1,2,ByteString("a")}, "t12a")

	d.Set(Tuple{1,2,Bytes("a")}, "t12b")
	assertData(None{},"n", Tuple{},"t", Tuple{1,2,"a"},"t12a", Tuple{1,2,Bytes("a")},"t12b")
	assertGet(None{},       "n")
	assertGet(Tuple{},      "t")
	assertGet(Tuple{1,2},   nil)
	assertGet(Tuple{1,2,"a"},             "t12a")
	assertGet(Tuple{1,2,Bytes("a")},      "t12b")
	assertGet(Tuple{1,2,ByteString("a")}, "t12a", "t12b")

	d.Del(Tuple{1,2,ByteString("a")})
	assertData(None{},"n", Tuple{},"t")
	assertGet(None{},       "n")
	assertGet(Tuple{},      "t")
	assertGet(Tuple{1,2},   nil)
	assertGet(Tuple{1,2,"a"},             nil)
	assertGet(Tuple{1,2,Bytes("a")},      nil)
	assertGet(Tuple{1,2,ByteString("a")}, nil)

	// structs
	d = NewDict()
	d.Set(Class{"a","b"}, 1)
	d.Set(Class{"c","d"}, 2)
	d.Set(Ref{"a"}, 3)
	d.Set(tStructWithPrivate{"x","y"}, 4)
	assertData(Class{"a","b"},1, Class{"c","d"},2, Ref{"a"},3, tStructWithPrivate{"x","y"},4)
	assertGet(Class{"a","b"},               1)
	assertGet(Class{"c","d"},               2)
	assertGet(Class{"x","y"},               nil)
	assertGet(Ref{"a"},                     3)
	assertGet(Ref{"x"},                     nil)
	assertGet(tStructWithPrivate{"x","y"},  4)
	assertGet(tStructWithPrivate{"p","q"},  nil)

	// pointers
	i := 1
	j := 1
	k := 1
	x := Class{"a","b"}
	y := Class{"a","b"}
	z := Class{"a","b"}
	d = NewDict()
	d.Set(&i, 1)
	d.Set(&j, 2)
	d.Set(&x, 3)
	d.Set(&y, 4)
	assertData(&i,1, &j,2, &x,3, &y,4)
	assertGet(&i, 1)
	assertGet(&j, 2)
	assertGet(&k, nil)
	assertGet(&x, 3)
	assertGet(&y, 4)
	assertGet(&z, nil)

	// NewDictWithSizeHint
	d = NewDictWithSizeHint(100)
	assertData()
	assertGet(1,   nil)
	assertGet(2,   nil)
	assertGet("a", nil)
	assertGet("b", nil)

	// NewDictWithData
	d = NewDictWithData("a",1, 2,"b")
	assertData("a",1, 2,"b")
	assertGet(1,   nil)
	assertGet(2,   "b")
	assertGet("a", 1)
	assertGet("b", nil)

	// unhashable types
	vbad := []any{
		[]any{},
		[]any{1,2,3},
		[]int{},
		[]int{1,2,3},
		NewDict(),
		map[any]any{},
		map[int]bool{},
		Ref{[]any{}},
		tStructWithPrivate{1,[]any{}},
		tStructWithPrivate{[]any{},1},
		tStructWithPrivate{[]any{},[]any{}},
	}

	assertPanics := func(subj any, errPrefix string, f func()) {
		t.Helper()
		defer func() {
			t.Helper()
			r := recover()
			if r == nil {
				t.Errorf("%#v: no panic", subj)
				return
			}
			s, ok := r.(string)
			if ok && strings.HasPrefix(s, errPrefix) {
				// ok
			} else {
				panic(r)
			}

		}()

		f()
	}

	for _, k := range vbad {
		assertUnhashable := func(f func()) {
			t.Helper()
			assertPanics(k, "unhashable type: ", f)
		}

		assertUnhashable(func() { d.Get(k) })
		assertUnhashable(func() { d.Set(k, 1) })
		assertUnhashable(func() { d.Del(k) })
		assertUnhashable(func() { NewDictWithData(k,1) })
	}

	// = ~nil
	d = Dict{}
	assertData()
	assertGet(1,   nil)
	assertGet(2,   nil)
	assertGet("a", nil)
	assertGet("b", nil)
	d.Del(1)
	assertData()
	assertGet(1,   nil)
	assertGet(2,   nil)
	assertGet("a", nil)
	assertGet("b", nil)

	assertPanics("nil.Set", "Set called on nil map", func() { d.Set(1, "x") })
}


// benchmarks for map and Dict compare them from performance point of view.

func BenchmarkMapGet(b *testing.B) {
	m := map[any]any{}
	for i := 0; i < 100; i++ {
		m[i] = i
	}
	m["abc"] = 777

	b.Run("string", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = m["abc"]
		}
	})

	b.Run("int", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = m[77]
		}
	})
}

func BenchmarkDictGet(b *testing.B) {
	d := NewDict()
	for i := 0; i < 100; i++ {
		d.Set(i, i)
	}
	d.Set("abc", 777)

	b.Run("string", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = d.Get("abc")
		}
	})

	b.Run("int", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = d.Get(77)
		}
	})
}
