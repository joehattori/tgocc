package main

import "log"

type ty interface {
	size() int
}

type ptr interface {
	ty
	base() ty
}

type tyArr struct {
	of  ty
	len int
}
type tyChar struct{}
type tyEmpty struct{}
type tyInt struct{}
type tyLong struct{}
type tyPtr struct {
	to ty
}
type tyShort struct{}
type tyStruct struct {
	members []*member
	sz      int
}

func newTyArr(of ty, len int) *tyArr              { return &tyArr{of, len} }
func newTyChar() *tyChar                          { return &tyChar{} }
func newTyEmpty() *tyEmpty                        { return &tyEmpty{} }
func newTyInt() *tyInt                            { return &tyInt{} }
func newTyLong() *tyLong                          { return &tyLong{} }
func newTyPtr(to ty) *tyPtr                       { return &tyPtr{to} }
func newTyShort() *tyShort                        { return &tyShort{} }
func newTyStruct(m []*member, size int) *tyStruct { return &tyStruct{m, size} }

func (a *tyArr) size() int    { return a.len * a.of.size() }
func (c *tyChar) size() int   { return 1 }
func (e *tyEmpty) size() int  { return 0 }
func (i *tyInt) size() int    { return 4 }
func (l *tyLong) size() int   { return 8 }
func (p *tyPtr) size() int    { return 8 }
func (s *tyShort) size() int  { return 2 }
func (s *tyStruct) size() int { return s.sz }

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
	case *tyChar, *tyInt, *tyShort, *tyLong:
		d.ty = v
	case *tyPtr:
		d.ty = v.to
	case *tyArr:
		d.ty = v.of
	default:
		log.Fatalf("unhandled type %T", d.ptr.loadType())
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
		n.ty = newTyLong()
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
