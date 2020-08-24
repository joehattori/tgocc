package main

type variable interface {
	getName() string
	getType() ty
}

type lVar struct {
	name   string
	ty     ty
	offset int
}

type gVar struct {
	name    string
	ty      ty
	content interface{}
}

func (v *lVar) getName() string { return v.name }
func (v *gVar) getName() string { return v.name }

func (v *lVar) getType() ty { return v.ty }
func (v *gVar) getType() ty { return v.ty }

func newLVar(name string, ty ty) *lVar {
	return &lVar{name: name, ty: ty}
}

func newGVar(name string, ty ty, content interface{}) *gVar {
	return &gVar{name, ty, content}
}
