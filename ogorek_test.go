package ogórek

import (
	"bytes"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"math/big"
	"strconv"
	"strings"
	"testing"
)

func bigInt(s string) *big.Int {
	i := new(big.Int)
	_, ok := i.SetString(s, 10)
	if !ok {
		panic("bigInt")
	}
	return i
}

func TestMarker(t *testing.T) {
	buf := bytes.Buffer{}
	dec := NewDecoder(&buf)
	dec.mark()
	k, err := dec.marker()
	if err != nil {
		t.Error(err)
	}
	if k != 0 {
		t.Error("no marker found")
	}
}

// hexInput decodes hex-encoded data into input TestPickle.
// it panics on decode errors.
func hexInput(hexdata string) TestPickle {
	data, err := hex.DecodeString(hexdata)
	if err != nil {
		panic(err)
	}
	return I(string(data))
}

var graphitePickle1 = hexInput("80025d71017d710228550676616c75657371035d71042847407d90000000000047407f100000000000474080e0000000000047409764000000000047409c40000000000047409d88000000000047409f74000000000047409c74000000000047409cdc00000000004740a10000000000004740a0d800000000004740938800000000004740a00e00000000004740988800000000004e4e655505737461727471054a00d87a5255047374657071064a805101005503656e6471074a00f08f5255046e616d657108552d5a5a5a5a2e55555555555555552e43434343434343432e4d4d4d4d4d4d4d4d2e5858585858585858582e545454710975612e")
var graphiteObject1 = []any{map[any]any{"values": []any{float64(473), float64(497), float64(540), float64(1497), float64(1808), float64(1890), float64(2013), float64(1821), float64(1847), float64(2176), float64(2156), float64(1250), float64(2055), float64(1570), None{}, None{}}, "start": int64(1383782400), "step": int64(86400), "end": int64(1385164800), "name": "ZZZZ.UUUUUUUU.CCCCCCCC.MMMMMMMM.XXXXXXXXX.TTT"}}

var graphitePickle2 = hexInput("286c70300a286470310a53277374617274270a70320a49313338333738323430300a73532773746570270a70330a4938363430300a735327656e64270a70340a49313338353136343830300a73532776616c756573270a70350a286c70360a463437332e300a61463439372e300a61463534302e300a6146313439372e300a6146313830382e300a6146313839302e300a6146323031332e300a6146313832312e300a6146313834372e300a6146323137362e300a6146323135362e300a6146313235302e300a6146323035352e300a6146313537302e300a614e614e617353276e616d65270a70370a5327757365722e6c6f67696e2e617265612e6d616368696e652e6d65747269632e6d696e757465270a70380a73612e")
var graphiteObject2 = []any{map[any]any{"values": []any{float64(473), float64(497), float64(540), float64(1497), float64(1808), float64(1890), float64(2013), float64(1821), float64(1847), float64(2176), float64(2156), float64(1250), float64(2055), float64(1570), None{}, None{}}, "start": int64(1383782400), "step": int64(86400), "end": int64(1385164800), "name": "user.login.area.machine.metric.minute"}}

var graphitePickle3 = hexInput("286c70310a286470320a5327696e74657276616c73270a70330a286c70340a7353276d65747269635f70617468270a70350a5327636172626f6e2e6167656e7473270a70360a73532769734c656166270a70370a4930300a7361286470380a67330a286c70390a7367350a5327636172626f6e2e61676772656761746f72270a7031300a7367370a4930300a736128647031310a67330a286c7031320a7367350a5327636172626f6e2e72656c617973270a7031330a7367370a4930300a73612e")
var graphiteObject3 = []any{map[any]any{"intervals": []any{}, "metric_path": "carbon.agents", "isLeaf": false}, map[any]any{"intervals": []any{}, "metric_path": "carbon.aggregator", "isLeaf": false}, map[any]any{"intervals": []any{}, "metric_path": "carbon.relays", "isLeaf": false}}

const longLine = "28,34,30,55,100,130,87,169,194,202,232,252,267,274,286,315,308,221,358,368,401,406,434,452,475,422,497,530,517,559,400,418,571,578,599,600,625,630,635,647,220,715,736,760,705,785,794,495,808,852,861,863,869,875,890,893,896,922,812,980,1074,1087,1145,1153,1163,1171,445,1195,1203,1242,1255,1274,52,1287,1319,636,1160,1339,1345,1353,1369,1391,1396,1405,1221,1410,1431,1451,1460,1470,1472,1492,1517,1528,419,1530,1532,1535,1573,1547,1574,1437,1594,1595,847,1551,983,1637,1647,1666,1672,1691,1726,1515,1731,1739,1741,1723,1776,1685,505,1624,1436,1890,728,1910,1931,1544,2013,2025,2030,2043,2069,1162,2129,2160,2199,2210,1911,2246,804,2276,1673,2299,2315,2322,2328,2355,2376,2405,1159,2425,2430,2452,1804,2442,2567,2577,1167,2611,2534,1879,2623,2682,2699,2652,2742,2754,2774,2782,2795,2431,2821,2751,2850,2090,513,2898,592,2932,2933,1555,2969,3003,3007,3010,2595,3064,3087,3105,3106,3110,151,3129,3132,304,3173,3205,3233,3245,3279,3302,3307,714,316,3331,3347,3360,3375,3380,3442,2620,3482,3493,3504,3516,3517,3518,3533,3511,2681,3530,3601,3606,3615,1210,3633,3651,3688,3690,3781,1907,3839,3840,3847,3867,3816,3899,3924,2345,3912,3966,982,4040,4056,4076,4084,4105,2649,4171,3873,1415,3567,4188,4221,4227,4231,2279,4250,4253,770,894,4343,4356,4289,4404,4438,2572,3124,4334,2114,3953,4522,4537,4561,4571,641,4629,4640,4664,4687,4702,4709,4740,4605,4746,4768,3856,3980,4814,2984,4895,4908,1249,4944,4947,4979,4988,4995,32,4066,5043,4956,5069,5072,5076,5084,5085,5137,4262,5152,479,5156,3114,1277,5183,5186,1825,5106,5216,963,5239,5252,5218,5284,1980,1972,5352,5364,5294,5379,5387,5391,5397,5419,5434,5468,5471,3350,5510,5522,5525,5538,5554,5573,5597,5610,5615,5624,842,2851,5641,5655,5656,5658,5678,5682,5696,5699,5709,5728,5753,851,5805,3528,5822,801,5855,2929,5871,5899,5918,5925,5927,5931,5935,5939,5958,778,5971,5980,5300,6009,6023,6030,6032,6016,6110,5009,6155,6197,1760,6253,6267,4886,5608,6289,6308,6311,6321,6316,6333,6244,6070,6349,6353,6186,6357,6366,6386,6387,6389,6399,6411,6421,6432,6437,6465,6302,6493,5602,6511,6529,6536,6170,6557,6561,6577,6581,6590,5290,5649,6231,6275,6635,6651,6652,5929,6692,6693,6695,6705,6711,6723,6738,6752,6753,3629,2975,6790,5845,338,6814,6826,6478,6860,6872,6882,880,356,6897,4102,6910,6611,1030,6934,6936,6987,6984,6999,827,6902,7027,7049,7051,4628,7084,7083,7071,7102,7137,5867,7152,6048,2410,3896,7168,7177,7224,6606,7233,1793,7261,7284,7290,7292,5212,7315,6964,3238,355,1969,4256,448,7325,908,2824,2981,3193,3363,3613,5325,6388,2247,1348,72,131,5414,7285,7343,7349,7362,7372,7381,7410,7418,7443,5512,7470,7487,7497,7516,7277,2622,2863,945,4344,3774,1024,2272,7523,4476,256,5643,3164,7539,7540,7489,1932,7559,7575,7602,7605,7609,7608,7619,7204,7652,7663,6907,7672,7654,7674,7687,7718,7745,1202,4030,7797,7801,7799,2924,7871,7873,7900,7907,7911,7912,7917,7923,7935,8007,8017,7636,8084,8087,3686,8114,8153,8158,8171,8175,8182,8205,8222,8225,8229,8232,8234,8244,8247,7256,8279,6929,8285,7040,8328,707,6773,7949,8468,5759,6344,8509,1635"

