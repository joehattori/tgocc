package main

import "log"

type (
	// ast node
	node interface {
		gen()
		loadType() ty
	}

	addressableNode interface {
		node
		genAddr()
	}

	addrNode struct {
		v  addressableNode
		ty ty
	}

	assignNode struct {
		lhs addressableNode
		rhs node
		ty  ty
	}

	binaryNode struct {
		op  nodeKind
		lhs node
		rhs node
		ty  ty
	}

	bitNotNode struct {
		body node
	}

	blkNode struct {
		body []node
	}

	breakNode struct{}

	caseNode struct {
		cmp  int
		body []node
		idx  int
	}

	castNode struct {
		base node
		toTy ty
	}

	continueNode struct{}

	decNode struct {
		body  addressableNode
		isPre bool
	}

	defaultNode struct {
		body []node
		idx  int
	}

	derefNode struct {
		ptr node
		ty  ty
	}

	doWhileNode struct {
		cond node
		then node
	}

	exprNode struct {
		body node
	}

	forNode struct {
		init node
		cond node
		inc  node
		body node
	}

	fnCallNode struct {
		name   string
		params []node
		retTy  ty
	}

	fnNode struct {
		isStatic  bool
		params    []*lVar
		body      []node
		lVars     []*lVar
		name      string
		stackSize int
		retTy     ty
	}

	ifNode struct {
		cond node
		then node
		els  node
	}

	incNode struct {
		body  addressableNode
		isPre bool
	}

	member struct {
		name   string
		offset int
		ty     ty
	}

	memberNode struct {
		lhs addressableNode
		mem *member
	}

	notNode struct {
		body node
	}

	nullNode struct{}

	numNode struct {
		val int64
	}

	retNode struct {
		rhs    node
		fnName string
		ty     ty
	}

	stmtExprNode struct {
		body []node
		ty   ty
	}

	switchNode struct {
		target node
		cases  []*caseNode
		dflt   *defaultNode
	}

	ternaryNode struct {
		cond node
		lhs  node
		rhs  node
	}

	varNode struct {
		v variable
	}

	whileNode struct {
		cond node
		then node
	}
)

func newMember(name string, offset int, t ty) *member {
	return &member{name, offset, t}
}

type nodeKind int

const (
	ndAdd = iota
	ndSub
	ndMul
	ndDiv
	ndNum
	ndEq
	ndNeq
	ndLt
	ndLeq
	ndGt
	ndGeq
	ndPtrAdd
	ndPtrSub
	ndPtrDiff
	ndAddEq
	ndSubEq
	ndMulEq
	ndDivEq
	ndPtrAddEq
	ndPtrSubEq
	ndBitOr
	ndBitXor
	ndBitAnd
	ndLogOr
	ndLogAnd
	ndShl
	ndShr
	ndShlEq
	ndShrEq
)

func newAddNode(lhs node, rhs node) *binaryNode {
	l := lhs.loadType()
	r := rhs.loadType()
	switch l.(type) {
	case *tyChar, *tyInt, *tyShort, *tyLong, *tyBool:
		switch r.(type) {
		case *tyChar, *tyInt, *tyShort, *tyLong, *tyBool:
			return &binaryNode{op: ndAdd, lhs: lhs, rhs: rhs}
		case *tyPtr, *tyArr:
			return &binaryNode{op: ndPtrAdd, lhs: rhs, rhs: lhs}
		}
	case *tyPtr, *tyArr:
		switch r.(type) {
		case *tyChar, *tyInt, *tyShort, *tyLong, *tyBool:
			return &binaryNode{op: ndPtrAdd, lhs: lhs, rhs: rhs}
		}
	}
	log.Fatalf("Unexpected type for addition: lhs: %T %T, rhs: %T %T", lhs, l, rhs, r)
	return nil
}

func newAddrNode(v addressableNode) *addrNode {
	return &addrNode{v: v}
}

func newAssignNode(lhs addressableNode, rhs node) *assignNode {
	return &assignNode{lhs: lhs, rhs: rhs}
}

func newBinaryNode(op nodeKind, lhs node, rhs node) *binaryNode {
	return &binaryNode{op: op, lhs: lhs, rhs: rhs}
}

func newBitNotNode(body node) *bitNotNode {
	return &bitNotNode{body}
}

func newBlkNode(body []node) *blkNode {
	return &blkNode{body}
}

func newBreakNode() *breakNode {
	return &breakNode{}
}

func newCaseNode(cmp int, body []node, idx int) *caseNode {
	return &caseNode{cmp, body, idx}
}

func newCastNode(base node, t ty) *castNode {
	return &castNode{base, t}
}

func newContinueNode() *continueNode {
	return &continueNode{}
}

func newDecNode(body addressableNode, isPre bool) *decNode {
	return &decNode{body, isPre}
}

func newDefaultNode(body []node, idx int) *defaultNode {
	return &defaultNode{body, idx}
}

func newDerefNode(ptr node) *derefNode {
	return &derefNode{ptr: ptr}
}

