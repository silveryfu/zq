package field

import (
	"github.com/brimsec/zq/expr"
	"github.com/brimsec/zq/streamfn"
	"github.com/brimsec/zq/zng"
)

type Int struct {
	fn *streamfn.Int64
}

func NewIntStreamfn(op string) Streamfn {
	return &Int{
		fn: streamfn.NewInt64(op),
	}
}

func (i *Int) Result() zng.Value {
	return zng.Value{zng.TypeInt64, zng.EncodeInt(i.fn.State)}
}

func (i *Int) Consume(v zng.Value) error {
	if v, ok := expr.CoerceToInt(v); ok {
		i.fn.Update(v)
		return nil
	}
	return zng.ErrTypeMismatch
}
