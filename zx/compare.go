package zx

import (
	"bytes"
	"fmt"
	"net"
	"regexp"
	"strings"

	"github.com/mccanne/zq/ast"
	"github.com/mccanne/zq/zng"
)

//XXX TBD:
// - change these comparisons to work in the zcode.Bytes domain
// - add timer, interval comparisons when we add time, interval literals to the language
// - add count comparisons when we add count literals to the language
// - add set/vector/record comparisons when we add composite literals to the language

// Predicate is a function that takes a Value and returns a boolean result
// based on the typed value.
type Predicate func(zng.Value) bool

var compareBool = map[string]func(bool, bool) bool{
	"eql":  func(a, b bool) bool { return a == b },
	"neql": func(a, b bool) bool { return a != b },
	"gt": func(a, b bool) bool {
		if a {
			return !b
		}
		return false
	},
	"gte": func(a, b bool) bool {
		if a {
			return true
		}
		return !b
	},
	"lt": func(a, b bool) bool {
		if a {
			return false
		}
		return b
	},
	"lte": func(a, b bool) bool {
		if a {
			return b
		}
		return !b
	},
}

// CompareBool returns a Predicate that compares zng.Values to a boolean literal
// that must be a boolean or coercible to an integer.  In the later case, the integer
// is converted to a boolean.
func CompareBool(op string, pattern bool) (Predicate, error) {
	compare, ok := compareBool[op]
	if !ok {
		return nil, fmt.Errorf("unknown bool comparator: %s", op)
	}
	return func(v zng.Value) bool {
		if v.Type != zng.TypeBool {
			return false
		}
		b, err := zng.DecodeBool(v.Bytes)
		if err != nil {
			return false
		}
		return compare(b, pattern)
	}, nil
	return nil, fmt.Errorf("bad comparator for boolean type: %s", op)
}

var compareInt = map[string]func(int64, int64) bool{
	"eql":  func(a, b int64) bool { return a == b },
	"neql": func(a, b int64) bool { return a != b },
	"gt":   func(a, b int64) bool { return a > b },
	"gte":  func(a, b int64) bool { return a >= b },
	"lt":   func(a, b int64) bool { return a < b },
	"lte":  func(a, b int64) bool { return a <= b }}

var compareFloat = map[string]func(float64, float64) bool{
	"eql":  func(a, b float64) bool { return a == b },
	"neql": func(a, b float64) bool { return a != b },
	"gt":   func(a, b float64) bool { return a > b },
	"gte":  func(a, b float64) bool { return a >= b },
	"lt":   func(a, b float64) bool { return a < b },
	"lte":  func(a, b float64) bool { return a <= b }}

// XXX fix command
// Return a predicate for comparing this value to one more typed
// byte slices by calling the predicate function with a Type and
// a byte slice.  Operand is one of "eql", "neql", "lt", "lte",
// "gt", "gte".  See the comments of the various implementation
// of this method as some types limit the operand to equality and
// the various types handle coercion in different ways.
func CompareInt64(op string, pattern int64) (Predicate, error) {
	CompareInt, ok1 := compareInt[op]
	CompareFloat, ok2 := compareFloat[op]
	if !ok1 || !ok2 {
		return nil, fmt.Errorf("unknown int comparator: %s", op)
	}
	// many different zeek data types can be compared with integers
	return func(val zng.Value) bool {
		zv := val.Bytes
		switch val.Type.(type) {
		case *zng.TypeOfInt:
			// we can parse counts and ports as an integer
			v, err := zng.DecodeInt(zv)
			if err == nil {
				return CompareInt(v, pattern)
			}
		case *zng.TypeOfCount:
			// we can parse counts and ports as an integer
			v, err := zng.DecodeCount(zv)
			if err == nil {
				return CompareInt(int64(v), pattern)
			}
		case *zng.TypeOfPort:
			// we can parse counts and ports as an integer
			v, err := zng.DecodePort(zv)
			if err == nil {
				return CompareInt(int64(v), pattern)
			}
		case *zng.TypeOfDouble:
			v, err := zng.DecodeDouble(zv)
			if err == nil {
				return CompareFloat(v, float64(pattern))
			}
		case *zng.TypeOfTime:
			ts, err := zng.DecodeTime(zv)
			if err == nil {
				return CompareInt(int64(ts), pattern*1e9)
			}
		case *zng.TypeOfInterval:
			v, err := zng.DecodeInt(zv)
			if err == nil {
				return CompareInt(int64(v), pattern*1e9)
			}
		}
		return false
	}, nil
}

