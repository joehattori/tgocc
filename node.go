package main

import "fmt"

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

	arithNode struct {
		op  nodeKind
		lhs node
		rhs node
		ty  ty
	}

	assignNode struct {
		lhs addressableNode
		rhs node
		ty  ty
	}

	blkNode struct {
		body []node
	}

	derefNode struct {
		ptr node
		ty  ty
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
		params []node
		name   string
		ty     ty
	}

	fnNode struct {
		params    []*lVar
		body      []node
		lVars     []*lVar
		name      string
		stackSize int
		ty        ty
	}

	varNode struct {
		v variable
	}

	ifNode struct {
		cond node
		then node
		els  node
	}

	nullNode struct{}

	numNode struct {
		val int
		ty  ty
	}

	retNode struct {
		rhs    node
		fnName string
		ty     ty
	}

	whileNode struct {
		cond node
		then node
	}
)

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
)

func newAddNode(lhs node, rhs node) node {
	l := lhs.loadType()
	r := rhs.loadType()
	switch l.(type) {
	case *tyInt, *tyChar:
		switch r.(type) {
		case *tyInt:
			return &arithNode{op: ndAdd, lhs: lhs, rhs: rhs}
		case *tyPtr, *tyArr:
			return &arithNode{op: ndPtrAdd, lhs: rhs, rhs: lhs}
		}
	case *tyPtr, *tyArr:
		switch r.(type) {
		case *tyInt, *tyChar:
			return &arithNode{op: ndPtrAdd, lhs: lhs, rhs: rhs}
		}
	}
	panic(fmt.Sprintf("Unexpected type: lhs: %T %T, rhs: %T %T", lhs, l, rhs, r))
}

func newAddrNode(v addressableNode) node {
	return &addrNode{v: v}
}

func newArithNode(op nodeKind, lhs node, rhs node) node {
	return &arithNode{op: op, lhs: lhs, rhs: rhs}
}

func newAssignNode(lhs addressableNode, rhs node) node {
	return &assignNode{lhs: lhs, rhs: rhs}
}

func newBlkNode(body []node) node {
	return &blkNode{body}
}

func newDerefNode(ptr node) node {
	return &derefNode{ptr: ptr}
}

func newExprNode(body node) node {
	return &exprNode{body}
}

func newForNode(init node, cond node, inc node, body node) node {
	return &forNode{init, cond, inc, body}
}

func newFnCallNode(name string, params []node) node {
	// TODO: change type dynamically
	return &fnCallNode{name: name, params: params, ty: &tyInt{}}
}

func newIfNode(cond node, then node, els node) node {
	return &ifNode{cond, then, els}
}

func newVarNode(v variable) node {
	return &varNode{v}
}

func newNumNode(val int) node {
	return &numNode{val: val}
}

func newSubNode(lhs node, rhs node) node {
	l := lhs.loadType()
	r := rhs.loadType()
	switch l.(type) {
	case *tyInt:
		switch r.(type) {
		case *tyInt:
			return &arithNode{op: ndSub, lhs: lhs, rhs: rhs}
		case *tyPtr, *tyArr:
			return &arithNode{op: ndPtrSub, lhs: rhs, rhs: lhs}
		}
	case *tyPtr, *tyArr:
		switch r.(type) {
		case *tyInt:
			return &arithNode{op: ndPtrSub, lhs: lhs, rhs: rhs}
		}
	}
	panic(fmt.Sprintf("Unexpected type: lhs: %T, rhs: %T", l, r))
}

func (t *tokenized) searchVar(varName string) variable {
	for _, lv := range t.curFn.lVars {
		if lv.name == varName {
			return lv
		}
	}
	for _, g := range t.res.gvars {
		if g.name == varName {
			return g
		}
	}
	return nil
}

func (t *tokenized) findVar(s string) variable {
	v := t.searchVar(s)
	if v == nil {
		panic(fmt.Sprintf("undefined variable %s", s))
	}
	return v
}

func (t *tokenized) buildLVarNode(s string, ty ty, isArg bool) node {
	if _, islVar := t.searchVar(s).(*lVar); islVar {
		panic(fmt.Sprintf("variable %s is already defined", s))
	}
	f := t.curFn
	offset := f.stackSize + ty.size()
	arg := &lVar{name: s, ty: ty, offset: offset}
	f.lVars = append(f.lVars, arg)
	if isArg {
		f.params = append(f.params, arg)
	}
	// TODO: align
	f.stackSize = offset
	return &nullNode{}
}

func newRetNode(rhs node, fnName string) node {
	return &retNode{rhs: rhs, fnName: fnName}
}

func newWhileNode(cond node, then node) node {
	return &whileNode{cond: cond, then: then}
}
