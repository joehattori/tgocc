package vars

import (
	"fmt"

	"github.com/joehattori/tgocc/types"
)

type (
	// interface for local variables, global variables, and typedefs
	Var interface {
		Name() string
		Type() types.Type
	}

	GVar struct {
		Emit bool
		Init GVarInit
		name string
		ty   types.Type
	}

	LVar struct {
		name   string
		Offset int
		ty     types.Type
	}

	TypeDef struct {
		name string
		ty   types.Type
	}

	Enum struct {
		name string
		ty   types.Type
		Val  int
	}
)

func (v *GVar) Name() string    { return v.name }
func (v *LVar) Name() string    { return v.name }
func (t *TypeDef) Name() string { return t.name }
func (e *Enum) Name() string    { return e.name }

func (v *GVar) Type() types.Type    { return v.ty }
func (v *LVar) Type() types.Type    { return v.ty }
func (t *TypeDef) Type() types.Type { return t.ty }
func (e *Enum) Type() types.Type    { return e.ty }

func NewGVar(emit bool, name string, t types.Type, init GVarInit) *GVar {
	return &GVar{emit, init, name, t}
}

func NewLVar(name string, t types.Type) *LVar {
	return &LVar{name: name, ty: t}
}

func NewTypeDef(name string, t types.Type) *TypeDef {
	return &TypeDef{name, t}
}

func NewEnum(name string, t types.Type, val int) *Enum {
	return &Enum{name, t, val}
}

func GenData(init GVarInit, t types.Type) {
	if init == nil {
		fmt.Printf("	.zero %d\n", t.Size())
	} else {
		init.gen(t)
	}
}