func CompareContainerLen(op string, len int64) (Predicate, error) {
	compare, ok := compareInt[op]
	if !ok {
		return nil, fmt.Errorf("unknown length comparator: %s", op)
	}
	return func(v zng.Value) bool {
		actual, err := v.ContainerLength()
		if err != nil {
			return false
		}
		return compare(int64(actual), len)
	}, nil
}

//XXX should just do equality and we should compare in the encoded domain
// and not make copies and have separate cases for len 4 and len 16
var compareAddr = map[string]func(net.IP, net.IP) bool{
	"eql":  func(a, b net.IP) bool { return a.Equal(b) },
	"neql": func(a, b net.IP) bool { return !a.Equal(b) },
	"gt":   func(a, b net.IP) bool { return bytes.Compare(a, b) > 0 },
	"gte":  func(a, b net.IP) bool { return bytes.Compare(a, b) >= 0 },
	"lt":   func(a, b net.IP) bool { return bytes.Compare(a, b) < 0 },
	"lte":  func(a, b net.IP) bool { return bytes.Compare(a, b) <= 0 },
}

// Comparison returns a Predicate that compares typed byte slices that must
// be TypeAddr with the value's address using a comparison based on op.
// Only equality operands are allowed.
func CompareIP(op string, pattern net.IP) (Predicate, error) {
	compare, ok := compareAddr[op]
	if !ok {
		return nil, fmt.Errorf("unknown addr comparator: %s", op)
	}
	return func(v zng.Value) bool {
		if v.Type != zng.TypeAddr {
			return false
		}
		ip, err := zng.DecodeAddr(v.Bytes)
		if err != nil {
			return false
		}
		return compare(ip, pattern)
	}, nil
}

// CompareFloat64 returns a Predicate that compares typed byte slices that must
// be coercible to an double with the value's double value using a comparison
// based on op.  Int, count, port, and double types can
// all be converted to the integer value.  XXX there are some overflow issues here.
func CompareFloat64(op string, pattern float64) (Predicate, error) {
	compare, ok := compareFloat[op]
	if !ok {
		return nil, fmt.Errorf("unknown double comparator: %s", op)
	}
	return func(val zng.Value) bool {
		zv := val.Bytes
		switch val.Type.(type) {
		// We allow comparison of float constant with integer-y
		// fields and just use typeDouble to parse since it will do
		// the right thing for integers.  XXX do we want to allow
		// integers that cause float64 overflow?  user can always
		// use an integer constant instead of a float constant to
		// compare with the integer-y field.
		case *zng.TypeOfDouble:
			v, err := zng.DecodeDouble(zv)
			if err == nil {
				return compare(v, pattern)
			}
		case *zng.TypeOfInt:
			v, err := zng.DecodeInt(zv)
			if err == nil {
				return compare(float64(v), pattern)
			}
		case *zng.TypeOfCount:
			v, err := zng.DecodeCount(zv)
			if err == nil {
				return compare(float64(v), pattern)
			}
		case *zng.TypeOfPort:
			v, err := zng.DecodePort(zv)
			if err == nil {
				return compare(float64(v), pattern)
			}

		case *zng.TypeOfTime:
			ts, err := zng.DecodeTime(zv)
			if err == nil {
				return compare(float64(ts)/1e9, pattern)
			}
		case *zng.TypeOfInterval:
			v, err := zng.DecodeInterval(zv)
			if err == nil {
				return compare(float64(v)/1e9, pattern)
			}
		}
		return false
	}, nil
}

