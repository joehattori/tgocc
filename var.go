package main

type (
	// interface for local variables, global variables, and typedefs
	variable interface {
		getName() string
		getType() ty
	}

	gVar struct {
		init gVarInit
		name string
		ty   ty
	}

	gVarInit interface{} // TODO: brush up

	gVarInitLabel struct { // reference to another global var
		label string
	}

	gVarInitStr struct {
		content string
	}

	gVarInitVal struct {
		val int
	}

	lVar struct {
		name   string
		offset int
		ty     ty
	}

	typeDef struct {
		name string
		ty   ty
	}

	enum struct {
		name string
		ty   ty
		val  int
	}
)

func (v *gVar) getName() string    { return v.name }
func (v *lVar) getName() string    { return v.name }
func (t *typeDef) getName() string { return t.name }
func (e *enum) getName() string    { return e.name }

func (v *gVar) getType() ty    { return v.ty }
func (v *lVar) getType() ty    { return v.ty }
func (t *typeDef) getType() ty { return t.ty }
func (e *enum) getType() ty    { return e.ty }

func newGVar(name string, t ty, init gVarInit) *gVar {
	return &gVar{init, name, t}
}

func newGVarInitLabel(label string) *gVarInitLabel {
	return &gVarInitLabel{label}
}

func newGVarInitStr(s string) *gVarInitStr {
	return &gVarInitStr{s}
}

func newGVarInitVal(i int) *gVarInitVal {
	return &gVarInitVal{i}
}

func newLVar(name string, t ty) *lVar {
	return &lVar{name: name, ty: t}
}

func newTypeDef(name string, t ty) *typeDef {
	return &typeDef{name, t}
}

func newEnum(name string, t ty, val int) *enum {
	return &enum{name, t, val}
}