// TestPickle represents a test pickle that ogórek encoder produces at particular protocols.
//
// If protov is empty there is no connection in between ogórek encoder and the
// data. However the test data can still be used to feed ogórek decoder.
type TestPickle struct {
	protov []int

	// pickle data without `PROTO <ver>` prefix.
	// optionally the prefix template (\x80\xff) could be given for cases
	// where initial `PROTO <ver>` presence affects decoding semantic.
	data   string

	err    error  // !nil if encoding should fail
}

// TestEntry represents one decode/encode test.
type TestEntry struct {
	name string

	// object(s) and []TestPickle. All pickles must decode to objectOut.
	// Encoding objectIn at particular protocol must give particular TestPickle.
	//
	// In the usual case objectIn == objectOut and they can differ if
	// e.g. objectIn contains a Go struct.
	objectIn  any
	picklev   []TestPickle
	objectOut any

	strictUnicodeN bool // whether to test with StrictUnicode=n while decoding/encoding
	strictUnicodeY bool // whether to test with StrictUnicode=y while decoding/encoding

	pyDictN bool // whether to test with PyDict=n while decoding/encoding
	pyDictY bool // ----//----           PyDict=y
}

// X, I, P0, P1, P* form a language to describe decode/encode tests:
//
// - X(name, object, ...) represents one test entry. All pickles from "..."
//   (see below) must decode to object. Encoding the object at particular
//   settings (e.g. at protocol=1 for P1 pickle) must give specified pickle data.
//
// - I denotes arbitrary input. Decoding it must produce the object.
//
// - P* denotes a TestPickle. Encoding the object at particular setting (e.g. P1
//   represents protocol=1, P1_ represents protocol >= 1) must give the pickle data.
//   Decoding the pickle data must give the object.

// X is syntactic sugar to prepare one TestEntry.
//
// the entry is tested under both StrictUnicode=n and StrictUnicode=y modes.
func X(name string, object any, picklev ...TestPickle) TestEntry {
	return TestEntry{name: name, objectIn: object, objectOut: object, picklev: picklev,
			 strictUnicodeN: true, strictUnicodeY: true,
			 pyDictN: true, pyDictY: true}
}

// Xuauto is syntactic sugar to prepare one TestEntry that is tested only under StrictUnicode=n mode.
func Xuauto(name string, object any, picklev ...TestPickle) TestEntry {
	x := X(name, object, picklev...)
	x.strictUnicodeY = false
	return x
}

// Xustrict is syntactic sugar to prepare one TestEntry that is tested only under StrictUnicode=y mode.
func Xustrict(name string, object any, picklev ...TestPickle) TestEntry {
	x := X(name, object, picklev...)
	x.strictUnicodeN = false
	return x
}

// Xdgo is syntactic sugar to prepare one TestEntry that is tested only under PyDict=n mode.
func Xdgo(name string, object any, picklev ...TestPickle) TestEntry {
	x := X(name, object, picklev...)
	x.pyDictY = false
	return x
}

// Xdpy is syntactic sugar to prepare one TestEntry that is tested only under PyDict=y mode.
func Xdpy(name string, object any, picklev ...TestPickle) TestEntry {
	x := X(name, object, picklev...)
	x.pyDictN = false
	return x
}

// Xuauto_dgo is syntactic sugar to prepare one TestEntry that is tested only
// under StrictUnicode=n ^ pyDict=n mode.
func Xuauto_dgo(name string, object any, picklev ...TestPickle) TestEntry {
	x := X(name, object, picklev...)
	x.strictUnicodeY = false
	x.pyDictY = false
	return x
}

// Xuauto_dpy is syntactic sugar to prepare one TestEntry that is tested only
// under StrictUnicode=n ^ pyDict=y mode.
func Xuauto_dpy(name string, object any, picklev ...TestPickle) TestEntry {
	x := X(name, object, picklev...)
	x.strictUnicodeY = false
	x.pyDictN = false
	return x
}

// Xloosy is syntactic sugar to prepare one TestEntry with loosy encoding.
func Xloosy(name string, objectIn, objectOut any, picklev ...TestPickle) TestEntry {
	x := X(name, objectIn, picklev...)
	x.objectOut = objectOut
	return x
}

// Xloosy_uauto_dgo is like Xuauto_dgo but for Xloosy.
func Xloosy_uauto_dgo(name string, objectIn, objectOut any, picklev ...TestPickle) TestEntry {
	x := Xuauto_dgo(name, objectIn, picklev...)
	x.objectOut = objectOut
	return x
}

func I(input string) TestPickle { return TestPickle{protov: nil, data: input, err: nil} }

// PP(protov) creates func PX(pickle) which in turn produces TestPickle{protocol: protov, pickle}.
func PP(protov ...int) func(xpickle any) TestPickle {
	return func(xpickle any) TestPickle {
		t := TestPickle{protov: protov}
		switch x := xpickle.(type) {
		case string:
			t.data = x
		case error:
			t.err = x

		default:
			panic(fmt.Sprintf("P* accept only string|error, not %T (%v)", xpickle, xpickle))
		}
		return t
	}
}

// PX  creates TestPickle with .protov={x} .
// PX_ creates TestPickle with .protov={x,x+1,...} .
var (
	P0 = PP(0)
	P1 = PP(1)
	P2 = PP(2)
	P3 = PP(3)
	P4 = PP(4)

	P01   = PP(0,1)
	P0123 = PP(0,1,2,3)
	P0_   = PP(0,1,2,3,4,5)
	P12   = PP(  1,2)
	P123  = PP(  1,2,3)
	P1_   = PP(  1,2,3,4,5)
	P23   = PP(    2,3)
	P2_   = PP(    2,3,4,5)
	P3_   = PP(      3,4,5)
	P4_   = PP(        4,5)
	P5_   = PP(          5)
)

// make sure we use test pickles in fuzz corpus
//go:generate go test -tags gofuzz -run TestFuzzGenerate