var compareString = map[string]func(string, string) bool{
	"eql":  func(a, b string) bool { return a == b },
	"neql": func(a, b string) bool { return a != b },
	"gt":   func(a, b string) bool { return a > b },
	"gte":  func(a, b string) bool { return a >= b },
	"lt":   func(a, b string) bool { return a < b },
	"lte":  func(a, b string) bool { return a <= b },
	//XXX this doesn't belong here.  scalar comparison vs composite operator
	"search": func(a, b string) bool { return strings.Contains(a, b) },
}

func CompareBstring(op string, pattern zng.Bstring) (Predicate, error) {
	compare, ok := compareString[op]
	if !ok {
		return nil, fmt.Errorf("unknown string comparator: %s", op)
	}
	s := string(pattern)
	return func(v zng.Value) bool {
		switch v.Type.(type) {
		case *zng.TypeOfString, *zng.TypeOfEnum:
			//XXX is this UTF safe?
			return compare(zng.UnsafeString(v.Bytes), s)
		}
		return false
	}, nil
}

// RegexpComparison returns a Predicate that compares values that must
// be a string or enum with the value's regular expression using a regex
// match comparison based on equality or inequality based on op.
func CompareRegexp(op, pattern string) (Predicate, error) {
	re, err := regexp.Compile(pattern)
	if err != nil {
		return nil, err
	}
	switch op {
	default:
		return nil, fmt.Errorf("unknown pattern comparator: %s", op)
	case "eql":
		return func(v zng.Value) bool {
			switch v.Type.(type) {
			case *zng.TypeOfString, *zng.TypeOfEnum:
				return re.Match(v.Bytes)
			}
			return false
		}, nil
	case "neql":
		return func(v zng.Value) bool {
			switch v.Type.(type) {
			case *zng.TypeOfString, *zng.TypeOfEnum:
				return !re.Match(v.Bytes)
			}
			return false
		}, nil
	}
}

// XXX Comparison returns a Predicate that compares typed byte slices that must
// be a port with the value's port value using a comparison based on op.
// Integer fields are not coerced (nor are any other types) so they never
// match the port literal here.
func ComparePort(op string, pattern uint32) (Predicate, error) {
	compare, ok := compareInt[op]
	if !ok {
		return nil, fmt.Errorf("unknown port comparator: %s", op)
	}
	// only a zeek port can be compared with a port type.  If the user went
	// to the trouble of specifying a port match (e.g., ":80" vs "80") then
	// we use strict typing here on the port comparison.
	return func(v zng.Value) bool {
		if v.Type != zng.TypePort {
			return false
		}
		p, err := zng.DecodePort(v.Bytes)
		if err != nil {
			return false
		}
		return compare(int64(p), int64(pattern))
	}, nil
}

func CompareUnset(op string) (Predicate, error) {
	switch op {
	default:
		return nil, fmt.Errorf("unknown unset comparator: %s", op)
	case "eql":
		return func(v zng.Value) bool {
			return v.IsUnset()
		}, nil
	case "neql":
		return func(v zng.Value) bool {
			return !v.IsUnset()
		}, nil
	}
}

// a better way to do this would be to compare IP's and mask's but
// go doesn't provide an easy way to compare masks so we do this
// hacky thing and compare strings
var compareSubnet = map[string]func(*net.IPNet, *net.IPNet) bool{
	"eql":  func(a, b *net.IPNet) bool { return bytes.Equal(a.IP, b.IP) },
	"neql": func(a, b *net.IPNet) bool { return bytes.Equal(a.IP, b.IP) },
	"lt":   func(a, b *net.IPNet) bool { return bytes.Compare(a.IP, b.IP) < 0 },
	"lte":  func(a, b *net.IPNet) bool { return bytes.Compare(a.IP, b.IP) <= 0 },
	"gt":   func(a, b *net.IPNet) bool { return bytes.Compare(a.IP, b.IP) > 0 },
	"gte":  func(a, b *net.IPNet) bool { return bytes.Compare(a.IP, b.IP) >= 0 },
}

