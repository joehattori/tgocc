package main

// Type represents type
type Type interface {
	size() int
}

// PtrType represents pointing type such as pointer and array
type PtrType interface {
	Type
	base() Type
}

// TyInt represents int type
type TyInt struct{}

// TyPtr represents pointer type
type TyPtr struct {
	to Type
}

// TyArr represents array type
type TyArr struct {
	of  Type
	len int
}

// TyEmpty represents empty type. e.g) Block expression has this type
type TyEmpty struct{}

func (i *TyInt) size() int {
	return 4
}

func (p *TyPtr) size() int {
	return 8
}

func (p *TyPtr) base() Type {
	return p.to
}

func (e *TyEmpty) size() int {
	return 0
}

func (a *TyArr) size() int {
	return a.len * a.of.size()
}

func (a *TyArr) base() Type {
	return a.of
}

func (a *AddrNode) loadType() Type {
	t := a.v.loadType()
	if arrT, ok := t.(*TyArr); ok {
		a.ty = &TyPtr{to: arrT.of}
	} else {
		a.ty = &TyPtr{to: t}
	}
	return a.ty
}

func (a *ArithNode) loadType() Type {
	a.ty = a.lhs.loadType()
	a.rhs.loadType()
	return a.ty
}

func (a *AssignNode) loadType() Type {
	a.rhs.loadType()
	a.ty = a.lhs.loadType()
	return a.ty
}

func (b *BlkNode) loadType() Type {
	for _, st := range b.body {
		st.loadType()
	}
	return &TyEmpty{}
}

func (d *DerefNode) loadType() Type {
	switch v := d.ptr.loadType().(type) {
	case *TyInt:
		d.ty = v
	case *TyPtr:
		d.ty = v.to
	case *TyArr:
		d.ty = v.of
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

func (v *VarNode) loadType() Type {
	return v.v.getType()
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