func newDoWhileNode(cond node, then node) *doWhileNode {
	return &doWhileNode{cond, then}
}

func newExprNode(body node) *exprNode {
	return &exprNode{body}
}

func newForNode(init node, cond node, inc node, body node) *forNode {
	return &forNode{init, cond, inc, body}
}

func newFnCallNode(name string, params []node, retTy ty) *fnCallNode {
	// TODO: change type dynamically
	return &fnCallNode{name, params, retTy}
}

func newFnNode(isStatic bool, name string, t ty) *fnNode {
	return &fnNode{isStatic: isStatic, name: name, retTy: t}
}

func newIfNode(cond node, then node, els node) *ifNode {
	return &ifNode{cond, then, els}
}

func newIncNode(body addressableNode, isPre bool) *incNode {
	return &incNode{body, isPre}
}

func newMemberNode(lhs addressableNode, m *member) *memberNode {
	return &memberNode{lhs, m}
}

func newNotNode(body node) *notNode {
	return &notNode{body}
}

func newNullNode() *nullNode {
	return &nullNode{}
}

func newNumNode(val int64) *numNode {
	return &numNode{val}
}

func newRetNode(rhs node, fnName string) *retNode {
	return &retNode{rhs: rhs, fnName: fnName}
}

func newStmtExprNode(body []node) *stmtExprNode {
	return &stmtExprNode{body: body}
}

func newSubNode(lhs node, rhs node) *binaryNode {
	l := lhs.loadType()
	r := rhs.loadType()
	switch l.(type) {
	case *tyChar, *tyInt, *tyLong, *tyShort, *tyBool:
		switch r.(type) {
		case *tyChar, *tyInt, *tyLong, *tyShort, *tyBool:
			return &binaryNode{op: ndSub, lhs: lhs, rhs: rhs}
		}
	case *tyPtr, *tyArr:
		switch r.(type) {
		case *tyChar, *tyInt, *tyLong, *tyShort, *tyBool:
			return &binaryNode{op: ndPtrSub, lhs: lhs, rhs: rhs}
		case *tyPtr, *tyArr:
			return &binaryNode{op: ndPtrDiff, lhs: lhs, rhs: rhs}
		}
	}
	log.Fatalf("Unexpected type for subtraction: lhs: %T, rhs: %T", l, r)
	return nil
}

func newSwitchNode(target node, cases []*caseNode, dflt *defaultNode) *switchNode {
	return &switchNode{target, cases, dflt}
}

func newTernaryNode(cond node, lhs node, rhs node) *ternaryNode {
	return &ternaryNode{cond, lhs, rhs}
}

func newVarNode(v variable) *varNode {
	return &varNode{v}
}

func newWhileNode(cond node, then node) *whileNode {
	return &whileNode{cond: cond, then: then}
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

func (a *assignNode) loadType() ty {
	if a.ty == nil {
		a.ty = a.lhs.loadType()
	}
	a.rhs.loadType()
	return a.ty
}

func (a *binaryNode) loadType() ty {
	if a.ty == nil {
		a.ty = a.lhs.loadType()
	}
	a.rhs.loadType()
	return a.ty
}

func (b *bitNotNode) loadType() ty {
	return b.body.loadType()
}

func (b *blkNode) loadType() ty {
	for _, st := range b.body {
		st.loadType()
	}
	return newTyEmpty()
}

func (b *breakNode) loadType() ty {
	return newTyEmpty()
}

func (c *caseNode) loadType() ty {
	return newTyEmpty()
}

func (c *castNode) loadType() ty {
	return c.toTy
}

func (c *continueNode) loadType() ty {
	return newTyEmpty()
}

func (d *decNode) loadType() ty {
	return d.body.loadType()
}

func (d *defaultNode) loadType() ty {
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
		log.Fatalf("Cannot dereference type %T", d.ptr.loadType())
	}
	return d.ty
}

func (d *doWhileNode) loadType() ty {
	return newTyEmpty()
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
	return f.retTy
}

func (f *fnNode) loadType() ty {
	for _, node := range f.body {
		node.loadType()
	}
	return f.retTy
}

func (i *ifNode) loadType() ty {
	i.cond.loadType()
	i.then.loadType()
	if i.els != nil {
		i.els.loadType()
	}
	return newTyEmpty()
}

func (i *incNode) loadType() ty {
	return i.body.loadType()
}

func (m *memberNode) loadType() ty {
	return m.mem.ty
}

func (n *notNode) loadType() ty {
	return newTyBool()
}

func (*nullNode) loadType() ty {
	return newTyEmpty()
}

func (n *numNode) loadType() ty {
	return newTyLong()
}

func (r *retNode) loadType() ty {
	if r.rhs == nil {
		return newTyEmpty()
	}
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

func (s *switchNode) loadType() ty {
	return newTyEmpty()
}

func (t *ternaryNode) loadType() ty {
	return t.lhs.loadType()
}

func (v *varNode) loadType() ty {
	return v.v.getType()
}

func (w *whileNode) loadType() ty {
	w.cond.loadType()
	w.then.loadType()
	return newTyEmpty()
}
