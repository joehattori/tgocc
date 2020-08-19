package main

// Type represents type
type Type interface {
	size() int
}

// TyInt represents int type
type TyInt struct{}

// TyPtr represents pointer type
type TyPtr struct {
	to Type
}

func (i *TyInt) size() int {
	return 4
}

func (p *TyPtr) size() int {
	return 8
}

func (e *TyEmpty) size() int {
	return 0
}

// TyEmpty represents empty type. e.g) Block expression has this type
type TyEmpty struct{}

func (a *AddrNode) loadType() Type {
	if a.ty != nil {
		return a.ty
	}
	a.ty = &TyPtr{to: a.v.loadType()}
	return a.ty
}

func (a *ArithNode) loadType() Type {
	if a.ty != nil {
		return a.ty
	}
	a.ty = a.lhs.loadType()
	a.rhs.loadType()
	return a.ty
}

func (a *AssignNode) loadType() Type {
	if a.ty != nil {
		return a.ty
	}
	a.ty = a.lhs.loadType()
	a.rhs.loadType()
	return a.ty
}

func (b *BlkNode) loadType() Type {
	for _, st := range b.body {
		st.loadType()
	}
	return &TyEmpty{}
}

func (d *DerefNode) loadType() Type {
	if d.ty != nil {
		return d.ty
	}
	switch v := d.ptr.loadType().(type) {
	case *TyInt:
		d.ty = &TyInt{}
	case *TyPtr:
		d.ty = v.to
	}
	return d.ty
}

func (e *ExprNode) loadType() Type {
	e.body.loadType()
	return &TyEmpty{}
}

func (f *ForNode) loadType() Type {
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
	return &TyEmpty{}
}

func (f *FnCallNode) loadType() Type {
	for _, param := range f.params {
		param.loadType()
	}
	return f.ty
}

func (f *FnNode) loadType() Type {
	for _, param := range f.params {
		param.loadType()
	}
	for _, node := range f.body {
		node.loadType()
	}
	return f.ty
}

func (i *IfNode) loadType() Type {
	i.cond.loadType()
	i.then.loadType()
	if i.els != nil {
		i.els.loadType()
	}
	return &TyEmpty{}
}

func (l *LVarNode) loadType() Type {
	return l.ty
}

func (*NullNode) loadType() Type {
	return &TyEmpty{}
}

func (n *NumNode) loadType() Type {
	n.ty = &TyInt{}
	return n.ty
}

func (r *RetNode) loadType() Type {
	r.ty = r.rhs.loadType()
	return r.ty
}

func (w *WhileNode) loadType() Type {
	w.cond.loadType()
	w.then.loadType()
	return &TyEmpty{}
}