// tests is the main registry for decode/encode tests.
//
// NOTE whenever you change something here - don't forget to run `go generate`
// to export test pickles to fuzzing corpus.
// XXX or better instead of `go generate`, automatically dump all test pickles
// on every `go test` run?
var tests = []TestEntry{
	X("None", None{},
		P0_("N.")), // NONE

	X("True", true,
		P01("I01\n."), // INT 01
		P2_("\x88.")), // NEWTRUE

	X("False", false,
		P01("I00\n."), // INT 00
		P2_("\x89.")), // NEWFALSE

	X("int(0)", int64(0),
		P0("I0\n."),    // INT
		P1_("K\x00.")), // BININT1

	X("int(5)", int64(5),
		P0("I5\n."),    // INT
		P1_("K\x05.")), // BININT1

	X("int(0xff)", int64(0xff),
		P0("I255\n."),  // INT
		P1_("K\xff.")), // BININT1

	X("int(0x123)", int64(0x123),
		P0("I291\n."),      // INT
		P1_("M\x23\x01.")), // BININT2

	X("int(0xffff)", int64(0xffff),
		P0("I65535\n."),    // INT
		P1_("M\xff\xff.")), // BININT2

	X("int(0x12345)", int64(0x12345),
		P0("I74565\n."),            // INT
		P1_("J\x45\x23\x01\x00.")), // BININT

	X("int(0x7fffffff)", int64(0x7fffffff),
		P0("I2147483647\n."),       // INT
		P1_("J\xff\xff\xff\x7f.")), // BININT

	X("int(-7)", int64(-7),
		P0("I-7\n."),               // INT
		P1_("J\xf9\xff\xff\xff.")), // BININT

	X("int(-0x80000000)", int64(-0x80000000),
		P0("I-2147483648\n."),      // INT
		P1_("J\x00\x00\x00\x80.")), // BININT

	X("int(0x1234ffffffff)", int64(0x1234ffffffff),
		P0_("I20018842566655\n.")), // INT

	X("int(0x7fffffffffffffff)", int64(0x7fffffffffffffff),
		P0_("I9223372036854775807\n.")), // INT

	Xloosy("uint(0)", uint64(0), int64(0),
		P0("I0\n."),    // INT
		P1_("K\x00.")), // BININT

	Xloosy("uint(0x80000000ffffffff)", uint64(0x80000000ffffffff), bigInt("9223372041149743103"),
		P0_("I9223372041149743103\n.")), // INT

	Xloosy("uint(0xffffffffffffffff)", uint64(0xffffffffffffffff), bigInt("18446744073709551615"),
		P0_("I18446744073709551615\n.")), // INT

	X("float", float64(1.23),
		P0("F1.23\n."),                    // FLOAT
		P1_("G?\xf3\xae\x14z\xe1G\xae.")), // BINFLOAT

	X("long", bigInt("12321231232131231231"),
		P0("L12321231232131231231L\n."),                           // LONG
		I("\x8a\x09\xffm\xa1b\x86\xce\xfd\xaa\x00.")),             // LONG1
		//I("\x8b\x09\x00\x00\x00\xffm\xa1b\x86\xce\xfd\xaa\x00.")), // LONG4 TODO

	X("tuple()", Tuple{},
		P0("(t."),  // MARK + TUPLE
		P1_(").")), // EMPTY_TUPLE

	X("tuple((1,))", Tuple{int64(1)},
		P0("(I1\nt."),     // MARK + TUPLE + INT
		P1("(K\x01t."),    // MARK + TUPLE + BININT1
		P2_("K\x01\x85."), // TUPLE1 + BININT1
		I("I1\n\x85.")),   // TUPLE1 + INT

	X("tuple((1,2))", Tuple{int64(1), int64(2)},
		P0("(I1\nI2\nt."),      // MARK + TUPLE + INT
		P1("(K\x01K\x02t."),    // MARK + TUPLE + BININT1
		P2_("K\x01K\x02\x86."), // TUPLE2 + BININT1
		I("I1\nI2\n\x86.")),    // TUPLE2 + INT

	X("tuple((1,2,3))", Tuple{int64(1), int64(2), int64(3)},
		P0("(I1\nI2\nI3\nt."),       // MARK + TUPLE + INT
		P1("(K\x01K\x02K\x03t."),    // MARK + TUPLE + BININT1
		P2_("K\x01K\x02K\x03\x87."), // TUPLE3 + BININT1
		I("I1\nI2\nI3\n\x87.")),     // TUPLE3 + INT

	X("tuple(((1,2), (3,4)))", Tuple{Tuple{int64(1), int64(2)}, Tuple{int64(3), int64(4)}},
		P0("((I1\nI2\nt(I3\nI4\ntt."),            // MARK + INT + TUPLE
		P1("((K\x01K\x02t(K\x03K\x04tt."),        // MARK + BININT1 + TUPLE
		P2_("K\x01K\x02\x86K\x03K\x04\x86\x86."), // BININT1 + TUPLE2
		I("((I1\nI2\ntp0\n(I3\nI4\ntp1\ntp2\n.")),

	X("list([])", []any{},
		P0("(l."), // MARK + LIST
		P1_("]."), // EMPTY_LIST
		I("(lp0\n.")),

	X("list([1,2,3,True])", []any{int64(1), int64(2), int64(3), true},
		P0("(I1\nI2\nI3\nI01\nl."),    // MARK + INT + INT(True) + LIST
		P1("(K\x01K\x02K\x03I01\nl."), // MARK + BININT1 + INT(True) + LIST
		P2_("(K\x01K\x02K\x03\x88l."), // MARK + BININT1 + NEW_TRUE + LIST
		I("(lp0\nI1\naI2\naI3\naI01\na.")),

	// strings in default StrictUnicode=n mode

	Xuauto("str('abc')", "abc",
		P0("S\"abc\"\n."),           // STRING
		P12("U\x03abc."),            // SHORT_BINSTRING
		P3("X\x03\x00\x00\x00abc."), // BINUNICODE
		P4_("\x8c\x03abc."),         // SHORT_BINUNICODE
		I("T\x03\x00\x00\x00abc."),  // BINSTRING
		I("S'abc'\np0\n."),
		I("S'abc'\n.")),

	Xuauto("unicode('日本語')", "日本語",
		P0("S\"日本語\"\n."),                                 // STRING
		P12("U\x09日本語."),                                  // SHORT_BINSTRING
		P3("X\x09\x00\x00\x00日本語."),                       // BINUNICODE
		P4_("\x8c\x09\xe6\x97\xa5\xe6\x9c\xac\xe8\xaa\x9e."), // SHORT_BINUNICODE

		I("V\\u65e5\\u672c\\u8a9e\np0\n."),                           // UNICODE
		I("X\x09\x00\x00\x00\xe6\x97\xa5\xe6\x9c\xac\xe8\xaa\x9e.")), // BINUNICODE
		// TODO BINUNICODE8

	Xuauto("unicode('\\' 知事少时烦恼少、识人多处是非多。')", "' 知事少时烦恼少、识人多处是非多。",
		// UNICODE
		I("V' \\u77e5\\u4e8b\\u5c11\\u65f6\\u70e6\\u607c\\u5c11\\u3001\\u8bc6\\u4eba\\u591a\\u5904\\u662f\\u975e\\u591a\\u3002\n."),

		// BINUNICODE
		P3("X\x32\x00\x00\x00' \xe7\x9f\xa5\xe4\xba\x8b\xe5\xb0\x91\xe6\x97\xb6\xe7\x83\xa6\xe6\x81\xbc\xe5\xb0\x91\xe3\x80\x81\xe8\xaf\x86\xe4\xba\xba\xe5\xa4\x9a\xe5\xa4\x84\xe6\x98\xaf\xe9\x9d\x9e\xe5\xa4\x9a\xe3\x80\x82."),

		// SHORT_BINUNICODE
		P4_("\x8c\x32' \xe7\x9f\xa5\xe4\xba\x8b\xe5\xb0\x91\xe6\x97\xb6\xe7\x83\xa6\xe6\x81\xbc\xe5\xb0\x91\xe3\x80\x81\xe8\xaf\x86\xe4\xba\xba\xe5\xa4\x9a\xe5\xa4\x84\xe6\x98\xaf\xe9\x9d\x9e\xe5\xa4\x9a\xe3\x80\x82.")),

		// TODO BINUNICODE8

	// strings in StrictUnicode=y mode

	Xustrict("str('abc')", ByteString("abc"),
		P0("S\"abc\"\n."),           // STRING
		P1_("U\x03abc."),            // SHORT_BINSTRING
		I("T\x03\x00\x00\x00abc."),  // BINSTRING
		I("S'abc'\np0\n."),
		I("S'abc'\n.")),

	Xustrict("unicode('abc')", "abc",
		P0("Vabc\n."),                 // UNICODE
		P123("X\x03\x00\x00\x00abc."), // BINUNICODE
		P4_("\x8c\x03abc.")),          // SHORT_BINUNICODE
		// TODO BINUNICODE8

	Xustrict("str('日本語')", ByteString("日本語"),
		P0("S\"日本語\"\n."), // STRING
		P1_("U\x09日本語.")), // SHORT_BINSTRING

	Xustrict("unicode('日本語')", "日本語",
		P0("V\\u65e5\\u672c\\u8a9e\n."),                      // UNICODE
		P123("X\x09\x00\x00\x00日本語."),                     // BINUNICODE
		P4_("\x8c\x09\xe6\x97\xa5\xe6\x9c\xac\xe8\xaa\x9e."), // SHORT_BINUNICODE

		I("V\\u65e5\\u672c\\u8a9e\np0\n."),                           // UNICODE
		I("X\x09\x00\x00\x00\xe6\x97\xa5\xe6\x9c\xac\xe8\xaa\x9e.")), // BINUNICODE
		// TODO BINUNICODE8

	Xustrict("unicode(non-utf8)", "\x93",
		P0(errP0UnicodeUTF8Only),       // UNICODE cannot represent non-UTF8 sequences
		P123("X\x01\x00\x00\x00\x93."), // BINUNICODE
		P4_("\x8c\x01\x93.")),          // SHORT_BINUNICODE

	// str/unicode with many control characters at P0
	// this exercises escape-based STRING/UNICODE coding

	Xustrict(`str('\x80ми\nр\r\u2028\\u1234\\U00004321') # text escape`, ByteString("\x80ми\nр\r\u2028\\u1234\\U00004321"),
		P0("S\"\\x80ми\\nр\\r\\xe2\\x80\\xa8\\\\u1234\\\\U00004321\"\n."),
		I("S\"\\x80ми\\nр\\r\\xe2\\x80\\xa8\\u1234\\U00004321\"\n.")), // \u and \U not decoded

	Xustrict(`str("hel'lo")`, ByteString("hel'lo"), I("S'hel'lo'\n.")),      // non-escaped ' inside '-quotes
	Xustrict(`str("hel\"lo")`, ByteString("hel\"lo"), I("S\"hel\"lo\"\n.")), // non-escaped " inside "-quotes


	Xuauto(`unicode(r'мир\n\r\x00'+'\r') # text escape`, `мир\n\r\x00`+"\r",
		I("V\\u043c\\u0438\\u0440\\n\\r\\x00" + // only \u and \U are decoded - not \n \r ...
			"\r" +                          // raw \r - ok, not lost
			"\n.")),


	X(`bytes(b"hello\nмир\x01")`, Bytes("hello\nмир\x01"),
		// GLOBAL + MARK + UNICODE + STRING + TUPLE + REDUCE
		P0("c_codecs\nencode\n(Vhello\\u000aмир\x01\nS\"latin1\"\ntR."),

		// GLOBAL + MARK + BINUNICODE + SHORT_BINSTRING + TUPLE + REDUCE
		P1("c_codecs\nencode\n(X\x13\x00\x00\x00hello\n\xc3\x90\xc2\xbc\xc3\x90\xc2\xb8\xc3\x91\xc2\x80\x01U\x06latin1tR."),

		// GLOBAL + BINUNICODE + SHORT_BINSTRING + TUPLE2 + REDUCE
		P2("c_codecs\nencode\nX\x13\x00\x00\x00hello\n\xc3\x90\xc2\xbc\xc3\x90\xc2\xb8\xc3\x91\xc2\x80\x01U\x06latin1\x86R."),

		P3_("C\x0dhello\nмир\x01."),            // SHORT_BINBYTES
		I("B\x0d\x00\x00\x00hello\nмир\x01.")), // BINBYTES

	X(`bytearray(b"hello\nмир\x01")`, []byte("hello\nмир\x01"),
		// GLOBAL + MARK + UNICODE + STRING + TUPLE + REDUCE
		P0("c__builtin__\nbytearray\n(c_codecs\nencode\n(Vhello\\u000aмир\x01\nS\"latin1\"\ntRtR."),

		// GLOBAL + MARK + BINUNICODE + SHORT_BINSTRING + TUPLE + REDUCE
		P1("c__builtin__\nbytearray\n(c_codecs\nencode\n(X\x13\x00\x00\x00hello\nÐ¼Ð¸Ñ\xc2\x80\x01U\x06latin1tRtR."),

		// GLOBAL + BINUNICODE + SHORT_BINSTRING + TUPLE{2,1} + REDUCE
		P2("c__builtin__\nbytearray\nc_codecs\nencode\nX\x13\x00\x00\x00hello\nÐ¼Ð¸Ñ\xc2\x80\x01U\x06latin1\x86R\x85R."),

		// PROTO + GLOBAL + SHORT_BINBYTES + TUPLE1 + REDUCE
		P3("\x80\xffcbuiltins\nbytearray\nC\rhello\nмир\x01\x85R."),

		// PROTO + SHORT_BINUNICODE + STACK_GLOBAL + SHORT_BINBYTES + TUPLE1 + REDUCE
		P4("\x80\xff\x8c\x08builtins\x8c\tbytearray\x93C\rhello\nмир\x01\x85R."),

		// PROTO + BYTEARRAY8
		P5_("\x80\xff\x96\x0d\x00\x00\x00\x00\x00\x00\x00hello\nмир\x01."),

		// bytearray(text, encoding); GLOBAL + BINUNICODE + TUPLE + REDUCE
		I("c__builtin__\nbytearray\nq\x00(X\x13\x00\x00\x00hello\n\xc3\x90\xc2\xbc\xc3\x90\xc2\xb8\xc3\x91\xc2\x80\x01q\x01X\x07\x00\x00\x00latin-1q\x02tq\x03Rq\x04.")),

	// dicts in default PyDict=n mode

	Xdgo("dict({})", make(map[any]any),
		P0("(d."), // MARK + DICT
		P1_("}."), // EMPTY_DICT
		I("(dp0\n.")),

	Xuauto_dgo("dict({'a': '1'})", map[any]any{"a": "1"},
		P0("(S\"a\"\nS\"1\"\nd."),                     // MARK + STRING + DICT
		P12("(U\x01aU\x011d."),                        // MARK + SHORT_BINSTRING + DICT
		P3("(X\x01\x00\x00\x00aX\x01\x00\x00\x001d."), // MARK + BINUNICODE + DICT
		P4_("(\x8c\x01a\x8c\x011d.")),                 // MARK + SHORT_BINUNICODE + DICT

	Xuauto_dgo("dict({'a': '1', 'b': '2'})", map[any]any{"a": "1", "b": "2"},
		// map iteration order is not stable - test only decoding
		I("(S\"a\"\nS\"1\"\nS\"b\"\nS\"2\"\nd."), // P0: MARK + STRING + DICT
		I("(U\x01aU\x011U\x01bU\x012d."),         // P12: MARK + SHORT_BINSTRING + DICT

		// P3: MARK + BINUNICODE + DICT
		I("(X\x01\x00\x00\x00aX\x01\x00\x00\x001X\x01\x00\x00\x00bX\x01\x00\x00\x002d."),

		I("(\x8c\x01a\x8c\x011\x8c\x01b\x8c\x012d."), // P4_: MARK + SHORT_BINUNICODE + DICT
		I("(dS'a'\nS'1'\nsS'b'\nS'2'\ns."),           // MARK + DICT + STRING + SETITEM
		I("}(U\x01aU\x011U\x01bU\x012u."),            // EMPTY_DICT + MARK + SHORT_BINSTRING + SETITEMS
		I("(dp0\nS'a'\np1\nS'1'\np2\nsS'b'\np3\nS'2'\np4\ns.")),

	// dicts in PyDict=y mode

	Xdpy("dict({})", NewDict(),
		P0("(d."), // MARK + DICT
		P1_("}."), // EMPTY_DICT
		I("(dp0\n.")),

	Xuauto_dpy("dict({'a': '1'})", NewDictWithData("a","1"),
		P0("(S\"a\"\nS\"1\"\nd."),                     // MARK + STRING + DICT
		P12("(U\x01aU\x011d."),                        // MARK + SHORT_BINSTRING + DICT
		P3("(X\x01\x00\x00\x00aX\x01\x00\x00\x001d."), // MARK + BINUNICODE + DICT
		P4_("(\x8c\x01a\x8c\x011d.")),                 // MARK + SHORT_BINUNICODE + DICT

	Xuauto_dpy("dict({'a': '1', 'b': '2'})", NewDictWithData("a","1", "b","2"),
		// map iteration order is not stable - test only decoding
		I("(S\"a\"\nS\"1\"\nS\"b\"\nS\"2\"\nd."), // P0: MARK + STRING + DICT
		I("(U\x01aU\x011U\x01bU\x012d."),         // P12: MARK + SHORT_BINSTRING + DICT

		// P3: MARK + BINUNICODE + DICT
		I("(X\x01\x00\x00\x00aX\x01\x00\x00\x001X\x01\x00\x00\x00bX\x01\x00\x00\x002d."),

		I("(\x8c\x01a\x8c\x011\x8c\x01b\x8c\x012d."), // P4_: MARK + SHORT_BINUNICODE + DICT
		I("(dS'a'\nS'1'\nsS'b'\nS'2'\ns."),           // MARK + DICT + STRING + SETITEM
		I("}(U\x01aU\x011U\x01bU\x012u."),            // EMPTY_DICT + MARK + SHORT_BINSTRING + SETITEMS
		I("(dp0\nS'a'\np1\nS'1'\np2\nsS'b'\np3\nS'2'\np4\ns.")),

	Xdpy("dict({123L: 0})", NewDictWithData(bigInt("123"), int64(0)),
		P0("(L123L\nI0\nd."),    // MARK + LONG + INT + DICT
		P1("(L123L\nK\x00d."),   // MARK + LONG + BININT1 + DICT
		I("(\x8a\x01{K\x00d.")), // MARK + LONG1 + BININT1 + DICT

	Xdpy("dict(tuple(): 0)", NewDictWithData(Tuple{}, int64(0)),
		P0("((tI0\nd."),   // MARK + MARK + TUPLE + INT + DICT
		P1_("()K\x00d.")), // MARK + EMPTY_TUPLE + BININT1 + DICT

	Xdpy("dict(tuple(1,2): 0)", NewDictWithData(Tuple{int64(1), int64(2)}, int64(0)),
		P0("((I1\nI2\ntI0\nd."),        // MARK + MARK + INT + INT + TUPLE + INT + DICT
		P1("((K\x01K\x02tK\x00d."),     // MARK + MARK + BININT1 + BININT1 + TUPLE + BININT1 + DICT
		P2_("(K\x01K\x02\x86K\x00d.")), // MARK + BININT1 + BININT1 + TUPLE2 + BININT1 + DICT


	Xuauto("foo.bar  # global", Class{Module: "foo", Name: "bar"},
		P0123("cfoo\nbar\n."),              // GLOBAL
		P4_("\x8c\x03foo\x8c\x03bar\x93."), // SHORT_BINUNICODE + STACK_GLOBAL
		I("S'foo'\nS'bar'\n\x93.")),        // STRING + STACK_GLOBAL

	X("foo\n2.bar  # global with \\n", Class{Module: "foo\n2", Name: "bar"},
		P0123(errP0123GlobalStringLineOnly),
		P4_("\x8c\x05foo\n2\x8c\x03bar\x93.")), // SHORT_BINUNICODE + STACK_GLOBAL

	Xuauto(`foo.bar("bing")  # global + reduce`, Call{Callable: Class{Module: "foo", Name: "bar"}, Args: []any{"bing"}},
		P0("cfoo\nbar\n(S\"bing\"\ntR."),                     // GLOBAL + MARK + STRING + TUPLE + REDUCE
		P1("cfoo\nbar\n(U\x04bingtR."),                       // GLOBAL + MARK + SHORT_BINSTRING + TUPLE + REDUCE
		P2("cfoo\nbar\nU\x04bing\x85R."),                     // GLOBAL + SHORT_BINSTRING + TUPLE1 + REDUCE
		P3("cfoo\nbar\nX\x04\x00\x00\x00bing\x85R."),         // GLOBAL + BINUNICODE + TUPLE1 + REDUCE
		P4_("\x8c\x03foo\x8c\x03bar\x93\x8c\x04bing\x85R.")), // SHORT_BINUNICODE + STACK_GLOBAL + TUPLE1 + REDUCE

	Xuauto(`persref("abc")`, Ref{"abc"},
		P0("Pabc\n."),                // PERSID
		P12("U\x03abcQ."),            // SHORT_BINSTRING + BINPERSID
		P3("X\x03\x00\x00\x00abcQ."), // BINUNICODE + BINPERSID
		P4_("\x8c\x03abcQ.")),        // SHORT_BINUNICODE + BINPERSID

	Xuauto(`persref("abc\nd")`, Ref{"abc\nd"},
		P0(errP0PersIDStringLineOnly),   // cannot be encoded
		P12("U\x05abc\ndQ."),            // SHORT_BINSTRING + BINPERSID
		P3("X\x05\x00\x00\x00abc\ndQ."), // BINUNICODE + BINPERSID
		P4_("\x8c\x05abc\ndQ.")),        // SHORT_BINUNICODE + BINPERSID

	X(`persref((1, 2))`, Ref{Tuple{int64(1), int64(2)}},
		P0(errP0PersIDStringLineOnly), // cannot be encoded
		P1("(K\x01K\x02tQ."),          // MARK + BININT1 + TUPLE + BINPERSID
		P2_("K\x01K\x02\x86Q."),       // BININT1 + TUPLE2 + BINPERSID
		I("(I1\nI2\ntQ.")),

	// decode only
	// TODO PUT + GET + BINGET + LONG_BINGET
	X("LONG_BINPUT", []any{int64(17)},
		I("(lr0000I17\na.")),

	Xuauto_dgo("graphite message1", graphiteObject1, graphitePickle1),
	Xuauto_dgo("graphite message2", graphiteObject2, graphitePickle2),
	Xuauto_dgo("graphite message3", graphiteObject3, graphitePickle3),
	Xuauto("too long line", longLine, I("V" + longLine + "\n.")),

	// opcodes from protocol 4

	X("FRAME opcode", int64(5),
		I("\x95\x00\x00\x00\x00\x00\x00\x00\x00I5\n.")), // FRAME is just skipped

	// loosy encode: decoding back gives another object.
	// the only case where ogórek encoding is loosy is for Go struct types.
	Xloosy_uauto_dgo("[]ogórek.foo{\"Qux\", 4}", []foo{{"Qux", 4}},
		[]any{map[any]any{"Foo": "Qux", "Bar": int64(4)}},

		// MARK + STRING + INT + DICT + LIST
		P0("((S\"Foo\"\nS\"Qux\"\nS\"Bar\"\nI4\ndl."),

		// MARK + SHORT_BINSTRING + BININT1 + DICT + LIST
		P12("((U\x03FooU\x03QuxU\x03BarK\x04dl."),

		// MARK + BINUNICODE + BININT1 + DICT + LIST
		P3("((X\x03\x00\x00\x00FooX\x03\x00\x00\x00QuxX\x03\x00\x00\x00BarK\x04dl."),

		// MARK + SHORT_BINUNICODE + BININT1 + DICT + LIST
		P4_("((\x8c\x03Foo\x8c\x03Qux\x8c\x03BarK\x04dl.")),
}

