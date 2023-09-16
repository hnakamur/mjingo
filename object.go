package mjingo

import "github.com/hnakamur/mjingo/option"

// Object is implemented by a dynamic object.
//
// The engine uses the [Value] type to represent values that the engine
// knows about.  Most of these values are primitives such as integers, strings
// or maps.  However it is also possible to expose custom types without
// undergoing a serialization step to the engine.  For this to work a type
// needs to implement the [Object] interface and be wrapped in a value with
// [ValueFromObject].
//
// Objects need to implement String() method which is used by
// the engine to convert the object into a string if needed.  Additionally
// DebugString() is required as well.
//
// The exact runtime characteristics of the object are influenced by the
// [Kind] of the object.  By default an object can just be
// stringified and methods can be called.
//
// For examples of how to implement objects refer to [SeqObject] and
// [StructObject].
type Object interface {
	// Kind describes the kind of an object.
	Kind() ObjectKind
}

// ObjectKind defines the object's behavior.
//
// When a dynamic [Object] is implemented, it can be of one of the kinds
// here.  The default behavior will be a [ObjectKindPlain] object which
// doesn't do much other than that it can be printed.  For an object to turn
// into a [StructObject] or [SeqObject] the necessary kind
// has to be returned with a pointer to itself.
//
// Today object's can have the behavior of structs and sequences but this
// might expand in the future.  It does mean that not all types of values can
// be represented by objects.
type ObjectKind uint

const (
	// ObjectKindPlain is a kind for a plain object.
	//
	// Such an object has no attributes but it might be callable and it
	// can be stringified.  When serialized it's serialized in it's
	// stringified form.
	ObjectKindPlain ObjectKind = iota + 1

	//ObjectKindSeq is a kind for a sequence.
	//
	// Requires that the object implements [SeqObject].
	ObjectKindSeq

	// ObjectKindStruct is a kind for a struct (map with string keys).
	//
	// Requires that the object implements [StructObject].
	ObjectKindStruct
)

// SeqObject provides the behavior of an [Object] holding sequence of values.
//
// An object holding a sequence of values (tuple, list etc.) can be
// represented by this interface.
type SeqObject interface {
	// GetItem looks up an item by index.
	//
	// Sequences should provide a value for all items in the range of `0..ItemCount`
	// but the engine will assume that items within the range are `Undefined`
	// if `None` is returned.
	GetItem(idx uint) option.Option[Value]

	// ItemCount returns the number of items in the sequence.
	ItemCount() uint
}

func newSliceSeqObject(values []Value) SeqObject {
	return &sliceSeqObject{values: values}
}

type sliceSeqObject struct {
	values []Value
}

func (s *sliceSeqObject) Kind() ObjectKind { return ObjectKindSeq }

func (s *sliceSeqObject) GetItem(idx uint) option.Option[Value] {
	if idx >= uint(len(s.values)) {
		return option.None[Value]()
	}
	return option.Some(s.values[idx])
}

func (s *sliceSeqObject) ItemCount() uint {
	return uint(len(s.values))
}

// StructObject provides the behavior of an [Object] holding a struct.
//
// An basic object with the shape and behavior of a struct (that means a
// map with string keys) can be represented by this interface.
//
// # Struct As context
//
// Structs can also be used as template rendering context.  This has a lot of
// benefits as it means that the serialization overhead can be largely to
// completely avoided.  This means that even if templates take hundreds of
// values, MiniJinja does not spend time eagerly converting them into values.
//
// Here is a very basic example of how a template can be rendered with a dynamic
// context.  Note that the implementation of [StructObject.Fields] is optional
// for this to work.  It's in fact not used by the engine during rendering but
// it is necessary for the DebugString() function to be
// able to show which values exist in the context.
type StructObject interface {
	// GetField is nvoked by the engine to get a field of a struct.
	//
	// Where possible it's a good idea for this to align with the return value
	// of [StructObject.Fields] but it's not necessary.
	//
	// If an field does not exist, `None` shall be returned.
	//
	// A note should be made here on side effects: unlike calling objects or
	// calling methods on objects, accessing fields is not supposed to
	// have side effects.  Neither does this API get access to the interpreter
	// [State] nor is there a channel to send out failures as only an option
	// can be returned.  If you do plan on doing something in field access
	// that is fallible, instead use a method call.
	GetField(name string) option.Option[Value]

	// StaticFields if possible returns a static vector of field names.
	//
	// If fields cannot be statically determined, then this must return `None`
	// and [StructObject.Fields] should be implemented instead.  If however
	// this method is implemented, then [StructObject.Fields] should be
	// implemented to return nil.
	StaticFields() option.Option[[]string]

	// Fields returns a vector of field names.
	//
	// This should be implemented if [StructObject.StaticFields] cannot
	// be implemented due to lifetime restrictions.
	Fields() []string
}

func fieldCount(s StructObject) uint {
	optFields := s.StaticFields()
	if optFields.IsSome() {
		return uint(len(optFields.Unwrap()))
	}
	return uint(len(s.Fields()))
}

func staticOrDynamicFields(s StructObject) []string {
	optFields := s.StaticFields()
	if optFields.IsSome() {
		return optFields.Unwrap()
	}
	return s.Fields()
}
