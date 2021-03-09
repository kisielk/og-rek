// Package ogórek(*) is a library for decoding/encoding Python's pickle format.
//
// Use Decoder to decode a pickle from input stream, for example:
//
//	d := ogórek.NewDecoder(r)
//	obj, err := d.Decode() // obj is interface{} representing decoded Python object
//
// Use Encoder to encode an object as pickle into output stream, for example:
//
//	e := ogórek.NewEncoder(w)
//	err := e.Encode(obj)
//
// The following table summarizes mapping of basic types in between Python and Go:
//
//	Python	   Go
//	------	   --
//
//	None	↔  ogórek.None
//	bool	↔  bool
//	int	↔  int64
//	int	←  int, intX, uintX
//	long	↔  *big.Int
//	float	↔  float64
//	float	←  floatX
//	list	↔  []interface{}
//	tuple	↔  ogórek.Tuple
//	dict	↔  map[interface{}]interface{}
//
//	str        ↔  string         (+)
//	bytes      ↔  ogórek.Bytes   (~)
//	bytearray  ↔  []byte
//
//
// Python classes and instances are mapped to Class and Call, for example:
//
//	Python				Go
//	------	   			--
//
//	decimal.Decimal            ↔    ogórek.Class{"decimal", "Decimal"}
//	decimal.Decimal("3.14")    ↔    ogórek.Call{
//						ogórek.Class{"decimal", "Decimal"},
//						ogórek.Tuple{"3.14"},
//					}
//
// In particular on Go side it is thus by default safe to decode pickles from
// untrusted sources(^).
//
//
// Pickle protocol versions
//
// Over the time the pickle stream format was evolving. The original protocol
// version 0 is human-readable with versions 1 and 2 extending the protocol in
// backward-compatible way with binary encodings for efficiency. Protocol
// version 2 is the highest protocol version that is understood by standard
// pickle module of Python2. Protocol version 3 added ways to represent Python
// bytes objects from Python3(~). Protocol version 4 further enhances on
// version 3 and completely switches to binary-only encoding. Protocol
// version 5 added support for out-of-band data(%). Please see
// https://docs.python.org/3/library/pickle.html#data-stream-format for details.
//
// On decoding ogórek detects which protocol is being used and automatically
// handles all necessary details.
//
// On encoding, for compatibility with Python2, by default ogórek produces
// pickles with protocol 2. Bytes thus, by default, will be unpickled as str on
// Python2 and as bytes on Python3. If an earlier protocol is desired, or on
// the other hand, if Bytes needs to be encoded efficiently (protocol 2
// encoding for bytes is far from optimal), and compatibility with pure Python2
// is not an issue, the protocol to use for encoding could be explicitly
// specified, for example:
//
//	e := ogórek.NewEncoderWithConfig(w, &ogórek.EncoderConfig{
//		Protocol: 3,
//	})
//	err := e.Encode(obj)
//
// See EncoderConfig.Protocol for details.
//
//
// Persistent references
//
// Pickle was originally created for serialization in ZODB (http://zodb.org)
// object database, where on-disk objects can reference each other similarly to
// how one in-RAM object can have a reference to another in-RAM object.
//
// When a pickle with such persistent reference is decoded, ogórek represents
// the reference with Ref placeholder similarly to Class and Call. However it
// is possible to hook into decoding and process such references in application
// specific way, for example loading the referenced object from the database:
//
//	d := ogórek.NewDecoderWithConfig(r, &ogórek.DecoderConfig{
//		PersistentLoad: ...
//	})
//	obj, err := d.Decode()
//
// Similarly, for encoding, an application can hook into serialization process
// and turn pointers to some in-RAM objects into persistent references.
//
// Please see DecoderConfig.PersistentLoad and EncoderConfig.PersistentRef for details.
//
//
// --------
//
// (*) ogórek is Polish for "pickle".
//
// (+) for Python2 both str and unicode are decoded into string with Python
// str being considered as UTF-8 encoded. Correspondingly for protocol ≤ 2 Go
// string is encoded as UTF-8 encoded Python str, and for protocol ≥ 3 as unicode.
//
// (~) bytes can be produced only by Python3 or zodbpickle (https://pypi.org/project/zodbpickle),
// not by standard Python2. Respectively, for protocol ≤ 2, what ogórek produces
// is unpickled as bytes by Python3 or zodbpickle, and as str by Python2.
//
// (^) contrary to Python implementation, where malicious pickle can cause the
// decoder to run arbitrary code, including e.g. os.system("rm -rf /").
//
// (%) ogórek currently does not support out-of-band data.
package ogórek
