package main

type variable interface {
	getType() ty
}

type lVar struct {
	name   string
	ty     ty
	offset int
}

type gVar struct {
	name string
	ty   ty
	// TODO: elaborate
	content interface{}
}

func (v *lVar) getType() ty { return v.ty }
func (v *gVar) getType() ty { return v.ty }

func newLVar(name string, ty ty, offset int) *lVar {
	return &lVar{name, ty, offset}
}

func newGVar(name string, ty ty, content interface{}) *gVar {
	return &gVar{name, ty, content}
}
