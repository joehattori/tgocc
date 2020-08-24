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

type tyStruct struct {
	members []*member
}

func newTyArr(of ty, len int) *tyArr    { return &tyArr{of, len} }
func newTyChar() *tyChar                { return &tyChar{} }
func newTyEmpty() *tyEmpty              { return &tyEmpty{} }
func newTyInt() *tyInt                  { return &tyInt{} }
func newTyPtr(to ty) *tyPtr             { return &tyPtr{to} }
func newTyStruct(m []*member) *tyStruct { return &tyStruct{m} }

func (a *tyArr) size() int   { return a.len * a.of.size() }
func (c *tyChar) size() int  { return 1 }
func (e *tyEmpty) size() int { return 0 }
func (i *tyInt) size() int   { return 4 }
func (p *tyPtr) size() int   { return 8 }
func (s *tyStruct) size() int {
	ret := 0
	for _, member := range s.members {
		ret += member.ty.size()
	}
	return ret
}

func (a *tyArr) base() ty { return a.of }
func (p *tyPtr) base() ty { return p.to }

func (s *tyStruct) findMember(name string) *member {
	for _, member := range s.members {
		if name == member.name {
			return member
		}
	}
	return nil
}

func (a *addrNode) loadType() ty {
	t := a.v.loadType()
	if arrT, ok := t.(*tyArr); ok {
		a.ty = newTyPtr(arrT.of)
	} else {
		a.ty = newTyPtr(t)
	}
	return a.ty
}

func (a *arithNode) loadType() ty {
	if a.ty == nil {
		a.ty = a.lhs.loadType()
	}
	a.rhs.loadType()
	return a.ty
}

func (a *assignNode) loadType() ty {
	if a.ty == nil {
		a.ty = a.lhs.loadType()
	}
	a.rhs.loadType()
	return a.ty
}

func (b *blkNode) loadType() ty {
	for _, st := range b.body {
		st.loadType()
	}
	return newTyEmpty()
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
	return newTyEmpty()
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
	return newTyEmpty()
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
	return newTyEmpty()
}

func (*nullNode) loadType() ty {
	return newTyEmpty()
}

func (m *memberNode) loadType() ty {
	return m.mem.ty
}

func (n *numNode) loadType() ty {
	if n.ty == nil {
		n.ty = newTyInt()
	}
	return n.ty
}

func (r *retNode) loadType() ty {
	if r.ty == nil {
		r.ty = r.rhs.loadType()
	}
	return r.ty
}

func (s *stmtExprNode) loadType() ty {
	if s.ty == nil {
		s.ty = s.body[len(s.body)-1].loadType()
	}
	return s.ty
}

func (v *varNode) loadType() ty {
	return v.v.getType()
}

func (w *whileNode) loadType() ty {
	w.cond.loadType()
	w.then.loadType()
	return newTyEmpty()
}