// foo is a type to test how encoder handles Go structs.
type foo struct {
	Foo string
	Bar int32
}

// if test pickle starts from protoPrefixTemplate, this prefix is changed to
// concrete `PROTO ver` when checking decoding. When checking encoding the
// protocol prefix is always automatically prepended and is always concrete.
var protoPrefixTemplate = string([]byte{opProto, 0xff})

// WithEachMode runs f under all decoding/encoding modes covered by test entry.
func (test TestEntry) WithEachMode(t *testing.T, f func(t *testing.T, decConfig DecoderConfig, encConfig EncoderConfig)) {
	for _, pyDict := range []bool{false, true} {
		if pyDict && !test.pyDictY {
			continue
		}
		if !pyDict && !test.pyDictN {
			continue
		}

		for _, strictUnicode := range []bool{false, true} {
			if  strictUnicode && !test.strictUnicodeY {
				continue
			}
			if !strictUnicode && !test.strictUnicodeN {
				continue
			}

			t.Run(fmt.Sprintf("%s/PyDict=%s/StrictUnicode=%s", test.name, yn(pyDict), yn(strictUnicode)),
			      func(t *testing.T) {
				decConfig := DecoderConfig{
					PyDict:        pyDict,
					StrictUnicode: strictUnicode,
				}
				encConfig := EncoderConfig{
					// no PyDict setting for encoder
					StrictUnicode: strictUnicode,
				}
				f(t, decConfig, encConfig)
			})

		}
	}
}

