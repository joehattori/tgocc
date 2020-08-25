package main

// interface for local variables, global variables, and typedefs
type variable interface {
	getName() string
	getType() ty
}

type lVar struct {
	name   string
	offset int
	ty     ty
}

type gVar struct {
	content interface{}
	name    string
	ty      ty
}

type typeDef struct {
	name string
	ty   ty
}

func (v *lVar) getName() string    { return v.name }
func (v *gVar) getName() string    { return v.name }
func (t *typeDef) getName() string { return t.name }

func (v *lVar) getType() ty    { return v.ty }
func (v *gVar) getType() ty    { return v.ty }
func (t *typeDef) getType() ty { return t.ty }

func newLVar(name string, ty ty) *lVar {
	return &lVar{name: name, ty: ty}
}

func newGVar(name string, ty ty, content interface{}) *gVar {
	return &gVar{content, name, ty}
}

func newTypeDef(name string, ty ty) *typeDef {
	return &typeDef{name, ty}
}
