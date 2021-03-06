package zng

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/brimsec/zq/ast"
	"github.com/brimsec/zq/zcode"
)

var (
	ErrNotNumber  = errors.New("not a number")
	ErrTypeSyntax = errors.New("syntax error parsing type string")
)

type Value struct {
	Type  Type
	Bytes zcode.Bytes
}

// Parse translates an Literal into a Value.
// This currently supports only primitive literals.
func Parse(v ast.Literal) (Value, error) {
	if v.Type == "null" {
		return Value{}, nil
	}
	t := LookupPrimitive(v.Type)
	if t == nil {
		return Value{}, fmt.Errorf("unsupported type %s in ast.Literal", v.Type)
	}
	zv, err := t.Parse([]byte(v.Value))
	if err != nil {
		return Value{}, err
	}
	return Value{t, zv}, nil
}

func parseContainer(containerType Type, elementType Type, b zcode.Bytes) ([]Value, error) {
	// We start out with a pointer instead of nil so that empty sets and arrays
	// are properly encoded etc., e.g., by json.Marshal.
	vals := []Value{}
	for it := b.Iter(); !it.Done(); {
		zv, _, err := it.Next()
		if err != nil {
			return nil, fmt.Errorf("parsing %s element %q: %w", containerType.String(), zv, err)
		}
		vals = append(vals, Value{elementType, zv})
	}
	return vals, nil
}

func (v Value) IsContainer() bool {
	return IsContainerType(v.Type)
}

func (v Value) MarshalJSON() ([]byte, error) {
	if v.Bytes == nil {
		return json.Marshal(nil)
	}
	object, err := v.Type.Marshal(v.Bytes)
	if err != nil {
		return nil, err
	}
	return json.Marshal(object)
}

func badZng(err error, t Type, zv zcode.Bytes) string {
	return fmt.Sprintf("<ZNG-ERR type %s [%s]: %s>", t, zv, err)
}

func (v Value) Format(fmt OutFmt) string {
	if v.Bytes == nil {
		return "-"
	}
	return v.Type.StringOf(v.Bytes, fmt, false)
}

// String implements the fmt.Stringer interface and returns a
// human-readable string representation of the value.
// This should only be used for logs, debugging, etc.  Any caller that
// requires a specific output format should use FormatAs() instead.
func (v Value) String() string {
	return v.Format(OutFormatDebug)
}

// Encode appends the ZNG representation of this value to the passed in
// argument and returns the resulting zcode.Bytes (which may or may not
// be the same underlying buffer, as with append(), depending on its capacity)
func (v Value) Encode(dst zcode.Bytes) zcode.Bytes {
	if IsContainerType(v.Type) {
		return zcode.AppendContainer(dst, v.Bytes)
	}
	return zcode.AppendPrimitive(dst, v.Bytes)
}

func (v Value) Iter() zcode.Iter {
	return v.Bytes.Iter()
}

// If the passed-in element is an array, attempt to get the idx'th
// element, and return its type and raw representation.  Returns an
// error if the passed-in element is not an array or if idx is
// outside the array bounds.
func (v Value) ArrayIndex(idx int64) (Value, error) {
	vec, ok := v.Type.(*TypeArray)
	if !ok {
		return Value{}, ErrNotArray
	}
	if idx < 0 {
		return Value{}, ErrIndex
	}
	for i, it := 0, v.Iter(); !it.Done(); i++ {
		zv, _, err := it.Next()
		if err != nil {
			return Value{}, err
		}
		if i == int(idx) {
			return Value{vec.Type, zv}, nil
		}
	}
	return Value{}, ErrIndex
}

// Elements returns an array of Values for the given container type.
// Returns an error if the element is not an array or set.
func (v Value) Elements() ([]Value, error) {
	innerType := InnerType(v.Type)
	if innerType == nil {
		return nil, ErrNotContainer
	}
	var elements []Value
	for it := v.Iter(); !it.Done(); {
		zv, _, err := it.Next()
		if err != nil {
			return nil, err
		}
		elements = append(elements, Value{innerType, zv})
	}
	return elements, nil
}

func (v Value) ContainerLength() (int, error) {
	switch v.Type.(type) {
	case *TypeSet, *TypeArray:
		if v.Bytes == nil {
			return -1, ErrLenUnset
		}
		var n int
		for it := v.Iter(); !it.Done(); {
			if _, _, err := it.Next(); err != nil {
				return -1, err
			}
			n++
		}
		return n, nil
	default:
		return -1, ErrNotContainer
	}
}

func (v Value) IsNil() bool {
	return v.Bytes == nil && v.Type == nil
}

// IsUnset returns true iff v is an unset value.  Unset values are represented
// with a zero-valued Value.  A zero-valued value that is not unset is represented
// by a non-nil slice for Bytes of zero length.
func (v Value) IsUnset() bool {
	return v.Bytes == nil && v.Type != nil
}

func (v Value) IsUnsetOrNil() bool {
	return v.Bytes == nil
}

func (v Value) Copy() Value {
	var b zcode.Bytes
	if v.Bytes != nil {
		b = make(zcode.Bytes, len(v.Bytes))
		copy(b, v.Bytes)
	}
	return Value{v.Type, b}
}
