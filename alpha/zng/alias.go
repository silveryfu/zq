package zng

import (
	"github.com/brimsec/zq/alpha/zcode"
)

type TypeAlias struct {
	id   int
	Name string
	Type Type
}

func NewTypeAlias(id int, name string, typ Type) *TypeAlias {
	return &TypeAlias{
		id:   id,
		Name: name,
		Type: typ,
	}
}

func (t *TypeAlias) ID() int {
	return t.Type.ID()
}

func (t *TypeAlias) AliasID() int {
	return t.id
}

//XXX get rid of this when we implement full ZNG
func (t *TypeAlias) SetID(id int) {
	t.id = id
}

func (t *TypeAlias) String() string {
	return t.Name
}
func (t *TypeAlias) Parse(in []byte) (zcode.Bytes, error) {
	return t.Type.Parse(in)
}

func (t *TypeAlias) StringOf(zv zcode.Bytes, out OutFmt, b bool) string {
	return t.Type.StringOf(zv, out, b)
}

func (t *TypeAlias) Marshal(zv zcode.Bytes) (interface{}, error) {
	return t.Type.Marshal(zv)
}

func AliasedType(typ Type) Type {
	alias, isAlias := typ.(*TypeAlias)
	if isAlias {
		return AliasedType(alias.Type)
	}
	return typ
}
