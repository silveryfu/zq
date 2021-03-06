package reducer

import (
	"github.com/brimsec/zq/expr"
	"github.com/brimsec/zq/zng"
	"github.com/brimsec/zq/zng/resolver"
)

type First struct {
	Reducer
	Resolver *expr.FieldExpr
	val      *zng.Value
}

func NewFirstReducer(resolver *expr.FieldExpr) Interface {
	return &First{Resolver: resolver}
}

func (f *First) Consume(r *zng.Record) {
	if f.val != nil {
		return
	}
	v, err := f.Resolver.Eval(r)
	if err != nil || v.Type == nil {
		return
	}
	f.val = &v
}

func (f *First) ConsumePart(p zng.Value) error {
	if f.val != nil || p.Type == zng.TypeNull {
		return nil
	}
	f.val = &p
	return nil
}

func (f *First) Result() zng.Value {
	if f.val == nil {
		return zng.Value{Type: zng.TypeNull, Bytes: nil}
	}
	return *f.val
}

func (f *First) ResultPart(*resolver.Context) (zng.Value, error) {
	return f.Result(), nil
}
