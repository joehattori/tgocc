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
		name   string
		params []node
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

	ifNode struct {
		cond node
		then node
		els  node
	}

	memberNode struct {
		lhs addressableNode
		mem *member
	}

	member struct {
		name   string
		offset int
		ty     ty
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

	stmtExprNode struct {
		body []node
		ty   ty
	}

	varNode struct {
		v variable
	}

	whileNode struct {
		cond node
		then node
	}
)

func newMember(name string, offset int, ty ty) *member {
	return &member{name, offset, ty}
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
)

func newAddNode(lhs node, rhs node) *arithNode {
	l := lhs.loadType()
	r := rhs.loadType()
	switch l.(type) {
	case *tyChar, *tyInt, *tyShort, *tyLong:
		switch r.(type) {
		case *tyChar, *tyInt, *tyShort, *tyLong:
			return &arithNode{op: ndAdd, lhs: lhs, rhs: rhs}
		case *tyPtr, *tyArr:
			return &arithNode{op: ndPtrAdd, lhs: rhs, rhs: lhs}
		}
	case *tyPtr, *tyArr:
		switch r.(type) {
		case *tyChar, *tyInt, *tyShort, *tyLong:
			return &arithNode{op: ndPtrAdd, lhs: lhs, rhs: rhs}
		}
	}
	log.Fatalf("Unexpected type: lhs: %T %T, rhs: %T %T", lhs, l, rhs, r)
	return nil
}

func newAddrNode(v addressableNode) *addrNode {
	return &addrNode{v: v}
}

func newArithNode(op nodeKind, lhs node, rhs node) *arithNode {
	return &arithNode{op: op, lhs: lhs, rhs: rhs}
}

func newAssignNode(lhs addressableNode, rhs node) *assignNode {
	return &assignNode{lhs: lhs, rhs: rhs}
}

func newBlkNode(body []node) *blkNode {
	return &blkNode{body}
}

func newDerefNode(ptr node) *derefNode {
	return &derefNode{ptr: ptr}
}

func newExprNode(body node) *exprNode {
	return &exprNode{body}
}

func newForNode(init node, cond node, inc node, body node) *forNode {
	return &forNode{init, cond, inc, body}
}

func newFnCallNode(name string, params []node) *fnCallNode {
	// TODO: change type dynamically
	return &fnCallNode{name, params, newTyLong()}
}

func newFnNode(name string, ty ty) *fnNode {
	return &fnNode{name: name, ty: ty}
}

func newIfNode(cond node, then node, els node) *ifNode {
	return &ifNode{cond, then, els}
}

func newMemberNode(lhs addressableNode, m *member) *memberNode {
	return &memberNode{lhs, m}
}

func newNullNode() *nullNode {
	return &nullNode{}
}

func newNumNode(val int) *numNode {
	return &numNode{val: val}
}

func newSubNode(lhs node, rhs node) *arithNode {
	l := lhs.loadType()
	r := rhs.loadType()
	switch l.(type) {
	case *tyChar, *tyInt, *tyLong, *tyShort:
		switch r.(type) {
		case *tyChar, *tyInt, *tyLong, *tyShort:
			return &arithNode{op: ndSub, lhs: lhs, rhs: rhs}
		}
	case *tyPtr, *tyArr:
		switch r.(type) {
		case *tyChar, *tyInt, *tyLong, *tyShort:
			return &arithNode{op: ndPtrSub, lhs: lhs, rhs: rhs}
		case *tyPtr, *tyArr:
			return &arithNode{op: ndPtrDiff, lhs: lhs, rhs: rhs}
		}
	}
	log.Fatalf("Unexpected type: lhs: %T, rhs: %T", l, r)
	return nil
}

func newRetNode(rhs node, fnName string) *retNode {
	return &retNode{rhs: rhs, fnName: fnName}
}

func newStmtExprNode(body []node) *stmtExprNode {
	return &stmtExprNode{body: body}
}

func newVarNode(v variable) *varNode {
	return &varNode{v}
}

func newWhileNode(cond node, then node) *whileNode {
	return &whileNode{cond: cond, then: then}
}