// TestDecode verifies ogórek decoder.
func TestDecode(t *testing.T) {
	for _, test := range tests {
		test.WithEachMode(t, func(t *testing.T, decConfig DecoderConfig, encConfig EncoderConfig) {
			for _, pickle := range test.picklev {
				if pickle.err != nil {
					continue
				}

				if strings.HasPrefix(pickle.data, protoPrefixTemplate) {
					// test case asked to have concrete `PROTO ver` prefix.
					// let's range over all pickle's protocols.
					for _, proto := range pickle.protov {
						data := string([]byte{opProto, byte(proto)}) +
							pickle.data[len(protoPrefixTemplate):]

						t.Run(fmt.Sprintf("%q/proto=%d", data, proto), func(t *testing.T) {
							testDecode(t, decConfig, test.objectOut, data)
						})
					}
				} else {
					t.Run(fmt.Sprintf("%q", pickle.data), func(t *testing.T) {
						testDecode(t, decConfig, test.objectOut, pickle.data)
					})
				}
			}
		})
	}
}

// TestEncode verifies ogórek encoder.
func TestEncode(t *testing.T) {
	for _, test := range tests {
		test.WithEachMode(t, func(t *testing.T, decConfig DecoderConfig, encConfig EncoderConfig) {
			alreadyTested := make(map[int]bool) // protocols we tested encode with so far
			for _, pickle := range test.picklev {
				for _, proto := range pickle.protov {
					dataOk := strings.TrimPrefix(pickle.data, protoPrefixTemplate)
					// protocols >= 2 must include "PROTO <ver>" prefix
					if proto >= 2 && pickle.err == nil {
						dataOk = string([]byte{opProto, byte(proto)}) + dataOk
					}

					t.Run(fmt.Sprintf("proto=%d", proto), func(t *testing.T) {
						testEncode(t, proto, encConfig, decConfig, test.objectIn, test.objectOut, dataOk, pickle.err)
					})

					alreadyTested[proto] = true
				}
			}

			// test encode-decode roundtrip on not yet tested protocols
			for proto := 0; proto <= highestProtocol; proto++ {
				if alreadyTested[proto] {
					continue
				}

				t.Run(fmt.Sprintf("proto=%d(roundtrip)", proto), func(t *testing.T) {
					testEncode(t, proto, encConfig, decConfig, test.objectIn, test.objectOut, "", nil)
				})
			}
		})
	}
}

