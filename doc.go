// Package ogórek(*) is a library for decoding/encoding Python's pickle format.
//
// Use [Decoder] to decode a pickle from input stream, for example:
//
//	d := ogórek.NewDecoder(r)
//	obj, err := d.Decode() // obj is any representing decoded Python object
//
// Use [Encoder] to encode an object as pickle into output stream, for example:
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
//	list	↔  []any
//	tuple	↔  ogórek.Tuple
//
//
// For dicts there are two modes. In the first, default, mode Python dicts are
// decoded into standard Go map. This mode tries to use builtin Go type, but
// cannot mirror py behaviour fully because e.g. int(1), big.Int(1) and
// float64(1.0) are all treated as different keys by Go, while Python treats
// them as being equal. It also does not support decoding dicts with tuple
// used in keys:
//
//      dict    ↔  map[any]any                       PyDict=n mode, default
//              ←  ogórek.Dict
//
// With PyDict=y mode, however, Python dicts are decoded as [ogórek.Dict] which
// mirrors behaviour of Python dict with respect to keys equality, and with
// respect to which types are allowed to be used as keys.
//
//      dict    ↔  ogórek.Dict                       PyDict=y mode
//              ←  map[any]any
//
//
// For strings there are also two modes. In the first, default, mode both py2/py3
// str and py2 unicode are decoded into string with py2 str being considered
// as UTF-8 encoded. Correspondingly for protocol ≤ 2 Go string is encoded as
// UTF-8 encoded py2 str, and for protocol ≥ 3 as py3 str / py2 unicode.
// [ogórek.ByteString] can be used to produce bytestring objects after encoding
// even for protocol ≥ 3. This mode tries to match Go string with str type of
// target Python depending on protocol version, but looses information after
// decoding/encoding cycle:
//
//	py2/py3 str  ↔  string                       StrictUnicode=n mode, default
//	py2 unicode  →  string
//	py2 str      ←  ogórek.ByteString
//
// However with StrictUnicode=y mode there is 1-1 mapping in between py2
// unicode / py3 str vs Go string, and between py2 str vs [ogórek.ByteString].
// In this mode decoding/encoding and encoding/decoding operations are always
// identity with respect to strings:
//
//	py2 unicode / py3 str  ↔  string             StrictUnicode=y mode
//	py2 str                ↔  ogórek.ByteString
//
//
// For bytes, unconditionally to string mode, there is direct 1-1 mapping in
// between Python and Go types:
//
//	bytes        ↔  ogórek.Bytes   (~)
//	bytearray    ↔  []byte
//
//
//
// Python classes and instances are mapped to [Class] and [Call], for example:
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
// the reference with [Ref] placeholder similarly to [Class] and [Call]. However it
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
// Handling unpickled values
//
// On Python two different objects with different types can represent
// essentially the same entity. For example 1 (int) and 1L (long) represent
// integer number one via two different types and are decoded by ogórek into Go
// types int64 and big.Int correspondingly. However on the Python side those
// two representations are often used interchangeably and programs are usually
// expected to handle both with the same effect. To help handling decoded
// values with such differences ogórek provides utilities that bring objects
// to common type irregardless of which type variant was used in the pickle
// stream. For example [AsInt64] tries to represent unpickled value as int64 if
// possible and errors if not.
//
// For strings the situation is similar, but a bit different.
// On Python3 strings are unicode strings and binary data is represented by
// bytes type. However on Python2 strings are bytestrings and could contain
// both text and binary data. In the default mode py2 strings, the same way as
// py2 unicode, are decoded into Go strings. However in StrictUnicode mode py2
// strings are decoded into [ByteString] - the type specially dedicated to
// represent them on Go side. There are two utilities to help programs handle
// all those bytes/string data in the pickle stream in uniform way:
//
//     - the program should use [AsString] if it expects text   data -
//       either unicode string, or byte string.
//     - the program should use [AsBytes]  if it expects binary data -
//	 either bytes, or byte string.
//
// Using the helpers fits into Python3 strings/bytes model but also allows to
// handle the data generated from under Python2.
//
// Similarly [Dict] considers [ByteString] to be equal to both string and [Bytes]
// with the same underlying content. This allows programs to access [Dict] via
// string/bytes keys following Python3 model, while still being able to handle
// dictionaries generated from under Python2.
//
//
// --------
//
// (*) ogórek is Polish for "pickle".
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