var matchSubnet = map[string]func(net.IP, *net.IPNet) bool{
	"eql": func(a net.IP, b *net.IPNet) bool {
		return b.IP.Equal(a.Mask(b.Mask))
	},
	"neql": func(a net.IP, b *net.IPNet) bool {
		return !b.IP.Equal(a.Mask(b.Mask))
	},
	"lt": func(a net.IP, b *net.IPNet) bool {
		net := a.Mask(b.Mask)
		return bytes.Compare(net, b.IP) < 0
	},
	"lte": func(a net.IP, b *net.IPNet) bool {
		net := a.Mask(b.Mask)
		return bytes.Compare(net, b.IP) <= 0
	},
	"gt": func(a net.IP, b *net.IPNet) bool {
		net := a.Mask(b.Mask)
		return bytes.Compare(net, b.IP) > 0
	},
	"gte": func(a net.IP, b *net.IPNet) bool {
		net := a.Mask(b.Mask)
		return bytes.Compare(net, b.IP) >= 0
	},
}

// Comparison returns a Predicate that compares typed byte slices that must
// be an addr or a subnet to the value's subnet value using a comparison
// based on op.  Onluy equalty and inequality are permitted.  If the typed
// byte slice is a subnet, then the comparison is based on strict equality.
// If the typed byte slice is an addr, then the comparison is performed by
// doing a CIDR match on the address with the subnet.
func CompareSubnet(op string, pattern *net.IPNet) (Predicate, error) {
	compare, ok1 := compareSubnet[op]
	match, ok2 := matchSubnet[op]
	if !ok1 || !ok2 {
		return nil, fmt.Errorf("unknown subnet comparator: %s", op)
	}
	return func(v zng.Value) bool {
		val := v.Bytes
		switch v.Type.(type) {
		case *zng.TypeOfAddr:
			ip, err := zng.DecodeAddr(val)
			if err == nil {
				return match(ip, pattern)
			}
		case *zng.TypeOfSubnet:
			subnet, err := zng.DecodeSubnet(val)
			if err == nil {
				return compare(subnet, pattern)
			}
		}
		return false
	}, nil
}

// Given a predicate for comparing individual elements, produce a new
// predicate that implements the "in" comparison.  The new predicate looks
// at the type of the value being compared, if it is a set or vector,
// the original predicate is applied to each element.  The new precicate
// returns true iff the predicate matched an element from the collection.
func Contains(compare Predicate) Predicate {
	return func(v zng.Value) bool {
		var el zng.Value
		el.Type = zng.InnerType(v.Type)
		if el.Type == nil {
			return false
		}
		for it := v.Iter(); !it.Done(); {
			var err error
			el.Bytes, _, err = it.Next()
			if err != nil {
				return false
			}
			if compare(el) {
				return true
			}
		}
		return false
	}
}

// Comparison returns a Predicate for comparing this value to other values.
// The op argument is one of "eql", "neql", "lt", "lte",
// "gt", "gte".  See the comments of the various type implementations
// of this method as some types limit the operand to equality and
// the various types handle coercion in different ways.
func Comparison(op string, literal ast.Literal) (Predicate, error) {
	if literal.Type == "unset" {
		return CompareUnset(op)
	}
	if literal.Type == "regexp" {
		return CompareRegexp(op, literal.Value)
	}
	v, err := zng.ParseLiteral(literal)
	if err != nil {
		return nil, err
	}
	switch v := v.(type) {
	default:
		return nil, fmt.Errorf("unknown type of constant: %s (%T)", literal.Type, v)
	case net.IP:
		return CompareIP(op, v)
	case *net.IPNet:
		return CompareSubnet(op, v)
	case bool:
		return CompareBool(op, v)
	case float64: //XXX
		return CompareFloat64(op, v)
	case zng.Bstring: //XXX
		return CompareBstring(op, v)
	case zng.Port:
		return ComparePort(op, uint32(v))
	case int64:
		return CompareInt64(op, v)
	}
}