// testDecode decodes input and verifies it is == object.
//
// It also verifies decoder robustness - via feeding it various kinds of
// corrupt data derived from input.
func testDecode(t *testing.T, decConfig DecoderConfig, object any, input string) {
	newDecoder := func(r io.Reader) *Decoder {
		return NewDecoderWithConfig(r, &decConfig)
	}

	// decode(input) -> expected
	buf := bytes.NewBufferString(input)
	dec := newDecoder(buf)
	v, err := dec.Decode()
	if err != nil {
		t.Error(err)
	}

	if !deepEqual(v, object) {
		t.Errorf("decode:\nhave: %#v\nwant: %#v", v, object)
	}

	// decode more -> EOF
	v, err = dec.Decode()
	if !(v == nil && err == io.EOF) {
		t.Errorf("decode: no EOF at end: v = %#v  err = %#v", v, err)
	}

	// decode(truncated input) -> must return io.ErrUnexpectedEOF
	for l := len(input) - 1; l > 0; l-- {
		buf := bytes.NewBufferString(input[:l])
		dec := newDecoder(buf)
		v, err := dec.Decode()
		if !(v == nil && err == io.ErrUnexpectedEOF) {
			t.Errorf("no ErrUnexpectedEOF on [:%d] truncated stream: v = %#v  err = %#v", l, v, err)
		}
	}

	// decode(input with omitted prefix) - tests how code handles pickle stack overflow:
	// it must not panic.
	for i := 0; i < len(input); i++ {
		buf := bytes.NewBufferString(input[i:])
		dec := newDecoder(buf)
		func() {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("panic on input[%d:]: %v", i, r)
				}
			}()
			dec.Decode()
		}()
	}
}

