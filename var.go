package main

type (
	// interface for local variables, global variables, and typedefs
	variable interface {
		getName() string
		getType() ty
	}

	gVar struct {
		emit bool
		init gVarInit
		name string
		ty   ty
	}

	gVarInit interface {
		genInit(ty)
	}

	gVarInitArr struct {
		body []gVarInit
	}

	gVarInitLabel struct { // reference to another global var
		label string
	}

	gVarInitStr struct {
		content string
	}

	gVarInitInt struct {
		val int64
		sz  int
	}

	gVarInitZero struct {
		len int
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

func newGVar(emit bool, name string, t ty, init gVarInit) *gVar {
	return &gVar{emit, init, name, t}
}

func newGVarInitArr(body []gVarInit) *gVarInitArr {
	return &gVarInitArr{body}
}

func newGVarInitLabel(label string) *gVarInitLabel {
	return &gVarInitLabel{label}
}

func newGVarInitStr(content string) *gVarInitStr {
	return &gVarInitStr{content}
}

func newGVarInitInt(i int64, sz int) *gVarInitInt {
	return &gVarInitInt{i, sz}
}

func newGVarInitZero(len int) *gVarInitZero {
	return &gVarInitZero{len}
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
