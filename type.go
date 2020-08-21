package main

type ty interface {
	size() int
}

type ptr interface {
	ty
	base() ty
}

// TODO: think of creating Number interface (tyInt and TyChar should implement this)

type tyArr struct {
	of  ty
	len int
}

type tyChar struct{}

type tyEmpty struct{}

type tyInt struct{}

type tyPtr struct {
	to ty
}

func newTyArr(of ty, len int) *tyArr {
	return &tyArr{of, len}
}

func newTyChar() *tyChar {
	return &tyChar{}
}

func newTyInt() *tyInt {
	return &tyInt{}
}

func newTyPtr(to ty) *tyPtr {
	return &tyPtr{to}
}

func (a *tyArr) size() int {
	return a.len * a.of.size()
}

func (c *tyChar) size() int {
	return 1
}

func (e *tyEmpty) size() int {
	return 0
}

func (i *tyInt) size() int {
	return 4
}

func (p *tyPtr) size() int {
	return 8
}

func (a *tyArr) base() ty {
	return a.of
}

func (p *tyPtr) base() ty {
	return p.to
}

func (a *addrNode) loadType() ty {
	t := a.v.loadType()
	if arrT, ok := t.(*tyArr); ok {
		a.ty = &tyPtr{to: arrT.of}
	} else {
		a.ty = &tyPtr{to: t}
	}
	return a.ty
}

func (a *arithNode) loadType() ty {
	a.ty = a.lhs.loadType()
	a.rhs.loadType()
	return a.ty
}

func (a *assignNode) loadType() ty {
	a.rhs.loadType()
	a.ty = a.lhs.loadType()
	return a.ty
}

func (b *blkNode) loadType() ty {
	for _, st := range b.body {
		st.loadType()
	}
	return &tyEmpty{}
}

func (d *derefNode) loadType() ty {
	switch v := d.ptr.loadType().(type) {
	case *tyInt:
		d.ty = v
	case *tyPtr:
		d.ty = v.to
	case *tyArr:
		d.ty = v.of
	}
	return d.ty
}

func (e *exprNode) loadType() ty {
	e.body.loadType()
	return &tyEmpty{}
}

func (f *forNode) loadType() ty {
	if f.init != nil {
		f.init.loadType()
	}
	if f.cond != nil {
		f.cond.loadType()
	}
	if f.body != nil {
		f.body.loadType()
	}
	if f.inc != nil {
		f.inc.loadType()
	}
	return &tyEmpty{}
}

func (f *fnCallNode) loadType() ty {
	for _, param := range f.params {
		param.loadType()
	}
	return f.ty
}

func (f *fnNode) loadType() ty {
	for _, node := range f.body {
		node.loadType()
	}
	return f.ty
}

func (i *ifNode) loadType() ty {
	i.cond.loadType()
	i.then.loadType()
	if i.els != nil {
		i.els.loadType()
	}
	return &tyEmpty{}
}

func (v *varNode) loadType() ty {
	return v.v.getType()
}

func (*nullNode) loadType() ty {
	return &tyEmpty{}
}

func (n *numNode) loadType() ty {
	n.ty = &tyInt{}
	return n.ty
}

func (r *retNode) loadType() ty {
	r.ty = r.rhs.loadType()
	return r.ty
}

func (w *whileNode) loadType() ty {
	w.cond.loadType()
	w.then.loadType()
	return &tyEmpty{}
}