// testEncode encodes object using proto for pickle protocol, and verifies the result == dataOk.
//
// It also verifies that encoder handles write errors via using it on all kinds
// of limited writers. The data, that encoder produces, must decode back to
// expected object.
//
// If dataOk == "" no `result == dataOk` check is done, but encoding + followup
// encode-back tests are still performed.
//
// If errOk != nil, object encoding must produce that error.
func testEncode(t *testing.T, proto int, encConfig EncoderConfig, decConfig DecoderConfig, object, objectDecodedBack any, dataOk string, errOk error) {
	newEncoder := func(w io.Writer) *Encoder {
		econf := EncoderConfig{}
		econf = encConfig
		econf.Protocol = proto
		return NewEncoderWithConfig(w, &econf)
	}

	newDecoder := func(r io.Reader) *Decoder {
		return NewDecoderWithConfig(r, &decConfig)
	}

	buf := &bytes.Buffer{}
	enc := newEncoder(buf)

	// encode(object) == expected data
	err := enc.Encode(object)
	if errOk != nil {
		if err != errOk {
			t.Errorf("encode: expected error:\nhave: %#v\nwant: %#v", err, errOk)
		}
		return
	}

	if err != nil {
		t.Fatalf("encode error: %s", err)
	}
	data := buf.String()
	if dataOk != "" && data != dataOk {
		t.Errorf("encode:\nhave: %s\nwant: %s", pyquote(data), pyquote(dataOk))
	}

	// encode | limited writer -> write error
	for l := int64(len(data))-1; l >= 0; l-- {
		buf.Reset()
		enc = newEncoder(LimitWriter(buf, l))

		err = enc.Encode(object)
		if err != io.EOF {
			t.Errorf("encoder did not handle write error @%d: got %#v", l, err)
		}
	}

	// decode(encode(object)) == object
	dec := newDecoder(bytes.NewBufferString(data))
	v, err := dec.Decode()
	if err != nil {
		t.Errorf("encode -> decode -> error: %s", err)
	} else {
		if !deepEqual(v, objectDecodedBack) {
			what := "identity"
			if !deepEqual(object, objectDecodedBack) {
				what = "expected object"
			}
			t.Errorf("encode -> decode != %s\nhave: %#v\nwant: %#v", what, v, objectDecodedBack)
		}
	}
}


// test that .Decode() decodes only until stop opcode, and can continue
// decoding further on next call
func TestDecodeMultiple(t *testing.T) {
	input := "I5\n.I7\n.N."
	expected := []any{int64(5), int64(7), None{}}

	buf := bytes.NewBufferString(input)
	dec := NewDecoder(buf)

	for i, objOk := range expected {
		obj, err := dec.Decode()
		if err != nil {
			t.Errorf("step #%v: %v", i, err)
		}

		if !deepEqual(obj, objOk) {
			t.Errorf("step #%v: %q  ; want %q", i, obj, objOk)
		}
	}

	obj, err := dec.Decode()
	if !(obj == nil && err == io.EOF) {
		t.Errorf("decode: no EOF at end: obj = %#v  err = %#v", obj, err)
	}
}

func TestDecodeLong(t *testing.T) {
	var testv = []struct {
		data  string
		value int64 // converted to big.Int by test driver
	}{
		{"", 0},
		{"\xff\x00", 255},
		{"\xff\x7f", 32767},
		{"\x00\xff", -256},
		{"\x00\x80", -32768},
		{"\x80", -128},
		{"\x7f", 127},
	}

	for _, tt := range testv {
		value, err := decodeLong(tt.data)
		if err != nil {
			t.Errorf("data %q: %s", tt.data, err)
			continue
		}

		valueOk := big.NewInt(tt.value)
		if valueOk.Cmp(value) != 0 {
			t.Errorf("data %q: ->long: got %s  ; want %s", tt.data, value, valueOk)
		}
	}
}

func BenchmarkDecodeLong(b *testing.B) {
	for i := 0; i < b.N; i++ {
		data := "\x00\x80"
		_, err := decodeLong(data)
		if err != nil {
			b.Errorf("Error from decodeLong - %v\n", err)
		}
	}
}

func TestMemoOpCode(t *testing.T) {
	buf := bytes.NewBufferString("I5\n\x94.")
	dec := NewDecoder(buf)
	_, err := dec.Decode()
	if err != nil {
		t.Errorf("Error from TestMemoOpCode - %v\n", err)
	}
	if dec.memo["0"] != int64(5) {
		t.Errorf("Error from TestMemoOpCode - Top stack value was not added to memo")
	}

}

// verify that decode of erroneous input produces error
func TestDecodeError(t *testing.T) {
	testv := []string{
		// all kinds of opcodes to read memo but key is not there
		"}g1\n.",
		"}h\x01.",
		"}j\x01\x02\x03\x04.",

		// invalid long format
		"L123\n.",
		"L12qL\n.",

		// invalid protocol version
		"\x80\xffI1\n.",

		// BINSTRING and BINUNICODE with big len and no data
		// (might cause out-of-memory DOS if buffer is preallocated blindly)
		"T\xff\xff\xff\xff.",
		"X\xff\xff\xff\xff.",


		// it is invalid to expose mark object
		"(.",                        // MARK
		"(\x85.",                    // MARK + TUPLE1
		"((\x86.",                   // MARK·2 + TUPLE2
		"(((\x87.",                  // MARK·3 + TUPLE3
		"](a.",                      // EMPTY_LIST + MARK + APPEND
		"(p0\n0g0\nt.",              // MARK + PUT + POP + GET + TUPLE
		"(q\x000g0\nt.",             // MARK + BINPUT + ...
		"(r\x00\x00\x00\x000g0\nt.", // MARK + LONG_BINPUT + ...
		"(\x940g0\nt.",              // MARK + MEMOIZE + ...
		"}I1\n(s.",                  // EMPTY_DICT + INT + MARK + SETITEM
		"}(I1\ns.",                  // EMPTY_DICT + MARK + INT + SETITEM
		"(Q.",                       // MARK + BINPERSID

		// \r\n should not be read as combined EOL - only \n is
		"L123L\r\n.",
		"S'abc'\r\n.",

		// out-of-band data  (TODO might consider to add support for it in the future)
		"\x97.", // NEXT_BUFFER
		"\x98.", // READONLY_BUFFER
	}
	for _, tt := range testv {
		buf := bytes.NewBufferString(tt)
		dec := NewDecoder(buf)
		v, err := dec.Decode()
		if !(v == nil && err != nil) {
			t.Errorf("%q: no decode error  ; got %#v, %#v", tt, v, err)
		}
	}
}

