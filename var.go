package main

// interface for local variables, global variables, and typedefs
type variable interface {
	getName() string
	getType() ty
}

type gVar struct {
	content interface{}
	name    string
	ty      ty
}

type lVar struct {
	name   string
	offset int
	ty     ty
}

type typeDef struct {
	name string
	ty   ty
}

func (v *gVar) getName() string    { return v.name }
func (v *lVar) getName() string    { return v.name }
func (t *typeDef) getName() string { return t.name }

func (v *gVar) getType() ty    { return v.ty }
func (v *lVar) getType() ty    { return v.ty }
func (t *typeDef) getType() ty { return t.ty }

func newGVar(name string, t ty, content interface{}) *gVar {
	return &gVar{content, name, t}
}

func newLVar(name string, t ty) *lVar {
	return &lVar{name: name, ty: t}
}

func newTypeDef(name string, t ty) *typeDef {
	return &typeDef{name, t}
}