// verify how decoder/encoder handle application-level settings wrt Refs.
func TestPersistentRefs(t *testing.T) {
	// ZBTree mimics BTree from ZODB.
	type ZBTree struct {
		oid string
	}

	errInvalidRef := errors.New("invalid reference")

	// Ref -> ? object
	loadref := func(ref Ref) (any, error) {
		// pretend we handle "zodb.BTree" -> ZBTree.
		t, ok := ref.Pid.(Tuple)
		if !ok || len(t) != 2 {
			return nil, errInvalidRef
		}

		class, ok1 := t[0].(Class)
		oid, ok2   := t[1].(string)
		if !(ok1 && ok2) {
			return nil, errInvalidRef
		}

		switch class {
		case Class{Module: "zodb", Name: "BTree"}:
			return &ZBTree{oid}, nil

		default:
			// leave it as is
			return nil, nil
		}
	}

	// object -> ? Ref
	getref := func(obj any) *Ref {
		// pretend we handle ZBTree.
		switch obj := obj.(type) {
		default:
			return nil

		case *ZBTree:
			return &Ref{Pid: Tuple{Class{Module: "zodb", Name: "BTree"}, obj.oid}}
		}
	}

	dconf := &DecoderConfig{PersistentLoad: loadref}
	econf := &EncoderConfig{PersistentRef: getref, Protocol: 1}

	testv := []struct {
		input    string
		expected any
	}{
		{"Pabc\n.", errInvalidRef},
		{"\x80\x01S'abc'\nQ.", errInvalidRef},
		{"\x80\x01S'abc'\nS'123'\n\x86Q.", errInvalidRef},
		{"\x80\x01cfoo\nbar\nS'123'\n\x86Q.", Ref{Tuple{Class{Module: "foo", Name: "bar"}, "123"}}},
		{"\x80\x01czodb\nBTree\nS'123'\n\x86Q.", &ZBTree{oid: "123"}},
	}

	for _, tt := range testv {
		// decode(input) -> expected
		buf := bytes.NewBufferString(tt.input)
		dec := NewDecoderWithConfig(buf, dconf)
		v, err := dec.Decode()

		expected := tt.expected
		errExpect := ""
		if e, iserr := expected.(error); iserr {
			expected = nil
			errExpect = "pickle: handleRef: " + e.Error()
		}

		if !(deepEqual(v, expected) &&
			((err == nil && errExpect == "") || err.Error() == errExpect)) {
			t.Errorf("%q: decode -> %#v, %q; want %#v, %q",
				tt.input, v, err, expected, errExpect)
		}

		if err != nil {
			continue
		}

		// expected -> encode -> decode = identity
		buf.Reset()
		enc := NewEncoderWithConfig(buf, econf)
		err = enc.Encode(tt.expected)
		if err != nil {
			t.Errorf("%q: encode(expected) -> %q", tt.input, err)
			continue
		}

		dec = NewDecoderWithConfig(buf, dconf)
		v, err = dec.Decode()
		if err != nil {
			t.Errorf("%q: expected -> encode -> decode: %q", tt.input, err)
			continue
		}

		if !deepEqual(v, tt.expected) {
			t.Errorf("%q: expected -> encode -> decode != identity\nhave: %#v\nwant: %#v",
				tt.input, v, tt.expected)
		}
	}
}

func TestFuzzCrashers(t *testing.T) {
	crashers := []string{
		"(dS''\n(lc\n\na2a2a22aasS''\na",
		"S\n",
		"((dd",
		"}}}s",
		"(((ld",
		"(dS''\n(lp4\nsg4\n(s",
		"}((tu",
		"}((du",
		"(c\n\nc\n\n\x85Rd",
		"}(U\x040000u",
		"(\x88d",
		"(]QNd.",          // PersID([])      -> dict
		"}]QNs.",          // PersID([])      -> setitem
		"}(]QNI1\nNu.",    // PersID([]) ...  -> setitems
		"\x960000000\xef", // BYTEARRAY8
	}

	for _, c := range crashers {
		buf := bytes.NewBufferString(c)
		dec := NewDecoder(buf)
		dec.Decode()
	}
}

func BenchmarkDecode(b *testing.B) {
	// prepare one large pickle stream from all test pickles
	input := make([]byte, 0)
	npickle := 0
	for _, test := range tests {
		for _, pickle := range test.picklev {
			if pickle.err != nil {
				continue
			}
			// not prepending `PROTO <ver>` - decoder should be
			// able to decode without it. But if the pickle already
			// comes with `PROTO 0xff` change it to `PROTO 3`.
			data := pickle.data
			if strings.HasPrefix(data, protoPrefixTemplate) {
				data = string([]byte{opProto, 3}) + data[len(protoPrefixTemplate):]
			}
			input = append(input, data...)
			npickle++
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf := bytes.NewBuffer(input)
		dec := NewDecoderWithConfig(buf, &DecoderConfig{
			PyDict: true, // so that decoding e.g. {(): 0} does not fail
		})

		j := 0
		for ; ; j++ {
			_, err := dec.Decode()
			if err != nil {
				if err == io.EOF {
					break
				}
				b.Fatal(err)
			}
		}

		if j != npickle {
			b.Fatalf("unexpected # of decode steps: got %v  ; want %v", j, npickle)
		}
	}
}

func BenchmarkEncode(b *testing.B) {
	// prepare one large slice from all test vector values
	input := make([]any, 0)
	approxOutSize := 0
	for _, test := range tests {
		if test.picklev[0].err == nil {
			input = append(input, test.objectIn)
			approxOutSize += len(test.picklev[0].data)
		}
	}

	buf := bytes.NewBuffer(make([]byte, approxOutSize))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf.Reset()
		enc := NewEncoder(buf)
		err := enc.Encode(input)
		if err != nil {
			b.Fatal(err)
		}
	}
}


var misquotedChars = []struct{
	in  string
	err error
}{
	{`\000`, nil}, // nil mean unquoteChar should be ok -> test for io.ErrUnexpectedEOF
	{`\x00`, nil}, // on truncated input
	{`\u0000`, nil},
	{`\U00000000`, nil},

	{`"`, strconv.ErrSyntax},
	{`\'`, strconv.ErrSyntax},
	{`\q`, strconv.ErrSyntax},
	{`\z`, strconv.ErrSyntax},
	{`\008`, strconv.ErrSyntax},
	{`\400`, strconv.ErrSyntax},
	{`\x0z`, strconv.ErrSyntax},
	{`\u000z`, strconv.ErrSyntax},
	{`\U0000000z`, strconv.ErrSyntax},
	{`\U12345678`, strconv.ErrSyntax},
}


// verify that our unquoteChar properly returns ErrUnexpectedEOF instead of ErrSyntax.
func TestUnquoteCharEOF(t *testing.T) {
	for _, tt := range misquotedChars {
		_, _, _, err := unquoteChar(tt.in, '"')
		if err != tt.err {
			t.Errorf("unquoteChar(%#q) -> err = %v want %v", tt.in, err, tt.err)
		}

		if tt.err != nil {
			continue
		}

		// truncated valid input should result in unexpected EOF
		for l := len(tt.in) - 1; l >= 0; l-- {
			_, _, _, err2 := unquoteChar(tt.in[:l], '"')
			if err2 != io.ErrUnexpectedEOF {
				t.Errorf("unquoteChar(%#q) -> err = %v want %v", tt.in[:l], err2, io.ErrUnexpectedEOF)
			}
		}
	}
}

func TestStringsFmt(t *testing.T) {
	tvhash := []struct{
		in      any
		vhashok string
	}{
		{"мир",             `"мир"`},
		{Bytes("мир"),      `ogórek.Bytes("мир")`},
		{ByteString("мир"), `ogórek.ByteString("мир")`},
		{unicode("мир"),    `ogórek.unicode("мир")`},
	}

	for _, tt := range tvhash {
		vhash := fmt.Sprintf("%#v", tt.in)
		if vhash != tt.vhashok {
			t.Errorf("%T %q: %%#v:\nhave: %s\nwant: %s", tt.in, tt.in, vhash, tt.vhashok)
		}
	}
}

// like io.LimitedReader but for writes
// XXX it would be good to have it in stdlib
type LimitedWriter struct {
	W io.Writer
	N int64
}

func (l *LimitedWriter) Write(p []byte) (n int, err error) {
	if l.N <= 0 {
		return 0, io.EOF
	}
	if int64(len(p)) > l.N {
		p = p[0:l.N]
	}
	n, err = l.W.Write(p)
	l.N -= int64(n)
	return
}

func LimitWriter(w io.Writer, n int64) io.Writer { return &LimitedWriter{w, n} }


// yn returns "y" or "n" for a boolean.
func yn(b bool) string {
	if b {
		return "y"
	}
	return "n"
}
